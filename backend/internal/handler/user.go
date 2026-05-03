package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niangao/backend/internal/middleware"
)

func RegisterUserRoutes(r *gin.RouterGroup, db *pgxpool.Pool) {
	user := r.Group("/user", middleware.RequireAuth())
	{
		user.GET("/profile", func(c *gin.Context) {
			getProfile(c, db)
		})
	}
}

func getProfile(c *gin.Context, db *pgxpool.Pool) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}

	var nickname string
	var avatarURL *string
	var expCount, bmCount, practicedCount int

	err := db.QueryRow(c.Request.Context(),
		`SELECT nickname, avatar_url, experience_count, bookmark_count, practiced_count
		 FROM users WHERE id = $1`, userID,
	).Scan(&nickname, &avatarURL, &expCount, &bmCount, &practicedCount)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":               userID,
		"nickname":         nickname,
		"avatar_url":       avatarURL,
		"experience_count": expCount,
		"bookmark_count":   bmCount,
		"practiced_count":  practicedCount,
	})
}
