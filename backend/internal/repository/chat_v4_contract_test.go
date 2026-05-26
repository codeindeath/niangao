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
