# acg-server-monitor

# 服务器监控系统

一个基于Go语言开发的完整服务器监控系统，提供实时系统指标监控、服务状态检查、告警管理等功能。
  
## 功能特性

###  系统监控
- **CPU使用率监控** - 实时监控CPU使用情况
- **内存使用率监控** - 监控内存使用和可用情况
- **磁盘使用率监控** - 监控磁盘空间使用情况
- **网络流量监控** - 实时监控网络上传/下载速度

### 🔧 服务监控
- **数据库服务检查** - 监控数据库连接状态
- **Web服务检查** - 监控Web服务可用性
- **邮件服务检查** - 监控邮件服务状态
- **云存储服务检查** - 监控存储服务状态

### 📊 数据管理
- **历史数据存储** - 自动保存监控数据到SQLite数据库
- **数据清理** - 自动清理过期数据
- **告警管理** - 支持CPU、内存、磁盘使用率告警

### 🌐 实时通信
- **WebSocket支持** - 实时推送监控数据到前端
- **RESTful API** - 提供完整的API接口
- **CORS支持** - 支持跨域请求

## 技术栈

- **后端**: Go 1.21+
- **Web框架**: Gin
- **数据库**: SQLite (GORM)
- **系统监控**: gopsutil
- **定时任务**: cron
- **WebSocket**: gorilla/websocket
- **配置管理**: Viper

## 项目结构

```
server-monitor/
├── api/                 # API处理器和路由
│   ├── handlers.go      # API处理器
│   └── routes.go        # 路由配置
├── config/              # 配置管理
│   ├── config.go        # 配置结构定义
│   └── config.yaml      # 配置文件
├── database/            # 数据库相关
│   └── database.go      # 数据库连接和初始化
├── models/              # 数据模型
│   └── models.go        # 数据库模型定义
├── monitor/             # 监控模块
│   ├── system_monitor.go # 系统监控
│   └── service_monitor.go # 服务监控
├── scheduler/           # 定时任务调度
│   └── scheduler.go     # 任务调度器
├── websocket/           # WebSocket模块
│   └── websocket.go     # WebSocket处理
├── go.mod              # Go模块文件
├── go.sum              # 依赖校验文件
├── main.go             # 主程序入口
├── index.html          # 前端页面
└── README.md           # 项目说明
```

## 快速开始

### 1. 环境要求

- Go 1.21 或更高版本
- Windows/Linux/macOS

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 配置

编辑 `config/config.yaml` 文件，根据需要调整配置：

```yaml
server:
  port: "8080" #启动端口
  host: "0.0.0.0" #允许all
  log_level: "info" #日志等级

monitor:
  interval: 5          # 监控间隔（秒）
  history_hours: 24    # 历史数据保留时间（小时）
  alert_cpu: 80        # CPU告警阈值
  alert_memory: 80     # 内存告警阈值
  alert_disk: 90       # 磁盘告警阈值

services:   
  database:   #mysql配置 没测试 应该可以正常运行？
    host: "localhost"
    port: "3306"
  web: #可以改为你其他域名 也就一个检测域名能不能访问成功
    url: "localhost"
    port: "80"
  mail: #没什么用
    host: "localhost"
    port: "25"
```

### 4. 运行

```bash
go run main.go
```

或者编译后运行：

```bash
go build -o server-monitor main.go
./server-monitor
```

### 5. 访问

- **Web界面**: http://localhost:8080
- **API文档**: http://localhost:8080/api/v1/
- **健康检查**: http://localhost:8080/health

## API接口

### 系统指标

- `GET /api/v1/metrics` - 获取系统指标历史数据
- `GET /api/v1/metrics/current` - 获取当前系统指标

### 服务状态

- `GET /api/v1/services` - 获取服务状态列表

### 系统日志

- `GET /api/v1/logs` - 获取系统日志
- `POST /api/v1/logs` - 添加系统日志

### 磁盘使用

- `GET /api/v1/disk` - 获取磁盘使用情况

### 告警管理

- `GET /api/v1/alerts` - 获取告警列表
- `PUT /api/v1/alerts/:id/resolve` - 解决告警

### 网络流量

- `GET /api/v1/network` - 获取网络流量数据

### 仪表板

- `GET /api/v1/dashboard` - 获取仪表板综合数据

## WebSocket接口

连接地址：`ws://localhost:8080/ws`

### 消息格式

```json
{
  "type": "system_metrics|service_status|alert|system_log",
  "data": {...}
}
```

### 客户端消息

```json
{
  "type": "subscribe",
  "data_type": "metrics"
}
```

## 监控指标

### 系统指标
- CPU使用率 (%)
- 内存使用率 (%)
- 磁盘使用率 (%)
- 网络上传速度 (MB/s)
- 网络下载速度 (MB/s)

### 服务状态
- 运行状态 (running/warning/error)
- 响应时间 (ms)
- 最后检查时间

### 告警类型
- CPU使用率过高
- 内存使用率过高
- 磁盘使用率过高
- 服务连接失败

## 定时任务

- **系统指标收集**: 每5秒（可配置）
- **服务状态检查**: 每30秒
- **磁盘使用收集**: 每5分钟
- **网络流量收集**: 每30秒
- **数据清理**: 每天凌晨2点

## 数据存储

系统使用SQLite数据库存储监控数据，包括：

- `system_metrics` - 系统指标数据
- `service_status` - 服务状态数据
- `system_logs` - 系统日志
- `disk_usage` - 磁盘使用情况
- `alerts` - 告警信息
- `network_traffic` - 网络流量数据

## 部署

### Docker部署

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o server-monitor main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server-monitor .
COPY --from=builder /app/config ./config
COPY --from=builder /app/index.html .
EXPOSE 8080
CMD ["./server-monitor"]
```

### 系统服务

创建systemd服务文件 `/etc/systemd/system/server-monitor.service`：

```ini
[Unit]
Description=Server Monitor
After=network.target

[Service]
Type=simple
User=monitor
WorkingDirectory=/opt/server-monitor
ExecStart=/opt/server-monitor/server-monitor
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

## 开发

### 添加新的监控指标

1. 在 `models/models.go` 中定义新的数据模型
2. 在 `monitor/system_monitor.go` 中实现数据收集逻辑
3. 在 `scheduler/scheduler.go` 中添加定时任务
4. 在 `api/handlers.go` 中添加API接口

### 添加新的服务监控

1. 在 `monitor/service_monitor.go` 中实现服务检查逻辑
2. 在配置文件中添加服务配置
3. 更新API接口

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request！

### v1.0.0
- 初始版本发布
- 支持基本的系统监控功能
- 提供WebSocket实时数据推送
- 完整的RESTful API接口

### 本项目不会继续更新 如有人继续接手此项目开发请创建新分支
