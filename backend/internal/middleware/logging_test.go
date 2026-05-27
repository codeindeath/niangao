package middleware

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestLoggerIncludesRequestIDStatusAndRoute(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	r := gin.New()
	r.Use(RequestID(), RequestLoggerWithLogger(logger))
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health?token=secret", nil)
	req.Header.Set(RequestIDHeader, "client-request-1")
	r.ServeHTTP(w, req)

	var entry map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &entry); err != nil {
		t.Fatalf("decode log entry: %v; raw=%q", err, buf.String())
	}
	if entry["request_id"] != "client-request-1" {
		t.Fatalf("request_id = %+v, want client-request-1", entry["request_id"])
	}
	if entry["status"] != float64(http.StatusOK) {
		t.Fatalf("status = %+v, want 200", entry["status"])
	}
	if entry["method"] != "GET" || entry["path"] != "/health" {
		t.Fatalf("method/path = %+v/%+v, want GET /health", entry["method"], entry["path"])
	}
	if _, ok := entry["latency_ms"].(float64); !ok {
		t.Fatalf("latency_ms = %+v, want number", entry["latency_ms"])
	}
	if _, ok := entry["token"]; ok {
		t.Fatalf("log entry leaked query token: %+v", entry)
	}
}
