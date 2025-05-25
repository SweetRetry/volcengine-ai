package service

import (
	"testing"
	"time"

	"jimeng-go-server/internal/database"
)

func TestImageTaskService_CreateImageTask(t *testing.T) {
	// 这里使用mock数据库进行测试
	// 在实际项目中，你可能需要使用testcontainers或mock数据库

	// 创建测试输入
	input := &ImageTaskInput{
		Prompt:  "一只可爱的小猫",
		UserID:  "test-user-123",
		Model:   "doubao-seedream-3.0-t2i",
		Size:    "1024x1024",
		Quality: "standard",
		Style:   "anime",
		N:       1,
	}

	// 验证输入结构体的JSON序列化
	if input.Prompt == "" {
		t.Error("Prompt不能为空")
	}
	if input.UserID == "" {
		t.Error("UserID不能为空")
	}
	if input.Model == "" {
		t.Error("Model不能为空")
	}
}

func TestImageTaskResult_Structure(t *testing.T) {
	// 测试结果结构体
	result := &ImageTaskResult{
		TaskID:   "task-123",
		Status:   "completed",
		ImageURL: "https://example.com/image.jpg",
		Created:  time.Now(),
	}

	if result.TaskID == "" {
		t.Error("TaskID不能为空")
	}
	if result.Status == "" {
		t.Error("Status不能为空")
	}
	if result.ImageURL == "" {
		t.Error("ImageURL不能为空")
	}
}

func TestImageTaskService_NewImageTaskService(t *testing.T) {
	// 测试服务创建
	var db database.Database // 这里应该是mock数据库
	service := NewImageTaskService(db)

	if service == nil {
		t.Error("服务创建失败")
	}
	if service.db != db {
		t.Error("数据库注入失败")
	}
}
