package main

import (
	"context"
	"time"

	"volcengine-go-server/pkg/logger"
)

func main() {
	// 初始化日志系统
	logger.Init()

	// 获取日志器实例
	log := logger.GetLogger()

	// 创建日志管理器
	logManager := logger.NewLogManager()

	// 设置较短的间隔用于测试
	logManager.SetRotateInterval(10 * time.Second)
	logManager.SetCleanInterval(30 * time.Second)
	logManager.SetKeepDays(1) // 只保留1天的日志用于测试

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动日志管理器
	go logManager.Start(ctx)

	// 测试不同级别的日志
	log.Info("这是一条信息日志")
	log.Warn("这是一条警告日志")
	log.Error("这是一条错误日志")
	log.Debug("这是一条调试日志")

	// 测试结构化日志
	log.WithFields(map[string]interface{}{
		"user_id": "12345",
		"action":  "login",
		"ip":      "192.168.1.1",
	}).Info("用户登录")

	log.WithFields(map[string]interface{}{
		"task_id": "task_67890",
		"model":   "doubao-image",
		"status":  "completed",
	}).Info("AI任务完成")

	// 模拟一些业务日志
	for i := 0; i < 10; i++ {
		log.Infof("处理第 %d 个请求", i+1)
		time.Sleep(1 * time.Second)
	}

	// 测试强制轮转
	log.Info("测试强制日志轮转")
	if err := logManager.ForceRotate(); err != nil {
		log.Errorf("强制轮转失败: %v", err)
	}

	// 再写一些日志到新文件
	for i := 0; i < 5; i++ {
		log.Infof("轮转后的日志 %d", i+1)
		time.Sleep(1 * time.Second)
	}

	// 测试强制清理
	log.Info("测试强制日志清理")
	if err := logManager.ForceClean(); err != nil {
		log.Errorf("强制清理失败: %v", err)
	}

	log.Info("日志测试完成")

	// 等待一段时间观察日志管理器的工作
	time.Sleep(5 * time.Second)
}
