package repository

import (
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
