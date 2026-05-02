package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LikeRepo struct {
	db *pgxpool.Pool
}

func NewLikeRepo(db *pgxpool.Pool) *LikeRepo {
	return &LikeRepo{db: db}
}

func (r *LikeRepo) Toggle(ctx context.Context, userID, experienceID string) (bool, error) {
	// Try delete first
	result, err := r.db.Exec(ctx,
		`DELETE FROM likes WHERE user_id=$1 AND experience_id=$2`,
		userID, experienceID,
	)
	if err != nil {
		return false, fmt.Errorf("toggle like: %w", err)
	}

	if result.RowsAffected() > 0 {
		return false, nil // unliked
	}

	// Not deleted, so insert
	_, err = r.db.Exec(ctx,
		`INSERT INTO likes (user_id, experience_id, created_at) VALUES ($1,$2,$3)`,
		userID, experienceID, time.Now(),
	)
	if err != nil {
		return false, fmt.Errorf("insert like: %w", err)
	}

	return true, nil // liked
}
