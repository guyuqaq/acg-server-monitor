package database

import (
	"log"
	"server-monitor/config"
	"server-monitor/models"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase() error {
	var err error
	
	// 配置GORM日志
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}
	
	// 连接SQLite数据库
	DB, err = gorm.Open(sqlite.Open(config.AppConfig.Database.Database), gormConfig)
	if err != nil {
		return err
	}
	
	// 获取底层的sql.DB对象
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	
	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	
	// 自动迁移数据库表
	err = autoMigrate()
	if err != nil {
		return err
	}
	
	// 初始化默认数据
	err = initDefaultData()
	if err != nil {
		return err
	}
	
	log.Println("Database initialized successfully")
	return nil
}

// autoMigrate 自动迁移数据库表
func autoMigrate() error {
	return DB.AutoMigrate(
		&models.SystemMetrics{},
		&models.ServiceStatus{},
		&models.SystemLog{},
		&models.DiskUsage{},
		&models.Alert{},
		&models.NetworkTraffic{},
		&models.ProcessInfo{},
	)
}

// initDefaultData 初始化默认数据
func initDefaultData() error {
	// 检查是否已有服务状态数据
	var count int64
	DB.Model(&models.ServiceStatus{}).Count(&count)
	
	if count == 0 {
		// 插入默认服务状态
		defaultServices := []models.ServiceStatus{
			{
				Name:      "数据库服务",
				Status:    "running",
				Host:      config.AppConfig.Services.Database.Host,
				Port:      config.AppConfig.Services.Database.Port,
				LastCheck: time.Now(),
				Response:  0,
			},
			{
				Name:      "Web服务",
				Status:    "running",
				Host:      config.AppConfig.Services.Web.URL,
				Port:      config.AppConfig.Services.Web.Port,
				LastCheck: time.Now(),
				Response:  0,
			},
			{
				Name:      "邮件服务",
				Status:    "warning",
				Host:      config.AppConfig.Services.Mail.Host,
				Port:      config.AppConfig.Services.Mail.Port,
				LastCheck: time.Now(),
				Response:  0,
			},
			{
				Name:      "云存储服务",
				Status:    "running",
				Host:      config.AppConfig.Services.Storage.Endpoint,
				Port:      "9000",
				LastCheck: time.Now(),
				Response:  0,
			},
		}
		
		for _, service := range defaultServices {
			if err := DB.Create(&service).Error; err != nil {
				return err
			}
		}
	}
	
	// 插入初始系统日志
	initialLogs := []models.SystemLog{
		{
			Level:     "info",
			Category:  "system",
			Message:   "监控系统启动成功",
			Timestamp: time.Now(),
		},
		{
			Level:     "info",
			Category:  "database",
			Message:   "数据库连接初始化完成",
			Timestamp: time.Now(),
		},
	}
	
	for _, log := range initialLogs {
		if err := DB.Create(&log).Error; err != nil {
			return err
		}
	}
	
	return nil
}

// CleanupOldData 清理旧数据
func CleanupOldData() {
	// 清理超过保留时间的系统指标数据
	retentionHours := config.AppConfig.Monitor.HistoryHours
	cutoffTime := time.Now().Add(-time.Duration(retentionHours) * time.Hour)
	
	DB.Where("created_at < ?", cutoffTime).Delete(&models.SystemMetrics{})
	DB.Where("created_at < ?", cutoffTime).Delete(&models.NetworkTraffic{})
	DB.Where("created_at < ?", cutoffTime).Delete(&models.ProcessInfo{})
	
	// 清理已解决的告警（保留7天）
	alertCutoffTime := time.Now().Add(-7 * 24 * time.Hour)
	DB.Where("status = ? AND updated_at < ?", "resolved", alertCutoffTime).Delete(&models.Alert{})
	
	// 清理旧日志（保留30天）
	logCutoffTime := time.Now().Add(-30 * 24 * time.Hour)
	DB.Where("created_at < ?", logCutoffTime).Delete(&models.SystemLog{})
} 