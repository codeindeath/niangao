package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/niangao/backend/internal/middleware"
)

func RegisterUserRoutes(r *gin.RouterGroup) {
	user := r.Group("/user", middleware.RequireAuth())
	{
		user.GET("/profile", getProfile)
	}
}

func getProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	// TODO: Implement full profile with stats
	c.JSON(http.StatusOK, gin.H{
		"id":       userID,
		"nickname": "",
	})
}
