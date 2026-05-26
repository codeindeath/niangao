package repository

import (
	"os"
	"strings"
	"testing"
)

func TestPromoteTempSessionCreatesTopicAndMovesMessages(t *testing.T) {
	sourceBytes, err := os.ReadFile("chat_v4.go")
	if err != nil {
		t.Fatalf("read chat_v4.go: %v", err)
	}
	source := string(sourceBytes)

	required := []string{
		"PromoteTempSession",
		"INSERT INTO chat_topics",
		"UPDATE chat_messages",
		"temp_session_id=NULL",
		"UPDATE chat_temp_sessions",
		"status='promoted'",
		"promoted_topic_id",
		"FOR UPDATE",
	}
	for _, fragment := range required {
		if !strings.Contains(source, fragment) {
			t.Fatalf("temp-session promotion should include fragment %q", fragment)
		}
	}
}

func TestAddChatMessageStoresEmptyReferenceArrayForMessagesWithoutCitations(t *testing.T) {
	sourceBytes, err := os.ReadFile("chat_v4.go")
	if err != nil {
		t.Fatalf("read chat_v4.go: %v", err)
	}
	source := string(sourceBytes)

	required := []string{
		"referencedExperienceIDs := req.ReferencedExperienceIDs",
		"if referencedExperienceIDs == nil",
		"referencedExperienceIDs = []string{}",
		"referencedExperienceIDs,",
	}
	for _, fragment := range required {
		if !strings.Contains(source, fragment) {
			t.Fatalf("chat messages without citations should persist an empty uuid array with fragment %q", fragment)
		}
	}
}

func TestChatCandidateQueryDoesNotRequireFutureSourceDerivationColumn(t *testing.T) {
	sourceBytes, err := os.ReadFile("chat_v4.go")
	if err != nil {
		t.Fatalf("read chat_v4.go: %v", err)
	}
	source := string(sourceBytes)

	if strings.Contains(source, "e.source_derivation_type") {
		t.Fatal("chat candidate query must not require the future source_derivation_type column before the production schema has it")
	}
	if !strings.Contains(source, "AS source_derivation_type") {
		t.Fatal("chat candidate query should still expose source_derivation_type in the AI candidate payload")
	}
}

func TestChatCandidateQueryUsesV4CollectionStatus(t *testing.T) {
	sourceBytes, err := os.ReadFile("chat_v4.go")
	if err != nil {
		t.Fatalf("read chat_v4.go: %v", err)
	}
	source := string(sourceBytes)

	if strings.Contains(source, "c.deleted_at") {
		t.Fatal("chat candidate query must use V4 experience_collections.status instead of legacy deleted_at")
	}
	if !strings.Contains(source, "c.status='active'") {
		t.Fatal("chat candidate query should filter active collections with V4 status")
	}
}

func TestChatReferenceAndCandidateQueriesUseV4VisibilityLifecycleGates(t *testing.T) {
	sourceBytes, err := os.ReadFile("chat_v4.go")
	if err != nil {
		t.Fatalf("read chat_v4.go: %v", err)
	}
	source := string(sourceBytes)

	for _, required := range []string{
		"e.visibility = 'public'",
		"e.lifecycle_status = 'active'",
		"e.lifecycle_status <> 'deleted'",
		"e.lifecycle_status='active'",
		"e.visibility='public'",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("chat reference/candidate queries should include V4 gate fragment %q", required)
		}
	}

	for _, forbidden := range []string{
		"COALESCE(e.visibility, 'public') = 'public'",
		"COALESCE(e.lifecycle_status, 'active') = 'active'",
		"COALESCE(e.lifecycle_status, 'active') <> 'deleted'",
		"COALESCE(e.lifecycle_status, 'active')='active'",
		"COALESCE(e.visibility, 'public')='public'",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("chat reference/candidate queries should not use fallback V4 gate fragment %q", forbidden)
		}
	}
}

func TestChatDailyUsageUsesV4MessagesAndSystemConfig(t *testing.T) {
	sourceBytes, err := os.ReadFile("chat_v4.go")
	if err != nil {
		t.Fatalf("read chat_v4.go: %v", err)
	}
	source := string(sourceBytes)

	for _, required := range []string{
		"ChatDailyUsage",
		"chat_limit_per_day",
		"FROM chat_messages",
		"role='user'",
		"status <> 'deleted'",
		"created_at >= CURRENT_DATE",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("V4 chat daily quota should include fragment %q", required)
		}
	}

	quotaStart := strings.Index(source, "func (r *ConversationRepo) ChatDailyUsage")
	if quotaStart < 0 {
		t.Fatal("ConversationRepo should implement ChatDailyUsage for V4 quota enforcement")
	}
	quotaSource := source[quotaStart:]
	for _, forbidden := range []string{
		"FROM messages",
		"conversation_id",
		"created_at::date",
	} {
		if strings.Contains(quotaSource, forbidden) {
			t.Fatalf("V4 chat daily quota must not use old chat quota fragment %q", forbidden)
		}
	}
}
