package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Clear env to test defaults
	os.Unsetenv("PORT")
	os.Unsetenv("DATABASE_URL")

	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("default port = %s, want 8080", cfg.Port)
	}
	if cfg.AIServiceURL != "http://localhost:8000" {
		t.Errorf("default AI URL = %s, want http://localhost:8000", cfg.AIServiceURL)
	}
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("PORT", "3000")
	os.Setenv("DATABASE_URL", "postgresql://test:test@localhost:5432/testdb")
	os.Setenv("JWT_SECRET", "test-jwt-secret")
	os.Setenv("WECHAT_APP_ID", "wx_test_app_id")
	os.Setenv("WECHAT_APP_SECRET", "test_app_secret")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("WECHAT_APP_ID")
		os.Unsetenv("WECHAT_APP_SECRET")
	}()

	cfg := Load()

	if cfg.Port != "3000" {
		t.Errorf("port = %s, want 3000", cfg.Port)
	}
	if cfg.DatabaseURL != "postgresql://test:test@localhost:5432/testdb" {
		t.Errorf("database URL not loaded from env")
	}
	if cfg.JWTSecret != "test-jwt-secret" {
		t.Errorf("JWT secret not loaded from env, got %q", cfg.JWTSecret)
	}
	if cfg.WechatAppID != "wx_test_app_id" {
		t.Errorf("WeChat AppID not loaded from env, got %q", cfg.WechatAppID)
	}
	if cfg.WechatAppSecret != "test_app_secret" {
		t.Errorf("WeChat AppSecret not loaded from env, got %q", cfg.WechatAppSecret)
	}
}

func TestGetEnvFallback(t *testing.T) {
	os.Unsetenv("NONEXISTENT_VAR")
	val := getEnv("NONEXISTENT_VAR", "fallback")
	if val != "fallback" {
		t.Errorf("getEnv fallback = %s, want fallback", val)
	}
}

func TestAllConfigFieldsExist(t *testing.T) {
	cfg := Load()

	// Every field should be non-nil string
	if cfg.Port == "" {
		t.Error("Port should not be empty")
	}
	if cfg.AIServiceURL == "" {
		t.Error("AIServiceURL should not be empty (has default)")
	}
}
