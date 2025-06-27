package models

import (
	"time"
	"gorm.io/gorm"
)

// SystemMetrics 系统指标数据
type SystemMetrics struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Timestamp time.Time `json:"timestamp"`
	CPU       float64   `json:"cpu"`        // CPU使用率
	Memory    float64   `json:"memory"`     // 内存使用率
	Disk      float64   `json:"disk"`       // 磁盘使用率
	Upload    float64   `json:"upload"`     // 上传速度 MB/s
	Download  float64   `json:"download"`   // 下载速度 MB/s
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ServiceStatus 服务状态
type ServiceStatus struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`       // 服务名称
	Status    string    `json:"status"`     // 状态: running, warning, error
	Host      string    `json:"host"`       // 服务地址
	Port      string    `json:"port"`       // 服务端口
	LastCheck time.Time `json:"last_check"` // 最后检查时间
	Response  int       `json:"response"`   // 响应时间(ms)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SystemLog 系统日志
type SystemLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Level     string    `json:"level"`      // 日志级别: info, warning, error
	Category  string    `json:"category"`   // 日志分类: system, security, database, network
	Message   string    `json:"message"`    // 日志消息
	Timestamp time.Time `json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
}

// DiskUsage 磁盘使用情况
type DiskUsage struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Path      string    `json:"path"`       // 磁盘路径
	Name      string    `json:"name"`       // 磁盘名称
	Total     uint64    `json:"total"`      // 总容量(GB)
	Used      uint64    `json:"used"`       // 已使用(GB)
	Free      uint64    `json:"free"`       // 可用空间(GB)
	Usage     float64   `json:"usage"`      // 使用率(%)
	Timestamp time.Time `json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Alert 告警信息
type Alert struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Type      string    `json:"type"`       // 告警类型: cpu, memory, disk, service
	Level     string    `json:"level"`      // 告警级别: info, warning, error
	Message   string    `json:"message"`    // 告警消息
	Value     float64   `json:"value"`      // 告警值
	Threshold float64   `json:"threshold"`  // 阈值
	Status    string    `json:"status"`     // 状态: active, resolved
	Timestamp time.Time `json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NetworkTraffic 网络流量数据
type NetworkTraffic struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Interface string    `json:"interface"`  // 网络接口
	Upload    uint64    `json:"upload"`     // 上传字节数
	Download  uint64    `json:"download"`   // 下载字节数
	UploadSpeed   float64 `json:"upload_speed"`   // 上传速度 MB/s
	DownloadSpeed float64 `json:"download_speed"` // 下载速度 MB/s
	Timestamp time.Time `json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProcessInfo 进程信息
type ProcessInfo struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	PID       int       `json:"pid"`
	Name      string    `json:"name"`
	CPU       float64   `json:"cpu"`
	Memory    float64   `json:"memory"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
}

// BeforeCreate GORM钩子，设置创建时间
func (m *SystemMetrics) BeforeCreate(tx *gorm.DB) error {
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	return nil
}

func (s *ServiceStatus) BeforeCreate(tx *gorm.DB) error {
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
	return nil
}

func (l *SystemLog) BeforeCreate(tx *gorm.DB) error {
	l.CreatedAt = time.Now()
	return nil
}

func (d *DiskUsage) BeforeCreate(tx *gorm.DB) error {
	d.CreatedAt = time.Now()
	d.UpdatedAt = time.Now()
	return nil
}

func (a *Alert) BeforeCreate(tx *gorm.DB) error {
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()
	return nil
}

func (n *NetworkTraffic) BeforeCreate(tx *gorm.DB) error {
	n.CreatedAt = time.Now()
	n.UpdatedAt = time.Now()
	return nil
}

func (p *ProcessInfo) BeforeCreate(tx *gorm.DB) error {
	p.CreatedAt = time.Now()
	return nil
} 