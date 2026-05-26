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

func TestExperienceActionGatesUseV4VisibilityLifecycleFacts(t *testing.T) {
	sourceBytes, err := os.ReadFile("experience_actions_v4.go")
	if err != nil {
		t.Fatalf("read experience_actions_v4.go: %v", err)
	}
	source := string(sourceBytes)

	for _, required := range []struct {
		fragment string
		minCount int
	}{
		{"e.visibility = 'public'", 4},
		{"e.lifecycle_status = 'active'", 4},
		{"e.lifecycle_status <> 'deleted'", 4},
	} {
		if count := strings.Count(source, required.fragment); count < required.minCount {
			t.Fatalf("experience action gates should include %q at least %d times, got %d", required.fragment, required.minCount, count)
		}
	}

	for _, legacy := range []string{
		"COALESCE(e.visibility, 'public') = 'public'",
		"COALESCE(e.lifecycle_status, 'active') = 'active'",
		"COALESCE(e.lifecycle_status, 'active') <> 'deleted'",
	} {
		if strings.Contains(source, legacy) {
			t.Fatalf("experience action gates should not use fallback gate fragment %q", legacy)
		}
	}
}
