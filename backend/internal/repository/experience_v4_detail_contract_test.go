package repository

import (
	"os"
	"strings"
	"testing"
)

func TestExperienceSelectColsExposeV4DetailOwnershipFields(t *testing.T) {
	wants := []string{
		"COALESCE(e.owner_user_id, e.author_id)",
		"e.creator_display_name",
		"COALESCE(e.experience_type",
		"COALESCE(e.visibility",
		"COALESCE(e.lifecycle_status",
		"COALESCE(e.inspiration_count",
		"COALESCE(e.collection_count",
	}

	for _, want := range wants {
		if !strings.Contains(experienceSelectCols, want) {
			t.Fatalf("experienceSelectCols should expose %q for V4 detail responses", want)
		}
	}
}

func TestSoftDeleteSynchronizesV4LifecycleFacts(t *testing.T) {
	source, err := os.ReadFile("experience.go")
	if err != nil {
		t.Fatalf("read experience.go: %v", err)
	}

	text := string(source)
	wants := []string{
		"lifecycle_status='deleted'",
		"recommendation_status='suppressed'",
		"ai_citable=FALSE",
		"updated_at=NOW()",
	}

	for _, want := range wants {
		if !strings.Contains(text, want) {
			t.Fatalf("soft delete should synchronize V4 lifecycle facts with %q", want)
		}
	}
}

func TestUpdateExperienceQueryPreservesSourceSceneAndSynchronizesLifecycle(t *testing.T) {
	if strings.Contains(updateExperienceQuery, "source_scene") {
		t.Fatal("experience updates should preserve original source_scene instead of rewriting chat notes into note-sourced experiences")
	}
	if !strings.Contains(updateExperienceQuery, "lifecycle_status=$11") {
		t.Fatal("experience updates should write V4 lifecycle_status explicitly")
	}
}

func TestUpdateLifecycleStatusForRequest(t *testing.T) {
	if got := updateLifecycleStatusForRequest(false); got != "needs_review" {
		t.Fatalf("public edits should become needs_review, got %q", got)
	}
	if got := updateLifecycleStatusForRequest(true); got != "active" {
		t.Fatalf("private edits should remain active private records, got %q", got)
	}
}
