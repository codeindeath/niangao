package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niangao/backend/internal/middleware"
	"github.com/niangao/backend/internal/model"
	"github.com/niangao/backend/internal/repository"
)

// ChatHandler orchestrates the chat feature: persistence, context loading, AI proxy.
type ChatHandler struct {
	convRepo     *repository.ConversationRepo
	bookmarkRepo *repository.BookmarkRepo
	aiBaseURL    string
	httpClient   *http.Client
}

// RegisterChatRoutes registers chat endpoints under the given router group.
func RegisterChatRoutes(r *gin.RouterGroup, convRepo *repository.ConversationRepo, bookmarkRepo *repository.BookmarkRepo, aiBaseURL string) {
	h := &ChatHandler{
		convRepo:     convRepo,
		bookmarkRepo: bookmarkRepo,
		aiBaseURL:    aiBaseURL,
		httpClient:   &http.Client{Timeout: 60 * time.Second},
	}

	chat := r.Group("/chat", middleware.RequireAuth())
	{
		chat.GET("", h.InitChat)
		chat.POST("/send", h.SendMessage)
	}
}

// InitChatResponse is the response for GET /chat.
type InitChatResponse struct {
	ConversationID string           `json:"conversation_id"`
	Messages       []model.Message  `json:"messages"`
}

// InitChat loads chat history and auto-generates a greeting if needed.
// GET /api/v1/chat
func (h *ChatHandler) InitChat(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}

	// Get or create the user's single conversation
	conv, err := h.convRepo.GetOrCreateByUser(c.Request.Context(), userID)
	if err != nil {
		log.Printf("[chat] get or create conversation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加载对话失败"})
		return
	}

	// Load messages from last 30 days
	messages, err := h.convRepo.GetMessagesSince(c.Request.Context(), conv.ID, 30*24*time.Hour)
	if err != nil {
		log.Printf("[chat] get messages: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加载消息失败"})
		return
	}

	// Determine if we should greet
	shouldGreet := false
	if len(messages) == 0 {
		shouldGreet = true
	} else {
		lastMsg := messages[len(messages)-1]
		if lastMsg.Role == "user" && time.Since(lastMsg.CreatedAt) > 2*time.Hour {
			shouldGreet = true
		}
	}

	if shouldGreet {
		var greeting string
		if len(messages) == 0 {
			greeting = pickWelcomeGreeting()
		} else {
			lastMsg := messages[len(messages)-1]
			greeting = pickGreeting(lastMsg.CreatedAt)
		}

		aiMsg, err := h.convRepo.AddMessage(c.Request.Context(), conv.ID, "assistant", greeting, nil)
		if err != nil {
			log.Printf("[chat] add greeting: %v", err)
			// Non-fatal: return messages without the greeting
		} else {
			messages = append(messages, *aiMsg)
		}
	}

	c.JSON(http.StatusOK, InitChatResponse{
		ConversationID: conv.ID,
		Messages:       messages,
	})
}

// sendMessageRequest is the request body for POST /chat/send.
type sendMessageRequest struct {
	ConversationID string `json:"conversation_id" binding:"required"`
	Message        string `json:"message" binding:"required"`
}

// sendMessageResponse is the response for POST /chat/send.
type sendMessageResponse struct {
	Reply                    string   `json:"reply"`
	ReferencedExperienceIDs  []string `json:"referenced_experience_ids"`
	MessageID                string   `json:"message_id"`
}

// SendMessage handles sending a user message and getting an AI reply.
// POST /api/v1/chat/send
func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}

	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "消息不能为空"})
		return
	}

	// Verify conversation belongs to user
	conv, err := h.convRepo.GetByID(c.Request.Context(), req.ConversationID)
	if err != nil || conv.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "对话不存在"})
		return
	}

	// Rate limit: 100 rounds per day
	todayCount, err := h.convRepo.CountTodayMessages(c.Request.Context(), req.ConversationID)
	if err != nil {
		log.Printf("[chat] count today: %v", err)
	}
	if todayCount >= 100 {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "今日对话已达上限（100轮），明天再来聊吧"})
		return
	}

	// Save user message
	_, err = h.convRepo.AddMessage(c.Request.Context(), req.ConversationID, "user", req.Message, nil)
	if err != nil {
		log.Printf("[chat] save user msg: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存消息失败"})
		return
	}

	// Load 48h context for AI
	history, err := h.convRepo.GetMessagesSince(c.Request.Context(), req.ConversationID, 48*time.Hour)
	if err != nil {
		log.Printf("[chat] get history: %v", err)
		history = make([]model.Message, 0)
	}

	// Load bookmarked experiences
	bookmarks, err := h.bookmarkRepo.ListBookmarkedExperiences(c.Request.Context(), userID)
	if err != nil {
		log.Printf("[chat] get bookmarks: %v", err)
		bookmarks = make([]repository.BookmarkedExperience, 0)
	}

	// Call AI service
	aiReply, refIDs, err := h.callAIService(userID, req.Message, history, bookmarks)
	if err != nil {
		log.Printf("[chat] AI service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "对话服务暂时不可用，请稍后再试"})
		return
	}

	// Save AI reply
	aiMsg, err := h.convRepo.AddMessage(c.Request.Context(), req.ConversationID, "assistant", aiReply, refIDs)
	if err != nil {
		log.Printf("[chat] save ai msg: %v", err)
		// Non-fatal: return reply anyway
	}

	msgID := ""
	if aiMsg != nil {
		msgID = aiMsg.ID
	}

	c.JSON(http.StatusOK, sendMessageResponse{
		Reply:                   aiReply,
		ReferencedExperienceIDs: refIDs,
		MessageID:               msgID,
	})
}

// aiChatRequest is the request sent to the AI service.
type aiChatRequest struct {
	Message                string                              `json:"message"`
	UserID                 string                              `json:"user_id"`
	History                []aiHistoryMsg                      `json:"history"`
	BookmarkedExperiences  []repository.BookmarkedExperience   `json:"bookmarked_experiences"`
}

type aiHistoryMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// aiChatResponse is the response from the AI service.
type aiChatResponse struct {
	Reply                   string   `json:"reply"`
	ReferencedExperienceIDs []string `json:"referenced_experience_ids"`
}

// callAIService calls the Python AI service with full context.
func (h *ChatHandler) callAIService(userID, message string, history []model.Message, bookmarks []repository.BookmarkedExperience) (string, []string, error) {
	// Convert history to AI service format
	aiHistory := make([]aiHistoryMsg, 0, len(history))
	for _, m := range history {
		aiHistory = append(aiHistory, aiHistoryMsg{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	reqBody := aiChatRequest{
		Message:               message,
		UserID:                userID,
		History:               aiHistory,
		BookmarkedExperiences: bookmarks,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("marshal request: %w", err)
	}

	url := h.aiBaseURL + "/api/v1/chat/send"
	resp, err := h.httpClient.Post(url, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", nil, fmt.Errorf("post to AI: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("read AI response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("AI service returned %d: %s", resp.StatusCode, string(respBytes))
	}

	var aiResp aiChatResponse
	if err := json.Unmarshal(respBytes, &aiResp); err != nil {
		return "", nil, fmt.Errorf("unmarshal AI response: %w", err)
	}

	return aiResp.Reply, aiResp.ReferencedExperienceIDs, nil
}
