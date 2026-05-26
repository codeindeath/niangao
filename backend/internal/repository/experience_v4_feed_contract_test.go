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
