package repository

import (
	"os"
	"strings"
	"testing"
)

func TestRecordExperienceEventDedupesAuthenticatedExposeWithinWindow(t *testing.T) {
	sourceBytes, err := os.ReadFile("experience_actions_v4.go")
	if err != nil {
		t.Fatalf("read experience_actions_v4.go: %v", err)
	}
	source := string(sourceBytes)

	required := []string{
		`event.EventType == "expose"`,
		"recordDedupedExposeEvent",
		"UPDATE experience_events",
		"NOW() - INTERVAL '30 minutes'",
		"ev.user_id = $1::uuid",
		"NULLIF($5::text, '')::uuid",
		"INSERT INTO experience_events",
	}
	for _, fragment := range required {
		if !strings.Contains(source, fragment) {
			t.Fatalf("authenticated expose events should dedupe with fragment %q", fragment)
		}
	}
	if strings.Contains(source, "ev.user_id = NULLIF($1, '')::uuid") {
		t.Fatalf("authenticated expose dedupe must not use optional-user NULLIF on $1 after $1 is inferred as uuid")
	}
}
