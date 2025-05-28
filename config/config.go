package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	Environment string
	LogLevel    string
	Database    DatabaseConfig
	Redis       RedisConfig
	AI          AIConfig
}

type DatabaseConfig struct {
	MongoURL string
}

type RedisConfig struct {
	URL string
}

type AIConfig struct {
	// 火山方舟AI配置
	VolcengineAPIKey    string // 火山方舟API Key
	VolcengineAccessKey string // Access Key ID (备用)
	VolcengineSecretKey string // Secret Access Key (备用)
	Timeout             string // 请求超时时间
}

func New() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		Database: DatabaseConfig{
			MongoURL: getEnv("MONGO_URL", "mongodb://localhost:27017/volcengine_db"),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", "redis://localhost:6379"),
		},
		AI: AIConfig{
			VolcengineAPIKey: getEnv("ARK_API_KEY", ""),
			Timeout:          getEnv("AI_TIMEOUT", "30s"),
		},
	}
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	if c.AI.VolcengineAPIKey == "" {
		return fmt.Errorf("ARK_API_KEY is required")
	}
	if c.Database.MongoURL == "" {
		return fmt.Errorf("MONGO_URL is required")
	}
	if c.Redis.URL == "" {
		return fmt.Errorf("REDIS_URL is required")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
