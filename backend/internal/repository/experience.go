package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niangao/backend/internal/model"
)

type ExperienceRepo struct {
	db *pgxpool.Pool
}

func NewExperienceRepo(db *pgxpool.Pool) *ExperienceRepo {
	return &ExperienceRepo{db: db}
}

func (r *ExperienceRepo) Create(ctx context.Context, authorID string, req model.CreateExperienceRequest) (*model.Experience, error) {
	exp := &model.Experience{
		ID:        newUUID(),
		AuthorID:  authorID,
		Content:   req.Content,
		Domain:    req.Domain,
		Status:    "published",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if req.Interpretation != "" {
		exp.Interpretation = &req.Interpretation
	}

	_, err := r.db.Exec(ctx,
		`INSERT INTO experiences (id, author_id, content, interpretation, domain, status, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		exp.ID, exp.AuthorID, exp.Content, exp.Interpretation, exp.Domain, exp.Status, exp.CreatedAt, exp.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert experience: %w", err)
	}

	return exp, nil
}

func (r *ExperienceRepo) GetByID(ctx context.Context, id string, viewerID string) (*model.Experience, error) {
	row := r.db.QueryRow(ctx,
		`SELECT e.id, e.author_id, e.content, e.interpretation, e.domain, e.is_official,
		        e.source_label, e.like_count, e.bookmark_count, e.interpretation_generated,
		        e.status, e.created_at, e.updated_at,
		        p.nickname, p.avatar_url,
		        EXISTS(SELECT 1 FROM likes WHERE user_id=$2 AND experience_id=e.id) as is_liked,
		        EXISTS(SELECT 1 FROM bookmarks WHERE user_id=$2 AND experience_id=e.id) as is_bookmarked
		 FROM experiences e
		 LEFT JOIN profiles p ON p.id = e.author_id
		 WHERE e.id = $1`,
		id, viewerID,
	)

	exp := &model.Experience{}
	err := row.Scan(
		&exp.ID, &exp.AuthorID, &exp.Content, &exp.Interpretation, &exp.Domain,
		&exp.IsOfficial, &exp.SourceLabel, &exp.LikeCount, &exp.BookmarkCount,
		&exp.InterpretationGenerated, &exp.Status, &exp.CreatedAt, &exp.UpdatedAt,
		&exp.AuthorName, &exp.AuthorAvatar, &exp.IsLiked, &exp.IsBookmarked,
	)
	if err != nil {
		return nil, fmt.Errorf("get experience: %w", err)
	}
	return exp, nil
}

func (r *ExperienceRepo) List(ctx context.Context, query model.ExperienceListQuery, viewerID string) ([]model.Experience, int, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 50 {
		query.PageSize = 20
	}

	var where []string
	var args []interface{}
	argIdx := 1

	where = append(where, "e.status = 'published'")
	if query.Domain != "" {
		where = append(where, fmt.Sprintf("e.domain = $%d", argIdx))
		args = append(args, string(query.Domain))
		argIdx++
	}
	if query.Search != "" {
		where = append(where, fmt.Sprintf("(e.content ILIKE $%d OR e.interpretation ILIKE $%d)", argIdx, argIdx+1))
		args = append(args, "%"+query.Search+"%", "%"+query.Search+"%")
		argIdx += 2
	}

	whereClause := strings.Join(where, " AND ")

	// Count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM experiences e WHERE %s", whereClause)
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count experiences: %w", err)
	}

	// Order
	orderClause := "e.created_at DESC"
	if query.Sort == "popular" {
		orderClause = "e.like_count DESC, e.created_at DESC"
	}

	// List
	selectQuery := fmt.Sprintf(
		`SELECT e.id, e.author_id, e.content, e.interpretation, e.domain, e.is_official,
		        e.source_label, e.like_count, e.bookmark_count, e.interpretation_generated,
		        e.status, e.created_at, e.updated_at,
		        p.nickname, p.avatar_url,
		        EXISTS(SELECT 1 FROM likes WHERE user_id=$%d AND experience_id=e.id) as is_liked,
		        EXISTS(SELECT 1 FROM bookmarks WHERE user_id=$%d AND experience_id=e.id) as is_bookmarked
		 FROM experiences e
		 LEFT JOIN profiles p ON p.id = e.author_id
		 WHERE %s
		 ORDER BY %s
		 LIMIT $%d OFFSET $%d`,
		argIdx, argIdx+1, whereClause, orderClause, argIdx+2, argIdx+3,
	)

	args = append(args, viewerID, viewerID, query.PageSize, (query.Page-1)*query.PageSize)

	rows, err := r.db.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list experiences: %w", err)
	}
	defer rows.Close()

	var experiences []model.Experience
	for rows.Next() {
		var e model.Experience
		err := rows.Scan(
			&e.ID, &e.AuthorID, &e.Content, &e.Interpretation, &e.Domain,
			&e.IsOfficial, &e.SourceLabel, &e.LikeCount, &e.BookmarkCount,
			&e.InterpretationGenerated, &e.Status, &e.CreatedAt, &e.UpdatedAt,
			&e.AuthorName, &e.AuthorAvatar, &e.IsLiked, &e.IsBookmarked,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan experience: %w", err)
		}
		experiences = append(experiences, e)
	}

	return experiences, total, nil
}

func (r *ExperienceRepo) SearchByEmbedding(ctx context.Context, embedding []float32, userID string, limit int) ([]model.Experience, error) {
	// For pgvector similarity search
	rows, err := r.db.Query(ctx,
		`SELECT e.id, e.content, e.domain, e.like_count,
		        p.nickname as author_name
		 FROM experiences e
		 LEFT JOIN profiles p ON p.id = e.author_id
		 WHERE e.status = 'published' AND e.author_id = $1
		 ORDER BY e.embedding <=> $2
		 LIMIT $3`,
		userID, embedding, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("search by embedding: %w", err)
	}
	defer rows.Close()

	var experiences []model.Experience
	for rows.Next() {
		var e model.Experience
		err := rows.Scan(&e.ID, &e.Content, &e.Domain, &e.LikeCount, &e.AuthorName)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		experiences = append(experiences, e)
	}

	return experiences, nil
}

func (r *ExperienceRepo) Update(ctx context.Context, id, authorID string, req model.CreateExperienceRequest) error {
	_, err := r.db.Exec(ctx,
		`UPDATE experiences SET content=$1, interpretation=$2, domain=$3, updated_at=$4
		 WHERE id=$5 AND author_id=$6`,
		req.Content, req.Interpretation, req.Domain, time.Now(), id, authorID,
	)
	return err
}

func (r *ExperienceRepo) Delete(ctx context.Context, id, authorID string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM experiences WHERE id=$1 AND author_id=$2`,
		id, authorID,
	)
	return err
}

func newUUID() string {
	// Use pgx type or UUID generation
	return "" // will be generated by DB
}
