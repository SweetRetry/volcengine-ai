package logger

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
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
	logrus.Info("日志管理器启动")

	// 启动日志轮转定时器
	rotateTicker := time.NewTicker(lm.rotateInterval)
	defer rotateTicker.Stop()

	// 启动清理定时器
	cleanTicker := time.NewTicker(lm.cleanInterval)
	defer cleanTicker.Stop()

	// 计算下一个午夜时间用于日志轮转
	now := time.Now()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	timeToMidnight := nextMidnight.Sub(now)

	// 设置首次轮转时间为下一个午夜
	firstRotateTimer := time.NewTimer(timeToMidnight)
	defer firstRotateTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			logrus.Info("日志管理器收到停止信号")
			return

		case <-lm.stopChan:
			logrus.Info("日志管理器手动停止")
			return

		case <-firstRotateTimer.C:
			// 首次轮转后，重置定时器为24小时间隔
			logrus.Info("执行首次日志轮转")
			if err := RotateLogFile(); err != nil {
				logrus.Errorf("日志轮转失败: %v", err)
			}
			// 重置为正常的24小时间隔
			rotateTicker.Reset(lm.rotateInterval)

		case <-rotateTicker.C:
			logrus.Info("执行定时日志轮转")
			if err := RotateLogFile(); err != nil {
				logrus.Errorf("日志轮转失败: %v", err)
			}

		case <-cleanTicker.C:
			logrus.Infof("执行日志清理，保留 %d 天", lm.keepDays)
			if err := CleanOldLogs(lm.keepDays); err != nil {
				logrus.Errorf("日志清理失败: %v", err)
			}
		}
	}
}

// Stop 停止日志管理器
func (lm *LogManager) Stop() {
	close(lm.stopChan)
}

// ForceRotate 强制轮转日志文件
func (lm *LogManager) ForceRotate() error {
	logrus.Info("强制执行日志轮转")
	return RotateLogFile()
}

// ForceClean 强制清理旧日志
func (lm *LogManager) ForceClean() error {
	logrus.Infof("强制执行日志清理，保留 %d 天", lm.keepDays)
	return CleanOldLogs(lm.keepDays)
}
