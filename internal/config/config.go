package config

import (
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
	// 火山引擎即梦AI配置
	VolcengineAPIKey   string
	VolcengineEndpoint string
	VolcengineRegion   string
	Timeout            string
}

func New() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		Database: DatabaseConfig{
			MongoURL: getEnv("MONGO_URL", "mongodb://localhost:27017/jimeng_db"),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", "redis://localhost:6379"),
		},
		AI: AIConfig{
			VolcengineAPIKey:   getEnv("VOLCENGINE_API_KEY", "WkRreU1HUmxNRFUzTXpJM05EQTBNVGsxTldVNE9XUmtaV0ZpWm1VeE0yWQ=="),
			VolcengineEndpoint: getEnv("VOLCENGINE_ENDPOINT", "https://visual.volcengineapi.com"),
			VolcengineRegion:   getEnv("VOLCENGINE_REGION", "cn-north-1"),
			Timeout:            getEnv("AI_TIMEOUT", "30s"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
