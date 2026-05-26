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

const (
	defaultFeedLimit = 20
	maxFeedLimit     = 50
)

type V4FeedStore interface {
	RecommendFeed(ctx context.Context, userID string, limit int, cursor string) (*model.FeedPage, error)
	CollectionsFeed(ctx context.Context, userID string, limit int, cursor string) (*model.FeedPage, error)
	MineFeed(ctx context.Context, userID string, limit int, cursor string) (*model.FeedPage, error)
}

type FeedHandler struct {
	store V4FeedStore
}

func RegisterFeedRoutes(r *gin.RouterGroup, store V4FeedStore) {
	h := &FeedHandler{store: store}

	feed := r.Group("/feed")
	{
		feed.GET("/recommend", h.Recommend)
		feed.GET("/collections", middleware.RequireAuth(), h.Collections)
		feed.GET("/mine", middleware.RequireAuth(), h.Mine)
	}
}

func (h *FeedHandler) Recommend(c *gin.Context) {
	h.respond(c, func(ctx context.Context, limit int, cursor string) (*model.FeedPage, error) {
		return h.store.RecommendFeed(ctx, getOptionalUserID(c), limit, cursor)
	})
}

func (h *FeedHandler) Collections(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	h.respond(c, func(ctx context.Context, limit int, cursor string) (*model.FeedPage, error) {
		return h.store.CollectionsFeed(ctx, userID, limit, cursor)
	})
}

func (h *FeedHandler) Mine(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	h.respond(c, func(ctx context.Context, limit int, cursor string) (*model.FeedPage, error) {
		return h.store.MineFeed(ctx, userID, limit, cursor)
	})
}

func (h *FeedHandler) respond(c *gin.Context, load func(context.Context, int, string) (*model.FeedPage, error)) {
	page, err := load(c.Request.Context(), parseFeedLimit(c), c.Query("cursor"))
	if err != nil {
		log.Printf("v4 feed failed path=%s: %v", c.FullPath(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load feed"})
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

func parseFeedLimit(c *gin.Context) int {
	limit := defaultFeedLimit
	if raw := c.Query("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}
	if limit < 1 {
		return defaultFeedLimit
	}
	if limit > maxFeedLimit {
		return maxFeedLimit
	}
	return limit
}
