package repository

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/niangao/backend/internal/model"
)

const recommendFeedQuery = `
  WITH viewer_domain_signals AS (
    SELECT domain, SUM(score) AS score
    FROM (
      SELECT e.domain::text AS domain, 4 AS score
      FROM experience_collections ec
      JOIN experiences e ON e.id = ec.experience_id
      WHERE ec.user_id = NULLIF($1, '')::uuid
        AND ec.status = 'active'
        AND e.deleted_at IS NULL
      UNION ALL
      SELECT e.domain::text AS domain, 3 AS score
      FROM experience_inspirations ei
      JOIN experiences e ON e.id = ei.experience_id
      WHERE ei.user_id = NULLIF($1, '')::uuid
        AND e.deleted_at IS NULL
      UNION ALL
      SELECT e.domain::text AS domain, 2 AS score
      FROM experience_events ev
      JOIN experiences e ON e.id = ev.experience_id
      WHERE ev.user_id = NULLIF($1, '')::uuid
        AND ev.event_type IN ('flip', 'search_click', 'chat_citation_click')
        AND ev.created_at >= NOW() - INTERVAL '90 days'
        AND e.deleted_at IS NULL
      UNION ALL
      SELECT e.domain::text AS domain, 1 AS score
      FROM experiences e
      WHERE e.owner_user_id = NULLIF($1, '')::uuid
        AND e.deleted_at IS NULL
    ) signals
    WHERE domain <> ''
    GROUP BY domain
  ),
  viewer_seen AS (
    SELECT DISTINCT ev.experience_id
    FROM experience_events ev
    WHERE ev.user_id = NULLIF($1, '')::uuid
      AND ev.event_type IN ('expose', 'flip', 'chat_citation_show', 'chat_citation_click', 'search_click')
      AND ev.created_at >= NOW() - INTERVAL '30 days'
      AND ev.experience_id IS NOT NULL
    UNION
    SELECT uv.experience_id
    FROM user_views uv
    WHERE uv.user_id = NULLIF($1, '')::uuid
      AND uv.viewed_at >= NOW() - INTERVAL '30 days'
  )
  SELECT
    e.id,
    COALESCE(e.owner_user_id, e.author_id),
    e.content,
    COALESCE(e.experience_type, 'platform_selected'),
    COALESCE(e.visibility, 'public'),
    COALESCE(e.lifecycle_status, 'active'),
    COALESCE(e.domain::text, ''),
    COALESCE(e.sub_domain, ''),
    COALESCE(e.topic, e.topics, ''),
    COALESCE(e.creator_display_name, e.creator_name, u.display_name, u.nickname, ''),
    COALESCE(e.interpretation_status, CASE WHEN e.interpretation_generated THEN 'ready' ELSE 'none' END),
    (e.interpretation IS NOT NULL AND e.interpretation <> ''),
    COALESCE(e.quality_tier, 'public_visible'),
    CASE COALESCE(e.quality_tier, 'public_visible')
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
    )
  FROM experiences e
  LEFT JOIN users u ON u.id = e.author_id
  LEFT JOIN viewer_domain_signals vds ON vds.domain = COALESCE(e.domain::text, '')
  LEFT JOIN viewer_seen vs ON vs.experience_id = e.id
  WHERE e.visibility = 'public'
    AND e.lifecycle_status = 'active'
    AND e.recommendation_status = 'eligible'
    AND e.quality_tier IN ('recommend_candidate', 'ai_citable', 'high_trust')
    AND e.deleted_at IS NULL
  ORDER BY
    CASE
      WHEN NULLIF($1, '') IS NULL THEN 0
      WHEN vs.experience_id IS NULL THEN 0
      ELSE 1
    END ASC,
    CASE COALESCE(e.quality_tier, 'public_visible')
      WHEN 'high_trust' THEN 5
      WHEN 'ai_citable' THEN 4
      WHEN 'recommend_candidate' THEN 3
      ELSE 0
    END DESC,
    COALESCE(vds.score, 0) DESC,
    (COALESCE(e.inspiration_count, e.like_count, 0) + COALESCE(e.collection_count, e.bookmark_count, 0)) DESC,
    CASE COALESCE(e.experience_type, 'user_original')
      WHEN 'platform_selected' THEN 0
      ELSE 1
    END ASC,
    e.created_at DESC,
    e.id DESC
  LIMIT $2 OFFSET $3`

const collectionsFeedQuery = `
  SELECT
    e.id,
    COALESCE(e.owner_user_id, e.author_id),
    e.content,
    COALESCE(e.experience_type, 'user_original'),
    COALESCE(e.visibility, 'public'),
    COALESCE(e.lifecycle_status, 'active'),
    COALESCE(e.domain::text, ''),
    COALESCE(e.sub_domain, ''),
    COALESCE(e.topic, e.topics, ''),
    COALESCE(e.creator_display_name, e.creator_name, u.display_name, u.nickname, ''),
    COALESCE(e.interpretation_status, CASE WHEN e.interpretation_generated THEN 'ready' ELSE 'none' END),
    (e.interpretation IS NOT NULL AND e.interpretation <> ''),
    COALESCE(e.quality_tier, 'public_visible'),
    CASE COALESCE(e.quality_tier, 'public_visible')
      WHEN 'high_trust' THEN 5
      WHEN 'ai_citable' THEN 4
      WHEN 'recommend_candidate' THEN 3
      WHEN 'public_visible' THEN 2
      ELSE 1
    END AS star_rating,
    COALESCE(e.inspiration_count, e.like_count, 0),
    COALESCE(e.collection_count, e.bookmark_count, 0),
    TRUE,
    EXISTS(
      SELECT 1 FROM experience_inspirations ei
      WHERE ei.user_id = $1::uuid
        AND ei.experience_id = e.id
    )
  FROM experience_collections ec
  JOIN experiences e ON e.id = ec.experience_id
  LEFT JOIN users u ON u.id = e.author_id
  WHERE ec.user_id = $1::uuid
    AND ec.status = 'active'
    AND e.deleted_at IS NULL
    AND (
      (COALESCE(e.visibility, 'public') = 'public' AND COALESCE(e.lifecycle_status, 'active') = 'active')
      OR COALESCE(e.owner_user_id, e.author_id) = $1::uuid
    )
  ORDER BY ec.collected_at DESC, ec.id DESC
  LIMIT $2 OFFSET $3`

const mineFeedQuery = `
  SELECT
    e.id,
    COALESCE(e.owner_user_id, e.author_id),
    e.content,
    COALESCE(e.experience_type, 'user_original'),
    COALESCE(e.visibility, CASE WHEN e.is_private THEN 'private' ELSE 'public' END),
    COALESCE(e.lifecycle_status, 'active'),
    COALESCE(e.domain::text, ''),
    COALESCE(e.sub_domain, ''),
    COALESCE(e.topic, e.topics, ''),
    COALESCE(e.creator_display_name, e.creator_name, u.display_name, u.nickname, ''),
    COALESCE(e.interpretation_status, CASE WHEN e.interpretation_generated THEN 'ready' ELSE 'none' END),
    (e.interpretation IS NOT NULL AND e.interpretation <> ''),
    COALESCE(e.quality_tier, CASE WHEN e.is_private THEN 'private_only' ELSE 'unreviewed' END),
    CASE COALESCE(e.quality_tier, 'unreviewed')
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
      WHERE ec.user_id = $1::uuid
        AND ec.experience_id = e.id
        AND ec.status = 'active'
    ),
    EXISTS(
      SELECT 1 FROM experience_inspirations ei
      WHERE ei.user_id = $1::uuid
        AND ei.experience_id = e.id
    )
  FROM experiences e
  LEFT JOIN users u ON u.id = e.author_id
  WHERE COALESCE(e.owner_user_id, e.author_id) = $1::uuid
    AND COALESCE(e.lifecycle_status, 'active') <> 'deleted'
    AND e.deleted_at IS NULL
  ORDER BY e.created_at DESC, e.id DESC
  LIMIT $2 OFFSET $3`

func (r *ExperienceRepo) RecommendFeed(ctx context.Context, userID string, limit int, cursor string) (*model.FeedPage, error) {
	offset := parseOffsetCursor(cursor)
	rows, err := r.db.Query(ctx, recommendFeedQuery,
		userID, limit+1, offset)
	if err != nil {
		return nil, fmt.Errorf("v4 recommend feed: %w", err)
	}
	defer rows.Close()

	return scanFeedPage(rows, limit, offset, "")
}

func (r *ExperienceRepo) CollectionsFeed(ctx context.Context, userID string, limit int, cursor string) (*model.FeedPage, error) {
	offset := parseOffsetCursor(cursor)
	rows, err := r.db.Query(ctx, collectionsFeedQuery,
		userID, limit+1, offset)
	if err != nil {
		return nil, fmt.Errorf("v4 collections feed: %w", err)
	}
	defer rows.Close()

	return scanFeedPage(rows, limit, offset, "")
}

func (r *ExperienceRepo) MineFeed(ctx context.Context, userID string, limit int, cursor string) (*model.FeedPage, error) {
	offset := parseOffsetCursor(cursor)
	rows, err := r.db.Query(ctx, mineFeedQuery,
		userID, limit+1, offset)
	if err != nil {
		return nil, fmt.Errorf("v4 mine feed: %w", err)
	}
	defer rows.Close()

	return scanFeedPage(rows, limit, offset, "")
}

func scanFeedPage(rows pgx.Rows, limit int, offset int, sessionID string) (*model.FeedPage, error) {
	cards := make([]model.ExperienceCard, 0, limit)
	for rows.Next() {
		var card model.ExperienceCard
		if err := rows.Scan(
			&card.ID,
			&card.OwnerUserID,
			&card.Content,
			&card.ExperienceType,
			&card.Visibility,
			&card.LifecycleStatus,
			&card.Domain,
			&card.SubDomain,
			&card.Topic,
			&card.CreatorDisplayName,
			&card.InterpretationStatus,
			&card.InterpretationSummaryAvailable,
			&card.QualityTier,
			&card.StarRating,
			&card.InspirationCount,
			&card.CollectionCount,
			&card.IsCollected,
			&card.IsInspired,
		); err != nil {
			return nil, fmt.Errorf("scan v4 feed card: %w", err)
		}
		cards = append(cards, card)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate v4 feed cards: %w", err)
	}

	hasMore := len(cards) > limit
	if hasMore {
		cards = cards[:limit]
	}

	nextCursor := ""
	if hasMore {
		nextCursor = strconv.Itoa(offset + limit)
	}

	return &model.FeedPage{
		Data:       cards,
		NextCursor: nextCursor,
		SessionID:  sessionID,
		HasMore:    hasMore,
	}, nil
}

func parseOffsetCursor(cursor string) int {
	if cursor == "" {
		return 0
	}
	offset, err := strconv.Atoi(cursor)
	if err != nil || offset < 0 {
		return 0
	}
	return offset
}
