package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"server-monitor/database"
	"server-monitor/models"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源
	},
}

// Client WebSocket客户端
type Client struct {
	ID       string
	Socket   *websocket.Conn
	Send     chan []byte
	Hub      *Hub
	mu       sync.Mutex
}

// Hub WebSocket中心
type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
}

// NewHub 创建新的Hub
func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// Run Hub运行
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			h.mu.Unlock()
			log.Printf("Client %s connected", client.ID)

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("Client %s disconnected", client.ID)

		case message := <-h.Broadcast:
			h.mu.RLock()
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// readPump 读取客户端消息
func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Socket.Close()
	}()

	c.Socket.SetReadLimit(512)
	c.Socket.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Socket.SetPongHandler(func(string) error {
		c.Socket.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Socket.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// 处理客户端消息
		c.handleMessage(message)
	}
}

// writePump 向客户端发送消息
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Socket.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Socket.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Socket.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Socket.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Socket.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 处理客户端消息
func (c *Client) handleMessage(message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		return
	}

	// 根据消息类型处理
	switch msg["type"] {
	case "subscribe":
		// 客户端订阅特定类型的数据
		if dataType, ok := msg["data_type"].(string); ok {
			log.Printf("Client %s subscribed to %s", c.ID, dataType)
		}
	case "ping":
		// 响应ping消息
		response := map[string]interface{}{
			"type": "pong",
			"timestamp": time.Now().Unix(),
		}
		if data, err := json.Marshal(response); err == nil {
			c.Send <- data
		}
	}
}

// ServeWebSocket WebSocket处理器
func ServeWebSocket(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}

		client := &Client{
			ID:     generateClientID(),
			Socket: conn,
			Send:   make(chan []byte, 256),
			Hub:    hub,
		}

		client.Hub.Register <- client

		// 启动读写协程
		go client.writePump()
		go client.readPump()
	}
}

// generateClientID 生成客户端ID
func generateClientID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString 生成随机字符串
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// BroadcastSystemMetrics 广播系统指标
func (h *Hub) BroadcastSystemMetrics(metrics *models.SystemMetrics) {
	data := map[string]interface{}{
		"type": "system_metrics",
		"data": metrics,
	}

	if message, err := json.Marshal(data); err == nil {
		h.Broadcast <- message
	}
}

// BroadcastServiceStatus 广播服务状态
func (h *Hub) BroadcastServiceStatus(services []models.ServiceStatus) {
	data := map[string]interface{}{
		"type": "service_status",
		"data": services,
	}

	if message, err := json.Marshal(data); err == nil {
		h.Broadcast <- message
	}
}

// BroadcastAlert 广播告警
func (h *Hub) BroadcastAlert(alert *models.Alert) {
	data := map[string]interface{}{
		"type": "alert",
		"data": alert,
	}

	if message, err := json.Marshal(data); err == nil {
		h.Broadcast <- message
	}
}

// BroadcastSystemLog 广播系统日志（支持单条或多条）
func (h *Hub) BroadcastSystemLog(logs interface{}) {
	data := map[string]interface{}{
		"type": "system_log",
		"data": logs,
	}

	if message, err := json.Marshal(data); err == nil {
		h.Broadcast <- message
	}
}

// StartMetricsBroadcaster 启动指标广播器
func (h *Hub) StartMetricsBroadcaster() {
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			// 获取最新系统指标
			var metrics models.SystemMetrics
			if err := database.DB.Order("timestamp desc").First(&metrics).Error; err == nil {
				h.BroadcastSystemMetrics(&metrics)
			}

			// 获取服务状态
			var services []models.ServiceStatus
			if err := database.DB.Find(&services).Error; err == nil {
				h.BroadcastServiceStatus(services)
			}
		}
	}()
} 