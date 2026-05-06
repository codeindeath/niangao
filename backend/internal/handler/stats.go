package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/niangao/backend/internal/middleware"
	"github.com/niangao/backend/internal/repository"
)

func RegisterStatsRoutes(r *gin.RouterGroup, repo *repository.StatsRepo) {
	r.GET("/user/stats", middleware.RequireAuth(), func(c *gin.Context) {
		getStats(c, repo)
	})
	// Record a view (fire-and-forget from client)
	r.POST("/experiences/:id/view", middleware.RequireAuth(), func(c *gin.Context) {
		recordView(c, repo)
	})
}

func getStats(c *gin.Context, repo *repository.StatsRepo) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}

	stats, err := repo.GetStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计数据失败"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func recordView(c *gin.Context, repo *repository.StatsRepo) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	experienceID := c.Param("id")
	if experienceID == "" {
		c.Status(http.StatusBadRequest)
		return
	}
	if err := repo.RecordView(c.Request.Context(), userID, experienceID); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusNoContent)
}
