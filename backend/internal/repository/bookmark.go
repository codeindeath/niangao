package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BookmarkRepo struct {
	db *pgxpool.Pool
}

func NewBookmarkRepo(db *pgxpool.Pool) *BookmarkRepo {
	return &BookmarkRepo{db: db}
}

func (r *BookmarkRepo) Toggle(ctx context.Context, userID, experienceID string) (bool, error) {
	result, err := r.db.Exec(ctx,
		`DELETE FROM bookmarks WHERE user_id=$1 AND experience_id=$2`,
		userID, experienceID,
	)
	if err != nil {
		return false, fmt.Errorf("toggle bookmark: %w", err)
	}

	if result.RowsAffected() > 0 {
		return false, nil
	}

	_, err = r.db.Exec(ctx,
		`INSERT INTO bookmarks (user_id, experience_id, created_at) VALUES ($1,$2,$3)`,
		userID, experienceID, time.Now(),
	)
	if err != nil {
		return false, fmt.Errorf("insert bookmark: %w", err)
	}

	return true, nil
}

func (r *BookmarkRepo) ListByUser(ctx context.Context, userID string, page, pageSize int) ([]string, int, error) {
	var total int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM bookmarks WHERE user_id=$1`, userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx,
		`SELECT experience_id FROM bookmarks WHERE user_id=$1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, pageSize, (page-1)*pageSize,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, 0, err
		}
		ids = append(ids, id)
	}

	return ids, total, nil
}

func (r *BookmarkRepo) MarkPracticed(ctx context.Context, userID, experienceID string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE bookmarks SET practiced=true, practiced_at=$1
		 WHERE user_id=$2 AND experience_id=$3`,
		time.Now(), userID, experienceID,
	)
	return err
}

// BookmarkedExperience is a lightweight experience struct for chat context.
type BookmarkedExperience struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Domain  string `json:"domain"`
}

// ListBookmarkedExperiences returns all bookmarked experiences with content for AI context.
func (r *BookmarkRepo) ListBookmarkedExperiences(ctx context.Context, userID string) ([]BookmarkedExperience, error) {
	rows, err := r.db.Query(ctx,
		`SELECT e.id, e.content, e.domain::text
		 FROM bookmarks b
		 JOIN experiences e ON e.id = b.experience_id
		 WHERE b.user_id=$1 AND e.deleted_at IS NULL
		 ORDER BY b.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exps []BookmarkedExperience
	for rows.Next() {
		var exp BookmarkedExperience
		if err := rows.Scan(&exp.ID, &exp.Content, &exp.Domain); err != nil {
			return nil, err
		}
		exps = append(exps, exp)
	}
	if exps == nil {
		exps = make([]BookmarkedExperience, 0)
	}
	return exps, nil
}
