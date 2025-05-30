package util

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// TaskResultChecker 任务结果检查器接口
type TaskResultChecker interface {
	// CheckResult 检查任务结果
	// 返回值：result(任务结果), isCompleted(是否完成), error(错误)
	CheckResult(ctx context.Context, taskID string) (interface{}, bool, error)
}

// PollConfig 轮询配置
type PollConfig struct {
	MaxRetries    int            // 最大重试次数，默认60
	RetryInterval time.Duration  // 重试间隔，默认10秒
	TaskType      string         // 任务类型，用于日志
	Logger        *logrus.Logger // 日志记录器，可选
}

// DefaultPollConfig 创建默认轮询配置
func DefaultPollConfig(taskType string) *PollConfig {
	return &PollConfig{
		MaxRetries:    60,
		RetryInterval: 10 * time.Second,
		TaskType:      taskType,
		Logger:        logrus.New(),
	}
}

// NewPollConfig 创建自定义轮询配置
func NewPollConfig(taskType string, maxRetries int, retryInterval time.Duration) *PollConfig {
	return &PollConfig{
		MaxRetries:    maxRetries,
		RetryInterval: retryInterval,
		TaskType:      taskType,
		Logger:        logrus.New(),
	}
}

// WithLogger 设置日志记录器
func (c *PollConfig) WithLogger(logger *logrus.Logger) *PollConfig {
	c.Logger = logger
	return c
}

// PollTaskResult 通用任务轮询方法
// 这个方法会阻塞当前goroutine直到任务完成或超时，但不会影响其他goroutine的执行
func PollTaskResult(ctx context.Context, taskID string, checker TaskResultChecker, config *PollConfig) (interface{}, error) {
	if config == nil {
		config = DefaultPollConfig("unknown")
	}

	if config.Logger == nil {
		config.Logger = logrus.New()
	}

	config.Logger.Infof("开始轮询%s任务结果: taskID=%s, maxRetries=%d, interval=%v",
		config.TaskType, taskID, config.MaxRetries, config.RetryInterval)

	for i := 0; i < config.MaxRetries; i++ {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("任务已取消: %v", ctx.Err())
		default:
		}

		// 检查任务结果
		result, isCompleted, err := checker.CheckResult(ctx, taskID)
		if err != nil {
			config.Logger.Errorf("查询%s任务结果失败: %v", config.TaskType, err)
			return nil, err
		}

		// 检查任务是否完成
		if isCompleted {
			config.Logger.Infof("%s任务完成: taskID=%s", config.TaskType, taskID)
			return result, nil
		}

		config.Logger.Infof("%s任务进行中: taskID=%s, 第%d次查询", config.TaskType, taskID, i+1)

		// 等待下次查询
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("任务已取消: %v", ctx.Err())
		case <-time.After(config.RetryInterval):
			// 继续下一次轮询
		}
	}

	return nil, fmt.Errorf("%s任务超时，taskID: %s", config.TaskType, taskID)
}

// PollTaskResultAsync 异步轮询任务结果
// 返回一个channel，可以非阻塞地获取结果
func PollTaskResultAsync(ctx context.Context, taskID string, checker TaskResultChecker, config *PollConfig) <-chan PollResult {
	resultChan := make(chan PollResult, 1)

	go func() {
		defer close(resultChan)

		result, err := PollTaskResult(ctx, taskID, checker, config)
		resultChan <- PollResult{
			Result: result,
			Error:  err,
		}
	}()

	return resultChan
}

// PollResult 轮询结果
type PollResult struct {
	Result interface{}
	Error  error
}
