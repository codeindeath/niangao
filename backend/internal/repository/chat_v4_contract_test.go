package repository

import (
	"os"
	"strings"
	"testing"
)

func TestChatMessagesHydratesReferenceCardsFromCitations(t *testing.T) {
	sourceBytes, err := os.ReadFile("chat_v4.go")
	if err != nil {
		t.Fatalf("read chat_v4.go: %v", err)
	}
	source := string(sourceBytes)

	for _, want := range []string{
		"attachChatReferenceCards",
		"chat_citations",
		"ReferenceCards",
		"unavailable_reason",
		"experience_unavailable",
		"visible_to_viewer",
		"COALESCE(e.owner_user_id, e.author_id) = $1::uuid",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("chat history messages should hydrate reference cards with %q", want)
		}
	}
}
