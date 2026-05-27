package handler

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/niangao/backend/internal/middleware"
	"github.com/niangao/backend/internal/model"
)

type V4MeStatsStore interface {
	AssetStats(ctx context.Context, userID string) (*model.AssetStats, error)
	ContributionStats(ctx context.Context, userID string) (*model.ContributionStats, error)
	ChangeStats(ctx context.Context, userID string) (*model.ChangeStats, error)
	RecentHarvestStats(ctx context.Context, userID string, rangeKey string) (*model.RecentHarvestStats, error)
	RecentRespondedExperiences(ctx context.Context, userID string, limit int) ([]model.RespondedExperienceCard, error)
}

type MeStatsHandler struct {
	store V4MeStatsStore
}

func RegisterMeStatsRoutes(r *gin.RouterGroup, store V4MeStatsStore) {
	h := &MeStatsHandler{store: store}

	stats := r.Group("/me/stats", middleware.RequireAuth())
	{
		stats.GET("/assets", h.Assets)
		stats.GET("/contribution", h.Contribution)
		stats.GET("/change", h.Change)
		stats.GET("/recent-harvest", h.RecentHarvest)
	}
	r.GET("/me/recent-responded-experiences", middleware.RequireAuth(), h.RecentRespondedExperiences)
}

func (h *MeStatsHandler) Assets(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	stats, err := h.store.AssetStats(c.Request.Context(), userID)
	if err != nil {
		log.Printf("v4 asset stats failed user=%s: %v", userID, err)
		respondError(c, http.StatusInternalServerError, "asset_stats_load_failed", "failed to load asset stats")
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *MeStatsHandler) Contribution(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	stats, err := h.store.ContributionStats(c.Request.Context(), userID)
	if err != nil {
		log.Printf("v4 contribution stats failed user=%s: %v", userID, err)
		respondError(c, http.StatusInternalServerError, "contribution_stats_load_failed", "failed to load contribution stats")
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *MeStatsHandler) Change(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	stats, err := h.store.ChangeStats(c.Request.Context(), userID)
	if err != nil {
		log.Printf("v4 change stats failed user=%s: %v", userID, err)
		respondError(c, http.StatusInternalServerError, "change_stats_load_failed", "failed to load change stats")
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *MeStatsHandler) RecentHarvest(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	rangeKey := c.DefaultQuery("range", "30d")
	if rangeKey != "7d" && rangeKey != "30d" && rangeKey != "all" {
		respondError(c, http.StatusBadRequest, "invalid_range", "invalid range")
		return
	}
	stats, err := h.store.RecentHarvestStats(c.Request.Context(), userID, rangeKey)
	if err != nil {
		log.Printf("v4 recent harvest stats failed user=%s range=%s: %v", userID, rangeKey, err)
		respondError(c, http.StatusInternalServerError, "recent_harvest_stats_load_failed", "failed to load recent harvest stats")
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *MeStatsHandler) RecentRespondedExperiences(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	limit := 3
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			respondError(c, http.StatusBadRequest, "invalid_limit", "invalid limit")
			return
		}
		limit = parsed
	}
	if limit > 10 {
		limit = 10
	}
	cards, err := h.store.RecentRespondedExperiences(c.Request.Context(), userID, limit)
	if err != nil {
		log.Printf("v4 recent responded experiences failed user=%s limit=%d: %v", userID, limit, err)
		respondError(c, http.StatusInternalServerError, "recent_responded_load_failed", "failed to load recent responded experiences")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": cards})
}
