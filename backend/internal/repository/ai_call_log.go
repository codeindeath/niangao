package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/niangao/backend/internal/ai"
)

type aiCallLogDB interface {
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
}

type AICallLogRepo struct {
	db aiCallLogDB
}

func NewAICallLogRepo(db aiCallLogDB) *AICallLogRepo {
	return &AICallLogRepo{db: db}
}

func (r *AICallLogRepo) RecordAICall(ctx context.Context, entry ai.CallLogEntry) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("ai call log db is nil")
	}
	tag, err := r.db.Exec(ctx, `
		INSERT INTO ai_call_logs (
			function_type,
			key_alias,
			provider,
			model,
			prompt_version,
			schema_version,
			call_source,
			queue_name,
			user_id,
			experience_id,
			chat_topic_id,
			chat_message_id,
			latency_ms,
			status,
			error_code,
			sanitized_input_summary,
			sanitized_output_summary,
			started_at,
			finished_at
		)
		SELECT
			cfg.function_type,
			cfg.key_alias,
			'deepseek',
			cfg.model,
			cfg.prompt_version,
			cfg.schema_version,
			$2,
			cfg.queue_name,
			NULLIF($3, '')::uuid,
			NULLIF($4, '')::uuid,
			NULLIF($5, '')::uuid,
			NULLIF($6, '')::uuid,
			$7,
			$8,
			NULLIF($9, ''),
			NULLIF($10, ''),
			NULLIF($11, ''),
			$12,
			$13
		FROM ai_function_configs cfg
		WHERE cfg.function_type = $1
		  AND cfg.status = 'active'
	`, entry.FunctionType,
		entry.CallSource,
		entry.UserID,
		entry.ExperienceID,
		entry.ChatTopicID,
		entry.ChatMessageID,
		entry.LatencyMS,
		entry.Status,
		entry.ErrorCode,
		entry.SanitizedInputSummary,
		entry.SanitizedOutputSummary,
		entry.StartedAt,
		entry.FinishedAt,
	)
	if err != nil {
		return fmt.Errorf("insert ai call log: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("active ai function config not found for %s", entry.FunctionType)
	}
	return nil
}
