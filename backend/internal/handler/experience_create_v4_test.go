package handler

import (
	"strings"
	"testing"

	"github.com/niangao/backend/internal/model"
)

func TestNormalizeCreateExperienceRequestV4DefaultsPublicNote(t *testing.T) {
	req := model.CreateExperienceRequest{Content: "先写下来"}

	if err := normalizeCreateExperienceRequest(&req); err != nil {
		t.Fatalf("normalize failed: %v", err)
	}
	if req.Visibility != model.VisibilityPublic {
		t.Fatalf("visibility = %q, want public", req.Visibility)
	}
	if req.SourceScene != string(model.SourceSceneNote) {
		t.Fatalf("source_scene = %q, want note", req.SourceScene)
	}
	if req.IsPrivate {
		t.Fatal("default note save should be public")
	}
}

func TestNormalizeCreateExperienceRequestAcceptsOptionalDomainSubdomainAndTopic(t *testing.T) {
	req := model.CreateExperienceRequest{Content: "今天想明白了", Topic: "#自我"}

	if err := normalizeCreateExperienceRequest(&req); err != nil {
		t.Fatalf("normalize failed: %v", err)
	}
	if req.Domain != "" || req.SubDomain != "" {
		t.Fatalf("domain/subdomain should remain optional: %+v", req)
	}
	if req.Topics != "#自我" {
		t.Fatalf("legacy topics mirror = %q, want #自我", req.Topics)
	}
}

func TestNormalizeCreateExperienceRequestLegacyPrivateMapsToVisibility(t *testing.T) {
	req := model.CreateExperienceRequest{Content: "这个先只给自己看", IsPrivate: true}

	if err := normalizeCreateExperienceRequest(&req); err != nil {
		t.Fatalf("normalize failed: %v", err)
	}
	if req.Visibility != model.VisibilityPrivate || !req.IsPrivate {
		t.Fatalf("private mapping failed: visibility=%q is_private=%v", req.Visibility, req.IsPrivate)
	}
}

func TestNormalizeCreateExperienceRequestChatSourceMessageIDs(t *testing.T) {
	firstID := "11111111-1111-1111-1111-111111111111"
	secondID := "22222222-2222-2222-2222-222222222222"
	req := model.CreateExperienceRequest{
		Content:          "先把这点记下来",
		SourceScene:      string(model.SourceSceneChat),
		Visibility:       model.VisibilityPrivate,
		SourceMessageIDs: []string{firstID, "", secondID},
	}

	if err := normalizeCreateExperienceRequest(&req); err != nil {
		t.Fatalf("normalize failed: %v", err)
	}
	if req.SourceChatMessageID != secondID {
		t.Fatalf("source_chat_message_id = %q, want latest source message %q", req.SourceChatMessageID, secondID)
	}
	if !strings.Contains(req.SourceChatMessageSnapshot, firstID) ||
		!strings.Contains(req.SourceChatMessageSnapshot, secondID) {
		t.Fatalf("source_chat_message_snapshot should preserve source ids, got %q", req.SourceChatMessageSnapshot)
	}
	if len(req.SourceMessageIDs) != 2 {
		t.Fatalf("empty source ids should be removed, got %+v", req.SourceMessageIDs)
	}
}

func TestNormalizeCreateExperienceRequestRejectsInvalidSubdomain(t *testing.T) {
	req := model.CreateExperienceRequest{
		Content:   "领域和子领域要匹配",
		Domain:    model.DomainMeaning,
		SubDomain: model.SubWorkComm,
	}

	if err := normalizeCreateExperienceRequest(&req); err == nil {
		t.Fatal("expected invalid sub_domain error")
	}
}

func TestNormalizeCreateExperienceRequestContentAndTopicLimits(t *testing.T) {
	tests := []struct {
		name string
		req  model.CreateExperienceRequest
	}{
		{name: "empty content", req: model.CreateExperienceRequest{Content: ""}},
		{name: "over 100 runes", req: model.CreateExperienceRequest{Content: strings.Repeat("年", 101)}},
		{name: "topic over 200 runes", req: model.CreateExperienceRequest{Content: "有效内容", Topic: strings.Repeat("题", 201)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := normalizeCreateExperienceRequest(&tt.req); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
