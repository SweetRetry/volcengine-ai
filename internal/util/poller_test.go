package util

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// MockTaskChecker 模拟任务检查器
type MockTaskChecker struct {
	completedAfter int  // 在第几次检查后完成
	currentCheck   int  // 当前检查次数
	shouldError    bool // 是否返回错误
}

func (m *MockTaskChecker) CheckResult(ctx context.Context, taskID string) (interface{}, bool, error) {
	m.currentCheck++

	if m.shouldError {
		return nil, false, fmt.Errorf("模拟检查错误")
	}

	if m.currentCheck >= m.completedAfter {
		return "任务完成", true, nil
	}

	return "任务进行中", false, nil
}

func TestCalculateWaitInterval(t *testing.T) {
	tests := []struct {
		name     string
		attempt  int
		expected time.Duration
		minTime  time.Duration
		maxTime  time.Duration
	}{
		{
			name:     "快速阶段第1次",
			attempt:  0,
			expected: 2 * time.Second,
			minTime:  1800 * time.Millisecond, // 考虑抖动
			maxTime:  2200 * time.Millisecond,
		},
		{
			name:     "快速阶段第5次",
			attempt:  4,
			expected: 2 * time.Second,
			minTime:  1800 * time.Millisecond,
			maxTime:  2200 * time.Millisecond,
		},
		{
			name:     "退避阶段第6次",
			attempt:  5,
			expected: 5 * time.Second,
			minTime:  4500 * time.Millisecond,
			maxTime:  5500 * time.Millisecond,
		},
		{
			name:    "退避阶段第10次",
			attempt: 9,
			minTime: 7 * time.Second, // 调整期望值，1.2^4 ≈ 2.07，所以约10秒
			maxTime: 11 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interval := calculateWaitInterval(tt.attempt)

			if interval < tt.minTime || interval > tt.maxTime {
				t.Errorf("calculateWaitInterval(%d) = %v, 期望在 %v 到 %v 之间",
					tt.attempt, interval, tt.minTime, tt.maxTime)
			}

			t.Logf("第%d次尝试，等待间隔: %v", tt.attempt+1, interval)
		})
	}
}

func TestPollTaskResult_Success(t *testing.T) {
	ctx := context.Background()

	// 模拟任务在第3次检查后完成
	checker := &MockTaskChecker{completedAfter: 3}
	config := DefaultPollConfig("测试任务")

	start := time.Now()
	result, err := PollTaskResult(ctx, "test-task-123", checker, config)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("期望成功，但得到错误: %v", err)
	}

	if result != "任务完成" {
		t.Errorf("期望结果为 '任务完成'，但得到: %v", result)
	}

	if checker.currentCheck != 3 {
		t.Errorf("期望检查3次，但实际检查了%d次", checker.currentCheck)
	}

	// 验证总耗时合理（前2次快速轮询，第3次成功）
	// 第1次立即检查，等待2秒，第2次检查，等待2秒，第3次检查成功
	expectedMinDuration := 3800 * time.Millisecond // 约4秒，考虑抖动
	if duration < expectedMinDuration {
		t.Errorf("总耗时 %v 小于期望的最小值 %v", duration, expectedMinDuration)
	}

	t.Logf("任务完成，总耗时: %v, 检查次数: %d", duration, checker.currentCheck)
}

func TestPollTaskResult_Timeout(t *testing.T) {
	ctx := context.Background()

	// 模拟任务永远不完成
	checker := &MockTaskChecker{completedAfter: 999}
	config := &PollConfig{
		MaxRetries: 3, // 只重试3次
		TaskType:   "测试超时任务",
		Logger:     nil,
	}

	start := time.Now()
	result, err := PollTaskResult(ctx, "test-timeout-task", checker, config)
	duration := time.Since(start)

	if err == nil {
		t.Fatal("期望超时错误，但得到成功结果")
	}

	if result != nil {
		t.Errorf("期望结果为nil，但得到: %v", result)
	}

	if checker.currentCheck != 3 {
		t.Errorf("期望检查3次，但实际检查了%d次", checker.currentCheck)
	}

	t.Logf("任务超时，总耗时: %v, 错误: %v", duration, err)
}

func TestPollTaskResult_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 模拟任务永远不完成
	checker := &MockTaskChecker{completedAfter: 999}
	config := DefaultPollConfig("测试取消任务")

	start := time.Now()
	result, err := PollTaskResult(ctx, "test-cancel-task", checker, config)
	duration := time.Since(start)

	if err == nil {
		t.Fatal("期望上下文取消错误，但得到成功结果")
	}

	if result != nil {
		t.Errorf("期望结果为nil，但得到: %v", result)
	}

	// 验证在合理时间内被取消（应该在3秒左右）
	if duration > 4*time.Second {
		t.Errorf("取消耗时 %v 超过期望值", duration)
	}

	t.Logf("任务被取消，总耗时: %v, 错误: %v", duration, err)
}

func TestPollTaskResultAsync(t *testing.T) {
	ctx := context.Background()

	// 模拟任务在第2次检查后完成
	checker := &MockTaskChecker{completedAfter: 2}
	config := DefaultPollConfig("测试异步任务")

	// 启动异步轮询
	resultChan := PollTaskResultAsync(ctx, "test-async-task", checker, config)

	// 等待结果
	select {
	case pollResult := <-resultChan:
		if pollResult.Error != nil {
			t.Fatalf("期望成功，但得到错误: %v", pollResult.Error)
		}

		if pollResult.Result != "任务完成" {
			t.Errorf("期望结果为 '任务完成'，但得到: %v", pollResult.Result)
		}

		t.Logf("异步任务完成，结果: %v", pollResult.Result)

	case <-time.After(10 * time.Second):
		t.Fatal("异步轮询超时")
	}
}
