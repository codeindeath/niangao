package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
    CASE e.quality_tier
      WHEN 'high_trust' THEN 5
      WHEN 'ai_citable' THEN 4
      WHEN 'recommend_candidate' THEN 3
      ELSE 0
    END DESC,
    COALESCE(vds.score, 0) DESC,
    (COALESCE(e.inspiration_count, e.like_count, 0) + COALESCE(e.collection_count, e.bookmark_count, 0)) DESC,
    CASE e.experience_type
      WHEN 'platform_selected' THEN 0
      ELSE 1
    END ASC,
    e.created_at DESC,
    e.id DESC
  LIMIT $2 OFFSET $3`

const recommendSessionCardsQuery = `
  WITH session AS (
    SELECT candidate_ids
    FROM recommendation_sessions
    WHERE session_id = $2::uuid
      AND expires_at > NOW()
      AND (
        (NULLIF($1::text, '') IS NULL AND user_id IS NULL)
        OR user_id = NULLIF($1::text, '')::uuid
      )
  ),
  candidate AS (
    SELECT ids.candidate_id, ids.ordinality::int AS ord
    FROM session,
      unnest(session.candidate_ids) WITH ORDINALITY AS ids(candidate_id, ordinality)
    ORDER BY ids.ordinality
    LIMIT $3 OFFSET $4
  )
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
      WHERE ec.user_id = NULLIF($1::text, '')::uuid
        AND ec.experience_id = e.id
        AND ec.status = 'active'
    ),
    EXISTS(
      SELECT 1 FROM experience_inspirations ei
      WHERE ei.user_id = NULLIF($1::text, '')::uuid
        AND ei.experience_id = e.id
    ),
    '' AS unavailable_reason
  FROM candidate c
  JOIN experiences e ON e.id = c.candidate_id
  LEFT JOIN users u ON u.id = e.author_id
  WHERE e.visibility = 'public'
    AND e.lifecycle_status = 'active'
    AND e.recommendation_status = 'eligible'
    AND e.quality_tier IN ('recommend_candidate', 'ai_citable', 'high_trust')
    AND e.deleted_at IS NULL
  ORDER BY c.ord`

const recommendSessionCandidateLimit = 160

const collectionsFeedQuery = `
  WITH collected AS (
    SELECT
      ec.collected_at,
      ec.id AS collection_id,
      e.*,
      u.display_name,
      u.nickname,
      (
        e.deleted_at IS NULL
        AND (
          (
            e.visibility = 'public'
            AND e.lifecycle_status = 'active'
          )
          OR (
            COALESCE(e.owner_user_id, e.author_id) = $1::uuid
            AND e.lifecycle_status <> 'deleted'
          )
        )
      ) AS visible_to_viewer
    FROM experience_collections ec
    JOIN experiences e ON e.id = ec.experience_id
    LEFT JOIN users u ON u.id = e.author_id
    WHERE ec.user_id = $1::uuid
      AND ec.status = 'active'
  )
  SELECT
    c.id,
    CASE WHEN c.visible_to_viewer THEN COALESCE(c.owner_user_id, c.author_id)::text ELSE '' END,
    CASE WHEN c.visible_to_viewer THEN c.content ELSE '' END AS content,
    CASE WHEN c.visible_to_viewer THEN c.experience_type ELSE '' END AS experience_type,
    CASE WHEN c.visible_to_viewer THEN c.visibility ELSE '' END AS visibility,
    CASE WHEN c.visible_to_viewer THEN c.lifecycle_status ELSE '' END AS lifecycle_status,
    CASE WHEN c.visible_to_viewer THEN COALESCE(c.domain::text, '') ELSE '' END AS domain,
    CASE WHEN c.visible_to_viewer THEN COALESCE(c.sub_domain, '') ELSE '' END AS sub_domain,
    CASE WHEN c.visible_to_viewer THEN c.topic ELSE '' END AS topic,
    CASE WHEN c.visible_to_viewer THEN COALESCE(c.creator_display_name, c.creator_name, c.display_name, c.nickname, '') ELSE '' END AS creator_display_name,
    CASE WHEN c.visible_to_viewer THEN c.interpretation_status ELSE '' END AS interpretation_status,
    CASE WHEN c.visible_to_viewer THEN (c.interpretation IS NOT NULL AND c.interpretation <> '') ELSE FALSE END AS interpretation_summary_available,
    CASE WHEN c.visible_to_viewer THEN c.quality_tier ELSE '' END AS quality_tier,
    CASE WHEN c.visible_to_viewer THEN CASE c.quality_tier
      WHEN 'high_trust' THEN 5
      WHEN 'ai_citable' THEN 4
      WHEN 'recommend_candidate' THEN 3
      WHEN 'public_visible' THEN 2
      ELSE 1
    END ELSE 0 END AS star_rating,
    CASE WHEN c.visible_to_viewer THEN COALESCE(c.inspiration_count, c.like_count, 0) ELSE 0 END AS inspiration_count,
    CASE WHEN c.visible_to_viewer THEN COALESCE(c.collection_count, c.bookmark_count, 0) ELSE 0 END AS collection_count,
    TRUE AS is_collected,
    CASE WHEN c.visible_to_viewer THEN EXISTS(
      SELECT 1 FROM experience_inspirations ei
      WHERE ei.user_id = $1::uuid
        AND ei.experience_id = c.id
    ) ELSE FALSE END AS is_inspired,
    CASE WHEN c.visible_to_viewer THEN '' ELSE 'experience_unavailable' END AS unavailable_reason
  FROM collected c
  ORDER BY c.collected_at DESC, c.collection_id DESC
  LIMIT $2 OFFSET $3`

const mineFeedQuery = `
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
      WHERE ec.user_id = $1::uuid
        AND ec.experience_id = e.id
        AND ec.status = 'active'
    ),
    EXISTS(
      SELECT 1 FROM experience_inspirations ei
      WHERE ei.user_id = $1::uuid
        AND ei.experience_id = e.id
    ),
    '' AS unavailable_reason
  FROM experiences e
  LEFT JOIN users u ON u.id = e.author_id
  WHERE COALESCE(e.owner_user_id, e.author_id) = $1::uuid
    AND e.lifecycle_status <> 'deleted'
    AND e.deleted_at IS NULL
  ORDER BY e.created_at DESC, e.id DESC
  LIMIT $2 OFFSET $3`

func (r *ExperienceRepo) RecommendFeed(ctx context.Context, userID string, limit int, cursor string) (*model.FeedPage, error) {
	if sessionID, offset, ok := parseRecommendationCursor(cursor); ok {
		page, found, err := r.recommendFeedFromSession(ctx, userID, sessionID, limit, offset)
		if err != nil {
			return nil, err
		}
		if found {
			return page, nil
		}
	}

	rows, err := r.db.Query(ctx, recommendFeedQuery,
		userID, recommendSessionCandidateLimit, 0)
	if err != nil {
		return nil, fmt.Errorf("v4 recommend feed: %w", err)
	}
	defer rows.Close()

	cards, err := scanFeedCards(rows, recommendSessionCandidateLimit)
	if err != nil {
		return nil, err
	}
	sessionID, err := r.createRecommendationSession(ctx, userID, cards)
	if err != nil {
		page := buildFeedPage(cards, limit, 0, "", func(int) string { return "" })
		page.HasMore = false
		return page, nil
	}
	return buildFeedPage(cards, limit, 0, sessionID, func(nextOffset int) string {
		return formatRecommendationCursor(sessionID, nextOffset)
	}), nil
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
	cards, err := scanFeedCards(rows, limit+1)
	if err != nil {
		return nil, err
	}
	return buildFeedPage(cards, limit, offset, sessionID, func(nextOffset int) string {
		return strconv.Itoa(nextOffset)
	}), nil
}

func scanFeedCards(rows pgx.Rows, capacity int) ([]model.ExperienceCard, error) {
	if capacity < 0 {
		capacity = 0
	}
	cards := make([]model.ExperienceCard, 0, capacity)
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
			&card.UnavailableReason,
		); err != nil {
			return nil, fmt.Errorf("scan v4 feed card: %w", err)
		}
		cards = append(cards, card)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate v4 feed cards: %w", err)
	}

	return cards, nil
}

func buildFeedPage(cards []model.ExperienceCard, limit int, offset int, sessionID string, nextCursorFor func(int) string) *model.FeedPage {
	hasMore := len(cards) > limit
	if hasMore {
		cards = cards[:limit]
	}
	nextCursor := ""
	if hasMore && nextCursorFor != nil {
		nextCursor = nextCursorFor(offset + limit)
	}
	return &model.FeedPage{
		Data:       cards,
		NextCursor: nextCursor,
		SessionID:  sessionID,
		HasMore:    hasMore,
	}
}

func (r *ExperienceRepo) recommendFeedFromSession(ctx context.Context, userID string, sessionID string, limit int, offset int) (*model.FeedPage, bool, error) {
	exists, err := r.recommendationSessionExists(ctx, userID, sessionID)
	if err != nil {
		return nil, false, fmt.Errorf("check recommendation session: %w", err)
	}
	if !exists {
		return nil, false, nil
	}

	rows, err := r.db.Query(ctx, recommendSessionCardsQuery, userID, sessionID, limit+1, offset)
	if err != nil {
		return nil, true, fmt.Errorf("v4 recommend session feed: %w", err)
	}
	defer rows.Close()

	cards, err := scanFeedCards(rows, limit+1)
	if err != nil {
		return nil, true, err
	}
	page := buildFeedPage(cards, limit, offset, sessionID, func(nextOffset int) string {
		return formatRecommendationCursor(sessionID, nextOffset)
	})
	r.markRecommendationSessionOffset(ctx, sessionID, offset+len(page.Data))
	return page, true, nil
}

func (r *ExperienceRepo) recommendationSessionExists(ctx context.Context, userID string, sessionID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
    SELECT EXISTS(
      SELECT 1
      FROM recommendation_sessions
      WHERE session_id = $2::uuid
        AND expires_at > NOW()
        AND (
          (NULLIF($1::text, '') IS NULL AND user_id IS NULL)
          OR user_id = NULLIF($1::text, '')::uuid
        )
    )`,
		userID, sessionID).Scan(&exists)
	return exists, err
}

func (r *ExperienceRepo) createRecommendationSession(ctx context.Context, userID string, cards []model.ExperienceCard) (string, error) {
	candidateIDs := make([]string, 0, len(cards))
	for _, card := range cards {
		if card.ID != "" {
			candidateIDs = append(candidateIDs, card.ID)
		}
	}

	var sessionID string
	err := r.db.QueryRow(ctx, `
    INSERT INTO recommendation_sessions (user_id, candidate_ids, metadata)
    SELECT
      NULLIF($1::text, '')::uuid,
      ARRAY(SELECT unnest($2::text[])::uuid),
      jsonb_build_object('source', 'v4_rule_recommend', 'candidate_count', cardinality($2::text[]))
    RETURNING session_id`,
		userID, candidateIDs).Scan(&sessionID)
	if err != nil {
		return "", fmt.Errorf("create recommendation session: %w", err)
	}
	return sessionID, nil
}

func (r *ExperienceRepo) markRecommendationSessionOffset(ctx context.Context, sessionID string, returnedOffset int) {
	if sessionID == "" || returnedOffset < 0 {
		return
	}
	_, _ = r.db.Exec(ctx, `
    UPDATE recommendation_sessions
    SET returned_offset = GREATEST(returned_offset, $2)
    WHERE session_id = $1::uuid`,
		sessionID, returnedOffset)
}

func formatRecommendationCursor(sessionID string, offset int) string {
	if !isUUIDString(sessionID) || offset < 0 {
		return ""
	}
	return "rec:" + sessionID + ":" + strconv.Itoa(offset)
}

func parseRecommendationCursor(cursor string) (string, int, bool) {
	parts := strings.Split(cursor, ":")
	if len(parts) != 3 || parts[0] != "rec" || !isUUIDString(parts[1]) {
		return "", 0, false
	}
	offset, err := strconv.Atoi(parts[2])
	if err != nil || offset < 0 {
		return "", 0, false
	}
	return parts[1], offset, true
}

func isUUIDString(value string) bool {
	if len(value) != 36 {
		return false
	}
	for i, r := range value {
		switch i {
		case 8, 13, 18, 23:
			if r != '-' {
				return false
			}
		default:
			if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
				return false
			}
		}
	}
	return true
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
