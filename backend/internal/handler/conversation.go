package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/niangao/backend/internal/middleware"
	"github.com/niangao/backend/internal/repository"
)

type ConversationHandler struct {
	repo *repository.ConversationRepo
}

func RegisterConversationRoutes(r *gin.RouterGroup, repo *repository.ConversationRepo) {
	h := &ConversationHandler{repo: repo}

	conv := r.Group("/conversations", middleware.RequireAuth())
	{
		conv.GET("", h.List)
		conv.POST("", h.Create)
		conv.GET("/:id/messages", h.GetMessages)
	}
}

func (h *ConversationHandler) List(c *gin.Context) {
	userID, _ := c.Get("user_id")

	convs, err := h.repo.ListByUser(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list conversations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": convs})
}

func (h *ConversationHandler) Create(c *gin.Context) {
	userID, _ := c.Get("user_id")

	conv, err := h.repo.Create(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create conversation"})
		return
	}

	c.JSON(http.StatusCreated, conv)
}

func (h *ConversationHandler) GetMessages(c *gin.Context) {
	userID, _ := c.Get("user_id")
	convID := c.Param("id")

	_ = userID // In production, verify conversation belongs to user

	messages, err := h.repo.GetMessages(c.Request.Context(), convID, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": messages})
}
