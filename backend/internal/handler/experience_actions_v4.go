package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/niangao/backend/internal/middleware"
	"github.com/niangao/backend/internal/model"
	"github.com/niangao/backend/internal/repository"
)

type V4ExperienceActionStore interface {
	InspireExperience(ctx context.Context, userID string, experienceID string) (bool, error)
	CollectExperience(ctx context.Context, userID string, experienceID string) (bool, error)
	UncollectExperience(ctx context.Context, userID string, experienceID string) error
	RecordExperienceEvent(ctx context.Context, userID string, experienceID string, event model.ExperienceEventRequest) error
}

type ExperienceActionHandler struct {
	store V4ExperienceActionStore
}

type ExperienceEventRequest = model.ExperienceEventRequest

const (
	experienceUnavailableMessage = "这条经验暂时看不了"
	inspirationFailedMessage     = "暂时标记不了，请稍后再试"
	collectFailedMessage         = "暂时收藏不了，请稍后再试"
	uncollectFailedMessage       = "暂时取消不了，请稍后再试"
	eventRequestInvalidMessage   = "操作格式不对"
	eventRecordFailedMessage     = "暂时记录不了操作"
)

func RegisterExperienceActionRoutes(r *gin.RouterGroup, store V4ExperienceActionStore) {
	h := &ExperienceActionHandler{store: store}

	exp := r.Group("/experiences")
	{
		exp.POST("/:id/inspire", middleware.RequireAuth(), h.Inspire)
		exp.POST("/:id/collect", middleware.RequireAuth(), h.Collect)
		exp.DELETE("/:id/collect", middleware.RequireAuth(), h.Uncollect)
		exp.POST("/:id/events", h.RecordEvent)
	}
}

func (h *ExperienceActionHandler) Inspire(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}

	already, err := h.store.InspireExperience(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		if errors.Is(err, repository.ErrExperienceUnavailable) {
			respondError(c, http.StatusNotFound, "experience_not_found", experienceUnavailableMessage)
			return
		}
		log.Printf("v4 inspire failed experience=%s user=%s: %v", c.Param("id"), userID, err)
		respondError(c, http.StatusInternalServerError, "inspiration_failed", inspirationFailedMessage)
		return
	}
	if already {
		c.JSON(http.StatusConflict, gin.H{"inspired": true, "code": "already_inspired"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"inspired": true})
}

func (h *ExperienceActionHandler) Collect(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}

	already, err := h.store.CollectExperience(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		if errors.Is(err, repository.ErrExperienceUnavailable) {
			respondError(c, http.StatusNotFound, "experience_not_found", experienceUnavailableMessage)
			return
		}
		log.Printf("v4 collect failed experience=%s user=%s: %v", c.Param("id"), userID, err)
		respondError(c, http.StatusInternalServerError, "collect_failed", collectFailedMessage)
		return
	}
	c.JSON(http.StatusOK, gin.H{"collected": true, "already_collected": already})
}

func (h *ExperienceActionHandler) Uncollect(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}

	if err := h.store.UncollectExperience(c.Request.Context(), userID, c.Param("id")); err != nil {
		if errors.Is(err, repository.ErrExperienceUnavailable) {
			respondError(c, http.StatusNotFound, "experience_not_found", experienceUnavailableMessage)
			return
		}
		log.Printf("v4 uncollect failed experience=%s user=%s: %v", c.Param("id"), userID, err)
		respondError(c, http.StatusInternalServerError, "uncollect_failed", uncollectFailedMessage)
		return
	}
	c.JSON(http.StatusOK, gin.H{"collected": false})
}

func (h *ExperienceActionHandler) RecordEvent(c *gin.Context) {
	var req ExperienceEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_event_request", eventRequestInvalidMessage)
		return
	}
	if err := normalizeExperienceEventRequest(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_event", err.Error())
		return
	}

	if err := h.store.RecordExperienceEvent(c.Request.Context(), getOptionalUserID(c), c.Param("id"), req); err != nil {
		if errors.Is(err, repository.ErrExperienceUnavailable) {
			respondError(c, http.StatusNotFound, "experience_not_found", experienceUnavailableMessage)
			return
		}
		log.Printf("v4 event failed event=%s experience=%s user=%s: %v", req.EventType, c.Param("id"), getOptionalUserID(c), err)
		respondError(c, http.StatusInternalServerError, "experience_event_failed", eventRecordFailedMessage)
		return
	}
	c.Status(http.StatusNoContent)
}

func normalizeExperienceEventRequest(req *ExperienceEventRequest) error {
	req.EventType = strings.TrimSpace(req.EventType)
	req.SourceContext = strings.TrimSpace(req.SourceContext)
	req.ContextID = strings.TrimSpace(req.ContextID)
	if req.Metadata == nil {
		req.Metadata = map[string]any{}
	}

	switch req.EventType {
	case "expose", "flip", "search_click", "chat_citation_show", "chat_citation_click":
	default:
		return errors.New("操作类型不支持")
	}

	if req.SourceContext == "" {
		if req.EventType == "search_click" {
			req.SourceContext = "search"
		} else {
			req.SourceContext = "app"
		}
	}
	if len(req.SourceContext) > 32 {
		return errors.New("来源信息过长")
	}
	if req.ContextID != "" && !isUUIDLike(req.ContextID) {
		return errors.New("上下文标识不正确")
	}
	return nil
}
