package logger

import (
	"context"
	"time"
)

// LogManager 日志管理器
type LogManager struct {
	rotateInterval time.Duration // 日志轮转间隔
	cleanInterval  time.Duration // 清理间隔
	keepDays       int           // 保留天数
	stopChan       chan struct{} // 停止信号
}

// NewLogManager 创建日志管理器
func NewLogManager() *LogManager {
	return &LogManager{
		rotateInterval: 24 * time.Hour, // 每24小时轮转一次
		cleanInterval:  24 * time.Hour, // 每24小时清理一次
		keepDays:       7,              // 保留7天的日志
		stopChan:       make(chan struct{}),
	}
}

// SetRotateInterval 设置日志轮转间隔
func (lm *LogManager) SetRotateInterval(interval time.Duration) {
	lm.rotateInterval = interval
}

// SetCleanInterval 设置清理间隔
func (lm *LogManager) SetCleanInterval(interval time.Duration) {
	lm.cleanInterval = interval
}

// SetKeepDays 设置日志保留天数
func (lm *LogManager) SetKeepDays(days int) {
	lm.keepDays = days
}

// Start 启动日志管理器
func (lm *LogManager) Start(ctx context.Context) {
	log := GetLogger()
	log.Info("日志管理器启动")

	// 执行首次日志轮转
	if err := RotateLogFile(); err != nil {
		log.Errorf("日志轮转失败: %v", err)
	} else {
		log.Info("执行首次日志轮转")
	}

	// 创建定时器
	ticker := time.NewTicker(24 * time.Hour) // 每24小时执行一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("日志管理器收到停止信号")
			return
		case <-lm.stopChan:
			log.Info("日志管理器手动停止")
			return
		case <-ticker.C:
			// 执行日志轮转
			if err := RotateLogFile(); err != nil {
				log.Errorf("日志轮转失败: %v", err)
			} else {
				log.Info("执行定时日志轮转")
			}

			// 执行日志清理
			if err := CleanOldLogs(lm.keepDays); err != nil {
				log.Errorf("日志清理失败: %v", err)
			} else {
				log.Infof("执行日志清理，保留 %d 天", lm.keepDays)
			}
		}
	}
}

// Stop 停止日志管理器
func (lm *LogManager) Stop() {
	close(lm.stopChan)
}

// ForceRotate 强制执行日志轮转
func (lm *LogManager) ForceRotate() error {
	log := GetLogger()
	log.Info("强制执行日志轮转")
	return RotateLogFile()
}

// ForceCleanup 强制执行日志清理
func (lm *LogManager) ForceCleanup() error {
	log := GetLogger()
	log.Infof("强制执行日志清理，保留 %d 天", lm.keepDays)
	return CleanOldLogs(lm.keepDays)
}
