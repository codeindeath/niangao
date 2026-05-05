package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// getAuthUserID extracts user_id from gin context.
// Returns empty string if not found (only happens when RequireAuth middleware is not applied).
func getAuthUserID(c *gin.Context) string {
	uid, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
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