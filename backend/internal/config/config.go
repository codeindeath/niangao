package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port             string
	DatabaseURL      string
	RedisURL         string
	JWTSecret        string
	AppleBundleID    string
	AIServiceURL     string
	AIGatewayTimeout time.Duration
	Env              string
}

func Load() *Config {
	return &Config{
		Port:             getEnv("PORT", "8080"),
		DatabaseURL:      getEnv("DATABASE_URL", ""),
		RedisURL:         getEnv("REDIS_URL", "localhost:6379"),
		JWTSecret:        getEnv("JWT_SECRET", ""),
		AppleBundleID:    getEnv("APPLE_BUNDLE_ID", "com.swt.niangaogao"),
		AIServiceURL:     getEnv("AI_SERVICE_URL", "http://localhost:8000"),
		AIGatewayTimeout: getEnvDurationSeconds("AI_GATEWAY_TIMEOUT_SECONDS", 65*time.Second),
		Env:              getEnv("ENV", "development"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvDurationSeconds(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	seconds, err := strconv.Atoi(v)
	if err != nil || seconds <= 0 {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}
