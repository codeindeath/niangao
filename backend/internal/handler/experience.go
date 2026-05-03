package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/niangao/backend/internal/middleware"
	"github.com/niangao/backend/internal/model"
	"github.com/niangao/backend/internal/repository"
)

type ExperienceHandler struct {
	repo      *repository.ExperienceRepo
	likeRepo  *repository.LikeRepo
	bookRepo  *repository.BookmarkRepo
}

func RegisterExperienceRoutes(r *gin.RouterGroup, expRepo *repository.ExperienceRepo, likeRepo *repository.LikeRepo, bookRepo *repository.BookmarkRepo) {
	h := &ExperienceHandler{repo: expRepo, likeRepo: likeRepo, bookRepo: bookRepo}

	exp := r.Group("/experiences")
	{
		exp.GET("", h.List)
		exp.GET("/:id", h.Get)
		exp.POST("", middleware.RequireAuth(), h.Create)
		exp.PUT("/:id", middleware.RequireAuth(), h.Update)
		exp.DELETE("/:id", middleware.RequireAuth(), h.Delete)
		exp.POST("/:id/like", middleware.RequireAuth(), h.ToggleLike)
		exp.POST("/:id/bookmark", middleware.RequireAuth(), h.ToggleBookmark)
	}
}

func (h *ExperienceHandler) List(c *gin.Context) {
	var query model.ExperienceListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	viewerID, _ := c.Get("user_id")
	viewerStr := ""
	if viewerID != nil {
		viewerStr = viewerID.(string)
	}

	experiences, total, err := h.repo.List(c.Request.Context(), query, viewerStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list experiences"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  experiences,
		"total": total,
		"page":  query.Page,
	})
}

func (h *ExperienceHandler) Get(c *gin.Context) {
	id := c.Param("id")
	viewerID, _ := c.Get("user_id")
	viewerStr := ""
	if viewerID != nil {
		viewerStr = viewerID.(string)
	}

	exp, err := h.repo.GetByID(c.Request.Context(), id, viewerStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "experience not found"})
		return
	}

	c.JSON(http.StatusOK, exp)
}

func (h *ExperienceHandler) Create(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}

	var req model.CreateExperienceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !model.IsValidDomain(req.Domain) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain"})
		return
	}

	exp, err := h.repo.Create(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create experience"})
		return
	}

	c.JSON(http.StatusCreated, exp)
}

func (h *ExperienceHandler) Update(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	id := c.Param("id")

	var req model.CreateExperienceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !model.IsValidDomain(req.Domain) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain"})
		return
	}

	if err := h.repo.Update(c.Request.Context(), id, userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *ExperienceHandler) Delete(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	id := c.Param("id")

	if err := h.repo.Delete(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *ExperienceHandler) ToggleLike(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	id := c.Param("id")

	liked, err := h.likeRepo.Toggle(c.Request.Context(), userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to toggle like"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"liked": liked})
}

func (h *ExperienceHandler) ToggleBookmark(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	id := c.Param("id")

	bookmarked, err := h.bookRepo.Toggle(c.Request.Context(), userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to toggle bookmark"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"bookmarked": bookmarked})
}

func parseIntParam(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
