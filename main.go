package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"server-monitor/api"
	"server-monitor/config"
	"server-monitor/database"
	"server-monitor/scheduler"
	"server-monitor/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	// 设置日志格式
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Server Monitor...")

	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 创建WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()

	// 创建调度器
	sched := scheduler.NewScheduler(hub)

	// 设置Gin模式
	if config.AppConfig.Server.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 设置路由
	router := api.SetupRoutes()

	// 添加WebSocket路由
	router.GET("/ws", websocket.ServeWebSocket(hub))

	// 添加静态文件服务（用于前端页面和静态资源）
	router.Static("/static", "./static")
	// 访问根路径/时返回index.html
	router.GET("/", func(c *gin.Context) {
		c.File("index.html")
	})

	// 创建HTTP服务器
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", config.AppConfig.Server.Host, config.AppConfig.Server.Port),
		Handler: router,
	}

	// 启动调度器
	sched.Start()

	// 启动HTTP服务器
	go func() {
		log.Printf("Server starting on %s:%s", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// 停止调度器
	sched.Stop()

	// 优雅关闭HTTP服务器
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
} 