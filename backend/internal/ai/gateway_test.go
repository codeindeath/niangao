package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/niangao/backend/internal/middleware"
	"github.com/niangao/backend/internal/model"
)

func TestNewGatewayWithTimeoutUsesConfiguredTimeout(t *testing.T) {
	gateway := NewGatewayWithTimeout("http://ai.local/", 90*time.Second)

	if gateway.baseURL != "http://ai.local" {
		t.Fatalf("baseURL = %q, want trimmed base URL", gateway.baseURL)
	}
	if gateway.httpClient.Timeout != 90*time.Second {
		t.Fatalf("gateway timeout = %s, want 90s", gateway.httpClient.Timeout)
	}
}

func TestGatewayPropagatesRequestIDHeader(t *testing.T) {
	var gotRequestID string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequestID = r.Header.Get(middleware.RequestIDHeader)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"result": {
				"can_rewrite": true,
				"content": "先把下一步写小，再看自己是不是还焦虑。",
				"domain": "meaning",
				"sub_domain": "self",
				"topic": "焦虑时的行动",
				"rewrite_level": "light",
				"source_preservation": "high",
				"needs_user_edit": false,
				"reason": "保留原意并压缩表达"
			},
			"confidence": 0.8,
			"warnings": []
		}`))
	}))
	defer server.Close()

	gateway := NewGateway(server.URL)
	ctx := middleware.ContextWithRequestID(context.Background(), "client-request-1")
	if _, err := gateway.RewriteExperience(ctx, model.ExperienceRewriteGatewayRequest{
		UserID:            "user-1",
		Source:            "manual_note",
		RawText:           "焦虑的时候先把下一步写小",
		DefaultVisibility: model.VisibilityPublic,
	}); err != nil {
		t.Fatalf("RewriteExperience returned error: %v", err)
	}

	if gotRequestID != "client-request-1" {
		t.Fatalf("X-Request-ID = %q, want client-request-1", gotRequestID)
	}
}

func TestClassifyChatTopicAcceptsNullCandidateTopicID(t *testing.T) {
	var gotFunctionType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		gotFunctionType, _ = body["function_type"].(string)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"result": {
				"clarity_score": 0.78,
				"should_create_topic": true,
				"title": "工作里的不甘心",
				"domain": "work",
				"sub_domain": "work-comm",
				"topic_keyword": "和上级沟通",
				"candidate_existing_topic_id": null,
				"should_bind_existing_topic": false,
				"discard_if_user_leaves": false,
				"reason": "用户已经围绕同一件工作冲突表达了明确困扰"
			},
			"confidence": 0.84,
			"warnings": []
		}`))
	}))
	defer server.Close()

	gateway := NewGateway(server.URL)
	result, err := gateway.ClassifyChatTopic(context.Background(), model.ChatTopicClassificationRequest{
		UserID:        "user-1",
		TempSessionID: "temp-1",
		Messages: []model.ChatMessage{{
			Role:    "user",
			Content: "我觉得在会上被上级当众否定这事过不去",
		}},
	})
	if err != nil {
		t.Fatalf("ClassifyChatTopic returned error: %v", err)
	}
	if gotFunctionType != "chat_topic_classify" {
		t.Fatalf("function_type = %q, want chat_topic_classify", gotFunctionType)
	}
	if result.CandidateExistingTopicID != "" {
		t.Fatalf("candidate_existing_topic_id = %q, want empty string for null", result.CandidateExistingTopicID)
	}
	if !result.ShouldCreateTopic || result.Title != "工作里的不甘心" {
		t.Fatalf("classification = %+v, want create-topic result", result)
	}
}
