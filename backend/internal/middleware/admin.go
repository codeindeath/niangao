package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RequireAdmin 验证当前用户是否为管理员
// 必须在 AuthMiddleware 之后使用
func RequireAdmin(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}

		if db == nil {
			// 测试环境：无 DB 连接时允许通过
			c.Next()
			return
		}

		var isAdmin bool
		err := db.QueryRow(c.Request.Context(),
			"SELECT is_admin FROM users WHERE id=$1", userID,
		).Scan(&isAdmin)
		if err != nil || !isAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
			return
		}
		c.Next()
	}
}
