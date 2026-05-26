package repository

import (
	"os"
	"strings"
	"testing"
)

func TestV4FeedQueriesExposeOwnerUserID(t *testing.T) {
	queries := map[string]string{
		"recommend":   recommendFeedQuery,
		"collections": collectionsFeedQuery,
		"mine":        mineFeedQuery,
		"search":      searchExperiencesQuery,
	}

	for name, query := range queries {
		t.Run(name, func(t *testing.T) {
			if !strings.Contains(query, "COALESCE(e.owner_user_id, e.author_id)") {
				t.Fatalf("%s feed query should expose owner_user_id for App owner actions", name)
			}
		})
	}
}

func TestV4RecommendQueryUsesLightPersonalizationAndSeenControl(t *testing.T) {
	required := []string{
		"viewer_domain_signals",
		"experience_collections",
		"experience_inspirations",
		"experience_events",
		"viewer_seen",
		"user_views",
		"COALESCE(vds.score, 0) DESC",
		"vs.experience_id IS NULL",
	}

	for _, fragment := range required {
		if !strings.Contains(recommendFeedQuery, fragment) {
			t.Fatalf("recommend feed query should include %q for V4 recommendation semantics", fragment)
		}
	}
}

func TestV4RecommendQueryKeepsPublicRecommendFilterIndexable(t *testing.T) {
	required := []string{
		"e.visibility = 'public'",
		"e.lifecycle_status = 'active'",
		"e.recommendation_status = 'eligible'",
		"e.quality_tier IN ('recommend_candidate', 'ai_citable', 'high_trust')",
		"e.owner_user_id = NULLIF($1, '')::uuid",
	}

	for _, fragment := range required {
		if !strings.Contains(recommendFeedQuery, fragment) {
			t.Fatalf("recommend feed query should include indexable fragment %q", fragment)
		}
	}

	forbidden := []string{
		"WHERE COALESCE(e.visibility, 'public') = 'public'",
		"AND COALESCE(e.lifecycle_status, 'active') = 'active'",
		"AND COALESCE(e.recommendation_status, 'ineligible') = 'eligible'",
		"WHERE COALESCE(e.owner_user_id, e.author_id) = NULLIF($1, '')::uuid",
	}

	for _, fragment := range forbidden {
		if strings.Contains(recommendFeedQuery, fragment) {
			t.Fatalf("recommend feed query should not use non-indexable fragment %q", fragment)
		}
	}
}

func TestV4SearchQueryUsesV4VisibilityLifecycleGate(t *testing.T) {
	required := []string{
		"e.visibility = 'public'",
		"e.lifecycle_status = 'active'",
		"e.quality_tier IN ('public_visible', 'recommend_candidate', 'ai_citable', 'high_trust')",
		"COALESCE(e.owner_user_id, e.author_id) = NULLIF($1, '')::uuid",
		"e.lifecycle_status <> 'deleted'",
	}

	for _, fragment := range required {
		if !strings.Contains(searchExperiencesQuery, fragment) {
			t.Fatalf("search query should include V4 visibility/lifecycle gate fragment %q", fragment)
		}
	}

	forbidden := []string{
		"COALESCE(e.visibility, 'public') = 'public'",
		"COALESCE(e.lifecycle_status, 'active') = 'active'",
		"COALESCE(e.lifecycle_status, 'active') <> 'deleted'",
	}
	for _, fragment := range forbidden {
		if strings.Contains(searchExperiencesQuery, fragment) {
			t.Fatalf("search query should not use fallback V4 gate fragment %q", fragment)
		}
	}
}

func TestV4CollectionsAndMineQueriesUseV4VisibilityLifecycleGates(t *testing.T) {
	for _, required := range []struct {
		name     string
		query    string
		fragment string
	}{
		{"collections", collectionsFeedQuery, "e.visibility = 'public'"},
		{"collections", collectionsFeedQuery, "e.lifecycle_status = 'active'"},
		{"collections", collectionsFeedQuery, "e.lifecycle_status <> 'deleted'"},
		{"mine", mineFeedQuery, "e.lifecycle_status <> 'deleted'"},
	} {
		t.Run(required.name+"/"+required.fragment, func(t *testing.T) {
			if !strings.Contains(required.query, required.fragment) {
				t.Fatalf("%s query should include V4 gate fragment %q", required.name, required.fragment)
			}
		})
	}

	for _, forbidden := range []struct {
		name     string
		query    string
		fragment string
	}{
		{"collections", collectionsFeedQuery, "COALESCE(e.visibility, 'public') = 'public'"},
		{"collections", collectionsFeedQuery, "COALESCE(e.lifecycle_status, 'active') = 'active'"},
		{"collections", collectionsFeedQuery, "COALESCE(e.lifecycle_status, 'active') <> 'deleted'"},
		{"mine", mineFeedQuery, "COALESCE(e.lifecycle_status, 'active') <> 'deleted'"},
	} {
		t.Run(forbidden.name+"/"+forbidden.fragment, func(t *testing.T) {
			if strings.Contains(forbidden.query, forbidden.fragment) {
				t.Fatalf("%s query should not use fallback V4 gate fragment %q", forbidden.name, forbidden.fragment)
			}
		})
	}
}

func TestRecommendationCursorRoundTrip(t *testing.T) {
	cursor := formatRecommendationCursor("11111111-1111-4111-8111-111111111111", 20)
	sessionID, offset, ok := parseRecommendationCursor(cursor)
	if !ok {
		t.Fatalf("parseRecommendationCursor(%q) ok=false, want true", cursor)
	}
	if sessionID != "11111111-1111-4111-8111-111111111111" || offset != 20 {
		t.Fatalf("parsed cursor = %q %d, want session and offset", sessionID, offset)
	}

	for _, invalid := range []string{"", "20", "rec:not-a-uuid:20", "rec:11111111-1111-4111-8111-111111111111:-1"} {
		if _, _, ok := parseRecommendationCursor(invalid); ok {
			t.Fatalf("parseRecommendationCursor(%q) ok=true, want false", invalid)
		}
	}
}

func TestRecommendFeedPersistsRecommendationSessions(t *testing.T) {
	sourceBytes, err := os.ReadFile("experience_v4.go")
	if err != nil {
		t.Fatalf("read experience_v4.go: %v", err)
	}
	source := string(sourceBytes)
	required := []string{
		"recommendation_sessions",
		"candidate_ids",
		"expires_at > NOW()",
		"formatRecommendationCursor",
		"parseRecommendationCursor",
		"recommendSessionCardsQuery",
	}
	for _, fragment := range required {
		if !strings.Contains(source, fragment) {
			t.Fatalf("recommend feed should persist and replay session cursor with fragment %q", fragment)
		}
	}
}

func TestV4FeedQueriesExposeUnavailableReasonForScanner(t *testing.T) {
	queries := map[string]string{
		"recommend":   recommendFeedQuery,
		"collections": collectionsFeedQuery,
		"mine":        mineFeedQuery,
		"search":      searchExperiencesQuery,
	}

	for name, query := range queries {
		t.Run(name, func(t *testing.T) {
			if !strings.Contains(query, "unavailable_reason") {
				t.Fatalf("%s feed query should expose unavailable_reason for the shared scanner", name)
			}
		})
	}

	source, err := os.ReadFile("experience_v4.go")
	if err != nil {
		t.Fatalf("read experience_v4.go: %v", err)
	}
	if !strings.Contains(string(source), "&card.UnavailableReason") {
		t.Fatal("scanFeedPage should scan unavailable_reason into ExperienceCard")
	}
}

func TestCollectionsFeedReturnsUnavailablePlaceholdersForInvisibleCollections(t *testing.T) {
	required := []string{
		"experience_unavailable",
		"AS unavailable_reason",
		"CASE WHEN",
		"visible_to_viewer",
		"TRUE AS is_collected",
		"COALESCE(e.owner_user_id, e.author_id) = $1::uuid",
	}

	for _, fragment := range required {
		if !strings.Contains(collectionsFeedQuery, fragment) {
			t.Fatalf("collections feed should retain invisible saved rows with placeholder fragment %q", fragment)
		}
	}

	forbidden := []string{
		"AND e.deleted_at IS NULL\n    AND (",
		"e.content,",
	}
	for _, fragment := range forbidden {
		if strings.Contains(collectionsFeedQuery, fragment) {
			t.Fatalf("collections feed should not filter out or expose original invisible content via fragment %q", fragment)
		}
	}
}
