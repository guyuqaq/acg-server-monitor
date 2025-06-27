package monitor

import (
	"fmt"
	"log"
	"math"
	"server-monitor/config"
	"server-monitor/database"
	"server-monitor/models"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type SystemMonitor struct {
	lastNetworkStats map[string]net.IOCountersStat
	lastNetworkTime  time.Time
}

// NewSystemMonitor 创建系统监控实例
func NewSystemMonitor() *SystemMonitor {
	return &SystemMonitor{
		lastNetworkStats: make(map[string]net.IOCountersStat),
		lastNetworkTime:  time.Now(),
	}
}

// CollectSystemMetrics 收集系统指标
func (sm *SystemMonitor) CollectSystemMetrics() (*models.SystemMetrics, error) {
	metrics := &models.SystemMetrics{
		Timestamp: time.Now(),
	}

	// 收集CPU使用率
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		log.Printf("Error collecting CPU metrics: %v", err)
		metrics.CPU = 0
	} else if len(cpuPercent) > 0 {
		metrics.CPU = math.Round(cpuPercent[0]*100) / 100
	}

	// 收集内存使用率
	memory, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("Error collecting memory metrics: %v", err)
		metrics.Memory = 0
	} else {
		metrics.Memory = math.Round(memory.UsedPercent*100) / 100
	}

	// 收集磁盘使用率
	partitions, err := disk.Partitions(false)
	if err != nil {
		log.Printf("Error collecting disk metrics: %v", err)
		metrics.Disk = 0
	} else {
		var totalUsage float64
		var partitionCount int
		
		for _, partition := range partitions {
			usage, err := disk.Usage(partition.Mountpoint)
			if err != nil {
				continue
			}
			totalUsage += usage.UsedPercent
			partitionCount++
		}
		
		if partitionCount > 0 {
			metrics.Disk = math.Round((totalUsage/float64(partitionCount))*100) / 100
		}
	}

	// 收集网络流量
	uploadSpeed, downloadSpeed, err := sm.getNetworkSpeed()
	if err != nil {
		log.Printf("Error collecting network metrics: %v", err)
		metrics.Upload = 0
		metrics.Download = 0
	} else {
		metrics.Upload = uploadSpeed
		metrics.Download = downloadSpeed
	}

	return metrics, nil
}

// getNetworkSpeed 获取网络速度
func (sm *SystemMonitor) getNetworkSpeed() (float64, float64, error) {
	netStats, err := net.IOCounters(false)
	if err != nil {
		return 0, 0, err
	}

	now := time.Now()
	timeDiff := now.Sub(sm.lastNetworkTime).Seconds()

	if timeDiff == 0 {
		return 0, 0, fmt.Errorf("time difference is zero")
	}

	var totalUploadBytes uint64
	var totalDownloadBytes uint64

	for _, stat := range netStats {
		if lastStat, exists := sm.lastNetworkStats[stat.Name]; exists {
			uploadDiff := stat.BytesSent - lastStat.BytesSent
			downloadDiff := stat.BytesRecv - lastStat.BytesRecv
			
			totalUploadBytes += uploadDiff
			totalDownloadBytes += downloadDiff
		}
		sm.lastNetworkStats[stat.Name] = stat
	}

	// 转换为MB/s
	uploadSpeed := float64(totalUploadBytes) / (1024 * 1024 * timeDiff)
	downloadSpeed := float64(totalDownloadBytes) / (1024 * 1024 * timeDiff)

	sm.lastNetworkTime = now

	return math.Round(uploadSpeed*100) / 100, math.Round(downloadSpeed*100) / 100, nil
}

// CollectDiskUsage 收集磁盘使用情况
func (sm *SystemMonitor) CollectDiskUsage() ([]models.DiskUsage, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}

	var diskUsages []models.DiskUsage
	now := time.Now()

	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}

		diskUsage := models.DiskUsage{
			Path:      partition.Mountpoint,
			Name:      partition.Device,
			Total:     usage.Total / (1024 * 1024 * 1024), // 转换为GB
			Used:      usage.Used / (1024 * 1024 * 1024),  // 转换为GB
			Free:      usage.Free / (1024 * 1024 * 1024),  // 转换为GB
			Usage:     math.Round(usage.UsedPercent*100) / 100,
			Timestamp: now,
		}

		diskUsages = append(diskUsages, diskUsage)
	}

	return diskUsages, nil
}

// CollectNetworkTraffic 收集网络流量数据
func (sm *SystemMonitor) CollectNetworkTraffic() ([]models.NetworkTraffic, error) {
	netStats, err := net.IOCounters(true)
	if err != nil {
		return nil, err
	}

	var networkTraffic []models.NetworkTraffic
	now := time.Now()

	for _, stat := range netStats {
		// 计算速度
		var uploadSpeed, downloadSpeed float64
		if lastStat, exists := sm.lastNetworkStats[stat.Name]; exists {
			timeDiff := now.Sub(sm.lastNetworkTime).Seconds()
			if timeDiff > 0 {
				uploadDiff := stat.BytesSent - lastStat.BytesSent
				downloadDiff := stat.BytesRecv - lastStat.BytesRecv
				
				uploadSpeed = float64(uploadDiff) / (1024 * 1024 * timeDiff)
				downloadSpeed = float64(downloadDiff) / (1024 * 1024 * timeDiff)
			}
		}

		traffic := models.NetworkTraffic{
			Interface:      stat.Name,
			Upload:         stat.BytesSent,
			Download:       stat.BytesRecv,
			UploadSpeed:    math.Round(uploadSpeed*100) / 100,
			DownloadSpeed:  math.Round(downloadSpeed*100) / 100,
			Timestamp:      now,
		}

		networkTraffic = append(networkTraffic, traffic)
	}

	return networkTraffic, nil
}

// SaveMetrics 保存监控指标到数据库
func (sm *SystemMonitor) SaveMetrics(metrics *models.SystemMetrics) error {
	return database.DB.Create(metrics).Error
}

// SaveDiskUsage 保存磁盘使用情况
func (sm *SystemMonitor) SaveDiskUsage(diskUsages []models.DiskUsage) error {
	for _, usage := range diskUsages {
		if err := database.DB.Create(&usage).Error; err != nil {
			return err
		}
	}
	return nil
}

// SaveNetworkTraffic 保存网络流量数据
func (sm *SystemMonitor) SaveNetworkTraffic(traffic []models.NetworkTraffic) error {
	for _, t := range traffic {
		if err := database.DB.Create(&t).Error; err != nil {
			return err
		}
	}
	return nil
}

// CheckAlerts 检查告警
func (sm *SystemMonitor) CheckAlerts(metrics *models.SystemMetrics) error {
	// 检查CPU告警
	if metrics.CPU > float64(config.AppConfig.Monitor.AlertCPU) {
		// 检查是否已有活跃的CPU告警
		var existingAlert models.Alert
		result := database.DB.Where("type = ? AND status = ?", "cpu", "active").First(&existingAlert)
		
		if result.Error != nil {
			// 没有活跃告警，创建新的
			alert := models.Alert{
				Type:      "cpu",
				Level:     "warning",
				Message:   fmt.Sprintf("CPU使用率过高: %.2f%%", metrics.CPU),
				Value:     metrics.CPU,
				Threshold: float64(config.AppConfig.Monitor.AlertCPU),
				Status:    "active",
				Timestamp: time.Now(),
			}
			database.DB.Create(&alert)
			
			// 同时创建系统日志
			systemLog := models.SystemLog{
				Level:     "warning",
				Category:  "system",
				Message:   fmt.Sprintf("CPU使用率过高: %.2f%%", metrics.CPU),
				Timestamp: time.Now(),
			}
			database.DB.Create(&systemLog)
		} else {
			// 已有活跃告警，只更新值
			existingAlert.Value = metrics.CPU
			existingAlert.Message = fmt.Sprintf("CPU使用率过高: %.2f%%", metrics.CPU)
			existingAlert.UpdatedAt = time.Now()
			database.DB.Save(&existingAlert)
		}
	} else {
		// CPU使用率正常，如果有活跃告警则标记为已解决
		var existingAlert models.Alert
		if database.DB.Where("type = ? AND status = ?", "cpu", "active").First(&existingAlert).Error == nil {
			existingAlert.Status = "resolved"
			existingAlert.UpdatedAt = time.Now()
			database.DB.Save(&existingAlert)
			
			// 创建解决日志
			systemLog := models.SystemLog{
				Level:     "info",
				Category:  "system",
				Message:   fmt.Sprintf("CPU使用率恢复正常: %.2f%%", metrics.CPU),
				Timestamp: time.Now(),
			}
			database.DB.Create(&systemLog)
		}
	}

	// 检查内存告警
	if metrics.Memory > float64(config.AppConfig.Monitor.AlertMemory) {
		// 检查是否已有活跃的内存告警
		var existingAlert models.Alert
		result := database.DB.Where("type = ? AND status = ?", "memory", "active").First(&existingAlert)
		
		if result.Error != nil {
			// 没有活跃告警，创建新的
			alert := models.Alert{
				Type:      "memory",
				Level:     "warning",
				Message:   fmt.Sprintf("内存使用率过高: %.2f%%", metrics.Memory),
				Value:     metrics.Memory,
				Threshold: float64(config.AppConfig.Monitor.AlertMemory),
				Status:    "active",
				Timestamp: time.Now(),
			}
			database.DB.Create(&alert)
			
			// 同时创建系统日志
			systemLog := models.SystemLog{
				Level:     "warning",
				Category:  "system",
				Message:   fmt.Sprintf("内存使用率过高: %.2f%%", metrics.Memory),
				Timestamp: time.Now(),
			}
			database.DB.Create(&systemLog)
		} else {
			// 已有活跃告警，只更新值
			existingAlert.Value = metrics.Memory
			existingAlert.Message = fmt.Sprintf("内存使用率过高: %.2f%%", metrics.Memory)
			existingAlert.UpdatedAt = time.Now()
			database.DB.Save(&existingAlert)
		}
	} else {
		// 内存使用率正常，如果有活跃告警则标记为已解决
		var existingAlert models.Alert
		if database.DB.Where("type = ? AND status = ?", "memory", "active").First(&existingAlert).Error == nil {
			existingAlert.Status = "resolved"
			existingAlert.UpdatedAt = time.Now()
			database.DB.Save(&existingAlert)
			
			// 创建解决日志
			systemLog := models.SystemLog{
				Level:     "info",
				Category:  "system",
				Message:   fmt.Sprintf("内存使用率恢复正常: %.2f%%", metrics.Memory),
				Timestamp: time.Now(),
			}
			database.DB.Create(&systemLog)
		}
	}

	// 检查磁盘告警
	if metrics.Disk > float64(config.AppConfig.Monitor.AlertDisk) {
		// 检查是否已有活跃的磁盘告警
		var existingAlert models.Alert
		result := database.DB.Where("type = ? AND status = ?", "disk", "active").First(&existingAlert)
		
		if result.Error != nil {
			// 没有活跃告警，创建新的
			alert := models.Alert{
				Type:      "disk",
				Level:     "warning",
				Message:   fmt.Sprintf("磁盘使用率过高: %.2f%%", metrics.Disk),
				Value:     metrics.Disk,
				Threshold: float64(config.AppConfig.Monitor.AlertDisk),
				Status:    "active",
				Timestamp: time.Now(),
			}
			database.DB.Create(&alert)
			
			// 同时创建系统日志
			systemLog := models.SystemLog{
				Level:     "warning",
				Category:  "system",
				Message:   fmt.Sprintf("磁盘使用率过高: %.2f%%", metrics.Disk),
				Timestamp: time.Now(),
			}
			database.DB.Create(&systemLog)
		} else {
			// 已有活跃告警，只更新值
			existingAlert.Value = metrics.Disk
			existingAlert.Message = fmt.Sprintf("磁盘使用率过高: %.2f%%", metrics.Disk)
			existingAlert.UpdatedAt = time.Now()
			database.DB.Save(&existingAlert)
		}
	} else {
		// 磁盘使用率正常，如果有活跃告警则标记为已解决
		var existingAlert models.Alert
		if database.DB.Where("type = ? AND status = ?", "disk", "active").First(&existingAlert).Error == nil {
			existingAlert.Status = "resolved"
			existingAlert.UpdatedAt = time.Now()
			database.DB.Save(&existingAlert)
			
			// 创建解决日志
			systemLog := models.SystemLog{
				Level:     "info",
				Category:  "system",
				Message:   fmt.Sprintf("磁盘使用率恢复正常: %.2f%%", metrics.Disk),
				Timestamp: time.Now(),
			}
			database.DB.Create(&systemLog)
		}
	}

	return nil
}

// HardwareInfo 结构体
type HardwareInfo struct {
	CPUModel   string  `json:"cpu_model"`
	CPUCores   int     `json:"cpu_cores"`
	CPUThreads int     `json:"cpu_threads"`
	CPUFreq    float64 `json:"cpu_freq"`
	MemorySize string  `json:"memory_size"`
	MemoryType string  `json:"memory_type"`
	MemorySpeed string `json:"memory_speed"`
	DiskModel  string  `json:"disk_model"`
	DiskSize   string  `json:"disk_size"`
	DiskType   string  `json:"disk_type"`
}

// GetHardwareInfo 采集硬件信息
func GetHardwareInfo() (*HardwareInfo, error) {
	info := &HardwareInfo{}
	// CPU信息
	cpuInfos, err := cpu.Info()
	if err == nil && len(cpuInfos) > 0 {
		info.CPUModel = cpuInfos[0].ModelName
		info.CPUCores = int(cpuInfos[0].Cores)
		info.CPUThreads = len(cpuInfos)
		info.CPUFreq = cpuInfos[0].Mhz
	}
	// 内存信息
	mem, err := mem.VirtualMemory()
	if err == nil {
		info.MemorySize = fmt.Sprintf("%.0fGB", float64(mem.Total)/1024/1024/1024)
		info.MemoryType = "N/A" // gopsutil不支持
		info.MemorySpeed = "N/A"
	}
	// 磁盘信息
	disks, err := disk.Partitions(false)
	if err == nil && len(disks) > 0 {
		usage, _ := disk.Usage(disks[0].Mountpoint)
		info.DiskModel = disks[0].Device
		info.DiskSize = fmt.Sprintf("%.0fGB", float64(usage.Total)/1024/1024/1024)
		info.DiskType = "N/A"
	}
	return info, nil
} 