package monitor

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"server-monitor/config"
	"server-monitor/database"
	"server-monitor/models"
	"time"
)

type ServiceMonitor struct {
	httpClient *http.Client
}

// NewServiceMonitor 创建服务监控实例
func NewServiceMonitor() *ServiceMonitor {
	return &ServiceMonitor{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CheckAllServices 检查所有服务状态
func (sm *ServiceMonitor) CheckAllServices() error {
	services := []struct {
		name string
		host string
		port string
		check func(string, string) (string, int, error)
	}{
		{
			name:  "数据库服务",
			host:  config.AppConfig.Services.Database.Host,
			port:  config.AppConfig.Services.Database.Port,
			check: sm.checkDatabaseService,
		},
		{
			name:  "Web服务",
			host:  config.AppConfig.Services.Web.URL,
			port:  config.AppConfig.Services.Web.Port,
			check: sm.checkWebService,
		},
		{
			name:  "邮件服务",
			host:  config.AppConfig.Services.Mail.Host,
			port:  config.AppConfig.Services.Mail.Port,
			check: sm.checkMailService,
		},
		{
			name:  "云存储服务",
			host:  config.AppConfig.Services.Storage.Endpoint,
			port:  "9000",
			check: sm.checkStorageService,
		},
	}

	for _, service := range services {
		status, responseTime, err := service.check(service.host, service.port)
		
		// 更新或创建服务状态记录
		var serviceStatus models.ServiceStatus
		result := database.DB.Where("name = ?", service.name).First(&serviceStatus)
		
		if result.Error != nil {
			// 创建新记录
			serviceStatus = models.ServiceStatus{
				Name:      service.name,
				Host:      service.host,
				Port:      service.port,
				Status:    status,
				LastCheck: time.Now(),
				Response:  responseTime,
			}
			database.DB.Create(&serviceStatus)
		} else {
			// 更新现有记录
			serviceStatus.Status = status
			serviceStatus.LastCheck = time.Now()
			serviceStatus.Response = responseTime
			database.DB.Save(&serviceStatus)
		}

		// 记录日志
		if err != nil {
			log.Printf("Service check failed for %s: %v", service.name, err)
			sm.logServiceEvent(service.name, "error", fmt.Sprintf("服务检查失败: %v", err))
		} else {
			sm.logServiceEvent(service.name, "info", fmt.Sprintf("服务状态: %s, 响应时间: %dms", status, responseTime))
		}
	}

	return nil
}

// checkDatabaseService 检查数据库服务
func (sm *ServiceMonitor) checkDatabaseService(host, port string) (string, int, error) {
	start := time.Now()
	
	// 尝试连接数据库端口
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", host, port), 5*time.Second)
	if err != nil {
		return "error", 0, err
	}
	defer conn.Close()
	
	responseTime := int(time.Since(start).Milliseconds())
	
	// 根据响应时间判断状态
	if responseTime < 100 {
		return "running", responseTime, nil
	} else if responseTime < 500 {
		return "warning", responseTime, nil
	} else {
		return "error", responseTime, fmt.Errorf("响应时间过长: %dms", responseTime)
	}
}

// checkWebService 检查Web服务
func (sm *ServiceMonitor) checkWebService(host, port string) (string, int, error) {
	start := time.Now()
	
	url := fmt.Sprintf("%s://%s:%s", config.AppConfig.Services.Web.Protocol, host, port)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "error", 0, err
	}
	
	resp, err := sm.httpClient.Do(req)
	if err != nil {
		return "error", 0, err
	}
	defer resp.Body.Close()
	
	responseTime := int(time.Since(start).Milliseconds())
	
	// 根据HTTP状态码和响应时间判断状态
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		if responseTime < 200 {
			return "running", responseTime, nil
		} else if responseTime < 1000 {
			return "warning", responseTime, nil
		} else {
			return "error", responseTime, fmt.Errorf("响应时间过长: %dms", responseTime)
		}
	} else {
		return "error", responseTime, fmt.Errorf("HTTP状态码错误: %d", resp.StatusCode)
	}
}

// checkMailService 检查邮件服务
func (sm *ServiceMonitor) checkMailService(host, port string) (string, int, error) {
	start := time.Now()
	
	// 尝试连接SMTP端口
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", host, port), 5*time.Second)
	if err != nil {
		return "error", 0, err
	}
	defer conn.Close()
	
	responseTime := int(time.Since(start).Milliseconds())
	
	// 根据响应时间判断状态
	if responseTime < 100 {
		return "running", responseTime, nil
	} else if responseTime < 500 {
		return "warning", responseTime, nil
	} else {
		return "error", responseTime, fmt.Errorf("响应时间过长: %dms", responseTime)
	}
}

// checkStorageService 检查云存储服务
func (sm *ServiceMonitor) checkStorageService(host, port string) (string, int, error) {
	start := time.Now()
	
	// 尝试连接存储服务端口
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", host, port), 5*time.Second)
	if err != nil {
		return "error", 0, err
	}
	defer conn.Close()
	
	responseTime := int(time.Since(start).Milliseconds())
	
	// 根据响应时间判断状态
	if responseTime < 100 {
		return "running", responseTime, nil
	} else if responseTime < 500 {
		return "warning", responseTime, nil
	} else {
		return "error", responseTime, fmt.Errorf("响应时间过长: %dms", responseTime)
	}
}

// logServiceEvent 记录服务事件
func (sm *ServiceMonitor) logServiceEvent(serviceName, level, message string) {
	log := models.SystemLog{
		Level:     level,
		Category:  "service",
		Message:   fmt.Sprintf("[%s] %s", serviceName, message),
		Timestamp: time.Now(),
	}
	
	database.DB.Create(&log)
}

// GetServiceStatus 获取服务状态列表
func (sm *ServiceMonitor) GetServiceStatus() ([]models.ServiceStatus, error) {
	var services []models.ServiceStatus
	err := database.DB.Find(&services).Error
	return services, err
}

// GetServiceStatusByName 根据名称获取服务状态
func (sm *ServiceMonitor) GetServiceStatusByName(name string) (*models.ServiceStatus, error) {
	var service models.ServiceStatus
	err := database.DB.Where("name = ?", name).First(&service).Error
	if err != nil {
		return nil, err
	}
	return &service, nil
} 