package handler

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/niangao/backend/internal/model"
)

type V4SearchStore interface {
	SearchExperiences(ctx context.Context, userID string, query string, limit int, cursor string) (*model.FeedPage, error)
}

type SearchHandler struct {
	store V4SearchStore
}

func RegisterSearchRoutes(r *gin.RouterGroup, store V4SearchStore) {
	h := &SearchHandler{store: store}
	r.GET("/search/experiences", h.SearchExperiences)
}

func (h *SearchHandler) SearchExperiences(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusOK, model.FeedPage{Data: []model.ExperienceCard{}})
		return
	}

	page, err := h.store.SearchExperiences(c.Request.Context(), getOptionalUserID(c), query, parseFeedLimit(c), c.Query("cursor"))
	if err != nil {
		log.Printf("v4 search failed query=%q: %v", query, err)
		respondError(c, http.StatusInternalServerError, "search_failed", "failed to search experiences")
		return
	}
	if page == nil {
		page = &model.FeedPage{Data: []model.ExperienceCard{}}
	}
	if page.Data == nil {
		page.Data = []model.ExperienceCard{}
	}
	c.JSON(http.StatusOK, page)
}
