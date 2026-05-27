package repository

import (
	"os"
	"strings"
	"testing"
)

func TestDeprecatedConversationRepositoryMethodsAreRemoved(t *testing.T) {
	sourceBytes, err := os.ReadFile("conversation.go")
	if err != nil {
		t.Fatalf("read conversation.go: %v", err)
	}
	source := string(sourceBytes)

	for _, forbidden := range []string{
		"func (r *ConversationRepo) Create(",
		"func (r *ConversationRepo) AddMessage(",
		"func (r *ConversationRepo) GetMessages(",
		"func (r *ConversationRepo) GetByID(",
		"func (r *ConversationRepo) ListByUser(",
		"func (r *ConversationRepo) GetOrCreateByUser(",
		"func (r *ConversationRepo) GetMessagesSince(",
		"func (r *ConversationRepo) CountTodayMessages(",
		"FROM messages",
		"FROM conversations",
		"INSERT INTO messages",
		"INSERT INTO conversations",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("deprecated conversation repository code should be removed: %s", forbidden)
		}
	}
}

func TestDeprecatedConversationModelsAreRemoved(t *testing.T) {
	sourceBytes, err := os.ReadFile("../model/models.go")
	if err != nil {
		t.Fatalf("read models.go: %v", err)
	}
	source := string(sourceBytes)

	for _, forbidden := range []string{
		"type Conversation struct",
		"type Message struct",
		"type ChatRequest struct",
		"`json:\"conversation_id",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("deprecated conversation model should be removed: %s", forbidden)
		}
	}
}
