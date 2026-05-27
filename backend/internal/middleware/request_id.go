package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	RequestIDHeader     = "X-Request-ID"
	RequestIDContextKey = "request_id"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := sanitizeRequestID(c.GetHeader(RequestIDHeader))
		if requestID == "" {
			requestID = newRequestID()
		}
		c.Set(RequestIDContextKey, requestID)
		c.Header(RequestIDHeader, requestID)
		c.Next()
	}
}

func sanitizeRequestID(value string) string {
	id := strings.TrimSpace(value)
	if id == "" || len(id) > 128 {
		return ""
	}
	for _, r := range id {
		if r < 0x20 || r == 0x7f {
			return ""
		}
	}
	return id
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "req-" + time.Now().UTC().Format("20060102T150405.000000000")
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	encoded := make([]byte, 36)
	hex.Encode(encoded[0:8], b[0:4])
	encoded[8] = '-'
	hex.Encode(encoded[9:13], b[4:6])
	encoded[13] = '-'
	hex.Encode(encoded[14:18], b[6:8])
	encoded[18] = '-'
	hex.Encode(encoded[19:23], b[8:10])
	encoded[23] = '-'
	hex.Encode(encoded[24:36], b[10:16])
	return "req-" + string(encoded)
}
