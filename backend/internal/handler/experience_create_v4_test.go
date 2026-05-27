package handler

import (
	"os"
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

func TestNormalizeCreateExperienceRequestUsesUserFacingCopyForInvalidV4Fields(t *testing.T) {
	tests := []struct {
		name string
		req  model.CreateExperienceRequest
		want string
	}{
		{name: "invalid visibility", req: model.CreateExperienceRequest{Content: "有效内容", Visibility: "friends"}, want: "可见性设置不支持"},
		{name: "invalid source scene", req: model.CreateExperienceRequest{Content: "有效内容", SourceScene: "legacy"}, want: "来源不支持"},
		{name: "invalid source chat topic id", req: model.CreateExperienceRequest{Content: "有效内容", SourceChatTopicID: "not-a-uuid"}, want: "聊天来源不正确"},
		{name: "invalid source chat message id", req: model.CreateExperienceRequest{Content: "有效内容", SourceChatMessageID: "not-a-uuid"}, want: "聊天来源不正确"},
		{name: "invalid domain", req: model.CreateExperienceRequest{Content: "有效内容", Domain: "legacy"}, want: "领域设置不支持"},
		{name: "invalid subdomain", req: model.CreateExperienceRequest{Content: "有效内容", SubDomain: "legacy"}, want: "子领域设置不支持"},
		{name: "subdomain without domain", req: model.CreateExperienceRequest{Content: "有效内容", SubDomain: model.SubSelf}, want: "领域和子领域不匹配"},
		{name: "subdomain parent mismatch", req: model.CreateExperienceRequest{Content: "有效内容", Domain: model.DomainMeaning, SubDomain: model.SubWorkComm}, want: "领域和子领域不匹配"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := normalizeCreateExperienceRequest(&tt.req)
			if err == nil {
				t.Fatal("expected validation error")
			}
			if err.Error() != tt.want {
				t.Fatalf("error = %q, want %q", err.Error(), tt.want)
			}
		})
	}
}

func TestExperienceHandlerAppFacingErrorsDoNotUseEnglishFallbackCopy(t *testing.T) {
	source, err := os.ReadFile("experience.go")
	if err != nil {
		t.Fatalf("read experience.go: %v", err)
	}
	text := string(source)

	for _, forbidden := range []string{
		"invalid rewrite payload",
		"content is required",
		"content is too long",
		"invalid source",
		"invalid default_visibility",
		"invalid user_selected_domain",
		"invalid user_selected_sub_domain",
		"sub_domain does not belong to domain",
		"experience not found",
		"failed to check display name",
		"failed to save experience",
		"failed to update",
		"failed to delete",
		"invalid visibility",
		"invalid source_scene",
		"invalid source_chat_topic_id",
		"invalid source_chat_message_id",
		"invalid domain",
		"invalid sub_domain",
		"sub_domain requires domain",
	} {
		if strings.Contains(text, `"`+forbidden+`"`) {
			t.Fatalf("experience.go still exposes English App-facing copy %q", forbidden)
		}
	}
}
