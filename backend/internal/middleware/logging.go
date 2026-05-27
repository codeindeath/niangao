package middleware

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

type accessLogEntry struct {
	Level     string  `json:"level"`
	Message   string  `json:"message"`
	RequestID string  `json:"request_id"`
	Method    string  `json:"method"`
	Path      string  `json:"path"`
	Status    int     `json:"status"`
	LatencyMS float64 `json:"latency_ms"`
	ClientIP  string  `json:"client_ip"`
}

func RequestLogger() gin.HandlerFunc {
	return RequestLoggerWithLogger(log.Default())
}

func RequestLoggerWithLogger(logger *log.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = log.Default()
	}
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		requestID, _ := c.Get(RequestIDContextKey)
		entry := accessLogEntry{
			Level:     "info",
			Message:   "http_request",
			RequestID: stringValue(requestID),
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			Status:    c.Writer.Status(),
			LatencyMS: float64(time.Since(start).Microseconds()) / 1000,
			ClientIP:  c.ClientIP(),
		}
		if entry.Status >= 500 {
			entry.Level = "error"
		} else if entry.Status >= 400 {
			entry.Level = "warn"
		}

		encoded, err := json.Marshal(entry)
		if err != nil {
			logger.Printf(`{"level":"error","message":"access_log_encode_failed","request_id":%q}`, entry.RequestID)
			return
		}
		logger.Print(string(encoded))
	}
}

func stringValue(value any) string {
	s, _ := value.(string)
	return s
}
