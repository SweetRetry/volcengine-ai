package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

// Init 初始化日志器
func Init() {
	Logger = logrus.New()

	// 设置输出
	Logger.SetOutput(os.Stdout)

	// 设置日志级别
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug":
		Logger.SetLevel(logrus.DebugLevel)
	case "info":
		Logger.SetLevel(logrus.InfoLevel)
	case "warn":
		Logger.SetLevel(logrus.WarnLevel)
	case "error":
		Logger.SetLevel(logrus.ErrorLevel)
	default:
		Logger.SetLevel(logrus.InfoLevel)
	}

	// 设置格式
	Logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
}

// GetLogger 获取日志器实例
func GetLogger() *logrus.Logger {
	if Logger == nil {
		Init()
	}
	return Logger
}
