package config

import (
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Monitor  MonitorConfig  `mapstructure:"monitor"`
	Services ServicesConfig `mapstructure:"services"`
}

type ServerConfig struct {
	Port    string `mapstructure:"port"`
	Host    string `mapstructure:"host"`
	LogLevel string `mapstructure:"log_level"`
}

type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

type MonitorConfig struct {
	Interval     int `mapstructure:"interval"`      // 监控间隔（秒）
	HistoryHours int `mapstructure:"history_hours"` // 历史数据保留小时数
	AlertCPU     int `mapstructure:"alert_cpu"`     // CPU告警阈值
	AlertMemory  int `mapstructure:"alert_memory"`  // 内存告警阈值
	AlertDisk    int `mapstructure:"alert_disk"`    // 磁盘告警阈值
}

type ServicesConfig struct {
	Database DatabaseServiceConfig `mapstructure:"database"`
	Web      WebServiceConfig      `mapstructure:"web"`
	Mail     MailServiceConfig     `mapstructure:"mail"`
	Storage  StorageServiceConfig  `mapstructure:"storage"`
}

type DatabaseServiceConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

type WebServiceConfig struct {
	URL      string `mapstructure:"url"`
	Port     string `mapstructure:"port"`
	Protocol string `mapstructure:"protocol"`
}

type MailServiceConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type StorageServiceConfig struct {
	Endpoint string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket   string `mapstructure:"bucket"`
}

var AppConfig Config

func LoadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// 设置默认值
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not read config file: %v", err)
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		return err
	}

	return nil
}

func setDefaults() {
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.log_level", "info")
	
	viper.SetDefault("database.driver", "sqlite")
	viper.SetDefault("database.database", "monitor.db")
	
	viper.SetDefault("monitor.interval", 5)
	viper.SetDefault("monitor.history_hours", 24)
	viper.SetDefault("monitor.alert_cpu", 80)
	viper.SetDefault("monitor.alert_memory", 80)
	viper.SetDefault("monitor.alert_disk", 90)
	
	viper.SetDefault("services.database.host", "localhost")
	viper.SetDefault("services.database.port", "3306")
	viper.SetDefault("services.web.url", "localhost")
	viper.SetDefault("services.web.port", "80")
	viper.SetDefault("services.web.protocol", "http")
	viper.SetDefault("services.mail.host", "localhost")
	viper.SetDefault("services.mail.port", "25")
} 