package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niangao/backend/internal/model"
	_ "github.com/niangao/backend/internal/repository" // imported per handler spec
)

// ============================================================
// Admin Review Routes
// ============================================================

// RegisterAdminReviewRoutes 注册管理后台 - 审核路由
func RegisterAdminReviewRoutes(admin *gin.RouterGroup, db *pgxpool.Pool) {
	r := admin.Group("/reviews")
	{
		r.GET("", func(c *gin.Context) {
			adminListReviews(c, db)
		})
		r.GET("/:id", func(c *gin.Context) {
			adminGetReview(c, db)
		})
		r.POST("/:id/approve", func(c *gin.Context) {
			adminApproveReview(c, db)
		})
		r.POST("/:id/reject", func(c *gin.Context) {
			adminRejectReview(c, db)
		})
		r.POST("/:id/retry", func(c *gin.Context) {
			adminRetryReview(c, db)
		})
		r.POST("/batch", func(c *gin.Context) {
			adminBatchReview(c, db)
		})
		r.POST("/:id/misjudge", func(c *gin.Context) {
			adminMisjudgeReview(c, db)
		})
	}
}

// ============================================================
// Review list select columns & scan helper
// ============================================================

const reviewSelectCols = `e.id, e.content, e.domain, e.sub_domain, e.source_type, e.review_status,
	e.quality_score, e.score_details, e.score_reason, e.interpretation, e.created_at,
	COALESCE(u.nickname, '') as author_name`

func scanReviewItem(row interface{ Scan(...interface{}) error }, item *model.ReviewItem) error {
	return row.Scan(
		&item.ID, &item.Content, &item.Domain, &item.SubDomain,
		&item.SourceType, &item.ReviewStatus, &item.AIScore, &item.AIScoreDetail,
		&item.AIVerdict, &item.AIInterpretation, &item.SubmittedAt, &item.AuthorName,
	)
}

// ============================================================
// Admin log helper
// ============================================================

func logAdminAction(c *gin.Context, db *pgxpool.Pool, actionType, targetID string, detail interface{}, result string) {
	adminID, exists := c.Get("user_id")
	if !exists {
		return
	}
	adminIDStr, ok := adminID.(string)
	if !ok {
		return
	}

	detailJSON, err := json.Marshal(detail)
	if err != nil {
		detailJSON = []byte("{}")
	}

	db.Exec(c.Request.Context(),
		`INSERT INTO admin_logs(admin_id, action_type, target_type, target_id, detail, result)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		adminIDStr, actionType, "experience", targetID, string(detailJSON), result,
	)
}

// ============================================================
// GET /admin/reviews — List pending reviews
// ============================================================

type adminReviewListQuery struct {
	Domain       string `form:"domain"`
	SourceType   string `form:"source_type"`
	ReviewStatus string `form:"review_status"`
	DateFrom     string `form:"date_from"`
	DateTo       string `form:"date_to"`
	AIVerdict    string `form:"ai_verdict"`
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
}

func adminListReviews(c *gin.Context, db *pgxpool.Pool) {
	var q adminReviewListQuery
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

	if q.ReviewStatus != "" {
		conditions = append(conditions, fmt.Sprintf("e.review_status = $%d", idx))
		args = append(args, q.ReviewStatus)
		idx++
	} else {
		conditions = append(conditions, "e.review_status = 'pending'")
	}

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
	if q.DateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("e.created_at >= $%d", idx))
		args = append(args, q.DateFrom)
		idx++
	}
	if q.DateTo != "" {
		conditions = append(conditions, fmt.Sprintf("e.created_at <= $%d", idx))
		args = append(args, q.DateTo)
		idx++
	}
	if q.AIVerdict != "" {
		conditions = append(conditions, fmt.Sprintf("e.score_reason ILIKE $%d", idx))
		args = append(args, "%"+q.AIVerdict+"%")
		idx++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM experiences e WHERE %s", whereClause)
	if err := db.QueryRow(c.Request.Context(), countQuery, args...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("count: %v", err)})
		return
	}

	// Select
	selectQuery := fmt.Sprintf(
		`SELECT %s FROM experiences e
		 LEFT JOIN users u ON u.id = e.author_id
		 WHERE %s
		 ORDER BY e.created_at ASC
		 LIMIT $%d OFFSET $%d`,
		reviewSelectCols, whereClause, idx, idx+1,
	)
	args = append(args, q.PageSize, (q.Page-1)*q.PageSize)

	rows, err := db.Query(c.Request.Context(), selectQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("list: %v", err)})
		return
	}
	defer rows.Close()

	items := make([]model.ReviewItem, 0)
	for rows.Next() {
		var item model.ReviewItem
		if err := scanReviewItem(rows, &item); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("scan: %v", err)})
			return
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, model.PaginatedResponse{
		Data:     items,
		Total:    total,
		Page:     q.Page,
		PageSize: q.PageSize,
	})
}

// ============================================================
// GET /admin/reviews/:id — Get review detail
// ============================================================

func adminGetReview(c *gin.Context, db *pgxpool.Pool) {
	id := c.Param("id")

	query := fmt.Sprintf(
		`SELECT %s FROM experiences e
		 LEFT JOIN users u ON u.id = e.author_id
		 WHERE e.id = $1`, reviewSelectCols,
	)

	var item model.ReviewItem
	if err := scanReviewItem(db.QueryRow(c.Request.Context(), query, id), &item); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
		return
	}

	c.JSON(http.StatusOK, item)
}

// ============================================================
// POST /admin/reviews/:id/approve — Approve review
// ============================================================

func adminApproveReview(c *gin.Context, db *pgxpool.Pool) {
	id := c.Param("id")

	result, err := db.Exec(c.Request.Context(),
		"UPDATE experiences SET review_status='approved', updated_at=NOW() WHERE id=$1",
		id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("approve: %v", err)})
		return
	}
	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "experience not found"})
		return
	}

	logAdminAction(c, db, "review_approve", id, gin.H{"review_status": "approved"}, "success")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ============================================================
// POST /admin/reviews/:id/reject — Reject review
// ============================================================

type adminRejectRequest struct {
	Reason string `json:"reason"`
}

func adminRejectReview(c *gin.Context, db *pgxpool.Pool) {
	id := c.Param("id")

	var req adminRejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	result, err := db.Exec(c.Request.Context(),
		"UPDATE experiences SET review_status='rejected', review_reason=$2, updated_at=NOW() WHERE id=$1",
		id, req.Reason,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("reject: %v", err)})
		return
	}
	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "experience not found"})
		return
	}

	logAdminAction(c, db, "review_reject", id, gin.H{"review_status": "rejected", "reason": req.Reason}, "success")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ============================================================
// POST /admin/reviews/:id/retry — Force re-review
// ============================================================

func adminRetryReview(c *gin.Context, db *pgxpool.Pool) {
	id := c.Param("id")

	result, err := db.Exec(c.Request.Context(),
		"UPDATE experiences SET review_status='pending', updated_at=NOW() WHERE id=$1",
		id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("retry: %v", err)})
		return
	}
	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "experience not found"})
		return
	}

	logAdminAction(c, db, "review_retry", id, gin.H{"review_status": "pending"}, "success")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ============================================================
// POST /admin/reviews/batch — Batch approve/reject
// ============================================================

type adminBatchReviewRequest struct {
	IDs    []string `json:"ids" binding:"required"`
	Action string   `json:"action" binding:"required"` // "approve" | "reject"
	Reason *string  `json:"reason"`
}

func adminBatchReview(c *gin.Context, db *pgxpool.Pool) {
	var req adminBatchReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误：需提供 ids 和 action"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids 不能为空"})
		return
	}

	var query string
	var args []interface{}
	detail := gin.H{"action": req.Action, "ids": req.IDs}

	switch req.Action {
	case "approve":
		query = "UPDATE experiences SET review_status='approved', updated_at=NOW() WHERE id = ANY($1)"
		args = append(args, req.IDs)
	case "reject":
		if req.Reason != nil && *req.Reason != "" {
			query = "UPDATE experiences SET review_status='rejected', review_reason=$1, updated_at=NOW() WHERE id = ANY($2)"
			args = append(args, *req.Reason, req.IDs)
			detail["reason"] = *req.Reason
		} else {
			query = "UPDATE experiences SET review_status='rejected', updated_at=NOW() WHERE id = ANY($1)"
			args = append(args, req.IDs)
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的操作类型，支持 approve 和 reject"})
		return
	}

	result, err := db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("batch %s: %v", req.Action, err)})
		return
	}

	// Log batch action
	adminID, _ := c.Get("user_id")
	adminIDStr, _ := adminID.(string)
	detailJSON, _ := json.Marshal(detail)
	db.Exec(c.Request.Context(),
		`INSERT INTO admin_logs(admin_id, action_type, target_type, target_id, detail, result)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		adminIDStr, "review_batch_"+req.Action, "experience", "", string(detailJSON), "success",
	)

	c.JSON(http.StatusOK, gin.H{
		"status":         "ok",
		"action":         req.Action,
		"affected_count": result.RowsAffected(),
	})
}

// ============================================================
// POST /admin/reviews/:id/misjudge — Mark an AI review as misjudged
// ============================================================

type adminMisjudgeRequest struct {
	Reason string `json:"reason"`
}

func adminMisjudgeReview(c *gin.Context, db *pgxpool.Pool) {
	id := c.Param("id")

	var req adminMisjudgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	// Reset review_status to pending so it can be re-evaluated
	result, err := db.Exec(c.Request.Context(),
		"UPDATE experiences SET review_status='pending', updated_at=NOW() WHERE id=$1",
		id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("misjudge: %v", err)})
		return
	}
	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "experience not found"})
		return
	}

	logAdminAction(c, db, "review_misjudge", id, gin.H{
		"review_status": "pending",
		"reason":        req.Reason,
	}, "success")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
