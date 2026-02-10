package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ListenAddr  string
	Environment string
	KubeConfig  string
	RedisAddr   string
	Rate        int
	RateWindow  time.Duration
	LogLevel    string
}

func Load() (*Config, error) {
	cfg := &Config{
		ListenAddr:  getEnv("LISTEN_ADDR", "127.0.0.1:8080"),
		Environment: getEnv("ENV", "dev"),
		KubeConfig:  getEnv("KUBECONFIG", ""),
		RedisAddr:   getEnv("REDIS_ADDR", ""),
		Rate:        getEnvAsInt("RATE_LIMIT", 3),
		RateWindow:  getEnvAsDuration("RATE_WINDOW", 1*time.Minute),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultVal
}