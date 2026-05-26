package repository

import (
	"os"
	"strings"
	"testing"
)

func TestMeStatsQueriesUseV4VisibilityFacts(t *testing.T) {
	sourceBytes, err := os.ReadFile("me_stats_v4.go")
	if err != nil {
		t.Fatalf("read me_stats_v4.go: %v", err)
	}
	source := string(sourceBytes)

	for _, required := range []string{
		"e.visibility='public'",
		"e.visibility='private'",
		"e.lifecycle_status <> 'deleted'",
		"e.lifecycle_status='active'",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("me stats queries should use V4 predicate %q", required)
		}
	}

	for _, legacy := range []string{
		"e.is_private",
		"CASE WHEN e.is_private",
		"COALESCE(e.lifecycle_status, 'active') <> 'deleted'",
		"COALESCE(e.lifecycle_status, 'active')='active'",
	} {
		if strings.Contains(source, legacy) {
			t.Fatalf("me stats queries should not fall back to legacy/V4 default field %q", legacy)
		}
	}
}
