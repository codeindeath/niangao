package handler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/niangao/backend/internal/middleware"
	"github.com/niangao/backend/internal/model"
	"github.com/niangao/backend/internal/repository"
)

type V4ChatStore interface {
	RecentChatTopics(ctx context.Context, userID string, limit int) ([]model.ChatTopic, error)
	ChatTopics(ctx context.Context, userID string, limit int, cursor string) (*model.ChatTopicPage, error)
	CreateTempSession(ctx context.Context, userID string, forcedNewTopic bool) (*model.ChatTempSession, error)
	CreateChatTopic(ctx context.Context, userID string, req model.CreateChatTopicRequest) (*model.ChatTopic, error)
	UpdateChatTopic(ctx context.Context, userID string, topicID string, req model.UpdateChatTopicRequest) (*model.ChatTopic, error)
	DeleteChatTopic(ctx context.Context, userID string, topicID string) error
	ChatMessages(ctx context.Context, userID string, scope model.ChatMessageScope, limit int, cursor string) (*model.ChatMessagePage, error)
	VerifyChatScope(ctx context.Context, userID string, scope model.ChatMessageScope) (*model.ChatScopeContext, error)
	AddChatMessage(ctx context.Context, userID string, req model.SaveChatMessageRequest) (*model.ChatMessage, error)
	ChatDailyUsage(ctx context.Context, userID string) (int, int, error)
	RecentChatMessages(ctx context.Context, userID string, scope model.ChatMessageScope, limit int) ([]model.ChatMessage, error)
	PromoteTempSession(ctx context.Context, userID string, tempSessionID string, req model.PromoteChatTempSessionRequest) (*model.ChatTopic, error)
	CandidateExperiencesForChat(ctx context.Context, userID string, scope model.ChatScopeContext, userMessage string, riskLevel string, limit int) ([]model.ChatCandidateExperience, error)
	SaveChatCitations(ctx context.Context, assistantMessageID string, cards []model.ChatReferenceCard) error
}

const defaultChatDailyLimit = 50

type ChatGateway interface {
	GenerateChatReply(ctx context.Context, req model.ChatGatewayRequest) (*model.ChatGatewayResponse, error)
	ClassifyChatTopic(ctx context.Context, req model.ChatTopicClassificationRequest) (*model.ChatTopicClassificationResponse, error)
}

type ChatV4Handler struct {
	store   V4ChatStore
	gateway ChatGateway
}

func RegisterChatV4Routes(r *gin.RouterGroup, store V4ChatStore, gateways ...ChatGateway) {
	var gateway ChatGateway
	if len(gateways) > 0 {
		gateway = gateways[0]
	}
	h := &ChatV4Handler{store: store, gateway: gateway}

	chat := r.Group("/chat", middleware.RequireAuth())
	{
		chat.GET("/recent-topics", h.RecentTopics)
		chat.GET("/topics", h.Topics)
		chat.POST("/temp-sessions", h.CreateTempSession)
		chat.POST("/topics", h.CreateTopic)
		chat.PATCH("/topics/:id", h.UpdateTopic)
		chat.DELETE("/topics/:id", h.DeleteTopic)
		chat.GET("/topics/:id/messages", h.TopicMessages)
		chat.POST("/topics/:id/messages", h.SendTopicMessage)
		chat.POST("/temp-sessions/:id/messages", h.SendTempSessionMessage)
	}
}

func (h *ChatV4Handler) RecentTopics(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	topics, err := h.store.RecentChatTopics(c.Request.Context(), userID, 10)
	if err != nil {
		log.Printf("v4 recent topics failed user=%s: %v", userID, err)
		respondError(c, http.StatusInternalServerError, "recent_topics_load_failed", "暂时加载不了最近聊过的话题")
		return
	}
	if topics == nil {
		topics = []model.ChatTopic{}
	}
	c.JSON(http.StatusOK, gin.H{"data": topics})
}

func (h *ChatV4Handler) Topics(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	page, err := h.store.ChatTopics(c.Request.Context(), userID, parseFeedLimit(c), c.Query("cursor"))
	if err != nil {
		log.Printf("v4 topics failed user=%s: %v", userID, err)
		respondError(c, http.StatusInternalServerError, "topics_load_failed", "暂时加载不了话题")
		return
	}
	if page == nil {
		page = &model.ChatTopicPage{Data: []model.ChatTopic{}}
	}
	if page.Data == nil {
		page.Data = []model.ChatTopic{}
	}
	c.JSON(http.StatusOK, page)
}

func (h *ChatV4Handler) CreateTempSession(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	var req struct {
		ForcedNewTopic bool `json:"forced_new_topic"`
	}
	_ = c.ShouldBindJSON(&req)
	session, err := h.store.CreateTempSession(c.Request.Context(), userID, req.ForcedNewTopic)
	if err != nil {
		log.Printf("v4 create temp session failed user=%s: %v", userID, err)
		respondError(c, http.StatusInternalServerError, "temp_session_create_failed", "暂时开始不了对话")
		return
	}
	c.JSON(http.StatusCreated, session)
}

func (h *ChatV4Handler) CreateTopic(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	var req model.CreateChatTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_topic_payload", "话题内容不完整")
		return
	}
	topic, err := h.store.CreateChatTopic(c.Request.Context(), userID, req)
	if err != nil {
		log.Printf("v4 create topic failed user=%s: %v", userID, err)
		respondError(c, http.StatusInternalServerError, "topic_create_failed", "暂时创建不了话题")
		return
	}
	c.JSON(http.StatusCreated, topic)
}

func (h *ChatV4Handler) UpdateTopic(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	var req model.UpdateChatTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_topic_payload", "话题内容不完整")
		return
	}
	topic, err := h.store.UpdateChatTopic(c.Request.Context(), userID, c.Param("id"), req)
	if err != nil {
		log.Printf("v4 update topic failed user=%s topic=%s: %v", userID, c.Param("id"), err)
		respondError(c, http.StatusInternalServerError, "topic_update_failed", "暂时更新不了话题")
		return
	}
	c.JSON(http.StatusOK, topic)
}

func (h *ChatV4Handler) DeleteTopic(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	if err := h.store.DeleteChatTopic(c.Request.Context(), userID, c.Param("id")); err != nil {
		log.Printf("v4 delete topic failed user=%s topic=%s: %v", userID, c.Param("id"), err)
		respondError(c, http.StatusInternalServerError, "topic_delete_failed", "暂时删除不了话题")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *ChatV4Handler) TopicMessages(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	scope := model.ChatMessageScope{Kind: model.ChatScopeTopic, ID: c.Param("id")}
	page, err := h.store.ChatMessages(c.Request.Context(), userID, scope, parseFeedLimit(c), c.Query("cursor"))
	if err != nil {
		if errors.Is(err, repository.ErrExperienceUnavailable) {
			respondError(c, http.StatusNotFound, "topic_not_found", "这个话题不存在或已不可用")
			return
		}
		log.Printf("v4 chat messages failed user=%s topic=%s: %v", userID, scope.ID, err)
		respondError(c, http.StatusInternalServerError, "chat_messages_load_failed", "暂时加载不了对话")
		return
	}
	if page == nil {
		page = &model.ChatMessagePage{Data: []model.ChatMessage{}}
	}
	if page.Data == nil {
		page.Data = []model.ChatMessage{}
	}
	c.JSON(http.StatusOK, page)
}

func (h *ChatV4Handler) SendTopicMessage(c *gin.Context) {
	h.sendMessage(c, model.ChatMessageScope{Kind: model.ChatScopeTopic, ID: c.Param("id")})
}

func (h *ChatV4Handler) SendTempSessionMessage(c *gin.Context) {
	h.sendMessage(c, model.ChatMessageScope{Kind: model.ChatScopeTempSession, ID: c.Param("id")})
}

func (h *ChatV4Handler) sendMessage(c *gin.Context, scope model.ChatMessageScope) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}

	var req model.SendChatMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_message_payload", "消息内容不完整")
		return
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		respondError(c, http.StatusBadRequest, "message_content_required", "先写点想聊的内容")
		return
	}
	if len([]rune(content)) > 2000 {
		respondError(c, http.StatusBadRequest, "message_content_too_long", "这次说得太长了，先拆成一小段")
		return
	}

	scopeContext, err := h.store.VerifyChatScope(c.Request.Context(), userID, scope)
	if err != nil {
		if errors.Is(err, repository.ErrExperienceUnavailable) {
			respondError(c, http.StatusNotFound, "chat_scope_not_found", "这段对话不存在或已不可用")
			return
		}
		log.Printf("v4 verify chat scope failed user=%s scope=%s/%s: %v", userID, scope.Kind, scope.ID, err)
		respondError(c, http.StatusInternalServerError, "chat_scope_verify_failed", "暂时确认不了对话状态")
		return
	}

	used, limit, err := h.store.ChatDailyUsage(c.Request.Context(), userID)
	if err != nil {
		log.Printf("v4 check chat daily quota failed user=%s scope=%s/%s: %v", userID, scope.Kind, scope.ID, err)
		respondErrorWith(c, http.StatusServiceUnavailable, "chat_quota_unavailable", "暂时没法确认今日对话额度，请稍后再试。", gin.H{
			"retryable": true,
		})
		return
	}
	if limit < 1 {
		limit = defaultChatDailyLimit
	}
	if used >= limit {
		respondErrorWith(c, http.StatusTooManyRequests, "chat_quota_exceeded", chatQuotaExceededMessage(limit), gin.H{
			"retryable": false,
		})
		return
	}

	pre := classifyChatMessage(content)
	userMsg, err := h.store.AddChatMessage(c.Request.Context(), userID, model.SaveChatMessageRequest{
		Scope:           scope,
		Role:            "user",
		Content:         content,
		Status:          "sent",
		RiskLevel:       riskLevelForStorage(pre.RiskLevel),
		ClientMessageID: strings.TrimSpace(req.ClientMessageID),
		Metadata: map[string]any{
			"pre_classification": pre,
		},
	})
	if err != nil {
		log.Printf("v4 save user chat message failed user=%s scope=%s/%s: %v", userID, scope.Kind, scope.ID, err)
		respondError(c, http.StatusInternalServerError, "chat_message_save_failed", "暂时发送不了这条消息")
		return
	}

	if h.gateway == nil {
		respondErrorWith(c, http.StatusServiceUnavailable, "chat_service_unavailable", "暂时聊不了，请稍后再试。", gin.H{
			"retryable":       true,
			"user_message_id": userMsg.ID,
		})
		return
	}

	recent, err := h.store.RecentChatMessages(c.Request.Context(), userID, scope, 12)
	if err != nil {
		log.Printf("v4 load recent chat messages failed user=%s scope=%s/%s: %v", userID, scope.Kind, scope.ID, err)
		recent = []model.ChatMessage{}
	}
	candidates, err := h.store.CandidateExperiencesForChat(c.Request.Context(), userID, *scopeContext, content, pre.RiskLevel, 5)
	if err != nil {
		log.Printf("v4 load chat candidate experiences failed user=%s scope=%s/%s: %v", userID, scope.Kind, scope.ID, err)
		candidates = []model.ChatCandidateExperience{}
	}

	gatewayReq := model.ChatGatewayRequest{
		UserID:               userID,
		UserMessageID:        userMsg.ID,
		UserMessage:          content,
		SessionState:         scopeContext.SessionState,
		Scope:                scope,
		Topic:                scopeContext.Topic,
		RecentMessages:       recent,
		PreClassification:    pre,
		CandidateExperiences: candidates,
		ContextFlags:         chatContextFlags(pre),
		Limits: map[string]int{
			"max_reply_chars_soft": 500,
			"max_citation_cards":   maxChatCitationCards(pre),
		},
	}
	aiResp, err := h.gateway.GenerateChatReply(c.Request.Context(), gatewayReq)
	if err != nil || aiResp == nil || strings.TrimSpace(aiResp.ReplyText) == "" {
		log.Printf("v4 chat gateway failed user=%s scope=%s/%s message=%s: %v", userID, scope.Kind, scope.ID, userMsg.ID, err)
		respondErrorWith(c, http.StatusServiceUnavailable, "chat_service_unavailable", "暂时聊不了，请稍后再试。", gin.H{
			"retryable":       true,
			"user_message_id": userMsg.ID,
		})
		return
	}

	cards := buildReferenceCards(candidates, aiResp.Citations, gatewayReq.Limits["max_citation_cards"])
	refIDs := make([]string, 0, len(cards))
	for _, card := range cards {
		refIDs = append(refIDs, card.ExperienceID)
	}
	assistantMsg, err := h.store.AddChatMessage(c.Request.Context(), userID, model.SaveChatMessageRequest{
		Scope:                   scope,
		Role:                    "assistant",
		Content:                 strings.TrimSpace(aiResp.ReplyText),
		Status:                  "sent",
		RiskLevel:               riskLevelForStorage(aiResp.RiskLevel),
		ReferencedExperienceIDs: refIDs,
		Metadata: map[string]any{
			"reply_mode":    aiResp.ReplyMode,
			"emotion_level": aiResp.EmotionLevel,
			"warnings":      aiResp.Warnings,
		},
	})
	if err != nil {
		log.Printf("v4 save assistant chat message failed user=%s scope=%s/%s: %v", userID, scope.Kind, scope.ID, err)
		respondError(c, http.StatusInternalServerError, "assistant_message_save_failed", "暂时保存不了回复")
		return
	}
	if len(cards) > 0 {
		if err := h.store.SaveChatCitations(c.Request.Context(), assistantMsg.ID, cards); err != nil {
			log.Printf("v4 save chat citations failed message=%s: %v", assistantMsg.ID, err)
		}
	}

	promotedTopic := h.promoteTempSessionIfClear(c.Request.Context(), userID, scopeContext, scope)
	sessionState := scopeContext.SessionState
	if promotedTopic != nil {
		sessionState = "stable_topic"
		userMsg.TopicID = &promotedTopic.ID
		userMsg.TempSessionID = nil
		assistantMsg.TopicID = &promotedTopic.ID
		assistantMsg.TempSessionID = nil
	}

	noteSuggestion := model.ChatNoteSuggestion{ShouldShow: false, SourceMessageIDs: []string{}}
	if aiResp.NoteSuggestion != nil {
		noteSuggestion = *aiResp.NoteSuggestion
	}
	c.JSON(http.StatusOK, model.SendChatMessageResponse{
		UserMessage:    *userMsg,
		Message:        *assistantMsg,
		ReferenceCards: cards,
		NoteSuggestion: noteSuggestion,
		SessionState:   sessionState,
		PromotedTopic:  promotedTopic,
	})
}

func chatQuotaExceededMessage(limit int) string {
	if limit < 1 {
		limit = defaultChatDailyLimit
	}
	return fmt.Sprintf("今日对话已达上限（%d轮），明天再来聊吧。", limit)
}

func (h *ChatV4Handler) promoteTempSessionIfClear(ctx context.Context, userID string, scopeContext *model.ChatScopeContext, scope model.ChatMessageScope) *model.ChatTopic {
	if h.gateway == nil || scopeContext == nil || scope.Kind != model.ChatScopeTempSession || scopeContext.TempSession == nil {
		return nil
	}
	messages, err := h.store.RecentChatMessages(ctx, userID, scope, 12)
	if err != nil {
		log.Printf("v4 load temp messages for topic classification failed user=%s temp=%s: %v", userID, scope.ID, err)
		return nil
	}
	if countChatMessagesByRole(messages, "user") == 0 {
		return nil
	}
	recentTopics, err := h.store.RecentChatTopics(ctx, userID, 5)
	if err != nil {
		log.Printf("v4 load recent topics for topic classification failed user=%s temp=%s: %v", userID, scope.ID, err)
		recentTopics = []model.ChatTopic{}
	}
	classification, err := h.gateway.ClassifyChatTopic(ctx, model.ChatTopicClassificationRequest{
		UserID:              userID,
		TempSessionID:       scope.ID,
		Messages:            messages,
		RecentTopics:        recentTopics,
		UserClickedNewTopic: scopeContext.TempSession.ForcedNewTopic,
		DomainTaxonomy:      chatDomainTaxonomy(),
	})
	if err != nil || classification == nil {
		log.Printf("v4 chat topic classification skipped user=%s temp=%s: %v", userID, scope.ID, err)
		return nil
	}
	if !classification.ShouldCreateTopic || classification.ClarityScore < 0.65 {
		return nil
	}

	req := promoteRequestFromClassification(*classification, messages)
	topic, err := h.store.PromoteTempSession(ctx, userID, scope.ID, req)
	if err != nil {
		log.Printf("v4 promote temp session failed user=%s temp=%s: %v", userID, scope.ID, err)
		return nil
	}
	return topic
}

func promoteRequestFromClassification(classification model.ChatTopicClassificationResponse, messages []model.ChatMessage) model.PromoteChatTempSessionRequest {
	domain := strings.TrimSpace(classification.Domain)
	subDomain := strings.TrimSpace(classification.SubDomain)
	if !model.IsValidDomain(model.Domain(domain)) {
		domain = ""
		subDomain = ""
	}
	if subDomain != "" {
		if !model.IsValidSubDomain(model.SubDomain(subDomain)) ||
			(domain != "" && !model.SubDomainBelongsToParent(model.Domain(domain), model.SubDomain(subDomain))) {
			subDomain = ""
		}
	}
	title := truncateRunes(strings.TrimSpace(classification.Title), 100)
	if title == "" {
		title = fallbackChatTopicTitle(messages)
	}
	return model.PromoteChatTempSessionRequest{
		Title:                title,
		Domain:               domain,
		SubDomain:            subDomain,
		Topic:                truncateRunes(strings.TrimSpace(classification.TopicKeyword), 200),
		ClarityScore:         clampFloat(classification.ClarityScore, 0, 1),
		ClassificationReason: strings.TrimSpace(classification.Reason),
	}
}

func fallbackChatTopicTitle(messages []model.ChatMessage) string {
	for _, message := range messages {
		if message.Role != "user" {
			continue
		}
		content := strings.TrimSpace(message.Content)
		if content == "" {
			continue
		}
		return truncateRunes(content, 18)
	}
	return "新的心事"
}

func countChatMessagesByRole(messages []model.ChatMessage, role string) int {
	count := 0
	for _, message := range messages {
		if message.Role == role {
			count++
		}
	}
	return count
}

func chatDomainTaxonomy() map[string][]string {
	taxonomy := make(map[string][]string, len(model.SubDomainsByParent))
	for domain, subDomains := range model.SubDomainsByParent {
		values := make([]string, 0, len(subDomains))
		for _, subDomain := range subDomains {
			values = append(values, string(subDomain))
		}
		taxonomy[string(domain)] = values
	}
	return taxonomy
}

func truncateRunes(value string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}

func clampFloat(value float64, min float64, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func classifyChatMessage(content string) model.ChatPreClassification {
	normalized := strings.ToLower(content)
	pre := model.ChatPreClassification{
		EmotionLevel: "low",
		UserIntent:   "unknown",
		RiskLevel:    "normal",
		RiskReasons:  []string{},
	}
	if containsAny(normalized, []string{"烦", "崩溃", "受不了", "难过", "痛苦", "委屈", "累到", "不想解释"}) {
		pre.EmotionLevel = "high"
		pre.UserIntent = "vent"
		pre.ShouldAvoidCitation = true
	}
	if containsAny(normalized, []string{"怎么办", "怎么做", "该不该", "要不要", "怎么选"}) {
		pre.UserIntent = "ask_advice"
	}
	if containsAny(normalized, []string{"辞职", "裸辞", "分手", "离婚", "借钱", "投资", "手术", "起诉", "报警"}) {
		pre.RiskLevel = "high_decision"
		pre.RiskReasons = append(pre.RiskReasons, "high_impact_decision")
	}
	if containsAny(normalized, []string{"自杀", "轻生", "杀了", "伤害自己", "不想活"}) {
		pre.RiskLevel = "safety_sensitive"
		pre.RiskReasons = append(pre.RiskReasons, "safety_sensitive")
		pre.ShouldAvoidCitation = true
	}
	if containsAny(normalized, []string{"诊断", "药", "律师", "合同", "股票", "基金", "保险"}) && pre.RiskLevel == "normal" {
		pre.RiskLevel = "professional_sensitive"
		pre.RiskReasons = append(pre.RiskReasons, "professional_sensitive")
	}
	if pre.UserIntent == "unknown" && containsAny(normalized, []string{"我发现", "我意识到", "突然明白", "原来"}) {
		pre.UserIntent = "reflect"
	}
	return pre
}

func containsAny(text string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(text, needle) {
			return true
		}
	}
	return false
}

func riskLevelForStorage(risk string) string {
	if risk == "high_decision" || risk == "professional_sensitive" || risk == "safety_sensitive" {
		return "high"
	}
	return "normal"
}

func chatContextFlags(pre model.ChatPreClassification) []string {
	flags := []string{}
	if pre.EmotionLevel == "high" {
		flags = append(flags, "strong_emotion")
	}
	if pre.RiskLevel == "high_decision" {
		flags = append(flags, "high_risk_decision")
	}
	if pre.RiskLevel == "professional_sensitive" || pre.RiskLevel == "safety_sensitive" {
		flags = append(flags, pre.RiskLevel)
	}
	if pre.UserIntent == "ask_advice" {
		flags = append(flags, "method_question")
	}
	if pre.ShouldAvoidCitation {
		flags = append(flags, "avoid_citation")
	}
	return flags
}

func maxChatCitationCards(pre model.ChatPreClassification) int {
	if pre.ShouldAvoidCitation || pre.EmotionLevel == "high" {
		return 0
	}
	return 1
}

func buildReferenceCards(candidates []model.ChatCandidateExperience, citations []model.ChatCitationDecision, maxCards int) []model.ChatReferenceCard {
	if maxCards <= 0 || len(candidates) == 0 || len(citations) == 0 {
		return []model.ChatReferenceCard{}
	}
	candidateByID := make(map[string]model.ChatCandidateExperience, len(candidates))
	for _, candidate := range candidates {
		candidateByID[candidate.ExperienceID] = candidate
	}
	cards := make([]model.ChatReferenceCard, 0, maxCards)
	seen := make(map[string]struct{})
	for _, citation := range citations {
		if !citation.ShowCard {
			continue
		}
		if _, ok := seen[citation.ExperienceID]; ok {
			continue
		}
		candidate, ok := candidateByID[citation.ExperienceID]
		if !ok {
			continue
		}
		cards = append(cards, model.ChatReferenceCard{
			ExperienceID:     candidate.ExperienceID,
			Content:          candidate.Content,
			IsCollected:      candidate.IsCollected || candidate.SourceRelation == "collected",
			CitationType:     citationTypeForCandidate(candidate),
			CitationSentence: citation.CitationSentence,
			ReasonCode:       citation.ReasonCode,
		})
		seen[citation.ExperienceID] = struct{}{}
		if len(cards) >= maxCards {
			break
		}
	}
	if cards == nil {
		return []model.ChatReferenceCard{}
	}
	return cards
}

func citationTypeForCandidate(candidate model.ChatCandidateExperience) string {
	switch candidate.SourceRelation {
	case "own":
		return "own"
	case "collected":
		return "favorite"
	case "public_original":
		return "public_original"
	default:
		return "public_featured"
	}
}
