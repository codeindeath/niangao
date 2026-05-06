package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niangao/backend/internal/auth"
)

// AuthMiddleware 验证自签 JWT，检查吊销表，注入 user_id 到 context
func AuthMiddleware(jwtSecret string, db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		claims, err := auth.ParseToken(jwtSecret, parts[1])
		if err != nil {
			c.Next()
			return
		}

		// Check token revocation (async log on error, don't block)
		if claims.ID != "" && db != nil {
			revoked, err := auth.IsTokenRevoked(c.Request.Context(), db, claims.ID)
			if err != nil {
				log.Printf("token revocation check failed: %v", err)
			}
			if revoked {
				c.Next() // let RequireAuth handle it if route is protected
				return
			}
		}

		// Check if user is disabled
		if db != nil {
			var deleted bool
			err := db.QueryRow(c.Request.Context(),
				`SELECT deleted_at IS NOT NULL FROM users WHERE id=$1`, claims.UserID,
			).Scan(&deleted)
			if err == nil && deleted {
				// User disabled — don't set identity so RequireAuth will 401
				c.Next()
				return
			}
		}

		c.Set("user_id", claims.UserID)
		c.Set("open_id", claims.OpenID)
		c.Set("nickname", claims.Nickname)
		c.Set("jti", claims.ID)
		c.Next()
	}
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists || userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetAuthUserID extracts user_id from gin context (set by AuthMiddleware)
func GetAuthUserID(c *gin.Context) string {
	uid, _ := c.Get("user_id")
	if uid == nil {
		return ""
	}
	return uid.(string)
}

// GetAuthJTI extracts jti from gin context
func GetAuthJTI(c *gin.Context) string {
	jti, _ := c.Get("jti")
	if jti == nil {
		return ""
	}
	return jti.(string)
}

// CleanupExpiredTokens periodically removes expired refresh_tokens and token_revocations
func CleanupExpiredTokens(ctx context.Context, db *pgxpool.Pool) {
	_, err := db.Exec(ctx, `SELECT cleanup_expired_tokens()`)
	if err != nil {
		log.Printf("token cleanup failed: %v", err)
	}
}
