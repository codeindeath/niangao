package repository

import (
	"context"
	"fmt"

	"github.com/niangao/backend/internal/model"
)

const searchExperiencesQuery = `
  SELECT
    e.id,
    COALESCE(e.owner_user_id, e.author_id),
    e.content,
    e.experience_type,
    e.visibility,
    e.lifecycle_status,
    COALESCE(e.domain::text, ''),
    COALESCE(e.sub_domain, ''),
    e.topic,
    COALESCE(e.creator_display_name, e.creator_name, u.display_name, u.nickname, ''),
    e.interpretation_status,
    (e.interpretation IS NOT NULL AND e.interpretation <> ''),
    e.quality_tier,
    CASE e.quality_tier
      WHEN 'high_trust' THEN 5
      WHEN 'ai_citable' THEN 4
      WHEN 'recommend_candidate' THEN 3
      WHEN 'public_visible' THEN 2
      ELSE 1
    END AS star_rating,
    COALESCE(e.inspiration_count, e.like_count, 0),
    COALESCE(e.collection_count, e.bookmark_count, 0),
    EXISTS(
      SELECT 1 FROM experience_collections ec
      WHERE ec.user_id = NULLIF($1, '')::uuid
        AND ec.experience_id = e.id
        AND ec.status = 'active'
    ),
    EXISTS(
      SELECT 1 FROM experience_inspirations ei
      WHERE ei.user_id = NULLIF($1, '')::uuid
        AND ei.experience_id = e.id
    ),
    '' AS unavailable_reason
  FROM experiences e
  LEFT JOIN users u ON u.id = e.author_id
  WHERE e.deleted_at IS NULL
    AND (
      (
        e.visibility = 'public'
        AND e.lifecycle_status = 'active'
        AND e.quality_tier IN ('public_visible', 'recommend_candidate', 'ai_citable', 'high_trust')
      )
      OR (
        COALESCE(e.owner_user_id, e.author_id) = NULLIF($1, '')::uuid
        AND e.lifecycle_status <> 'deleted'
      )
    )
    AND (
      e.content ILIKE $2
      OR COALESCE(e.creator_display_name, e.creator_name, u.display_name, u.nickname, '') ILIKE $2
      OR COALESCE(e.domain::text, '') ILIKE $2
      OR COALESCE(e.sub_domain, '') ILIKE $2
      OR e.topic ILIKE $2
    )
  ORDER BY
    CASE
      WHEN COALESCE(e.creator_display_name, e.creator_name, u.display_name, u.nickname, '') ILIKE $2 THEN 0
      WHEN e.topic ILIKE $2 THEN 1
      WHEN e.content ILIKE $2 THEN 2
      ELSE 3
    END,
    CASE e.quality_tier
      WHEN 'high_trust' THEN 5
      WHEN 'ai_citable' THEN 4
      WHEN 'recommend_candidate' THEN 3
      WHEN 'public_visible' THEN 2
      ELSE 1
    END DESC,
    e.created_at DESC,
    e.id DESC
  LIMIT $3 OFFSET $4`

func (r *ExperienceRepo) SearchExperiences(ctx context.Context, userID string, query string, limit int, cursor string) (*model.FeedPage, error) {
	offset := parseOffsetCursor(cursor)
	pattern := "%" + query + "%"

	rows, err := r.db.Query(ctx, searchExperiencesQuery,
		userID, pattern, limit+1, offset)
	if err != nil {
		return nil, fmt.Errorf("v4 search experiences: %w", err)
	}
	defer rows.Close()

	return scanFeedPage(rows, limit, offset, "")
}
