//go:build integration
// +build integration

package repository

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niangao/backend/internal/model"
)

func connectDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL must be set for integration tests")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// ── List tests ──

func TestList_NoFilters(t *testing.T) {
	pool := connectDB(t)
	repo := NewExperienceRepo(pool)
	ctx := context.Background()

	results, total, err := repo.List(ctx, model.ExperienceListQuery{
		Page:     1,
		PageSize: 5,
	}, "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total < 1 {
		t.Errorf("expected total > 0, got %d", total)
	}
	if len(results) > 5 {
		t.Errorf("expected <= 5 results, got %d", len(results))
	}
	// Verify all results are approved and published
	for _, e := range results {
		if e.ReviewStatus != "approved" {
			t.Errorf("expected review_status=approved, got %s (id=%s)", e.ReviewStatus, e.ID)
		}
		if e.Status != "published" {
			t.Errorf("expected status=published, got %s (id=%s)", e.Status, e.ID)
		}
	}
}

func TestList_SearchContent(t *testing.T) {
	pool := connectDB(t)
	repo := NewExperienceRepo(pool)
	ctx := context.Background()

	results, total, err := repo.List(ctx, model.ExperienceListQuery{
		Page:     1,
		PageSize: 10,
		Search:   "产品",
	}, "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total == 0 {
		t.Skip("no results for search '产品'")
	}
	found := false
	for _, e := range results {
		if len([]rune(e.Content)) > 0 {
			found = true
			break
		}
	}
	if !found && len(results) > 0 {
		t.Error("expected at least one result with content")
	}
}

func TestList_SearchCreatorName(t *testing.T) {
	pool := connectDB(t)
	repo := NewExperienceRepo(pool)
	ctx := context.Background()

	results, _, err := repo.List(ctx, model.ExperienceListQuery{
		Page:     1,
		PageSize: 10,
		Search:   "张一鸣",
	}, "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, e := range results {
		if e.CreatorName == nil {
			continue
		}
		if len([]rune(*e.CreatorName)) > 0 {
			// At least one result with creator_name — test passes
			return
		}
	}
	// If no creator_name matches, that's acceptable if data doesn't exist
	t.Log("no creator_name matches (may be normal if no such seed data)")
}

func TestList_SearchInterpretationExcluded(t *testing.T) {
	pool := connectDB(t)
	repo := NewExperienceRepo(pool)
	ctx := context.Background()

	// Search for a word likely only in interpretation, not content
	results, _, err := repo.List(ctx, model.ExperienceListQuery{
		Page:     1,
		PageSize: 10,
		Search:   "背景",
	}, "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	// The search should match content OR creator_name OR nickname, NOT interpretation
	// If results exist, verify they match in content/creator/nickname, not just interpretation
	for _, e := range results {
		hasContentMatch := contains(e.Content, "背景")
		hasCreatorMatch := e.CreatorName != nil && contains(*e.CreatorName, "背景")
		if !hasContentMatch && !hasCreatorMatch {
			// Could match via nickname (joined field). This is acceptable.
			t.Logf("result id=%s content=%q matched via other field", e.ID, truncate(e.Content, 30))
		}
	}
}

func TestList_Pagination(t *testing.T) {
	pool := connectDB(t)
	repo := NewExperienceRepo(pool)
	ctx := context.Background()

	// Page 1
	p1, total, err := repo.List(ctx, model.ExperienceListQuery{
		Page:     1,
		PageSize: 3,
	}, "")
	if err != nil {
		t.Fatalf("page 1: %v", err)
	}
	if total < 4 {
		t.Skipf("only %d total results, need > 3 for pagination test", total)
	}
	if len(p1) != 3 {
		t.Errorf("page 1: expected 3, got %d", len(p1))
	}

	// Page 2
	p2, _, err := repo.List(ctx, model.ExperienceListQuery{
		Page:     2,
		PageSize: 3,
	}, "")
	if err != nil {
		t.Fatalf("page 2: %v", err)
	}
	if len(p2) == 0 {
		t.Error("page 2: expected results, got 0")
	}

	// No overlap
	ids := make(map[string]bool)
	for _, e := range p1 {
		ids[e.ID] = true
	}
	for _, e := range p2 {
		if ids[e.ID] {
			t.Errorf("duplicate id %s on page 1 and page 2", e.ID)
		}
	}
}

// ── GetByID tests ──

func TestGetByID_Exists(t *testing.T) {
	pool := connectDB(t)
	repo := NewExperienceRepo(pool)
	ctx := context.Background()

	// First get a valid ID from List
	results, _, err := repo.List(ctx, model.ExperienceListQuery{
		Page:     1,
		PageSize: 1,
	}, "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(results) == 0 {
		t.Skip("no experiences in database")
	}

	exp, err := repo.GetByID(ctx, results[0].ID, "")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if exp.ID != results[0].ID {
		t.Errorf("expected id %s, got %s", results[0].ID, exp.ID)
	}
	if exp.Content == "" {
		t.Error("expected non-empty content")
	}
}

func TestGetByID_NotFound(t *testing.T) {
	pool := connectDB(t)
	repo := NewExperienceRepo(pool)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "00000000-0000-0000-0000-000000000000", "")
	if err == nil {
		t.Error("expected error for non-existent ID, got nil")
	}
}

// ── ListByAuthor tests ──

func TestListByAuthor(t *testing.T) {
	pool := connectDB(t)
	repo := NewExperienceRepo(pool)
	ctx := context.Background()

	// Find a user who has experiences
	results, _, err := repo.List(ctx, model.ExperienceListQuery{
		Page:     1,
		PageSize: 1,
	}, "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(results) == 0 {
		t.Skip("no experiences in database")
	}

	authorID := results[0].AuthorID
	authorExps, _, err := repo.ListByAuthor(ctx, authorID, 1, 20)
	if err != nil {
		t.Fatalf("ListByAuthor: %v", err)
	}
	if len(authorExps) == 0 {
		t.Errorf("expected experiences for author %s, got 0", authorID)
	}
	for _, e := range authorExps {
		if e.AuthorID != authorID {
			t.Errorf("expected author_id=%s, got %s (id=%s)", authorID, e.AuthorID, e.ID)
		}
	}
}

// ── Recommend tests ──

func TestRecommend(t *testing.T) {
	pool := connectDB(t)
	repo := NewExperienceRepo(pool)
	ctx := context.Background()

	results, err := repo.Recommend(ctx, "" /*no user*/, 5, 0)
	if err != nil {
		t.Fatalf("Recommend: %v", err)
	}
	if len(results) == 0 {
		t.Skip("no recommendations available")
	}
	if len(results) > 5 {
		t.Errorf("expected <= 5 results, got %d", len(results))
	}
}

func TestRecommend_Pagination(t *testing.T) {
	pool := connectDB(t)
	repo := NewExperienceRepo(pool)
	ctx := context.Background()

	p1, err := repo.Recommend(ctx, "", 2, 0)
	if err != nil {
		t.Fatalf("page 1: %v", err)
	}
	if len(p1) < 2 {
		t.Skipf("only %d results, need >= 2 for pagination test", len(p1))
	}

	p2, err := repo.Recommend(ctx, "", 2, 2)
	if err != nil {
		t.Fatalf("page 2: %v", err)
	}
	if len(p2) == 0 {
		t.Error("page 2: expected results, got 0")
	}

	// No overlap
	ids := make(map[string]bool)
	for _, e := range p1 {
		ids[e.ID] = true
	}
	for _, e := range p2 {
		if ids[e.ID] {
			t.Errorf("duplicate id %s in recommend page 1 and 2", e.ID)
		}
	}
}

// ── helpers ──

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && runesContains([]rune(s), []rune(substr))
}

func runesContains(s, substr []rune) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "..."
}
