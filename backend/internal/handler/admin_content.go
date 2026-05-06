package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niangao/backend/internal/model"
	"github.com/niangao/backend/internal/repository"
)

// RegisterAdminContentRoutes 注册管理后台 - 内容管理路由
func RegisterAdminContentRoutes(admin *gin.RouterGroup, expRepo *repository.ExperienceRepo, db *pgxpool.Pool) {
	r := admin.Group("/experiences")
	{
		r.GET("", func(c *gin.Context) {
			adminListExperiences(c, expRepo, db)
		})
		r.GET("/:id", func(c *gin.Context) {
			adminGetExperience(c, expRepo, db)
		})
		r.PUT("/:id", func(c *gin.Context) {
			adminUpdateExperience(c, expRepo, db)
		})
		r.DELETE("/:id", func(c *gin.Context) {
			adminDeleteExperience(c, expRepo, db)
		})
		r.POST("/:id/restore", func(c *gin.Context) {
			adminRestoreExperience(c, expRepo, db)
		})
		r.POST("/:id/unpublish", func(c *gin.Context) {
			adminUnpublishExperience(c, expRepo, db)
		})
		r.POST("/:id/hard-delete", func(c *gin.Context) {
			adminHardDeleteExperience(c, expRepo, db)
		})
		r.PUT("/:id/review-status", func(c *gin.Context) {
			adminUpdateReviewStatus(c, expRepo, db)
		})
		r.POST("/batch", func(c *gin.Context) {
			adminBatchExperience(c, expRepo, db)
		})
	}
}

// ============================================================
// Admin Experience List Query
// ============================================================

type adminExpListQuery struct {
	Domain       string `form:"domain"`
	SourceType   string `form:"source_type"`
	ReviewStatus string `form:"review_status"`
	Search       string `form:"search"`
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
}

const adminExpSelectCols = `e.id, e.author_id, e.content, e.interpretation, e.domain, e.sub_domain,
	e.is_private, e.review_status, e.review_reason, e.quality_score, e.score_details, e.is_official,
	e.source_label, e.like_count, e.bookmark_count, e.interpretation_generated,
	e.creator_name, e.source_type, e.score_reason,
	e.status, e.created_at, e.updated_at, e.deleted_at,
	COALESCE(u.nickname, '') as author_name,
	COALESCE(u.avatar_url, '') as author_avatar,
	COALESCE(u.title, '') as author_title`

func scanAdminExperience(row interface{ Scan(...interface{}) error }, e *model.Experience) error {
	return row.Scan(
		&e.ID, &e.AuthorID, &e.Content, &e.Interpretation, &e.Domain,
		&e.SubDomain, &e.IsPrivate, &e.ReviewStatus, &e.ReviewReason, &e.QualityScore, &e.ScoreDetails,
		&e.IsOfficial, &e.SourceLabel, &e.LikeCount, &e.BookmarkCount,
		&e.InterpretationGenerated, &e.CreatorName, &e.SourceType, &e.ScoreReason,
		&e.Status, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
		&e.AuthorName, &e.AuthorAvatar, &e.AuthorTitle,
	)
}

// adminListExperiences 管理员查看所有经验（含已删除、未审核）
func adminListExperiences(c *gin.Context, _ *repository.ExperienceRepo, db *pgxpool.Pool) {
	var q adminExpListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "查询参数错误"})
		return
	}

	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 50 {
		q.PageSize = 20
	}

	// 动态构建 WHERE
	var conditions []string
	var args []interface{}
	idx := 1

	// 管理员列表不过滤 status，允许查看所有
	conditions = append(conditions, "1=1")

	if q.Domain != "" {
		conditions = append(conditions, fmt.Sprintf("e.domain = $%d", idx))
		args = append(args, q.Domain)
		idx++
	}
	if q.SourceType != "" {
		conditions = append(conditions, fmt.Sprintf("e.source_type = $%d", idx))
		args = append(args, q.SourceType)
		idx++
	}
	if q.ReviewStatus != "" {
		conditions = append(conditions, fmt.Sprintf("e.review_status = $%d", idx))
		args = append(args, q.ReviewStatus)
		idx++
	}
	if q.Search != "" {
		conditions = append(conditions, fmt.Sprintf("e.content ILIKE $%d", idx))
		args = append(args, "%"+q.Search+"%")
		idx++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count
	var total int
	countQuery := fmt.Sprintf(
		"SELECT COUNT(*) FROM experiences e LEFT JOIN users u ON u.id = e.author_id WHERE %s",
		whereClause,
	)
	if err := db.QueryRow(c.Request.Context(), countQuery, args...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("count: %v", err)})
		return
	}

	// Select
	selectQuery := fmt.Sprintf(
		`SELECT %s FROM experiences e
		 LEFT JOIN users u ON u.id = e.author_id
		 WHERE %s
		 ORDER BY e.created_at DESC
		 LIMIT $%d OFFSET $%d`,
		adminExpSelectCols, whereClause, idx, idx+1,
	)
	args = append(args, q.PageSize, (q.Page-1)*q.PageSize)

	rows, err := db.Query(c.Request.Context(), selectQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("list: %v", err)})
		return
	}
	defer rows.Close()

	var experiences []model.Experience
	for rows.Next() {
		var e model.Experience
		if err := scanAdminExperience(rows, &e); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("scan: %v", err)})
			return
		}
		experiences = append(experiences, e)
	}

	c.JSON(http.StatusOK, model.PaginatedResponse{
		Data:     experiences,
		Total:    total,
		Page:     q.Page,
		PageSize: q.PageSize,
	})
}

// ============================================================
// Admin Get Experience Detail (含互动统计)
// ============================================================

type adminExpDetail struct {
	model.Experience
	LikeCountTotal     int `json:"like_count_total"`
	BookmarkCountTotal int `json:"bookmark_count_total"`
}

// adminGetExperience 获取经验完整详情，包含互动计数
func adminGetExperience(c *gin.Context, expRepo *repository.ExperienceRepo, db *pgxpool.Pool) {
	id := c.Param("id")

	// 复用 repo 获取基础数据（不带 viewerID，不查 is_liked/is_bookmarked）
	exp, err := expRepo.GetByID(c.Request.Context(), id, "")
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "experience not found"})
		return
	}

	detail := adminExpDetail{
		Experience:         *exp,
		LikeCountTotal:     exp.LikeCount,
		BookmarkCountTotal: exp.BookmarkCount,
	}

	// 可选：补充互动深度统计
	// 当前 like_count / bookmark_count 已在 Experience 中

	c.JSON(http.StatusOK, detail)
}

// ============================================================
// Admin Update Experience
// ============================================================

type adminUpdateExpRequest struct {
	Content     *string `json:"content"`
	Domain      *string `json:"domain"`
	SubDomain   *string `json:"sub_domain"`
	SourceType  *string `json:"source_type"`
	CreatorName *string `json:"creator_name"`
	ScoreReason *string `json:"score_reason"`
}

// adminUpdateExperience 管理员编辑经验内容/元数据
func adminUpdateExperience(c *gin.Context, _ *repository.ExperienceRepo, db *pgxpool.Pool) {
	id := c.Param("id")

	var req adminUpdateExpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	// 动态构建 SET
	var sets []string
	var args []interface{}
	idx := 1

	args = append(args, id) // $1
	idx++

	if req.Content != nil {
		sets = append(sets, fmt.Sprintf("content = $%d", idx))
		args = append(args, *req.Content)
		idx++
	}
	if req.Domain != nil {
		sets = append(sets, fmt.Sprintf("domain = $%d", idx))
		args = append(args, *req.Domain)
		idx++
	}
	if req.SubDomain != nil {
		sets = append(sets, fmt.Sprintf("sub_domain = $%d", idx))
		args = append(args, *req.SubDomain)
		idx++
	}
	if req.SourceType != nil {
		sets = append(sets, fmt.Sprintf("source_type = $%d", idx))
		args = append(args, *req.SourceType)
		idx++
	}
	if req.CreatorName != nil {
		sets = append(sets, fmt.Sprintf("creator_name = $%d", idx))
		args = append(args, *req.CreatorName)
		idx++
	}
	if req.ScoreReason != nil {
		sets = append(sets, fmt.Sprintf("score_reason = $%d", idx))
		args = append(args, *req.ScoreReason)
		idx++
	}

	if len(sets) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "至少需要提供一个要更新的字段"})
		return
	}

	sets = append(sets, "updated_at = NOW()")

	query := fmt.Sprintf(
		"UPDATE experiences SET %s WHERE id = $1",
		strings.Join(sets, ", "),
	)

	result, err := db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("update: %v", err)})
		return
	}
	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "experience not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ============================================================
// Admin Soft Delete
// ============================================================

// adminDeleteExperience 管理员软删除经验
func adminDeleteExperience(c *gin.Context, _ *repository.ExperienceRepo, db *pgxpool.Pool) {
	id := c.Param("id")

	result, err := db.Exec(c.Request.Context(),
		"UPDATE experiences SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL",
		id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("delete: %v", err)})
		return
	}
	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "experience not found or already deleted"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ============================================================
// Admin Restore
// ============================================================

// adminRestoreExperience 管理员恢复已软删除的经验
func adminRestoreExperience(c *gin.Context, _ *repository.ExperienceRepo, db *pgxpool.Pool) {
	id := c.Param("id")

	result, err := db.Exec(c.Request.Context(),
		"UPDATE experiences SET deleted_at = NULL, updated_at = NOW() WHERE id = $1 AND deleted_at IS NOT NULL",
		id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("restore: %v", err)})
		return
	}
	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "experience not found or not deleted"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ============================================================
// Admin Batch Operations
// ============================================================

type adminBatchRequest struct {
	IDs    []string `json:"ids" binding:"required"`
	Action string   `json:"action" binding:"required"` // "update_domain" | "delete"
	Domain *string  `json:"domain"`                    // for update_domain
}

// adminBatchExperience 批量更新领域或批量软删除
func adminBatchExperience(c *gin.Context, _ *repository.ExperienceRepo, db *pgxpool.Pool) {
	var req adminBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误：需提供 ids 和 action"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids 不能为空"})
		return
	}

	switch req.Action {
	case "update_domain":
		if req.Domain == nil || *req.Domain == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "update_domain 操作需要提供 domain 字段"})
			return
		}
		if !model.IsValidDomain(model.Domain(*req.Domain)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain"})
			return
		}
		// 批量更新领域
		result, err := db.Exec(c.Request.Context(),
			"UPDATE experiences SET domain = $1, updated_at = NOW() WHERE id = ANY($2)",
			*req.Domain, req.IDs,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("batch update_domain: %v", err)})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":         "ok",
			"action":         "update_domain",
			"domain":         *req.Domain,
			"affected_count": result.RowsAffected(),
		})

	case "delete":
		// 批量软删除
		result, err := db.Exec(c.Request.Context(),
			"UPDATE experiences SET deleted_at = NOW(), updated_at = NOW() WHERE id = ANY($1) AND deleted_at IS NULL",
			req.IDs,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("batch delete: %v", err)})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":         "ok",
			"action":         "delete",
			"affected_count": result.RowsAffected(),
		})

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的操作类型，支持 update_domain 和 delete"})
	}
}

// ============================================================
// Admin Unpublish (set status='hidden')
// ============================================================

func adminUnpublishExperience(c *gin.Context, _ *repository.ExperienceRepo, db *pgxpool.Pool) {
	id := c.Param("id")
	result, err := db.Exec(c.Request.Context(),
		"UPDATE experiences SET status='hidden', updated_at=NOW() WHERE id=$1 AND status='published'",
		id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unpublish: %v", err)})
		return
	}
	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "experience not found or already hidden"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ============================================================
// Admin Hard Delete (permanent, for non-approved experiences)
// ============================================================

func adminHardDeleteExperience(c *gin.Context, _ *repository.ExperienceRepo, db *pgxpool.Pool) {
	id := c.Param("id")
	// Only allow hard delete for non-approved (not in public pool)
	result, err := db.Exec(c.Request.Context(),
		"DELETE FROM experiences WHERE id=$1 AND review_status != 'approved'",
		id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("hard delete: %v", err)})
		return
	}
	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "experience not found or is approved (use soft delete instead)"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ============================================================
// Admin Update Review Status
// ============================================================

type adminReviewStatusRequest struct {
	ReviewStatus string `json:"review_status"` // approved | rejected | pending | private
	Reason       string `json:"reason"`
}

func adminUpdateReviewStatus(c *gin.Context, _ *repository.ExperienceRepo, db *pgxpool.Pool) {
	id := c.Param("id")
	var req adminReviewStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}
	valid := map[string]bool{"approved": true, "rejected": true, "pending": true, "private": true}
	if !valid[req.ReviewStatus] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的审核状态，支持 approved/rejected/pending/private"})
		return
	}

	var query string
	var args []interface{}
	if req.Reason != "" {
		query = "UPDATE experiences SET review_status=$1, review_reason=$2, updated_at=NOW() WHERE id=$3"
		args = []interface{}{req.ReviewStatus, req.Reason, id}
	} else {
		query = "UPDATE experiences SET review_status=$1, updated_at=NOW() WHERE id=$2"
		args = []interface{}{req.ReviewStatus, id}
	}

	result, err := db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("update review status: %v", err)})
		return
	}
	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "experience not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "review_status": req.ReviewStatus})
}
