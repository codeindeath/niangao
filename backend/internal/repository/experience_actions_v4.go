package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/niangao/backend/internal/model"
)

var ErrExperienceUnavailable = errors.New("experience unavailable")

func (r *ExperienceRepo) InspireExperience(ctx context.Context, userID string, experienceID string) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin inspire: %w", err)
	}
	defer tx.Rollback(ctx)

	var insertedID string
	err = tx.QueryRow(ctx, `
    INSERT INTO experience_inspirations (user_id, experience_id, source_context, inspired_at, created_at)
    SELECT $1::uuid, e.id, 'app', NOW(), NOW()
    FROM experiences e
    WHERE e.id = $2::uuid
      AND e.deleted_at IS NULL
      AND (
        (e.visibility = 'public' AND e.lifecycle_status = 'active')
        OR (
          COALESCE(e.owner_user_id, e.author_id) = $1::uuid
          AND e.lifecycle_status <> 'deleted'
        )
      )
    ON CONFLICT (user_id, experience_id) DO NOTHING
    RETURNING id`,
		userID, experienceID).Scan(&insertedID)
	if errors.Is(err, pgx.ErrNoRows) {
		var exists bool
		if checkErr := tx.QueryRow(ctx,
			`SELECT EXISTS(
        SELECT 1 FROM experience_inspirations
        WHERE user_id=$1::uuid AND experience_id=$2::uuid
      )`,
			userID, experienceID).Scan(&exists); checkErr != nil {
			return false, fmt.Errorf("check existing inspiration: %w", checkErr)
		}
		if exists {
			return true, tx.Commit(ctx)
		}
		return false, ErrExperienceUnavailable
	}
	if err != nil {
		return false, fmt.Errorf("insert inspiration: %w", err)
	}

	if _, err := tx.Exec(ctx, `
    UPDATE experiences
    SET inspiration_count = COALESCE(inspiration_count, like_count, 0) + 1,
        updated_at = NOW()
    WHERE id=$1::uuid`,
		experienceID); err != nil {
		return false, fmt.Errorf("increment inspiration count: %w", err)
	}

	if _, err := tx.Exec(ctx, `
    INSERT INTO experience_events (user_id, experience_id, event_type, source_context, created_at)
    VALUES ($1::uuid, $2::uuid, 'inspire', 'app', NOW())`,
		userID, experienceID); err != nil {
		return false, fmt.Errorf("insert inspire event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit inspiration: %w", err)
	}
	return false, nil
}

func (r *ExperienceRepo) CollectExperience(ctx context.Context, userID string, experienceID string) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin collect: %w", err)
	}
	defer tx.Rollback(ctx)

	var alreadyActive bool
	if err := tx.QueryRow(ctx, `
    SELECT EXISTS(
      SELECT 1 FROM experience_collections
      WHERE user_id=$1::uuid AND experience_id=$2::uuid AND status='active'
    )`,
		userID, experienceID).Scan(&alreadyActive); err != nil {
		return false, fmt.Errorf("check existing collection: %w", err)
	}
	if alreadyActive {
		return true, tx.Commit(ctx)
	}

	var visible bool
	if err := tx.QueryRow(ctx, `
    SELECT EXISTS(
      SELECT 1 FROM experiences e
      WHERE e.id=$2::uuid
        AND e.deleted_at IS NULL
        AND (
          (e.visibility = 'public' AND e.lifecycle_status = 'active')
          OR (
            COALESCE(e.owner_user_id, e.author_id) = $1::uuid
            AND e.lifecycle_status <> 'deleted'
          )
        )
    )`,
		userID, experienceID).Scan(&visible); err != nil {
		return false, fmt.Errorf("check collect visibility: %w", err)
	}
	if !visible {
		return false, ErrExperienceUnavailable
	}

	if _, err := tx.Exec(ctx, `
    INSERT INTO experience_collections (user_id, experience_id, status, source_context, collected_at, created_at, updated_at)
    VALUES ($1::uuid, $2::uuid, 'active', 'app', NOW(), NOW(), NOW())
    ON CONFLICT (user_id, experience_id)
    DO UPDATE SET status='active', source_context='app', collected_at=NOW(), updated_at=NOW()`,
		userID, experienceID); err != nil {
		return false, fmt.Errorf("upsert collection: %w", err)
	}

	if _, err := tx.Exec(ctx, `
    UPDATE experiences
    SET collection_count = COALESCE(collection_count, bookmark_count, 0) + 1,
        updated_at = NOW()
    WHERE id=$1::uuid`,
		experienceID); err != nil {
		return false, fmt.Errorf("increment collection count: %w", err)
	}

	if _, err := tx.Exec(ctx, `
    INSERT INTO experience_events (user_id, experience_id, event_type, source_context, created_at)
    VALUES ($1::uuid, $2::uuid, 'collect', 'app', NOW())`,
		userID, experienceID); err != nil {
		return false, fmt.Errorf("insert collect event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit collection: %w", err)
	}
	return false, nil
}

func (r *ExperienceRepo) UncollectExperience(ctx context.Context, userID string, experienceID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin uncollect: %w", err)
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx, `
    UPDATE experience_collections
    SET status='removed', updated_at=NOW()
    WHERE user_id=$1::uuid AND experience_id=$2::uuid AND status='active'`,
		userID, experienceID)
	if err != nil {
		return fmt.Errorf("remove collection: %w", err)
	}

	if result.RowsAffected() > 0 {
		if _, err := tx.Exec(ctx, `
      UPDATE experiences
      SET collection_count = GREATEST(COALESCE(collection_count, bookmark_count, 0) - 1, 0),
          updated_at = NOW()
      WHERE id=$1::uuid`,
			experienceID); err != nil {
			return fmt.Errorf("decrement collection count: %w", err)
		}
	}

	if _, err := tx.Exec(ctx, `
    INSERT INTO experience_events (user_id, experience_id, event_type, source_context, created_at)
    VALUES ($1::uuid, $2::uuid, 'uncollect', 'app', NOW())`,
		userID, experienceID); err != nil {
		return fmt.Errorf("insert uncollect event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit uncollection: %w", err)
	}
	return nil
}

func (r *ExperienceRepo) RecordExperienceEvent(ctx context.Context, userID string, experienceID string, event model.ExperienceEventRequest) error {
	metadata := event.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal event metadata: %w", err)
	}

	if event.EventType == "expose" && userID != "" {
		return r.recordDedupedExposeEvent(ctx, userID, experienceID, event, string(metadataJSON))
	}

	var insertedID string
	err = r.db.QueryRow(ctx, `
    INSERT INTO experience_events (user_id, experience_id, event_type, source_context, context_id, metadata, created_at)
    SELECT NULLIF($1, '')::uuid, e.id, $3, $4, NULLIF($5::text, '')::uuid, $6::jsonb, NOW()
    FROM experiences e
    WHERE e.id = $2::uuid
      AND e.deleted_at IS NULL
      AND (
        (e.visibility = 'public' AND e.lifecycle_status = 'active')
        OR (
          NULLIF($1, '') IS NOT NULL
          AND COALESCE(e.owner_user_id, e.author_id) = NULLIF($1, '')::uuid
          AND e.lifecycle_status <> 'deleted'
        )
      )
    RETURNING id`,
		userID, experienceID, event.EventType, event.SourceContext, event.ContextID, string(metadataJSON)).Scan(&insertedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrExperienceUnavailable
	}
	if err != nil {
		return fmt.Errorf("insert experience event: %w", err)
	}
	return nil
}

func (r *ExperienceRepo) recordDedupedExposeEvent(ctx context.Context, userID string, experienceID string, event model.ExperienceEventRequest, metadataJSON string) error {
	var eventID string
	err := r.db.QueryRow(ctx, `
    WITH visible AS (
      SELECT e.id
      FROM experiences e
      WHERE e.id = $2::uuid
        AND e.deleted_at IS NULL
        AND (
          (e.visibility = 'public' AND e.lifecycle_status = 'active')
          OR (
            COALESCE(e.owner_user_id, e.author_id) = $1::uuid
            AND e.lifecycle_status <> 'deleted'
          )
        )
    ),
    updated AS (
      UPDATE experience_events ev
      SET source_context = $4,
          context_id = NULLIF($5::text, '')::uuid,
          metadata = $6::jsonb,
          created_at = NOW()
      FROM visible
      WHERE ev.experience_id = visible.id
        AND ev.event_type = 'expose'
		AND ev.user_id = $1::uuid
        AND ev.created_at >= NOW() - INTERVAL '30 minutes'
      RETURNING ev.id
    ),
    inserted AS (
      INSERT INTO experience_events (user_id, experience_id, event_type, source_context, context_id, metadata, created_at)
      SELECT $1::uuid, visible.id, $3, $4, NULLIF($5::text, '')::uuid, $6::jsonb, NOW()
      FROM visible
      WHERE NOT EXISTS (SELECT 1 FROM updated)
      RETURNING id
    )
    SELECT id FROM updated
    UNION ALL
    SELECT id FROM inserted
    LIMIT 1`,
		userID, experienceID, event.EventType, event.SourceContext, event.ContextID, metadataJSON).Scan(&eventID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrExperienceUnavailable
	}
	if err != nil {
		return fmt.Errorf("dedupe expose event: %w", err)
	}
	return nil
}
