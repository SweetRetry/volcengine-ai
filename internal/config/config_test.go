package config

import (
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("PORT", "9000")
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("MONGO_URL", "mongodb://test:27017/test_db")

	cfg := New()

	if cfg.Port != "9000" {
		t.Errorf("期望端口为 9000，实际为 %s", cfg.Port)
	}

	if cfg.Environment != "test" {
		t.Errorf("期望环境为 test，实际为 %s", cfg.Environment)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("期望日志级别为 debug，实际为 %s", cfg.LogLevel)
	}

	if cfg.Database.MongoURL != "mongodb://test:27017/test_db" {
		t.Errorf("期望MongoDB URL为 mongodb://test:27017/test_db，实际为 %s", cfg.Database.MongoURL)
	}
}

func TestGetEnv(t *testing.T) {
	// 测试默认值
	value := getEnv("NON_EXISTENT_KEY", "default_value")
	if value != "default_value" {
		t.Errorf("期望默认值为 default_value，实际为 %s", value)
	}

	// 测试环境变量
	os.Setenv("TEST_KEY", "test_value")
	value = getEnv("TEST_KEY", "default_value")
	if value != "test_value" {
		t.Errorf("期望值为 test_value，实际为 %s", value)
	}
}
