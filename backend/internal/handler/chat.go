package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
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
	ConversationID string          `json:"conversation_id"`
	Messages       []model.Message `json:"messages"`
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
	Reply                   string   `json:"reply"`
	ReferencedExperienceIDs []string `json:"referenced_experience_ids"`
	MessageID               string   `json:"message_id"`
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

	// Load bookmarked experiences closest to the current chat topic.
	inferredDomain := inferChatDomain(req.Message, history)
	bookmarks, err := h.bookmarkRepo.ListBookmarkedExperiencesForChat(c.Request.Context(), userID, inferredDomain, 50)
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
	Message               string                            `json:"message"`
	UserID                string                            `json:"user_id"`
	History               []aiHistoryMsg                    `json:"history"`
	BookmarkedExperiences []repository.BookmarkedExperience `json:"bookmarked_experiences"`
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

var chatDomainOrder = []string{
	string(model.DomainWork),
	string(model.DomainRelationship),
	string(model.DomainVitality),
	string(model.DomainMeaning),
	string(model.DomainCognition),
	string(model.DomainLiving),
}

var chatDomainKeywords = map[string][]string{
	string(model.DomainVitality): {
		"身体", "健康", "睡眠", "失眠", "运动", "锻炼", "健身", "疲惫", "疲劳",
		"生病", "医院", "饮食", "吃饭", "体能", "疼", "痛", "焦虑", "压力",
		"health", "sleep", "exercise", "diet",
	},
	string(model.DomainLiving): {
		"生活", "租房", "房子", "居住", "通勤", "出行", "旅行", "宠物", "猫",
		"狗", "购物", "衣服", "穿搭", "护肤", "娱乐", "家务", "收纳",
		"life", "home", "travel", "shopping",
	},
	string(model.DomainWork): {
		"工作", "职场", "老板", "上司", "同事", "项目", "绩效", "晋升", "升职",
		"加班", "会议", "汇报", "管理", "创业", "产品", "客户", "面试", "求职",
		"简历", "效率", "deadline", "work", "job", "career", "startup", "manager",
	},
	string(model.DomainRelationship): {
		"关系", "恋爱", "分手", "伴侣", "对象", "男朋友", "女朋友", "老公", "老婆",
		"婚姻", "夫妻", "朋友", "友情", "父母", "家人", "孩子", "亲子", "吵架",
		"沟通", "relationship", "friend", "partner", "family",
	},
	string(model.DomainCognition): {
		"学习", "思考", "认知", "信息", "知识", "工具", "表达", "写作", "决策",
		"判断", "思维", "模型", "创造", "复盘", "方法", "阅读",
		"learn", "think", "decision", "writing", "tool",
	},
	string(model.DomainMeaning): {
		"意义", "自我", "幸福", "情绪", "信仰", "使命", "孤独", "迷茫", "价值", "人生",
		"归属", "内耗", "空虚", "方向", "存在", "identity", "meaning", "purpose",
	},
}

func inferChatDomain(message string, history []model.Message) string {
	scores := make(map[string]int)
	scoreChatDomainText(scores, message, 4)

	start := 0
	if len(history) > 8 {
		start = len(history) - 8
	}
	for _, msg := range history[start:] {
		weight := 1
		if msg.Role == "user" {
			weight = 2
		}
		scoreChatDomainText(scores, msg.Content, weight)
	}

	bestDomain := ""
	bestScore := 0
	for _, domain := range chatDomainOrder {
		if scores[domain] > bestScore {
			bestDomain = domain
			bestScore = scores[domain]
		}
	}
	return bestDomain
}

func scoreChatDomainText(scores map[string]int, text string, weight int) {
	if text == "" || weight <= 0 {
		return
	}
	normalized := strings.ToLower(text)
	for domain, keywords := range chatDomainKeywords {
		for _, keyword := range keywords {
			if strings.Contains(normalized, strings.ToLower(keyword)) {
				scores[domain] += weight
			}
		}
	}
}
