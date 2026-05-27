package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	// Clear env to test defaults
	os.Unsetenv("PORT")
	os.Unsetenv("DATABASE_URL")
	t.Setenv("AI_GATEWAY_TIMEOUT_SECONDS", "")

	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("default port = %s, want 8080", cfg.Port)
	}
	if cfg.AIServiceURL != "http://localhost:8000" {
		t.Errorf("default AI URL = %s, want http://localhost:8000", cfg.AIServiceURL)
	}
	if cfg.AIGatewayTimeout != 65*time.Second {
		t.Errorf("default AI gateway timeout = %s, want 65s", cfg.AIGatewayTimeout)
	}
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("PORT", "3000")
	os.Setenv("DATABASE_URL", "postgresql://test:test@localhost:5432/testdb")
	os.Setenv("JWT_SECRET", "test-jwt-secret")
	os.Setenv("APPLE_BUNDLE_ID", "com.test.app")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("APPLE_BUNDLE_ID")
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
	if cfg.AppleBundleID != "com.test.app" {
		t.Errorf("Apple Bundle ID not loaded from env, got %q", cfg.AppleBundleID)
	}
}

func TestLoadAIGatewayTimeoutFromEnv(t *testing.T) {
	t.Setenv("AI_GATEWAY_TIMEOUT_SECONDS", "90")

	cfg := Load()

	if cfg.AIGatewayTimeout != 90*time.Second {
		t.Errorf("AI gateway timeout = %s, want 90s", cfg.AIGatewayTimeout)
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
