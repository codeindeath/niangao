package handler

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/niangao/backend/internal/model"
)

func TestExperienceDetailResponseUsesV4AppContract(t *testing.T) {
	ownerID := "user-1"
	subDomain := "self"
	creatorName := "旧创建者"
	authorName := "旧作者"
	exp := &model.Experience{
		ID:               "exp-1",
		AuthorID:         "legacy-author",
		OwnerUserID:      &ownerID,
		Content:          "先把复杂事写成一句能行动的话",
		Domain:           model.DomainMeaning,
		SubDomain:        &subDomain,
		Topics:           "#旧话题",
		Topic:            "#自我",
		IsPrivate:        true,
		IsOfficial:       true,
		SourceType:       "platform",
		ReviewStatus:     "approved",
		LikeCount:        99,
		BookmarkCount:    88,
		InspirationCount: 7,
		CollectionCount:  6,
		IsLiked:          true,
		IsBookmarked:     false,
		ExperienceType:   string(model.ExperienceTypeUserOriginal),
		Visibility:       string(model.VisibilityPrivate),
		LifecycleStatus:  string(model.LifecycleActive),
		QualityTier:      string(model.QualityTierPrivateOnly),
		CreatorName:      &creatorName,
		AuthorName:       &authorName,
		CreatedAt:        time.Date(2026, 5, 27, 9, 0, 0, 0, time.UTC),
	}

	body, err := json.Marshal(toExperienceDetailResponse(exp))
	if err != nil {
		t.Fatalf("marshal detail response: %v", err)
	}
	text := string(body)

	for _, want := range []string{
		`"owner_user_id":"user-1"`,
		`"topic":"#自我"`,
		`"inspiration_count":7`,
		`"collection_count":6`,
		`"is_inspired":true`,
		`"is_collected":false`,
		`"creator_display_name":"旧创建者"`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("detail response should include V4 field %s in %s", want, text)
		}
	}

	for _, legacy := range []string{
		"author_id",
		"topics",
		"is_private",
		"is_official",
		"source_type",
		"review_status",
		"like_count",
		"bookmark_count",
		"is_liked",
		"is_bookmarked",
		"creator_name",
		"author_name",
	} {
		if strings.Contains(text, legacy) {
			t.Fatalf("detail response should not expose legacy field %q in %s", legacy, text)
		}
	}
}
