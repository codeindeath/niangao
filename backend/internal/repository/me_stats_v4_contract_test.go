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
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("me stats queries should use V4 visibility predicate %q", required)
		}
	}

	for _, legacy := range []string{
		"e.is_private",
		"CASE WHEN e.is_private",
	} {
		if strings.Contains(source, legacy) {
			t.Fatalf("me stats queries should not fall back to legacy visibility field %q", legacy)
		}
	}
}
