package util

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"

	"volcengine-go-server/pkg/logger"
)

// TaskResultChecker 任务结果检查器接口
type TaskResultChecker interface {
	// CheckResult 检查任务结果
	// 返回值：result(任务结果), isCompleted(是否完成), error(错误)
	CheckResult(ctx context.Context, taskID string) (interface{}, bool, error)
}

// PollConfig 轮询配置
type PollConfig struct {
	MaxRetries int            // 最大重试次数，默认60
	TaskType   string         // 任务类型，用于日志
	Logger     *logrus.Logger // 日志记录器，可选
}

// DefaultPollConfig 创建默认轮询配置
func DefaultPollConfig(taskType string) *PollConfig {
	return &PollConfig{
		MaxRetries: 60,
		TaskType:   taskType,
		Logger:     logger.GetLogger(),
	}
}

// NewPollConfig 创建自定义轮询配置
func NewPollConfig(taskType string, maxRetries int, _ time.Duration) *PollConfig {
	return &PollConfig{
		MaxRetries: maxRetries,
		TaskType:   taskType,
		Logger:     logger.GetLogger(),
	}
}

// WithLogger 设置日志记录器
func (c *PollConfig) WithLogger(logger *logrus.Logger) *PollConfig {
	c.Logger = logger
	return c
}

// calculateWaitInterval 计算等待间隔 - 使用自适应策略
// 前5次快速轮询(2秒)，之后逐渐增加间隔，最大30秒
func calculateWaitInterval(attempt int) time.Duration {
	const (
		fastPhaseAttempts = 5
		fastInterval      = 2 * time.Second
		baseInterval      = 5 * time.Second
		maxInterval       = 30 * time.Second
		growthFactor      = 1.2 // 降低增长因子，使增长更温和
	)

	var interval time.Duration

	if attempt < fastPhaseAttempts {
		// 快速阶段：前5次使用2秒间隔
		interval = fastInterval
	} else {
		// 逐渐增长阶段：5秒 * 1.2^(attempt-5)
		backoffAttempt := attempt - fastPhaseAttempts
		multiplier := 1.0
		for i := 0; i < backoffAttempt; i++ {
			multiplier *= growthFactor
		}
		interval = time.Duration(float64(baseInterval) * multiplier)

		// 限制最大间隔
		if interval > maxInterval {
			interval = maxInterval
		}
	}

	// 添加10%的随机抖动，避免惊群效应
	if interval > time.Second {
		jitterRange := int64(interval / 10)
		if jitterRange > 0 {
			jitter := time.Duration(rand.Int63n(jitterRange*2) - jitterRange)
			interval += jitter
		}
	}

	// 确保间隔不小于1秒
	if interval < time.Second {
		interval = time.Second
	}

	return interval
}

// PollTaskResult 通用任务轮询方法
// 使用自适应轮询策略：前期快速轮询，后期逐渐增加间隔
func PollTaskResult(ctx context.Context, taskID string, checker TaskResultChecker, config *PollConfig) (interface{}, error) {
	if config == nil {
		config = DefaultPollConfig("unknown")
	}

	if config.Logger == nil {
		config.Logger = logger.GetLogger()
	}

	config.Logger.WithFields(logrus.Fields{
		"task_id":     taskID,
		"task_type":   config.TaskType,
		"max_retries": config.MaxRetries,
		"strategy":    "自适应轮询",
	}).Info("开始轮询任务结果")

	var lastError error
	startTime := time.Now()

	for attempt := 0; attempt < config.MaxRetries; attempt++ {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("任务已取消: %v", ctx.Err())
		default:
		}

		attemptStart := time.Now()

		// 检查任务结果
		result, isCompleted, err := checker.CheckResult(ctx, taskID)
		attemptDuration := time.Since(attemptStart)

		if err != nil {
			lastError = err
			config.Logger.WithFields(logrus.Fields{
				"task_id":     taskID,
				"attempt":     attempt + 1,
				"duration_ms": attemptDuration.Milliseconds(),
				"error":       err.Error(),
			}).Warn("查询任务结果失败，将重试")

			// 对于查询错误，也需要等待后重试
		} else {
			// 检查任务是否完成
			if isCompleted {
				totalDuration := time.Since(startTime)
				config.Logger.WithFields(logrus.Fields{
					"task_id":        taskID,
					"task_type":      config.TaskType,
					"total_attempts": attempt + 1,
					"total_duration": totalDuration,
				}).Info("任务完成")
				return result, nil
			}

			// 任务未完成，记录进度
			config.Logger.WithFields(logrus.Fields{
				"task_id":     taskID,
				"attempt":     attempt + 1,
				"duration_ms": attemptDuration.Milliseconds(),
			}).Debug("任务进行中")
		}

		// 如果不是最后一次尝试，计算等待时间
		if attempt < config.MaxRetries-1 {
			waitInterval := calculateWaitInterval(attempt)

			config.Logger.WithFields(logrus.Fields{
				"task_id":      taskID,
				"next_attempt": attempt + 2,
				"wait_seconds": waitInterval.Seconds(),
			}).Debug("等待下次轮询")

			// 等待下次查询
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("任务已取消: %v", ctx.Err())
			case <-time.After(waitInterval):
				// 继续下一次轮询
			}
		}
	}

	totalDuration := time.Since(startTime)

	// 构建超时错误信息
	errorMsg := fmt.Sprintf("%s任务轮询超时: taskID=%s, 总尝试次数=%d, 总耗时=%v",
		config.TaskType, taskID, config.MaxRetries, totalDuration)

	if lastError != nil {
		errorMsg += fmt.Sprintf(", 最后错误: %v", lastError)
	}

	config.Logger.WithFields(logrus.Fields{
		"task_id":        taskID,
		"task_type":      config.TaskType,
		"total_attempts": config.MaxRetries,
		"total_duration": totalDuration,
		"last_error":     lastError,
	}).Error("任务轮询超时")

	return nil, errors.New(errorMsg)
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
