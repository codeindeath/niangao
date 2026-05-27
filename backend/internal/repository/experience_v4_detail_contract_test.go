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
		"e.experience_type",
		"e.visibility",
		"e.lifecycle_status",
		"e.source_scene",
		"e.topic",
		"e.quality_tier",
		"e.recommendation_status",
		"e.ai_citable",
		"e.interpretation_status",
		"COALESCE(e.inspiration_count",
		"COALESCE(e.collection_count",
	}

	for _, want := range wants {
		if !strings.Contains(experienceSelectCols, want) {
			t.Fatalf("experienceSelectCols should expose %q for V4 detail responses", want)
		}
	}

	for _, forbidden := range []string{
		"COALESCE(e.experience_type",
		"COALESCE(e.visibility",
		"COALESCE(e.lifecycle_status",
		"COALESCE(e.source_scene",
		"COALESCE(e.topic, e.topics",
		"COALESCE(e.quality_tier",
		"COALESCE(e.recommendation_status",
		"COALESCE(e.ai_citable",
		"COALESCE(e.interpretation_status",
	} {
		if strings.Contains(experienceSelectCols, forbidden) {
			t.Fatalf("experienceSelectCols should expose canonical V4 detail fields without fallback fragment %q", forbidden)
		}
	}
}

func TestExperienceDetailUsesV4InteractionTables(t *testing.T) {
	for _, want := range []string{
		"experience_collections",
		"experience_inspirations",
		"ec.status = 'active'",
	} {
		if !strings.Contains(experienceLikedBookmark, want) {
			t.Fatalf("detail interaction state should use V4 table/query fragment %q", want)
		}
	}

	for _, legacy := range []string{
		" FROM likes ",
		" FROM bookmarks ",
	} {
		if strings.Contains(experienceLikedBookmark, legacy) {
			t.Fatalf("detail interaction state should not use legacy interaction table fragment %q", legacy)
		}
	}
}

func TestExperienceDetailUsesV4VisibilityLifecycleGate(t *testing.T) {
	source, err := os.ReadFile("experience.go")
	if err != nil {
		t.Fatalf("read experience.go: %v", err)
	}
	text := string(source)
	start := strings.Index(text, "func (r *ExperienceRepo) GetByID(")
	if start < 0 {
		t.Fatal("ExperienceRepo.GetByID not found")
	}
	end := strings.Index(text[start:], "func (r *ExperienceRepo) GetByIDForAdmin(")
	if end < 0 {
		t.Fatal("ExperienceRepo.GetByID end marker not found")
	}
	body := text[start : start+end]

	for _, want := range []string{
		"e.visibility = 'public'",
		"e.lifecycle_status = 'active'",
		"e.lifecycle_status <> 'deleted'",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("GetByID public visibility gate should use V4 condition %q", want)
		}
	}
	for _, legacy := range []string{
		"e.status = 'published' AND e.review_status = 'approved' AND e.is_private = FALSE",
		"COALESCE(e.visibility, CASE WHEN e.is_private THEN 'private' ELSE 'public' END) = 'public'",
		"COALESCE(e.lifecycle_status, CASE WHEN e.deleted_at IS NOT NULL THEN 'deleted' WHEN e.review_status = 'pending' THEN 'needs_review' ELSE 'active' END) = 'active'",
	} {
		if strings.Contains(body, legacy) {
			t.Fatalf("GetByID public visibility gate should not use legacy condition %q", legacy)
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

func TestCreateWithReviewSynchronizesLifecycleForPublicReview(t *testing.T) {
	source, err := os.ReadFile("experience.go")
	if err != nil {
		t.Fatalf("read experience.go: %v", err)
	}
	text := string(source)
	start := strings.Index(text, "func (r *ExperienceRepo) CreateWithReview(")
	if start < 0 {
		t.Fatal("ExperienceRepo.CreateWithReview not found")
	}
	end := strings.Index(text[start:], "const experienceSelectCols")
	if end < 0 {
		t.Fatal("ExperienceRepo.CreateWithReview end marker not found")
	}
	body := text[start : start+end]

	if !strings.Contains(body, "LifecycleStatus:           updateLifecycleStatusForRequest(req.IsPrivate)") {
		t.Fatal("CreateWithReview should derive lifecycle_status from public/private visibility")
	}
	if strings.Contains(body, "LifecycleStatus:           string(model.LifecycleActive)") {
		t.Fatal("CreateWithReview should not mark public pending-review creates as active")
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
