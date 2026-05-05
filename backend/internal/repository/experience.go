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

// nilUUID replaces empty viewerID for PostgreSQL UUID compatibility
const nilUUID = "00000000-0000-0000-0000-000000000000"

func NewExperienceRepo(db *pgxpool.Pool) *ExperienceRepo {
	return &ExperienceRepo{db: db}
}

func (r *ExperienceRepo) Create(ctx context.Context, authorID string, req model.CreateExperienceRequest) (*model.Experience, error) {
	exp := &model.Experience{
		AuthorID:     authorID,
		Content:      req.Content,
		Domain:       req.Domain,
		SubDomain:    string(req.SubDomain),
		IsPrivate:    req.IsPrivate,
		Status:       "published",
		ReviewStatus: string(model.ReviewPrivate),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if req.Interpretation != "" {
		exp.Interpretation = &req.Interpretation
	}

	err := r.db.QueryRow(ctx,
		`INSERT INTO experiences (author_id, content, interpretation, domain, sub_domain, is_private,
		 review_status, status, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`,
		exp.AuthorID, exp.Content, exp.Interpretation, exp.Domain, exp.SubDomain, exp.IsPrivate,
		exp.ReviewStatus, exp.Status, exp.CreatedAt, exp.UpdatedAt,
	).Scan(&exp.ID)
	if err != nil {
		return nil, fmt.Errorf("insert experience: %w", err)
	}

	return exp, nil
}

// CreateWithReview creates an experience with review fields (sub_domain, is_private, review_status, quality_score).
func (r *ExperienceRepo) CreateWithReview(ctx context.Context, authorID string, req model.CreateExperienceRequest, reviewStatus string, reviewReason *string, qualityScore *float64, scoreDetails *string) (*model.Experience, error) {
	exp := &model.Experience{
		AuthorID:     authorID,
		Content:      req.Content,
		Domain:       req.Domain,
		SubDomain:    string(req.SubDomain),
		IsPrivate:    req.IsPrivate,
		Status:       "published",
		ReviewStatus: reviewStatus,
		ReviewReason: reviewReason,
		QualityScore: qualityScore,
		ScoreDetails: scoreDetails,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if req.Interpretation != "" {
		exp.Interpretation = &req.Interpretation
	}

	err := r.db.QueryRow(ctx,
		`INSERT INTO experiences (author_id, content, interpretation, domain, sub_domain, is_private,
		 review_status, review_reason, quality_score, score_details, status, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING id`,
		exp.AuthorID, exp.Content, exp.Interpretation, exp.Domain, exp.SubDomain, exp.IsPrivate,
		exp.ReviewStatus, exp.ReviewReason, exp.QualityScore, exp.ScoreDetails,
		exp.Status, exp.CreatedAt, exp.UpdatedAt,
	).Scan(&exp.ID)
	if err != nil {
		return nil, fmt.Errorf("insert experience with review: %w", err)
	}

	return exp, nil
}

func (r *ExperienceRepo) GetByID(ctx context.Context, id string, viewerID string) (*model.Experience, error) {
	if viewerID == "" {
		viewerID = nilUUID
	}
	row := r.db.QueryRow(ctx,
		`SELECT e.id, e.author_id, e.content, e.interpretation, e.domain, e.is_official,
		        e.source_label, e.like_count, e.bookmark_count, e.interpretation_generated,
		        e.status, e.created_at, e.updated_at,
		        u.nickname, u.avatar_url,
		        EXISTS(SELECT 1 FROM likes WHERE user_id=$2 AND experience_id=e.id) as is_liked,
		        EXISTS(SELECT 1 FROM bookmarks WHERE user_id=$2 AND experience_id=e.id) as is_bookmarked
		 FROM experiences e
		 LEFT JOIN users u ON u.id = e.author_id
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

	var conditions []string
	var args []interface{}
	idx := 1

	conditions = append(conditions, "e.status = 'published' AND e.review_status = 'approved' AND e.is_private = FALSE")
	if query.Domain != "" {
		conditions = append(conditions, fmt.Sprintf("e.domain = $%d", idx))
		args = append(args, string(query.Domain))
		idx++
	}
	if query.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(e.content ILIKE $%d OR e.interpretation ILIKE $%d)", idx, idx+1))
		args = append(args, "%"+query.Search+"%", "%"+query.Search+"%")
		idx += 2
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM experiences e WHERE %s", whereClause)
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count: %w", err)
	}

	// Order
	orderClause := "e.created_at DESC"
	if query.Sort == "popular" {
		orderClause = "e.like_count DESC, e.created_at DESC"
	}

	selectQuery := fmt.Sprintf(
		`SELECT e.id, e.author_id, e.content, e.interpretation, e.domain, e.is_official,
		        e.source_label, e.like_count, e.bookmark_count, e.interpretation_generated,
		        e.status, e.created_at, e.updated_at,
		        u.nickname, u.avatar_url,
		        EXISTS(SELECT 1 FROM likes WHERE user_id=$%d AND experience_id=e.id) as is_liked,
		        EXISTS(SELECT 1 FROM bookmarks WHERE user_id=$%d AND experience_id=e.id) as is_bookmarked
		 FROM experiences e
		 LEFT JOIN users u ON u.id = e.author_id
		 WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		idx, idx+1, whereClause, orderClause, idx+2, idx+3,
	)

	if viewerID == "" {
		viewerID = nilUUID
	}
	args = append(args, viewerID, viewerID, query.PageSize, (query.Page-1)*query.PageSize)

	rows, err := r.db.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list: %w", err)
	}
	defer rows.Close()

	var experiences []model.Experience
	for rows.Next() {
		var e model.Experience
		if err := rows.Scan(
			&e.ID, &e.AuthorID, &e.Content, &e.Interpretation, &e.Domain,
			&e.IsOfficial, &e.SourceLabel, &e.LikeCount, &e.BookmarkCount,
			&e.InterpretationGenerated, &e.Status, &e.CreatedAt, &e.UpdatedAt,
			&e.AuthorName, &e.AuthorAvatar, &e.IsLiked, &e.IsBookmarked,
		); err != nil {
			return nil, 0, fmt.Errorf("scan: %w", err)
		}
		experiences = append(experiences, e)
	}

	return experiences, total, nil
}

// TODO: 推荐系统 — 需要 pgvector 扩展 + embedding 列，当前不可用
func (r *ExperienceRepo) SearchByEmbedding(ctx context.Context, embedding []float32, userID string, limit int) ([]model.Experience, error) {
	rows, err := r.db.Query(ctx,
		`SELECT e.id, e.content, e.domain, e.like_count, u.nickname as author_name
		 FROM experiences e
		 LEFT JOIN users u ON u.id = e.author_id
		 WHERE e.status = 'published' AND e.author_id = $1
		 ORDER BY e.embedding <=> $2 LIMIT $3`,
		userID, embedding, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("search embedding: %w", err)
	}
	defer rows.Close()

	var experiences []model.Experience
	for rows.Next() {
		var e model.Experience
		if err := rows.Scan(&e.ID, &e.Content, &e.Domain, &e.LikeCount, &e.AuthorName); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		experiences = append(experiences, e)
	}

	return experiences, nil
}

// Recommend returns experiences the user hasn't seen,
// ranked by domain preference × hotness (like_count + bookmark_count).
// Domain preference = published_count×2 + bookmarked_count×1.
// Falls back to pure hotness for users without history.
func (r *ExperienceRepo) Recommend(ctx context.Context, userID string, limit int) ([]model.Experience, error) {
	if limit < 1 || limit > 50 {
		limit = 20
	}

	query := `
		WITH user_domain_stats AS (
			SELECT domain, COUNT(*) * 2 AS weight
			FROM experiences
			WHERE author_id = $1 AND status = 'published'
			GROUP BY domain
			UNION ALL
			SELECT e.domain, COUNT(*) AS weight
			FROM bookmarks b
			JOIN experiences e ON e.id = b.experience_id
			WHERE b.user_id = $1 AND e.status = 'published'
			GROUP BY e.domain
		),
		domain_scores AS (
			SELECT domain, SUM(weight) AS score FROM user_domain_stats GROUP BY domain
		)
		SELECT e.id, e.author_id, e.content, e.interpretation, e.domain, e.is_official,
		       e.source_label, e.like_count, e.bookmark_count, e.interpretation_generated,
		       e.status, e.created_at, e.updated_at,
		       u.nickname, u.avatar_url,
		       EXISTS(SELECT 1 FROM likes WHERE user_id = $1 AND experience_id = e.id) AS is_liked,
		       EXISTS(SELECT 1 FROM bookmarks WHERE user_id = $1 AND experience_id = e.id) AS is_bookmarked,
		       COALESCE(ds.score, 1) * (e.like_count + e.bookmark_count + 1) AS rec_score
		FROM experiences e
		LEFT JOIN users u ON u.id = e.author_id
		LEFT JOIN domain_scores ds ON ds.domain = e.domain
		WHERE e.status = 'published' AND e.review_status = 'approved' AND e.is_private = FALSE
		  AND e.author_id != $1
		  AND e.id NOT IN (SELECT experience_id FROM bookmarks WHERE user_id = $1)
		ORDER BY rec_score DESC
		LIMIT $2`

	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("recommend: %w", err)
	}
	defer rows.Close()

	var experiences []model.Experience
	for rows.Next() {
		var e model.Experience
		var recScore float64
		if err := rows.Scan(
			&e.ID, &e.AuthorID, &e.Content, &e.Interpretation, &e.Domain,
			&e.IsOfficial, &e.SourceLabel, &e.LikeCount, &e.BookmarkCount,
			&e.InterpretationGenerated, &e.Status, &e.CreatedAt, &e.UpdatedAt,
			&e.AuthorName, &e.AuthorAvatar, &e.IsLiked, &e.IsBookmarked,
			&recScore,
		); err != nil {
			return nil, fmt.Errorf("recommend scan: %w", err)
		}
		experiences = append(experiences, e)
	}

	return experiences, nil
}

func (r *ExperienceRepo) Update(ctx context.Context, id, authorID string, req model.CreateExperienceRequest) error {
	result, err := r.db.Exec(ctx,
		`UPDATE experiences SET content=$1, interpretation=$2, domain=$3, updated_at=NOW()
		 WHERE id=$4 AND author_id=$5`,
		req.Content, req.Interpretation, req.Domain, id, authorID,
	)
	if err != nil {
		return fmt.Errorf("update experience: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("experience not found or permission denied")
	}
	return nil
}

func (r *ExperienceRepo) Delete(ctx context.Context, id, authorID string) error {
	result, err := r.db.Exec(ctx, `DELETE FROM experiences WHERE id=$1 AND author_id=$2`, id, authorID)
	if err != nil {
		return fmt.Errorf("delete experience: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("experience not found or permission denied")
	}
	return nil
}

// ListByAuthor — 用户自己发布的经验
func (r *ExperienceRepo) ListByAuthor(ctx context.Context, authorID string, page, pageSize int) ([]model.Experience, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	var total int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM experiences WHERE author_id=$1 AND status='published'`, authorID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count by author: %w", err)
	}

	rows, err := r.db.Query(ctx,
		`SELECT e.id, e.author_id, e.content, e.interpretation, e.domain, e.is_official,
		        e.source_label, e.like_count, e.bookmark_count, e.interpretation_generated,
		        e.status, e.created_at, e.updated_at,
		        u.nickname, u.avatar_url,
		        EXISTS(SELECT 1 FROM likes WHERE user_id=$1 AND experience_id=e.id) as is_liked,
		        EXISTS(SELECT 1 FROM bookmarks WHERE user_id=$1 AND experience_id=e.id) as is_bookmarked
		 FROM experiences e
		 LEFT JOIN users u ON u.id = e.author_id
		 WHERE e.author_id=$1 AND e.status='published'
		 ORDER BY e.created_at DESC LIMIT $2 OFFSET $3`,
		authorID, pageSize, (page-1)*pageSize,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list by author: %w", err)
	}
	defer rows.Close()

	return scanExperiences(rows, total)
}

// ListBookmarked — 用户收藏的经验
func (r *ExperienceRepo) ListBookmarked(ctx context.Context, userID string, page, pageSize int) ([]model.Experience, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	var total int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM bookmarks b
		 JOIN experiences e ON e.id = b.experience_id
		 WHERE b.user_id=$1 AND e.status='published' AND e.review_status IN ('approved', 'private')`, userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count bookmarks: %w", err)
	}

	rows, err := r.db.Query(ctx,
		`SELECT e.id, e.author_id, e.content, e.interpretation, e.domain, e.is_official,
		        e.source_label, e.like_count, e.bookmark_count, e.interpretation_generated,
		        e.status, e.created_at, e.updated_at,
		        u.nickname, u.avatar_url,
		        EXISTS(SELECT 1 FROM likes WHERE user_id=$1 AND experience_id=e.id) as is_liked,
		        true as is_bookmarked
		 FROM bookmarks b
		 JOIN experiences e ON e.id = b.experience_id
		 LEFT JOIN users u ON u.id = e.author_id
		 WHERE b.user_id=$1 AND e.status='published' AND e.review_status IN ('approved', 'private')
		 ORDER BY b.created_at DESC LIMIT $2 OFFSET $3`,
		userID, pageSize, (page-1)*pageSize,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list bookmarks: %w", err)
	}
	defer rows.Close()

	return scanExperiences(rows, total)
}

func scanExperiences(rows pgx.Rows, total int) ([]model.Experience, int, error) {
	var experiences []model.Experience
	for rows.Next() {
		var e model.Experience
		if err := rows.Scan(
			&e.ID, &e.AuthorID, &e.Content, &e.Interpretation, &e.Domain,
			&e.IsOfficial, &e.SourceLabel, &e.LikeCount, &e.BookmarkCount,
			&e.InterpretationGenerated, &e.Status, &e.CreatedAt, &e.UpdatedAt,
			&e.AuthorName, &e.AuthorAvatar, &e.IsLiked, &e.IsBookmarked,
		); err != nil {
			return nil, 0, fmt.Errorf("scan: %w", err)
		}
		experiences = append(experiences, e)
	}
	return experiences, total, nil
}
