package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niangao/backend/internal/model"
)

type fakeV4ChatStore struct {
	gotUserID        string
	gotID            string
	fail             bool
	addedMessages    []model.SaveChatMessageRequest
	savedCitations   []model.ChatReferenceCard
	candidates       []model.ChatCandidateExperience
	recentMessages   []model.ChatMessage
	assistantMessage *model.ChatMessage
}

func (f *fakeV4ChatStore) RecentChatTopics(ctx context.Context, userID string, limit int) ([]model.ChatTopic, error) {
	f.gotUserID = userID
	if f.fail {
		return nil, errors.New("store failed")
	}
	return []model.ChatTopic{{ID: "topic-1", Title: "工作里的不甘心", UpdatedAt: time.Now()}}, nil
}

func (f *fakeV4ChatStore) ChatTopics(ctx context.Context, userID string, limit int, cursor string) (*model.ChatTopicPage, error) {
	f.gotUserID = userID
	if f.fail {
		return nil, errors.New("store failed")
	}
	return &model.ChatTopicPage{Data: []model.ChatTopic{{ID: "topic-1", Title: "工作里的不甘心"}}, HasMore: false}, nil
}

func (f *fakeV4ChatStore) CreateTempSession(ctx context.Context, userID string, forcedNewTopic bool) (*model.ChatTempSession, error) {
	f.gotUserID = userID
	if f.fail {
		return nil, errors.New("store failed")
	}
	return &model.ChatTempSession{ID: "temp-1", Status: "active", ForcedNewTopic: forcedNewTopic}, nil
}

func (f *fakeV4ChatStore) CreateChatTopic(ctx context.Context, userID string, req model.CreateChatTopicRequest) (*model.ChatTopic, error) {
	f.gotUserID = userID
	if f.fail {
		return nil, errors.New("store failed")
	}
	return &model.ChatTopic{ID: "topic-1", Title: req.Title, Status: "active"}, nil
}

func (f *fakeV4ChatStore) UpdateChatTopic(ctx context.Context, userID string, topicID string, req model.UpdateChatTopicRequest) (*model.ChatTopic, error) {
	f.gotUserID = userID
	f.gotID = topicID
	if f.fail {
		return nil, errors.New("store failed")
	}
	return &model.ChatTopic{ID: topicID, Title: valueOr(req.Title, "工作里的不甘心"), Status: "active"}, nil
}

func (f *fakeV4ChatStore) DeleteChatTopic(ctx context.Context, userID string, topicID string) error {
	f.gotUserID = userID
	f.gotID = topicID
	if f.fail {
		return errors.New("store failed")
	}
	return nil
}

func (f *fakeV4ChatStore) ChatMessages(ctx context.Context, userID string, scope model.ChatMessageScope, limit int, cursor string) (*model.ChatMessagePage, error) {
	f.gotUserID = userID
	f.gotID = scope.ID
	if f.fail {
		return nil, errors.New("store failed")
	}
	return &model.ChatMessagePage{Data: f.recentMessages, HasMore: false}, nil
}

func (f *fakeV4ChatStore) VerifyChatScope(ctx context.Context, userID string, scope model.ChatMessageScope) (*model.ChatScopeContext, error) {
	f.gotUserID = userID
	f.gotID = scope.ID
	if f.fail {
		return nil, errors.New("store failed")
	}
	return &model.ChatScopeContext{
		Scope:        scope,
		SessionState: "stable_topic",
		Topic: &model.ChatTopic{
			ID:        scope.ID,
			Title:     "工作里的不甘心",
			Domain:    string(model.DomainWork),
			SubDomain: "communication",
			Topic:     "不想继续硬撑",
			UpdatedAt: time.Now(),
		},
	}, nil
}

func (f *fakeV4ChatStore) AddChatMessage(ctx context.Context, userID string, req model.SaveChatMessageRequest) (*model.ChatMessage, error) {
	f.gotUserID = userID
	if f.fail {
		return nil, errors.New("store failed")
	}
	f.addedMessages = append(f.addedMessages, req)
	message := &model.ChatMessage{
		ID:                      "msg-user",
		UserID:                  userID,
		Role:                    req.Role,
		Content:                 req.Content,
		Status:                  "sent",
		RiskLevel:               req.RiskLevel,
		ReferencedExperienceIDs: req.ReferencedExperienceIDs,
		CreatedAt:               time.Now(),
	}
	if req.Role == "assistant" {
		message.ID = "msg-ai"
		f.assistantMessage = message
	}
	return message, nil
}

func (f *fakeV4ChatStore) RecentChatMessages(ctx context.Context, userID string, scope model.ChatMessageScope, limit int) ([]model.ChatMessage, error) {
	f.gotUserID = userID
	if f.fail {
		return nil, errors.New("store failed")
	}
	return f.recentMessages, nil
}

func (f *fakeV4ChatStore) CandidateExperiencesForChat(ctx context.Context, userID string, scope model.ChatScopeContext, userMessage string, riskLevel string, limit int) ([]model.ChatCandidateExperience, error) {
	f.gotUserID = userID
	if f.fail {
		return nil, errors.New("store failed")
	}
	if f.candidates != nil {
		return f.candidates, nil
	}
	return []model.ChatCandidateExperience{
		{
			ExperienceID:   "exp-1",
			Content:        "先把触发你想离开的具体点写下来，再决定是不是离开。",
			CreatorName:    "某个认真生活的人",
			SourceRelation: "collected",
			Visibility:     "public",
			QualityTier:    string(model.QualityTierAICitable),
		},
	}, nil
}

func (f *fakeV4ChatStore) SaveChatCitations(ctx context.Context, assistantMessageID string, cards []model.ChatReferenceCard) error {
	if f.fail {
		return errors.New("store failed")
	}
	f.savedCitations = append([]model.ChatReferenceCard{}, cards...)
	return nil
}

type fakeChatGateway struct {
	fail        bool
	gotRequest  *model.ChatGatewayRequest
	citations   []model.ChatCitationDecision
	replyText   string
	noteSuggest *model.ChatNoteSuggestion
}

func (f *fakeChatGateway) GenerateChatReply(ctx context.Context, req model.ChatGatewayRequest) (*model.ChatGatewayResponse, error) {
	f.gotRequest = &req
	if f.fail {
		return nil, errors.New("gateway failed")
	}
	reply := f.replyText
	if reply == "" {
		reply = "先别急着把它变成一个必须马上解决的问题。你可以先把最刺痛你的那一点拎出来看。"
	}
	return &model.ChatGatewayResponse{
		ReplyText:      reply,
		Citations:      f.citations,
		NoteSuggestion: f.noteSuggest,
		EmotionLevel:   "medium",
		RiskLevel:      "normal",
	}, nil
}

func TestV4ChatTopicRoutesRequireAuth(t *testing.T) {
	tests := []struct {
		method string
		path   string
		body   string
	}{
		{"GET", "/api/v1/chat/recent-topics", ""},
		{"GET", "/api/v1/chat/topics", ""},
		{"POST", "/api/v1/chat/temp-sessions", `{}`},
		{"POST", "/api/v1/chat/topics", `{"title":"工作里的不甘心"}`},
		{"PATCH", "/api/v1/chat/topics/topic-1", `{"title":"新标题"}`},
		{"DELETE", "/api/v1/chat/topics/topic-1", ""},
		{"GET", "/api/v1/chat/topics/topic-1/messages", ""},
		{"POST", "/api/v1/chat/topics/topic-1/messages", `{"content":"我想辞职"}`},
		{"POST", "/api/v1/chat/temp-sessions/temp-1/messages", `{"content":"我想辞职"}`},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			r := gin.New()
			RegisterChatV4Routes(r.Group("/api/v1"), &fakeV4ChatStore{})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestV4ChatTopicRoutesReturnExpectedShapes(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		key    string
	}{
		{"recent", "GET", "/api/v1/chat/recent-topics", "", "data"},
		{"topics", "GET", "/api/v1/chat/topics", "", "data"},
		{"temp", "POST", "/api/v1/chat/temp-sessions", `{"forced_new_topic":true}`, "id"},
		{"create topic", "POST", "/api/v1/chat/topics", `{"title":"工作里的不甘心"}`, "id"},
		{"update topic", "PATCH", "/api/v1/chat/topics/topic-1", `{"title":"新标题"}`, "id"},
		{"delete topic", "DELETE", "/api/v1/chat/topics/topic-1", "", "status"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			store := &fakeV4ChatStore{}
			v1 := r.Group("/api/v1", func(c *gin.Context) {
				c.Set("user_id", "user-1")
			})
			RegisterChatV4Routes(v1, store)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK && w.Code != http.StatusCreated {
				t.Fatalf("status = %d, want 200/201: %s", w.Code, w.Body.String())
			}
			if store.gotUserID != "user-1" {
				t.Fatalf("userID = %q, want user-1", store.gotUserID)
			}
			var body map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if _, ok := body[tt.key]; !ok {
				t.Fatalf("response missing %q: %+v", tt.key, body)
			}
		})
	}
}

func TestV4ChatTopicRouteFailureReturnsServerError(t *testing.T) {
	r := gin.New()
	v1 := r.Group("/api/v1", func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	RegisterChatV4Routes(v1, &fakeV4ChatStore{fail: true})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/chat/recent-topics", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500: %s", w.Code, w.Body.String())
	}
}

func TestV4ChatSendTopicMessageSavesUserBeforeCallingGateway(t *testing.T) {
	r := gin.New()
	store := &fakeV4ChatStore{}
	gateway := &fakeChatGateway{citations: []model.ChatCitationDecision{{
		ExperienceID: "exp-1",
		UsageType:    "card",
		ShowCard:     true,
		ReasonCode:   "high_relevance",
		Strength:     "strong",
	}}}
	v1 := r.Group("/api/v1", func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	RegisterChatV4Routes(v1, store, gateway)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/chat/topics/topic-1/messages", strings.NewReader(`{"content":"我最近很想辞职，但又怕后悔","client_message_id":"c-1"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
	}
	if len(store.addedMessages) < 2 {
		t.Fatalf("added messages = %d, want user and assistant", len(store.addedMessages))
	}
	if store.addedMessages[0].Role != "user" || store.addedMessages[0].Content == "" {
		t.Fatalf("first saved message = %+v, want user message", store.addedMessages[0])
	}
	if gateway.gotRequest == nil {
		t.Fatal("gateway was not called")
	}
	if gateway.gotRequest.UserMessageID != "msg-user" {
		t.Fatalf("gateway user message id = %q, want msg-user", gateway.gotRequest.UserMessageID)
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, ok := body["message"].(map[string]any); !ok {
		t.Fatalf("response missing assistant message: %+v", body)
	}
	cards, ok := body["reference_cards"].([]any)
	if !ok || len(cards) != 1 {
		t.Fatalf("reference_cards = %+v, want one card", body["reference_cards"])
	}
}

func TestV4ChatSendTempMessageUsesTempScope(t *testing.T) {
	r := gin.New()
	store := &fakeV4ChatStore{}
	v1 := r.Group("/api/v1", func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	RegisterChatV4Routes(v1, store, &fakeChatGateway{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/chat/temp-sessions/temp-1/messages", strings.NewReader(`{"content":"我还没想清楚聊什么"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
	}
	if len(store.addedMessages) == 0 || store.addedMessages[0].Scope.Kind != model.ChatScopeTempSession {
		t.Fatalf("first saved message scope = %+v, want temp session", store.addedMessages)
	}
}

func TestV4ChatSendKeepsUserMessageWhenGatewayFails(t *testing.T) {
	r := gin.New()
	store := &fakeV4ChatStore{}
	v1 := r.Group("/api/v1", func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	RegisterChatV4Routes(v1, store, &fakeChatGateway{fail: true})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/chat/topics/topic-1/messages", strings.NewReader(`{"content":"我想辞职"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503: %s", w.Code, w.Body.String())
	}
	if len(store.addedMessages) != 1 || store.addedMessages[0].Role != "user" {
		t.Fatalf("added messages = %+v, want only saved user message", store.addedMessages)
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if retryable, _ := body["retryable"].(bool); !retryable {
		t.Fatalf("retryable = %+v, want true", body["retryable"])
	}
}

func TestV4ChatSendDropsOutOfScopeCitations(t *testing.T) {
	r := gin.New()
	store := &fakeV4ChatStore{}
	v1 := r.Group("/api/v1", func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	RegisterChatV4Routes(v1, store, &fakeChatGateway{citations: []model.ChatCitationDecision{{
		ExperienceID: "exp-not-candidate",
		UsageType:    "card",
		ShowCard:     true,
		ReasonCode:   "high_relevance",
	}}})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/chat/topics/topic-1/messages", strings.NewReader(`{"content":"我想辞职"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
	}
	if len(store.savedCitations) != 0 {
		t.Fatalf("saved citations = %+v, want none for out-of-scope citation", store.savedCitations)
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if cards, _ := body["reference_cards"].([]any); len(cards) != 0 {
		t.Fatalf("reference_cards = %+v, want empty", cards)
	}
}

func TestV4ChatSendRejectsEmptyContent(t *testing.T) {
	r := gin.New()
	v1 := r.Group("/api/v1", func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	RegisterChatV4Routes(v1, &fakeV4ChatStore{}, &fakeChatGateway{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/chat/topics/topic-1/messages", strings.NewReader(`{"content":"   "}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", w.Code, w.Body.String())
	}
}
