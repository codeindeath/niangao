package repository

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/niangao/backend/internal/model"
)

func (r *ExperienceRepo) AssetStats(ctx context.Context, userID string) (*model.AssetStats, error) {
	stats := &model.AssetStats{}
	err := r.db.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*)
			 FROM experiences e
			 WHERE COALESCE(e.owner_user_id, e.author_id)=$1::uuid
			   AND COALESCE(e.lifecycle_status, 'active') <> 'deleted'
			   AND e.deleted_at IS NULL),
			(SELECT COUNT(*)
			 FROM experience_collections ec
			 WHERE ec.user_id=$1::uuid AND ec.status='active'),
			(SELECT COUNT(*)
			 FROM experiences e
			 WHERE COALESCE(e.owner_user_id, e.author_id)=$1::uuid
			   AND COALESCE(e.source_scene, 'note') IN ('note', 'chat')
			   AND e.created_at >= date_trunc('month', NOW())
			   AND COALESCE(e.lifecycle_status, 'active') <> 'deleted'
			   AND e.deleted_at IS NULL),
			(SELECT COUNT(*)
			 FROM experiences e
			 WHERE COALESCE(e.owner_user_id, e.author_id)=$1::uuid
			   AND COALESCE(e.visibility, CASE WHEN e.is_private THEN 'private' ELSE 'public' END)='public'
			   AND COALESCE(e.lifecycle_status, 'active') <> 'deleted'
			   AND e.deleted_at IS NULL),
			(SELECT COUNT(*)
			 FROM experiences e
			 WHERE COALESCE(e.owner_user_id, e.author_id)=$1::uuid
			   AND COALESCE(e.visibility, CASE WHEN e.is_private THEN 'private' ELSE 'public' END)='private'
			   AND COALESCE(e.lifecycle_status, 'active') <> 'deleted'
			   AND e.deleted_at IS NULL),
			(SELECT COUNT(*)
			 FROM experiences e
			 WHERE COALESCE(e.owner_user_id, e.author_id)=$1::uuid
			   AND COALESCE(e.source_scene, 'note')='note'
			   AND COALESCE(e.lifecycle_status, 'active') <> 'deleted'
			   AND e.deleted_at IS NULL),
			(SELECT COUNT(*)
			 FROM experiences e
			 WHERE COALESCE(e.owner_user_id, e.author_id)=$1::uuid
			   AND COALESCE(e.source_scene, '')='chat'
			   AND COALESCE(e.lifecycle_status, 'active') <> 'deleted'
			   AND e.deleted_at IS NULL)`,
		userID,
	).Scan(
		&stats.MyExperiences,
		&stats.Collections,
		&stats.MonthAdded,
		&stats.PublicExperiences,
		&stats.PrivateExperiences,
		&stats.FromNote,
		&stats.FromChat,
	)
	if err != nil {
		return nil, fmt.Errorf("asset stats: %w", err)
	}
	return stats, nil
}

func (r *ExperienceRepo) ContributionStats(ctx context.Context, userID string) (*model.ContributionStats, error) {
	stats := &model.ContributionStats{}
	err := r.db.QueryRow(ctx, `
		WITH owned_public AS (
			SELECT id
			FROM experiences e
			WHERE COALESCE(e.owner_user_id, e.author_id)=$1::uuid
			  AND COALESCE(e.experience_type, 'user_original')='user_original'
			  AND COALESCE(e.visibility, CASE WHEN e.is_private THEN 'private' ELSE 'public' END)='public'
			  AND COALESCE(e.lifecycle_status, 'active')='active'
			  AND e.deleted_at IS NULL
		)
		SELECT
			(SELECT COUNT(DISTINCT ei.user_id)
			 FROM experience_inspirations ei
			 JOIN owned_public op ON op.id=ei.experience_id),
			(SELECT COUNT(*)
			 FROM experience_collections ec
			 JOIN owned_public op ON op.id=ec.experience_id
			 WHERE ec.status='active'),
			(SELECT COUNT(DISTINCT ei.user_id)
			 FROM experience_inspirations ei
			 JOIN owned_public op ON op.id=ei.experience_id
			 WHERE ei.created_at >= date_trunc('month', NOW())),
			(SELECT COUNT(*)
			 FROM experience_collections ec
			 JOIN owned_public op ON op.id=ec.experience_id
			 WHERE ec.status='active' AND ec.collected_at >= date_trunc('month', NOW()))`,
		userID,
	).Scan(
		&stats.InspiredUsers,
		&stats.CollectedCount,
		&stats.MonthInspiredUsers,
		&stats.MonthCollected,
	)
	if err != nil {
		return nil, fmt.Errorf("contribution stats: %w", err)
	}
	return stats, nil
}

func (r *ExperienceRepo) ChangeStats(ctx context.Context, userID string) (*model.ChangeStats, error) {
	stats := &model.ChangeStats{}
	err := r.db.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*)
			 FROM chat_topics
			 WHERE user_id=$1::uuid AND status='active' AND deleted_at IS NULL),
			0,
			(SELECT COUNT(*)
			 FROM experiences e
			 WHERE COALESCE(e.owner_user_id, e.author_id)=$1::uuid
			   AND COALESCE(e.source_scene, '')='chat'
			   AND e.created_at >= date_trunc('month', NOW())
			   AND COALESCE(e.lifecycle_status, 'active') <> 'deleted'
			   AND e.deleted_at IS NULL)`,
		userID,
	).Scan(&stats.ChatTopics, &stats.ClearerCount, &stats.MonthChatExperiences)
	if err != nil {
		return nil, fmt.Errorf("change stats: %w", err)
	}
	return stats, nil
}

func (r *ExperienceRepo) RecentHarvestStats(ctx context.Context, userID string, rangeKey string) (*model.RecentHarvestStats, error) {
	experienceTimeFilter, inspirationTimeFilter, collectionTimeFilter, err := recentRangeFilters(rangeKey)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf(`
		WITH owned_public AS (
			SELECT id
			FROM experiences e
			WHERE COALESCE(e.owner_user_id, e.author_id)=$1::uuid
			  AND COALESCE(e.experience_type, 'user_original')='user_original'
			  AND COALESCE(e.visibility, CASE WHEN e.is_private THEN 'private' ELSE 'public' END)='public'
			  AND COALESCE(e.lifecycle_status, 'active')='active'
			  AND e.deleted_at IS NULL
		)
		SELECT
			(SELECT COUNT(*)
			 FROM experiences e
			 WHERE COALESCE(e.owner_user_id, e.author_id)=$1::uuid
			   AND COALESCE(e.source_scene, 'note')='note'
			   AND COALESCE(e.lifecycle_status, 'active') <> 'deleted'
			   AND e.deleted_at IS NULL
			   %s),
			(SELECT COUNT(*)
			 FROM experiences e
			 WHERE COALESCE(e.owner_user_id, e.author_id)=$1::uuid
			   AND COALESCE(e.source_scene, '')='chat'
			   AND COALESCE(e.lifecycle_status, 'active') <> 'deleted'
			   AND e.deleted_at IS NULL
			   %s),
			(SELECT COUNT(DISTINCT ei.user_id)
			 FROM experience_inspirations ei
			 JOIN owned_public op ON op.id=ei.experience_id
			 WHERE TRUE %s),
			(SELECT COUNT(*)
			 FROM experience_collections ec
			 JOIN owned_public op ON op.id=ec.experience_id
			 WHERE ec.status='active' %s)`,
		experienceTimeFilter,
		experienceTimeFilter,
		inspirationTimeFilter,
		collectionTimeFilter,
	)
	stats := &model.RecentHarvestStats{Range: rangeKey}
	if err := r.db.QueryRow(ctx, query, userID).Scan(
		&stats.NoteAdded,
		&stats.ChatExperiences,
		&stats.InspiredUsers,
		&stats.CollectedCount,
	); err != nil {
		return nil, fmt.Errorf("recent harvest stats: %w", err)
	}
	return stats, nil
}

func (r *ExperienceRepo) RecentRespondedExperiences(ctx context.Context, userID string, limit int) ([]model.RespondedExperienceCard, error) {
	if limit < 1 {
		limit = 1
	}
	if limit > 10 {
		limit = 10
	}
	rows, err := r.db.Query(ctx, `
		WITH owned_public AS (
			SELECT
				id,
				content,
				COALESCE(domain, '') AS domain,
				COALESCE(sub_domain, '') AS sub_domain,
				COALESCE(quality_score, 0) AS quality_score
			FROM experiences e
			WHERE COALESCE(e.owner_user_id, e.author_id)=$1::uuid
			  AND COALESCE(e.experience_type, 'user_original')='user_original'
			  AND COALESCE(e.visibility, CASE WHEN e.is_private THEN 'private' ELSE 'public' END)='public'
			  AND COALESCE(e.lifecycle_status, 'active')='active'
			  AND e.deleted_at IS NULL
		),
		inspire AS (
			SELECT experience_id, COUNT(DISTINCT user_id) AS inspiration_count, MAX(created_at) AS last_inspired_at
			FROM experience_inspirations
			GROUP BY experience_id
		),
		collect AS (
			SELECT experience_id, COUNT(*) AS collection_count, MAX(collected_at) AS last_collected_at
			FROM experience_collections
			WHERE status='active'
			GROUP BY experience_id
		)
		SELECT
			op.id::text,
			op.content,
			op.domain,
			op.sub_domain,
			op.quality_score,
			COALESCE(i.inspiration_count, 0),
			COALESCE(c.collection_count, 0),
			GREATEST(
				COALESCE(i.last_inspired_at, '1970-01-01'::timestamptz),
				COALESCE(c.last_collected_at, '1970-01-01'::timestamptz)
			) AS last_responded_at
		FROM owned_public op
		LEFT JOIN inspire i ON i.experience_id=op.id
		LEFT JOIN collect c ON c.experience_id=op.id
		WHERE COALESCE(i.inspiration_count, 0) + COALESCE(c.collection_count, 0) > 0
		ORDER BY
			COALESCE(i.inspiration_count, 0) + COALESCE(c.collection_count, 0) DESC,
			last_responded_at DESC,
			op.id DESC
		LIMIT $2`,
		userID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("recent responded experiences: %w", err)
	}
	defer rows.Close()

	cards := []model.RespondedExperienceCard{}
	for rows.Next() {
		var card model.RespondedExperienceCard
		var qualityScore float64
		var lastRespondedAt time.Time
		if err := rows.Scan(
			&card.ID,
			&card.Content,
			&card.Domain,
			&card.SubDomain,
			&qualityScore,
			&card.InspirationCount,
			&card.CollectionCount,
			&lastRespondedAt,
		); err != nil {
			return nil, fmt.Errorf("scan recent responded experience: %w", err)
		}
		card.StarRating = starRatingFromQualityScore(qualityScore)
		card.LastRespondedAt = lastRespondedAt
		cards = append(cards, card)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recent responded experiences: %w", err)
	}
	return cards, nil
}

func recentRangeFilters(rangeKey string) (experienceFilter string, inspirationFilter string, collectionFilter string, err error) {
	switch rangeKey {
	case "7d":
		return "AND e.created_at >= NOW() - INTERVAL '7 days'",
			"AND ei.created_at >= NOW() - INTERVAL '7 days'",
			"AND ec.collected_at >= NOW() - INTERVAL '7 days'",
			nil
	case "30d":
		return "AND e.created_at >= NOW() - INTERVAL '30 days'",
			"AND ei.created_at >= NOW() - INTERVAL '30 days'",
			"AND ec.collected_at >= NOW() - INTERVAL '30 days'",
			nil
	case "all":
		return "", "", "", nil
	default:
		return "", "", "", fmt.Errorf("invalid recent stats range: %s", rangeKey)
	}
}

func starRatingFromQualityScore(score float64) int {
	if score <= 0 {
		return 0
	}
	stars := int(math.Round(score / 2))
	if stars < 1 {
		return 1
	}
	if stars > 5 {
		return 5
	}
	return stars
}
