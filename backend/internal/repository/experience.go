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
		SubDomain:    strPtr(string(req.SubDomain)),
		IsPrivate:    req.IsPrivate,
		SourceType:   "user",
		Status:       "published",
		ReviewStatus: string(model.ReviewPrivate),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if req.Interpretation != "" {
		exp.Interpretation = &req.Interpretation
	}

	err := r.db.QueryRow(ctx,
		`INSERT INTO experiences (author_id, content, interpretation, domain, sub_domain, is_private, source_type,
		 review_status, status, original_text, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING id`,
		exp.AuthorID, exp.Content, exp.Interpretation, exp.Domain, exp.SubDomain, exp.IsPrivate, exp.SourceType,
		exp.ReviewStatus, exp.Status, exp.OriginalText, exp.CreatedAt, exp.UpdatedAt,
	).Scan(&exp.ID)
	if err != nil {
		return nil, fmt.Errorf("insert experience: %w", err)
	}

	return exp, nil
}

// CreateWithReview creates an experience with review fields.
// originalText is set when the content is a translation (e.g., classical→modern Chinese).
func (r *ExperienceRepo) CreateWithReview(ctx context.Context, authorID string, req model.CreateExperienceRequest, reviewStatus string, reviewReason *string, qualityScore *float64, scoreDetails *string, originalText *string) (*model.Experience, error) {
	exp := &model.Experience{
		AuthorID:     authorID,
		Content:      req.Content,
		Domain:       req.Domain,
		SubDomain:    strPtr(string(req.SubDomain)),
		IsPrivate:    req.IsPrivate,
		SourceType:   "user",
		Status:       "published",
		ReviewStatus: reviewStatus,
		ReviewReason: reviewReason,
		QualityScore: qualityScore,
		ScoreDetails: scoreDetails,
		OriginalText: originalText,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if req.Interpretation != "" {
		exp.Interpretation = &req.Interpretation
	}

	err := r.db.QueryRow(ctx,
		`INSERT INTO experiences (author_id, content, interpretation, domain, sub_domain, is_private, source_type,
		 review_status, review_reason, quality_score, score_details, status, original_text, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) RETURNING id`,
		exp.AuthorID, exp.Content, exp.Interpretation, exp.Domain, exp.SubDomain, exp.IsPrivate, exp.SourceType,
		exp.ReviewStatus, exp.ReviewReason, exp.QualityScore, exp.ScoreDetails,
		exp.Status, exp.OriginalText, exp.CreatedAt, exp.UpdatedAt,
	).Scan(&exp.ID)
	if err != nil {
		return nil, fmt.Errorf("insert experience with review: %w", err)
	}

	return exp, nil
}

// CreateOfficial inserts a platform-generated experience with full metadata.
func (r *ExperienceRepo) CreateOfficial(ctx context.Context, authorID, content, interpretation, domain, subDomain, creatorName, sourceLabel, scoreReason string, qualityScore float64) (*model.Experience, error) {
	exp := &model.Experience{
		AuthorID:                authorID,
		Content:                 content,
		Domain:                  model.Domain(domain),
		SubDomain:               &subDomain,
		IsOfficial:              true,
		SourceType:              "platform",
		SourceLabel:             &sourceLabel,
		CreatorName:             &creatorName,
		Status:                  "published",
		ReviewStatus:            string(model.ReviewApproved),
		QualityScore:            &qualityScore,
		ScoreReason:             &scoreReason,
		InterpretationGenerated: true,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}

	if interpretation != "" {
		exp.Interpretation = &interpretation
	}

	err := r.db.QueryRow(ctx,
		`INSERT INTO experiences (author_id, content, interpretation, domain, sub_domain,
		 is_official, source_type, source_label, creator_name, score_reason,
		 review_status, quality_score, interpretation_generated, status, original_text, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17) RETURNING id`,
		exp.AuthorID, exp.Content, exp.Interpretation, exp.Domain, exp.SubDomain,
		exp.IsOfficial, exp.SourceType, exp.SourceLabel, exp.CreatorName, exp.ScoreReason,
		exp.ReviewStatus, exp.QualityScore, exp.InterpretationGenerated,
		exp.Status, exp.OriginalText, exp.CreatedAt, exp.UpdatedAt,
	).Scan(&exp.ID)
	if err != nil {
		return nil, fmt.Errorf("insert official: %w", err)
	}

	return exp, nil
}

const experienceSelectCols = `e.id, e.author_id, e.content, e.interpretation, e.domain, e.sub_domain, e.is_private, e.review_status, e.review_reason, e.quality_score, e.score_details, e.is_official,
		e.source_label, e.like_count, e.bookmark_count, e.interpretation_generated,
		e.creator_name, e.source_type, e.score_reason, e.original_text,
		e.status, e.created_at, e.updated_at, e.random_sort,
		u.nickname, u.avatar_url, u.title as author_title`

const experienceLikedBookmark = `EXISTS(SELECT 1 FROM likes WHERE user_id=$2 AND experience_id=e.id) as is_liked,
		EXISTS(SELECT 1 FROM bookmarks WHERE user_id=$2 AND experience_id=e.id) as is_bookmarked`

func scanExperience(row pgx.Row, e *model.Experience) error {
	return row.Scan(
		&e.ID, &e.AuthorID, &e.Content, &e.Interpretation, &e.Domain,
		&e.SubDomain, &e.IsPrivate, &e.ReviewStatus, &e.ReviewReason, &e.QualityScore, &e.ScoreDetails,
		&e.IsOfficial, &e.SourceLabel, &e.LikeCount, &e.BookmarkCount,
		&e.InterpretationGenerated, &e.CreatorName, &e.SourceType, &e.ScoreReason, &e.OriginalText,
		&e.Status, &e.CreatedAt, &e.UpdatedAt, &e.RandomSort,
		&e.AuthorName, &e.AuthorAvatar, &e.AuthorTitle, &e.IsLiked, &e.IsBookmarked,
	)
}

func (r *ExperienceRepo) GetByID(ctx context.Context, id string, viewerID string) (*model.Experience, error) {
	if viewerID == "" {
		viewerID = nilUUID
	}
	query := fmt.Sprintf(`SELECT %s, %s FROM experiences e
		 LEFT JOIN users u ON u.id = e.author_id
		WHERE e.id = $1`, experienceSelectCols, experienceLikedBookmark)

	exp := &model.Experience{}
	err := scanExperience(r.db.QueryRow(ctx, query, id, viewerID), exp)
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

	conditions = append(conditions, "e.status = 'published' AND e.review_status = 'approved' AND e.is_private = FALSE AND e.deleted_at IS NULL")
	if query.Domain != "" {
		conditions = append(conditions, fmt.Sprintf("e.domain = $%d", idx))
		args = append(args, string(query.Domain))
		idx++
	}
	if query.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(e.content ILIKE $%d OR e.creator_name ILIKE $%d OR u.nickname ILIKE $%d)", idx, idx+1, idx+2))
		args = append(args, "%"+query.Search+"%", "%"+query.Search+"%", "%"+query.Search+"%")
		idx += 3
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM experiences e LEFT JOIN users u ON u.id = e.author_id WHERE %s", whereClause)
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count: %w", err)
	}

	// Order
	orderClause := "e.created_at DESC"
	if query.Sort == "popular" {
		orderClause = "e.like_count DESC, e.created_at DESC"
	}

	selectQuery := fmt.Sprintf(
		`SELECT %s, %s
		 FROM experiences e
		 LEFT JOIN users u ON u.id = e.author_id
		 WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		experienceSelectCols, fmt.Sprintf(`EXISTS(SELECT 1 FROM likes WHERE user_id=$%d AND experience_id=e.id) as is_liked,
		EXISTS(SELECT 1 FROM bookmarks WHERE user_id=$%d AND experience_id=e.id) as is_bookmarked`, idx, idx+1),
		whereClause, orderClause, idx+2, idx+3,
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
			&e.SubDomain, &e.IsPrivate, &e.ReviewStatus, &e.ReviewReason, &e.QualityScore, &e.ScoreDetails,
			&e.IsOfficial, &e.SourceLabel, &e.LikeCount, &e.BookmarkCount,
			&e.InterpretationGenerated, &e.CreatorName, &e.SourceType, &e.ScoreReason, &e.OriginalText,
			&e.Status, &e.CreatedAt, &e.UpdatedAt, &e.RandomSort,
				&e.AuthorName, &e.AuthorAvatar, &e.AuthorTitle, &e.IsLiked, &e.IsBookmarked,
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

// Recommend returns personalized recommendations.
// - bookmarks < 100: purely random order (discovery mode)
// - bookmarks >= 100: domain preference + unseen priority + random_sort
func (r *ExperienceRepo) Recommend(ctx context.Context, userID string, limit, offset int) ([]model.Experience, error) {
	if limit < 1 || limit > 200 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	// Count user bookmarks to decide strategy
	var bookmarkCount int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM bookmarks WHERE user_id = $1`, userID,
	).Scan(&bookmarkCount)
	if err != nil {
		bookmarkCount = 0
	}

	var query string
	var args []interface{}

	baseSelect := `SELECT e.id, e.author_id, e.content, e.interpretation, e.domain, e.sub_domain, e.is_private, e.review_status, e.review_reason, e.quality_score, e.score_details, e.is_official,
		e.source_label, e.like_count, e.bookmark_count, e.interpretation_generated,
		e.creator_name, e.source_type, e.score_reason, e.original_text,
		e.status, e.created_at, e.updated_at, e.random_sort,
		u.nickname, u.avatar_url, u.title as author_title,
		EXISTS(SELECT 1 FROM likes WHERE user_id = $1 AND experience_id = e.id) AS is_liked,
		EXISTS(SELECT 1 FROM bookmarks WHERE user_id = $1 AND experience_id = e.id) AS is_bookmarked`

	baseFrom := `FROM experiences e
		LEFT JOIN users u ON u.id = e.author_id`

	baseWhere := `WHERE e.status = 'published' AND e.review_status = 'approved' AND e.is_private = FALSE AND e.deleted_at IS NULL
		AND e.author_id != $1`

	if bookmarkCount < 100 {
		// Pure random discovery
		query = baseSelect + " " + baseFrom + " " + baseWhere + `
		ORDER BY e.random_sort
		LIMIT $2 OFFSET $3`
		args = []interface{}{userID, limit, offset}
	} else {
		// Domain preference + unseen priority + random_sort
		query = baseSelect + " " + baseFrom + " " + baseWhere + `
		ORDER BY
			CASE WHEN e.id NOT IN (SELECT experience_id FROM user_views WHERE user_id = $1) THEN 0 ELSE 1 END,
			e.random_sort
		LIMIT $2 OFFSET $3`
		args = []interface{}{userID, limit, offset}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("recommend: %w", err)
	}
	defer rows.Close()

	var experiences []model.Experience
	for rows.Next() {
		var e model.Experience
		if err := rows.Scan(
			&e.ID, &e.AuthorID, &e.Content, &e.Interpretation, &e.Domain,
			&e.SubDomain, &e.IsPrivate, &e.ReviewStatus, &e.ReviewReason, &e.QualityScore, &e.ScoreDetails,
			&e.IsOfficial, &e.SourceLabel, &e.LikeCount, &e.BookmarkCount,
			&e.InterpretationGenerated, &e.CreatorName, &e.SourceType, &e.ScoreReason, &e.OriginalText,
			&e.Status, &e.CreatedAt, &e.UpdatedAt, &e.RandomSort,
			&e.AuthorName, &e.AuthorAvatar, &e.AuthorTitle, &e.IsLiked, &e.IsBookmarked,
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
	// First verify the experience exists and is owned by the user, and check review_status
	var reviewStatus string
	err := r.db.QueryRow(ctx,
		`SELECT review_status FROM experiences WHERE id=$1 AND author_id=$2 AND deleted_at IS NULL`,
		id, authorID,
	).Scan(&reviewStatus)
	if err != nil {
		return fmt.Errorf("experience not found or permission denied")
	}

	if reviewStatus == "approved" {
		// Soft-delete: approved experiences keep their row (stay in public pool logically)
		_, err = r.db.Exec(ctx,
			`UPDATE experiences SET deleted_at = NOW() WHERE id=$1 AND author_id=$2`,
			id, authorID)
		if err != nil {
			return fmt.Errorf("soft-delete experience: %w", err)
		}
	} else {
		// Hard-delete: the experience wasn't in the public pool anyway
		_, err = r.db.Exec(ctx,
			`DELETE FROM experiences WHERE id=$1 AND author_id=$2`,
			id, authorID)
		if err != nil {
			return fmt.Errorf("delete experience: %w", err)
		}
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
		`SELECT COUNT(*) FROM experiences WHERE author_id=$1 AND status='published' AND deleted_at IS NULL`, authorID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count by author: %w", err)
	}

	rows, err := r.db.Query(ctx,
		`SELECT e.id, e.author_id, e.content, e.interpretation, e.domain, e.sub_domain, e.is_private, e.review_status, e.review_reason, e.quality_score, e.score_details, e.is_official,
		        e.source_label, e.like_count, e.bookmark_count, e.interpretation_generated,
		        e.creator_name, e.source_type, e.score_reason, e.original_text,
		e.status, e.created_at, e.updated_at, e.random_sort,
		        u.nickname, u.avatar_url, u.title as author_title,
		        EXISTS(SELECT 1 FROM likes WHERE user_id=$1 AND experience_id=e.id) as is_liked,
		        EXISTS(SELECT 1 FROM bookmarks WHERE user_id=$1 AND experience_id=e.id) as is_bookmarked
		 FROM experiences e
		 LEFT JOIN users u ON u.id = e.author_id
		 WHERE e.author_id=$1 AND e.status='published' AND e.deleted_at IS NULL
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
		 WHERE b.user_id=$1 AND e.status='published' AND e.review_status IN ('approved', 'private') AND e.deleted_at IS NULL`, userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count bookmarks: %w", err)
	}

	rows, err := r.db.Query(ctx,
		`SELECT e.id, e.author_id, e.content, e.interpretation, e.domain, e.sub_domain, e.is_private, e.review_status, e.review_reason, e.quality_score, e.score_details, e.is_official,
		        e.source_label, e.like_count, e.bookmark_count, e.interpretation_generated,
		        e.creator_name, e.source_type, e.score_reason, e.original_text,
		e.status, e.created_at, e.updated_at, e.random_sort,
		        u.nickname, u.avatar_url, u.title as author_title,
		        EXISTS(SELECT 1 FROM likes WHERE user_id=$1 AND experience_id=e.id) as is_liked,
		        true as is_bookmarked
		 FROM bookmarks b
		 JOIN experiences e ON e.id = b.experience_id
		 LEFT JOIN users u ON u.id = e.author_id
		 WHERE b.user_id=$1 AND e.status='published' AND e.review_status IN ('approved', 'private') AND e.deleted_at IS NULL
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
			&e.SubDomain, &e.IsPrivate, &e.ReviewStatus, &e.ReviewReason, &e.QualityScore, &e.ScoreDetails,
			&e.IsOfficial, &e.SourceLabel, &e.LikeCount, &e.BookmarkCount,
			&e.InterpretationGenerated, &e.CreatorName, &e.SourceType, &e.ScoreReason, &e.OriginalText,
			&e.Status, &e.CreatedAt, &e.UpdatedAt, &e.RandomSort,
				&e.AuthorName, &e.AuthorAvatar, &e.AuthorTitle, &e.IsLiked, &e.IsBookmarked,
		); err != nil {
			return nil, 0, fmt.Errorf("scan: %w", err)
		}
		experiences = append(experiences, e)
	}
	return experiences, total, nil
}

func strPtr(s string) *string { return &s }
