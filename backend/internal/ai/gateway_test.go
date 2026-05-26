package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/niangao/backend/internal/model"
)

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
