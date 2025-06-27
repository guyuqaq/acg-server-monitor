package scheduler

import (
	"fmt"
	"log"
	"server-monitor/config"
	"server-monitor/database"
	"server-monitor/monitor"
	"server-monitor/websocket"
	"time"
	"server-monitor/models"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron     *cron.Cron
	hub      *websocket.Hub
	sysMon   *monitor.SystemMonitor
	svcMon   *monitor.ServiceMonitor
}

// NewScheduler 创建新的调度器
func NewScheduler(hub *websocket.Hub) *Scheduler {
	return &Scheduler{
		cron:   cron.New(cron.WithSeconds()),
		hub:    hub,
		sysMon: monitor.NewSystemMonitor(),
		svcMon: monitor.NewServiceMonitor(),
	}
}

// Start 启动调度器
func (s *Scheduler) Start() {
	log.Println("Starting scheduler...")

	// 启动WebSocket指标广播器
	s.hub.StartMetricsBroadcaster()

	// 添加定时任务
	s.addSystemMetricsJob()
	s.addServiceCheckJob()
	s.addDataCleanupJob()
	s.addDiskUsageJob()
	s.addNetworkTrafficJob()
	s.addSystemLogPushJob()

	// 启动cron调度器
	s.cron.Start()

	log.Println("Scheduler started successfully")
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	log.Println("Stopping scheduler...")
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Println("Scheduler stopped")
}

// addSystemMetricsJob 添加系统指标收集任务
func (s *Scheduler) addSystemMetricsJob() {
	interval := config.AppConfig.Monitor.Interval
	schedule := fmt.Sprintf("*/%d * * * * *", interval)
	
	_, err := s.cron.AddFunc(schedule, func() {
		s.collectSystemMetrics()
	})
	
	if err != nil {
		log.Printf("Error adding system metrics job: %v", err)
	} else {
		log.Printf("System metrics job scheduled every %d seconds", interval)
	}
}

// addServiceCheckJob 添加服务检查任务
func (s *Scheduler) addServiceCheckJob() {
	// 每30秒检查一次服务状态
	_, err := s.cron.AddFunc("*/30 * * * * *", func() {
		s.checkServices()
	})
	
	if err != nil {
		log.Printf("Error adding service check job: %v", err)
	} else {
		log.Println("Service check job scheduled every 30 seconds")
	}
}

// addDataCleanupJob 添加数据清理任务
func (s *Scheduler) addDataCleanupJob() {
	// 每天凌晨2点清理旧数据
	_, err := s.cron.AddFunc("0 0 2 * * *", func() {
		s.cleanupOldData()
	})
	
	if err != nil {
		log.Printf("Error adding data cleanup job: %v", err)
	} else {
		log.Println("Data cleanup job scheduled daily at 2:00 AM")
	}
}

// addDiskUsageJob 添加磁盘使用情况收集任务
func (s *Scheduler) addDiskUsageJob() {
	// 每5分钟收集一次磁盘使用情况
	_, err := s.cron.AddFunc("0 */5 * * * *", func() {
		s.collectDiskUsage()
	})
	
	if err != nil {
		log.Printf("Error adding disk usage job: %v", err)
	} else {
		log.Println("Disk usage job scheduled every 5 minutes")
	}
}

// addNetworkTrafficJob 添加网络流量收集任务
func (s *Scheduler) addNetworkTrafficJob() {
	// 每30秒收集一次网络流量
	_, err := s.cron.AddFunc("*/30 * * * * *", func() {
		s.collectNetworkTraffic()
	})
	
	if err != nil {
		log.Printf("Error adding network traffic job: %v", err)
	} else {
		log.Println("Network traffic job scheduled every 30 seconds")
	}
}

// addSystemLogPushJob 添加系统日志推送任务
func (s *Scheduler) addSystemLogPushJob() {
	_, err := s.cron.AddFunc("*/10 * * * * *", func() {
		var logs []models.SystemLog
		database.DB.Order("timestamp desc").Limit(5).Find(&logs)
		s.hub.BroadcastSystemLog(logs)
	})
	if err != nil {
		log.Printf("Error adding system log push job: %v", err)
	} else {
		log.Println("System log push job scheduled every 10 seconds")
	}
}

// collectSystemMetrics 收集系统指标
func (s *Scheduler) collectSystemMetrics() {
	metrics, err := s.sysMon.CollectSystemMetrics()
	if err != nil {
		log.Printf("Error collecting system metrics: %v", err)
		return
	}

	// 保存到数据库
	err = s.sysMon.SaveMetrics(metrics)
	if err != nil {
		log.Printf("Error saving system metrics: %v", err)
		return
	}

	// 检查告警
	err = s.sysMon.CheckAlerts(metrics)
	if err != nil {
		log.Printf("Error checking alerts: %v", err)
	}

	// 广播到WebSocket客户端
	s.hub.BroadcastSystemMetrics(metrics)

	log.Printf("System metrics collected: CPU=%.2f%%, Memory=%.2f%%, Disk=%.2f%%, Upload=%.2fMB/s, Download=%.2fMB/s",
		metrics.CPU, metrics.Memory, metrics.Disk, metrics.Upload, metrics.Download)
}

// checkServices 检查服务状态
func (s *Scheduler) checkServices() {
	err := s.svcMon.CheckAllServices()
	if err != nil {
		log.Printf("Error checking services: %v", err)
		return
	}

	// 获取服务状态并广播
	services, err := s.svcMon.GetServiceStatus()
	if err != nil {
		log.Printf("Error getting service status: %v", err)
		return
	}

	s.hub.BroadcastServiceStatus(services)

	log.Printf("Service status checked: %d services", len(services))
}

// cleanupOldData 清理旧数据
func (s *Scheduler) cleanupOldData() {
	log.Println("Starting data cleanup...")
	
	start := time.Now()
	database.CleanupOldData()
	
	duration := time.Since(start)
	log.Printf("Data cleanup completed in %v", duration)
}

// collectDiskUsage 收集磁盘使用情况
func (s *Scheduler) collectDiskUsage() {
	diskUsages, err := s.sysMon.CollectDiskUsage()
	if err != nil {
		log.Printf("Error collecting disk usage: %v", err)
		return
	}

	// 保存到数据库
	err = s.sysMon.SaveDiskUsage(diskUsages)
	if err != nil {
		log.Printf("Error saving disk usage: %v", err)
		return
	}

	log.Printf("Disk usage collected: %d partitions", len(diskUsages))
}

// collectNetworkTraffic 收集网络流量
func (s *Scheduler) collectNetworkTraffic() {
	traffic, err := s.sysMon.CollectNetworkTraffic()
	if err != nil {
		log.Printf("Error collecting network traffic: %v", err)
		return
	}

	// 保存到数据库
	err = s.sysMon.SaveNetworkTraffic(traffic)
	if err != nil {
		log.Printf("Error saving network traffic: %v", err)
		return
	}

	log.Printf("Network traffic collected: %d interfaces", len(traffic))
}

// GetJobStatus 获取任务状态
func (s *Scheduler) GetJobStatus() []cron.Entry {
	return s.cron.Entries()
}

// AddCustomJob 添加自定义任务
func (s *Scheduler) AddCustomJob(schedule string, job func()) (cron.EntryID, error) {
	return s.cron.AddFunc(schedule, job)
}

// RemoveJob 移除任务
func (s *Scheduler) RemoveJob(id cron.EntryID) {
	s.cron.Remove(id)
} 