package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes() *gin.Engine {
	r := gin.Default()

	// 配置CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(config))

	// API路由组
	api := r.Group("/api/v1")
	{
		// 系统指标相关
		api.GET("/metrics", GetSystemMetrics)
		api.GET("/metrics/current", GetCurrentMetrics)
		
		// 服务状态相关
		api.GET("/services", GetServiceStatus)
		
		// 系统日志相关
		api.GET("/logs", GetSystemLogs)
		api.POST("/logs", AddSystemLog)
		
		// 磁盘使用情况
		api.GET("/disk", GetDiskUsage)
		
		// 告警相关
		api.GET("/alerts", GetAlerts)
		api.PUT("/alerts/:id/resolve", ResolveAlert)
		
		// 网络流量
		api.GET("/network", GetNetworkTraffic)
		
		// 硬件信息
		api.GET("/hardware", GetHardwareInfoHandler)
		
		// 仪表板数据
		api.GET("/dashboard", GetDashboardData)
		r.Static("/css", "./css")
		r.Static("/js", "./js")
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"message": "Server is running",
		})
	})

	return r
} 