package config

import "os"

type Config struct {
	Port          string
	DatabaseURL   string
	RedisURL      string
	JWTSecret     string
	AppleBundleID string
	AIServiceURL  string
	Env           string
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", ""),
		RedisURL:      getEnv("REDIS_URL", "localhost:6379"),
		JWTSecret:     getEnv("JWT_SECRET", ""),
		AppleBundleID: getEnv("APPLE_BUNDLE_ID", "com.swt.niangaogao"),
		AIServiceURL:  getEnv("AI_SERVICE_URL", "http://localhost:8000"),
		Env:           getEnv("ENV", "development"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
