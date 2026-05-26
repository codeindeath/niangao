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

func (r *ExperienceRepo) GetUserDisplayName(ctx context.Context, userID string) (string, error) {
	var displayName string
	err := r.db.QueryRow(ctx,
		`SELECT COALESCE(NULLIF(TRIM(display_name), ''), NULLIF(TRIM(nickname), ''), '')
		 FROM users WHERE id=$1`,
		userID,
	).Scan(&displayName)
	if err != nil {
		return "", fmt.Errorf("get user display name: %w", err)
	}
	return displayName, nil
}

func (r *ExperienceRepo) Create(ctx context.Context, authorID string, req model.CreateExperienceRequest) (*model.Experience, error) {
	exp := &model.Experience{
		AuthorID:     authorID,
		Content:      req.Content,
		Domain:       req.Domain,
		SubDomain:    strPtrNilIfEmpty(string(req.SubDomain)),
		Topics:       req.Topics,
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
		`INSERT INTO experiences (author_id, content, interpretation, domain, sub_domain, topics, is_private, source_type,
		 review_status, status, original_text, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING id`,
		exp.AuthorID, exp.Content, exp.Interpretation, exp.Domain, exp.SubDomain, exp.Topics, exp.IsPrivate, exp.SourceType,
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
	qualityTier := string(model.QualityTierUnreviewed)
	if req.IsPrivate {
		qualityTier = string(model.QualityTierPrivateOnly)
	}
	interpretationStatus := string(model.InterpretationNone)
	if req.Interpretation != "" {
		interpretationStatus = string(model.InterpretationReady)
	}

	exp := &model.Experience{
		AuthorID:                  authorID,
		OwnerUserID:               strPtr(authorID),
		Content:                   req.Content,
		Domain:                    req.Domain,
		SubDomain:                 strPtrNilIfEmpty(string(req.SubDomain)),
		Topics:                    req.Topics,
		Topic:                     req.Topic,
		IsPrivate:                 req.IsPrivate,
		ExperienceType:            string(model.ExperienceTypeUserOriginal),
		Visibility:                string(req.Visibility),
		LifecycleStatus:           string(model.LifecycleActive),
		SourceType:                "user",
		SourceScene:               req.SourceScene,
		SourceChatTopicID:         strPtrNilIfEmpty(req.SourceChatTopicID),
		SourceChatMessageID:       strPtrNilIfEmpty(req.SourceChatMessageID),
		SourceChatMessageSnapshot: strPtrNilIfEmpty(req.SourceChatMessageSnapshot),
		Status:                    "published",
		ReviewStatus:              reviewStatus,
		ReviewReason:              reviewReason,
		QualityScore:              qualityScore,
		QualityTier:               qualityTier,
		ScoreDetails:              scoreDetails,
		OriginalText:              originalText,
		RecommendationStatus:      string(model.RecommendationIneligible),
		AICitable:                 false,
		InterpretationStatus:      interpretationStatus,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}

	if req.Interpretation != "" {
		exp.Interpretation = &req.Interpretation
		exp.InterpretationGenerated = true
	}

	err := r.db.QueryRow(ctx,
		`INSERT INTO experiences (author_id, content, interpretation, domain, sub_domain, topics, is_private, source_type,
		 review_status, review_reason, quality_score, score_details, status, original_text, interpretation_generated, created_at, updated_at,
		 owner_user_id, creator_display_name, experience_type, visibility, lifecycle_status, source_scene, topic, quality_tier,
		 recommendation_status, ai_citable, inspiration_count, collection_count, interpretation_status,
		 source_chat_topic_id, source_chat_message_id, source_chat_message_snapshot)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,
		 $18, COALESCE((SELECT display_name FROM users WHERE id=$1), (SELECT nickname FROM users WHERE id=$1), ''),
		 $19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32) RETURNING id`,
		exp.AuthorID, exp.Content, exp.Interpretation, exp.Domain, exp.SubDomain, exp.Topics, exp.IsPrivate, exp.SourceType,
		exp.ReviewStatus, exp.ReviewReason, exp.QualityScore, exp.ScoreDetails,
		exp.Status, exp.OriginalText, exp.InterpretationGenerated, exp.CreatedAt, exp.UpdatedAt,
		exp.OwnerUserID, exp.ExperienceType, exp.Visibility, exp.LifecycleStatus, exp.SourceScene, exp.Topic, exp.QualityTier,
		exp.RecommendationStatus, exp.AICitable, exp.InspirationCount, exp.CollectionCount, exp.InterpretationStatus,
		exp.SourceChatTopicID, exp.SourceChatMessageID, exp.SourceChatMessageSnapshot,
	).Scan(&exp.ID)
	if err != nil {
		return nil, fmt.Errorf("insert experience with review: %w", err)
	}

	return exp, nil
}

// ExistsByContent checks if an experience with the same content already exists (not deleted).
func (r *ExperienceRepo) ExistsByContent(ctx context.Context, content string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM experiences WHERE content = $1 AND deleted_at IS NULL)`,
		content,
	).Scan(&exists)
	return exists, err
}

// ExistsByContentExcluding checks for duplicate content, excluding a specific experience ID.
// Used when an experience is already saved and we want to check for OTHER duplicates.
func (r *ExperienceRepo) ExistsByContentExcluding(ctx context.Context, content string, excludeID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM experiences WHERE content = $1 AND id != $2 AND deleted_at IS NULL)`,
		content, excludeID,
	).Scan(&exists)
	return exists, err
}

// UpdateReviewResult updates an existing experience with review results
// (after pipeline: normalize → dedup → hard_policy → AI review → translate → interpret).
func (r *ExperienceRepo) UpdateReviewResult(ctx context.Context, id string, content string, interpretation *string, originalText *string, score *float64, scoreDetails *string, reviewStatus string, reviewReason *string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE experiences SET content=$1, interpretation=COALESCE($2, interpretation),
		 original_text=COALESCE($3, original_text), quality_score=COALESCE($4, quality_score),
		 score_details=COALESCE($5, score_details), review_status=$6, review_reason=$7,
		 interpretation_generated=CASE WHEN $2 IS NOT NULL THEN TRUE ELSE interpretation_generated END,
		 updated_at=NOW() WHERE id=$8`,
		content, interpretation, originalText, score, scoreDetails, reviewStatus, reviewReason, id)
	if err != nil {
		return fmt.Errorf("update review result: %w", err)
	}
	return nil
}

const experienceSelectCols = `e.id, e.author_id, COALESCE(e.owner_user_id, e.author_id), e.content, e.interpretation, e.domain, e.sub_domain, COALESCE(e.topics, ''), e.is_private, e.review_status, e.review_reason, e.quality_score, e.score_details, e.is_official,
		e.source_label, COALESCE(e.inspiration_count, e.like_count), COALESCE(e.collection_count, e.bookmark_count), e.like_count, e.bookmark_count, e.interpretation_generated,
		e.creator_name, e.creator_display_name, e.source_type,
		COALESCE(e.experience_type, CASE WHEN e.is_official THEN 'platform_selected' ELSE 'user_original' END),
		COALESCE(e.visibility, CASE WHEN e.is_private THEN 'private' ELSE 'public' END),
		COALESCE(e.lifecycle_status, CASE WHEN e.deleted_at IS NOT NULL THEN 'deleted' WHEN e.review_status = 'pending' THEN 'needs_review' ELSE 'active' END),
		COALESCE(e.source_scene, ''), COALESCE(e.topic, e.topics, ''),
		COALESCE(e.quality_tier, ''), COALESCE(e.recommendation_status, ''), COALESCE(e.ai_citable, FALSE),
		COALESCE(e.interpretation_status, ''), e.score_reason, e.original_text,
		e.status, e.created_at, e.updated_at, e.random_sort,
		u.nickname, u.avatar_url, u.title as author_title`

const experienceLikedBookmark = `EXISTS(SELECT 1 FROM likes WHERE user_id=$2 AND experience_id=e.id) as is_liked,
		EXISTS(SELECT 1 FROM bookmarks WHERE user_id=$2 AND experience_id=e.id) as is_bookmarked`

const updateExperienceQuery = `UPDATE experiences
		 SET content=$1,
		     interpretation=NULLIF($2, ''),
		     domain=$3,
		     sub_domain=NULLIF($4, ''),
		     is_private=$5,
		     topics=$6,
		     visibility=$7,
		     topic=$8,
		     review_status=$9,
		     review_reason=NULL,
		     quality_tier=$10,
		     lifecycle_status=$11,
		     recommendation_status='ineligible',
		     ai_citable=FALSE,
		     interpretation_status=CASE WHEN NULLIF($2, '') IS NULL THEN 'none' ELSE 'stale' END,
		     updated_at=NOW()
		 WHERE id=$12 AND COALESCE(owner_user_id, author_id)=$13`

func scanExperience(row pgx.Row, e *model.Experience) error {
	return row.Scan(
		&e.ID, &e.AuthorID, &e.OwnerUserID, &e.Content, &e.Interpretation, &e.Domain,
		&e.SubDomain, &e.Topics, &e.IsPrivate, &e.ReviewStatus, &e.ReviewReason, &e.QualityScore, &e.ScoreDetails,
		&e.IsOfficial, &e.SourceLabel, &e.InspirationCount, &e.CollectionCount, &e.LikeCount, &e.BookmarkCount,
		&e.InterpretationGenerated, &e.CreatorName, &e.CreatorDisplayName, &e.SourceType,
		&e.ExperienceType, &e.Visibility, &e.LifecycleStatus, &e.SourceScene, &e.Topic,
		&e.QualityTier, &e.RecommendationStatus, &e.AICitable, &e.InterpretationStatus, &e.ScoreReason, &e.OriginalText,
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
		WHERE e.id = $1 AND e.deleted_at IS NULL
		  AND ((e.status = 'published' AND e.review_status = 'approved' AND e.is_private = FALSE)
		       OR COALESCE(e.owner_user_id, e.author_id) = $2)`, experienceSelectCols, experienceLikedBookmark)

	exp := &model.Experience{}
	err := scanExperience(r.db.QueryRow(ctx, query, id, viewerID), exp)
	if err != nil {
		return nil, fmt.Errorf("get experience: %w", err)
	}
	return exp, nil
}

func (r *ExperienceRepo) GetByIDForAdmin(ctx context.Context, id string) (*model.Experience, error) {
	query := fmt.Sprintf(`SELECT %s, %s FROM experiences e
		 LEFT JOIN users u ON u.id = e.author_id
		WHERE e.id = $1`, experienceSelectCols, experienceLikedBookmark)

	exp := &model.Experience{}
	err := scanExperience(r.db.QueryRow(ctx, query, id, nilUUID), exp)
	if err != nil {
		return nil, fmt.Errorf("get admin experience: %w", err)
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
		if err := scanExperience(rows, &e); err != nil {
			return nil, 0, fmt.Errorf("scan: %w", err)
		}
		experiences = append(experiences, e)
	}

	return experiences, total, nil
}

func (r *ExperienceRepo) Update(ctx context.Context, id, authorID string, req model.CreateExperienceRequest) error {
	reviewStatus := string(model.ReviewPending)
	qualityTier := string(model.QualityTierUnreviewed)
	lifecycleStatus := updateLifecycleStatusForRequest(req.IsPrivate)
	if req.IsPrivate {
		reviewStatus = string(model.ReviewPrivate)
		qualityTier = string(model.QualityTierPrivateOnly)
	}

	result, err := r.db.Exec(ctx,
		updateExperienceQuery,
		req.Content, req.Interpretation, req.Domain, string(req.SubDomain), req.IsPrivate, req.Topics,
		string(req.Visibility), req.Topic, reviewStatus, qualityTier, lifecycleStatus, id, authorID,
	)
	if err != nil {
		return fmt.Errorf("update experience: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("experience not found or permission denied")
	}
	return nil
}

func updateLifecycleStatusForRequest(isPrivate bool) string {
	if isPrivate {
		return string(model.LifecycleActive)
	}
	return string(model.LifecycleNeedsReview)
}

func (r *ExperienceRepo) Delete(ctx context.Context, id, authorID string) error {
	// First verify the experience exists and is owned by the user, and check review_status
	var reviewStatus string
	err := r.db.QueryRow(ctx,
		`SELECT review_status FROM experiences WHERE id=$1 AND COALESCE(owner_user_id, author_id)=$2 AND deleted_at IS NULL`,
		id, authorID,
	).Scan(&reviewStatus)
	if err != nil {
		return fmt.Errorf("experience not found or permission denied")
	}

	if reviewStatus == "approved" {
		// Soft-delete approved rows so historic references can stay auditable while V4 public surfaces suppress them.
		_, err = r.db.Exec(ctx,
			`UPDATE experiences
			 SET deleted_at = NOW(),
			     lifecycle_status='deleted',
			     recommendation_status='suppressed',
			     ai_citable=FALSE,
			     updated_at=NOW()
			 WHERE id=$1 AND COALESCE(owner_user_id, author_id)=$2`,
			id, authorID)
		if err != nil {
			return fmt.Errorf("soft-delete experience: %w", err)
		}
	} else {
		// Hard-delete: the experience wasn't in the public pool anyway
		_, err = r.db.Exec(ctx,
			`DELETE FROM experiences WHERE id=$1 AND COALESCE(owner_user_id, author_id)=$2`,
			id, authorID)
		if err != nil {
			return fmt.Errorf("delete experience: %w", err)
		}
	}
	return nil
}

func strPtr(s string) *string { return &s }

func strPtrNilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
