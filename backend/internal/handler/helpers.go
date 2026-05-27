package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// getAuthUserID extracts user_id from gin context.
// Returns empty string if not found (only happens when RequireAuth middleware is not applied).
func getAuthUserID(c *gin.Context) string {
	uid, exists := c.Get("user_id")
	if !exists {
		respondError(c, http.StatusUnauthorized, "auth_required", "请先登录")
		c.Abort()
		return ""
	}
	return uid.(string)
}

// getOptionalUserID extracts user_id from context without requiring auth.
// Returns empty string if not present (for public endpoints that optionally use viewer info).
func getOptionalUserID(c *gin.Context) string {
	uid, exists := c.Get("user_id")
	if !exists {
		return ""
	}
	s, ok := uid.(string)
	if !ok {
		return ""
	}
	return s
}

// getAuthJTI extracts jti from context.
func getAuthJTI(c *gin.Context) string {
	jti, _ := c.Get("jti")
	if jti == nil {
		return ""
	}
	return jti.(string)
}

// nilIfEmpty returns nil pointer if string is empty
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// parseInt parses string to int, returns (0, false) on error
func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func deprecatedMobileEndpoint(c *gin.Context) {
	respondError(c, http.StatusGone, "deprecated_endpoint", "这个入口已下线，请更新到新版年糕")
}

func respondError(c *gin.Context, status int, code string, message string) {
	respondErrorWith(c, status, code, message, nil)
}

func respondErrorWith(c *gin.Context, status int, code string, message string, extra gin.H) {
	errorBody := gin.H{
		"code":    code,
		"message": message,
	}
	if requestID := requestIDFromContext(c); requestID != "" {
		errorBody["request_id"] = requestID
	}
	body := gin.H{"error": errorBody}
	for key, value := range extra {
		body[key] = value
	}
	c.JSON(status, body)
}

func requestIDFromContext(c *gin.Context) string {
	raw, exists := c.Get("request_id")
	if !exists {
		return ""
	}
	requestID, _ := raw.(string)
	return requestID
}
