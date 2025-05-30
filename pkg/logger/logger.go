package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

// Init 初始化日志器
func Init() {
	Logger = logrus.New()

	// 创建日志目录
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		logrus.Fatalf("创建日志目录失败: %v", err)
	}

	// 生成当前日期的日志文件名
	currentDate := time.Now().Format("2006-01-02")
	logFileName := fmt.Sprintf("app-%s.log", currentDate)
	logFilePath := filepath.Join(logDir, logFileName)

	// 打开或创建日志文件
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		logrus.Fatalf("打开日志文件失败: %v", err)
	}

	// 创建多输出器，同时输出到控制台和文件
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	Logger.SetOutput(multiWriter)

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

	Logger.Infof("日志系统初始化完成，日志文件: %s", logFilePath)
}

// GetLogger 获取日志器实例
func GetLogger() *logrus.Logger {
	if Logger == nil {
		Init()
	}
	return Logger
}

// SetLevel 设置日志级别
func SetLevel(level string) {
	if Logger == nil {
		Init()
	}

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
}

// RotateLogFile 轮转日志文件（可选的手动轮转功能）
func RotateLogFile() error {
	if Logger == nil {
		return fmt.Errorf("日志器未初始化")
	}

	// 创建日志目录
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	// 生成新的日志文件名
	currentDate := time.Now().Format("2006-01-02")
	logFileName := fmt.Sprintf("app-%s.log", currentDate)
	logFilePath := filepath.Join(logDir, logFileName)

	// 打开或创建新的日志文件
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}

	// 更新输出器
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	Logger.SetOutput(multiWriter)

	Logger.Infof("日志文件已轮转到: %s", logFilePath)
	return nil
}

// CleanOldLogs 清理旧日志文件（保留指定天数的日志）
func CleanOldLogs(keepDays int) error {
	logDir := "logs"

	// 读取日志目录
	files, err := os.ReadDir(logDir)
	if err != nil {
		return fmt.Errorf("读取日志目录失败: %v", err)
	}

	// 计算截止日期
	cutoffDate := time.Now().AddDate(0, 0, -keepDays)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// 检查文件名是否符合日志文件格式
		fileName := file.Name()
		if len(fileName) < 14 || fileName[:4] != "app-" || fileName[len(fileName)-4:] != ".log" {
			continue
		}

		// 提取日期部分
		dateStr := fileName[4 : len(fileName)-4] // 去掉 "app-" 前缀和 ".log" 后缀
		fileDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue // 跳过无法解析日期的文件
		}

		// 如果文件日期早于截止日期，则删除
		if fileDate.Before(cutoffDate) {
			filePath := filepath.Join(logDir, fileName)
			if err := os.Remove(filePath); err != nil {
				Logger.Warnf("删除旧日志文件失败: %s, 错误: %v", filePath, err)
			} else {
				Logger.Infof("已删除旧日志文件: %s", filePath)
			}
		}
	}

	return nil
}
