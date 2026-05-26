package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niangao/backend/internal/middleware"
)

func RegisterMeAccountRoutes(r *gin.RouterGroup, db *pgxpool.Pool) {
	me := r.Group("/me", middleware.RequireAuth())
	{
		me.DELETE("/account", func(c *gin.Context) {
			deleteMeAccount(c, db)
		})
	}
}

func deleteMeAccount(c *gin.Context, db *pgxpool.Pool) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}

	ct, err := db.Exec(c.Request.Context(),
		`DELETE FROM users WHERE id = $1`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除账号失败"})
		return
	}
	if ct.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "账号已删除"})
}
