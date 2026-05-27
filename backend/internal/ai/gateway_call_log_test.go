package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/niangao/backend/internal/model"
)

type fakeCallLogger struct {
	entries []CallLogEntry
}

func (f *fakeCallLogger) RecordAICall(_ context.Context, entry CallLogEntry) error {
	f.entries = append(f.entries, entry)
	return nil
}

func TestGatewayRecordsSuccessfulAICall(t *testing.T) {
	logger := &fakeCallLogger{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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

	gateway := NewGatewayWithTimeoutAndLogger(server.URL, 2*time.Second, logger)
	if _, err := gateway.RewriteExperience(context.Background(), model.ExperienceRewriteGatewayRequest{
		UserID:            "11111111-1111-4111-8111-111111111111",
		Source:            "manual_note",
		RawText:           "焦虑的时候先把下一步写小",
		DefaultVisibility: model.VisibilityPublic,
	}); err != nil {
		t.Fatalf("RewriteExperience returned error: %v", err)
	}

	if len(logger.entries) != 1 {
		t.Fatalf("logged entries = %d, want 1", len(logger.entries))
	}
	entry := logger.entries[0]
	if entry.FunctionType != "experience_rewrite" || entry.CallSource != "app_rewrite" {
		t.Fatalf("entry function/source = %q/%q", entry.FunctionType, entry.CallSource)
	}
	if entry.UserID != "11111111-1111-4111-8111-111111111111" {
		t.Fatalf("entry user = %q", entry.UserID)
	}
	if entry.Status != "success" || entry.ErrorCode != "" {
		t.Fatalf("entry status/error = %q/%q", entry.Status, entry.ErrorCode)
	}
	if entry.LatencyMS < 0 || entry.FinishedAt.Before(entry.StartedAt) {
		t.Fatalf("entry timing = %+v", entry)
	}
}

func TestGatewayRecordsFailedAICall(t *testing.T) {
	logger := &fakeCallLogger{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "provider unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	gateway := NewGatewayWithTimeoutAndLogger(server.URL, 2*time.Second, logger)
	_, err := gateway.GenerateChatReply(context.Background(), model.ChatGatewayRequest{
		UserID:        "11111111-1111-4111-8111-111111111111",
		UserMessageID: "22222222-2222-4222-8222-222222222222",
		UserMessage:   "今天有点乱",
		RecentMessages: []model.ChatMessage{{
			Role:    "user",
			Content: "今天有点乱",
		}},
	})
	if err == nil {
		t.Fatal("GenerateChatReply returned nil error")
	}

	if len(logger.entries) != 1 {
		t.Fatalf("logged entries = %d, want 1", len(logger.entries))
	}
	entry := logger.entries[0]
	if entry.FunctionType != "chat" || entry.CallSource != "app_chat" {
		t.Fatalf("entry function/source = %q/%q", entry.FunctionType, entry.CallSource)
	}
	if entry.ChatMessageID != "22222222-2222-4222-8222-222222222222" {
		t.Fatalf("entry chat_message_id = %q", entry.ChatMessageID)
	}
	if entry.Status != "failed" || entry.ErrorCode != "http_503" {
		t.Fatalf("entry status/error = %q/%q", entry.Status, entry.ErrorCode)
	}
}
