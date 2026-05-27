package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/niangao/backend/internal/middleware"
)

type V4MeFeedbackStore interface {
	CreateMeFeedback(ctx context.Context, userID, feedbackType, content, appVersion, device, osVersion string) error
}

type MeFeedbackHandler struct {
	store V4MeFeedbackStore
}

type meFeedbackRequest struct {
	Type       string `json:"type"`
	Content    string `json:"content"`
	AppVersion string `json:"app_version"`
	Device     string `json:"device"`
	OSVersion  string `json:"os_version"`
}

func RegisterMeFeedbackRoutes(r *gin.RouterGroup, store V4MeFeedbackStore) {
	h := &MeFeedbackHandler{store: store}
	me := r.Group("/me", middleware.RequireAuth())
	{
		me.POST("/feedback", h.Create)
	}
}

func (h *MeFeedbackHandler) Create(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}

	var req meFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_feedback_payload", "invalid feedback payload")
		return
	}
	if err := normalizeMeFeedbackRequest(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_feedback", err.Error())
		return
	}

	if err := h.store.CreateMeFeedback(c.Request.Context(), userID, req.Type, req.Content, req.AppVersion, req.Device, req.OSVersion); err != nil {
		log.Printf("v4 me feedback failed user=%s: %v", userID, err)
		respondError(c, http.StatusInternalServerError, "feedback_submit_failed", "failed to submit feedback")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "ok"})
}

func normalizeMeFeedbackRequest(req *meFeedbackRequest) error {
	req.Type = strings.TrimSpace(req.Type)
	if req.Type == "" {
		req.Type = "general"
	}
	req.Content = strings.TrimSpace(req.Content)
	req.AppVersion = strings.TrimSpace(req.AppVersion)
	req.Device = strings.TrimSpace(req.Device)
	req.OSVersion = strings.TrimSpace(req.OSVersion)

	if req.Content == "" {
		return errors.New("反馈内容不能为空")
	}
	if len([]rune(req.Content)) > 1000 {
		return errors.New("反馈内容不超过 1000 字")
	}
	if len([]rune(req.Type)) > 32 {
		return errors.New("反馈类型不超过 32 字")
	}
	if len([]rune(req.AppVersion)) > 64 || len([]rune(req.Device)) > 128 || len([]rune(req.OSVersion)) > 64 {
		return errors.New("客户端信息过长")
	}
	return nil
}
