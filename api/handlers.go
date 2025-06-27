package api

import (
	"net/http"
	"server-monitor/database"
	"server-monitor/models"
	"server-monitor/monitor"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// GetSystemMetrics 获取系统指标数据
func GetSystemMetrics(c *gin.Context) {
	// 获取查询参数
	limitStr := c.DefaultQuery("limit", "100")
	hoursStr := c.Query("hours")
	daysStr := c.Query("days")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}

	query := database.DB.Order("timestamp desc")
	
	// 处理时间范围查询
	if hoursStr != "" {
		if hours, err := strconv.Atoi(hoursStr); err == nil {
			startTime := time.Now().Add(-time.Duration(hours) * time.Hour)
			query = query.Where("timestamp >= ?", startTime)
		}
	} else if daysStr != "" {
		if days, err := strconv.Atoi(daysStr); err == nil {
			startTime := time.Now().Add(-time.Duration(days*24) * time.Hour)
			query = query.Where("timestamp >= ?", startTime)
		}
	} else {
		// 如果没有指定时间范围，使用limit限制数量
		query = query.Limit(limit)
	}

	var metrics []models.SystemMetrics
	err = query.Find(&metrics).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "获取系统指标失败",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data:    metrics,
	})
}

// GetCurrentMetrics 获取当前系统指标
func GetCurrentMetrics(c *gin.Context) {
	var metric models.SystemMetrics
	err := database.DB.Order("timestamp desc").First(&metric).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "获取当前指标失败",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data:    metric,
	})
}

// GetServiceStatus 获取服务状态
func GetServiceStatus(c *gin.Context) {
	var services []models.ServiceStatus
	err := database.DB.Find(&services).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "获取服务状态失败",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data:    services,
	})
}

// GetSystemLogs 获取系统日志
func GetSystemLogs(c *gin.Context) {
	// 获取查询参数
	limitStr := c.DefaultQuery("limit", "50")
	level := c.DefaultQuery("level", "")
	category := c.DefaultQuery("category", "")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	query := database.DB.Order("timestamp desc").Limit(limit)
	
	if level != "" {
		query = query.Where("level = ?", level)
	}
	
	if category != "" {
		query = query.Where("category = ?", category)
	}

	var logs []models.SystemLog
	err = query.Find(&logs).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "获取系统日志失败",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data:    logs,
	})
}

// GetDiskUsage 获取磁盘使用情况
func GetDiskUsage(c *gin.Context) {
	var diskUsages []models.DiskUsage
	err := database.DB.Order("timestamp desc").Find(&diskUsages).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "获取磁盘使用情况失败",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data:    diskUsages,
	})
}

// GetAlerts 获取告警信息
func GetAlerts(c *gin.Context) {
	// 获取查询参数
	status := c.DefaultQuery("status", "")
	level := c.DefaultQuery("level", "")
	
	query := database.DB.Order("timestamp desc")
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	if level != "" {
		query = query.Where("level = ?", level)
	}

	var alerts []models.Alert
	err := query.Find(&alerts).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "获取告警信息失败",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data:    alerts,
	})
}

// GetNetworkTraffic 获取网络流量数据
func GetNetworkTraffic(c *gin.Context) {
	// 获取查询参数
	limitStr := c.DefaultQuery("limit", "100")
	interfaceName := c.DefaultQuery("interface", "")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}

	query := database.DB.Order("timestamp desc").Limit(limit)
	
	if interfaceName != "" {
		query = query.Where("interface = ?", interfaceName)
	}

	var traffic []models.NetworkTraffic
	err = query.Find(&traffic).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "获取网络流量数据失败",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data:    traffic,
	})
}

// GetDashboardData 获取仪表板数据
func GetDashboardData(c *gin.Context) {
	// 获取当前系统指标
	var currentMetric models.SystemMetrics
	database.DB.Order("timestamp desc").First(&currentMetric)

	// 获取服务状态
	var services []models.ServiceStatus
	database.DB.Find(&services)

	// 获取最近的系统日志
	var recentLogs []models.SystemLog
	database.DB.Order("timestamp desc").Limit(10).Find(&recentLogs)

	// 获取活跃告警
	var activeAlerts []models.Alert
	database.DB.Where("status = ?", "active").Order("timestamp desc").Limit(10).Find(&activeAlerts)

	// 获取历史数据（最近24小时，每小时一个数据点）
	var historicalData []models.SystemMetrics
	startTime := time.Now().Add(-24 * time.Hour)
	database.DB.Where("timestamp >= ?", startTime).Order("timestamp asc").Find(&historicalData)

	dashboardData := map[string]interface{}{
		"current_metrics":   currentMetric,
		"services":          services,
		"recent_logs":       recentLogs,
		"active_alerts":     activeAlerts,
		"historical_data":   historicalData,
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data:    dashboardData,
	})
}

// ResolveAlert 解决告警
func ResolveAlert(c *gin.Context) {
	alertID := c.Param("id")
	
	var alert models.Alert
	err := database.DB.First(&alert, alertID).Error
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: "告警不存在",
			Data:    nil,
		})
		return
	}

	alert.Status = "resolved"
	alert.UpdatedAt = time.Now()
	
	err = database.DB.Save(&alert).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "更新告警状态失败",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "告警已解决",
		Data:    alert,
	})
}

// AddSystemLog 添加系统日志
func AddSystemLog(c *gin.Context) {
	var log models.SystemLog
	if err := c.ShouldBindJSON(&log); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误",
			Data:    nil,
		})
		return
	}

	log.Timestamp = time.Now()
	
	err := database.DB.Create(&log).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "添加系统日志失败",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "日志添加成功",
		Data:    log,
	})
}

// GetHardwareInfo 获取硬件信息
func GetHardwareInfoHandler(c *gin.Context) {
	info, err := monitor.GetHardwareInfo()
	if err != nil {
		c.JSON(500, Response{
			Code: 500,
			Message: "获取硬件信息失败",
			Data: nil,
		})
		return
	}
	c.JSON(200, Response{
		Code: 200,
		Message: "success",
		Data: info,
	})
}

// GetCssboardData 处理 /api/v1/css 路由，返回css静态文件
func GetCssboardData(c *gin.Context) {
	c.File("css/remixicon.min.css")
}

// GetJsboardData 处理 /api/v1/js 路由，返回js静态文件
func GetJsboardData(c *gin.Context) {
	c.File("js/echarts.min.js")
} 