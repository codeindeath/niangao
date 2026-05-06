package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niangao/backend/internal/model"
)

// RegisterAdminUserRoutes registers admin user management routes.
func RegisterAdminUserRoutes(admin *gin.RouterGroup, db *pgxpool.Pool) {
	r := admin.Group("/users")
	{
		r.GET("", func(c *gin.Context) {
			adminListUsers(c, db)
		})
		r.GET("/:id", func(c *gin.Context) {
			adminGetUser(c, db)
		})
		r.GET("/:id/experiences", func(c *gin.Context) {
			adminListUserExperiences(c, db)
		})
		r.PUT("/:id/status", func(c *gin.Context) {
			adminUpdateUserStatus(c, db)
		})
		r.POST("/batch-status", func(c *gin.Context) {
			adminBatchUpdateUserStatus(c, db)
		})
	}
}

// ============================================================
// GET /admin/users
// ============================================================

type adminUserListQuery struct {
	Search       string `form:"search"`
	AuthProvider string `form:"auth_provider"`
	IsActive     *bool  `form:"is_active"`
	DateFrom     string `form:"date_from"`
	DateTo       string `form:"date_to"`
	Sort         string `form:"sort"` // "activity" or "newest" (default)
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
}

func adminListUsers(c *gin.Context, db *pgxpool.Pool) {
	ctx := c.Request.Context()
	var q adminUserListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "查询参数错误"})
		return
	}

	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 20
	}

	// Dynamic WHERE clause
	var conditions []string
	var args []interface{}
	idx := 1

	conditions = append(conditions, "1=1")

	if q.Search != "" {
		conditions = append(conditions, fmt.Sprintf("u.nickname ILIKE $%d", idx))
		args = append(args, "%"+q.Search+"%")
		idx++
	}

	if q.AuthProvider != "" {
		if q.AuthProvider == "dev" {
			conditions = append(conditions, "u.apple_user_id LIKE 'dev-%'")
		} else if q.AuthProvider == "apple" {
			conditions = append(conditions, "u.apple_user_id NOT LIKE 'dev-%'")
		}
	}

	if q.IsActive != nil {
		if *q.IsActive {
			conditions = append(conditions, "u.deleted_at IS NULL")
		} else {
			conditions = append(conditions, "u.deleted_at IS NOT NULL")
		}
	}

	if q.DateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("u.created_at >= $%d::timestamptz", idx))
		args = append(args, q.DateFrom)
		idx++
	}
	if q.DateTo != "" {
		conditions = append(conditions, fmt.Sprintf("u.created_at <= $%d::timestamptz", idx))
		args = append(args, q.DateTo)
		idx++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count query (simpler: just count users without the LEFT JOIN)
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users u WHERE %s", whereClause)
	var total int
	if err := db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("count: %v", err)})
		return
	}

	orderClause := "u.created_at DESC"
	if q.Sort == "activity" {
		orderClause = "exp_count DESC, u.created_at DESC"
	}

	// Select query
	selectQuery := fmt.Sprintf(
		`SELECT u.id, u.nickname, u.avatar_url, u.title, u.is_admin, u.created_at,
			CASE WHEN u.apple_user_id LIKE 'dev-%%' THEN 'dev' ELSE 'apple' END as auth_provider,
			(u.deleted_at IS NULL) as is_active,
			COUNT(e.id) as exp_count
		 FROM users u
		 LEFT JOIN experiences e ON e.author_id = u.id
		 WHERE %s
		 GROUP BY u.id
		 ORDER BY %s
		 LIMIT $%d OFFSET $%d`,
		whereClause, orderClause, idx, idx+1,
	)
	args = append(args, q.PageSize, (q.Page-1)*q.PageSize)

	rows, err := db.Query(ctx, selectQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("list: %v", err)})
		return
	}
	defer rows.Close()

	var users []model.AdminUserItem
	for rows.Next() {
		var u model.AdminUserItem
		if err := rows.Scan(
			&u.ID, &u.Nickname, &u.AvatarURL, &u.Title,
			&u.IsAdmin, &u.CreatedAt, &u.AuthProvider,
			&u.IsActive, &u.ExpCount,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("scan: %v", err)})
			return
		}
		users = append(users, u)
	}

	if users == nil {
		users = []model.AdminUserItem{}
	}

	c.JSON(http.StatusOK, model.PaginatedResponse{
		Data:     users,
		Total:    total,
		Page:     q.Page,
		PageSize: q.PageSize,
	})
}

// ============================================================
// GET /admin/users/:id
// ============================================================

func adminGetUser(c *gin.Context, db *pgxpool.Pool) {
	ctx := c.Request.Context()
	userID := c.Param("id")

	var detail model.AdminUserDetail

	// Fetch base user info
	err := db.QueryRow(ctx, `
		SELECT u.id, u.nickname, u.avatar_url, u.bio, u.title, u.is_admin, u.created_at,
			CASE WHEN u.apple_user_id LIKE 'dev-%' THEN 'dev' ELSE 'apple' END,
			u.deleted_at IS NULL,
			COUNT(e.id)
		FROM users u
		LEFT JOIN experiences e ON e.author_id = u.id
		WHERE u.id = $1
		GROUP BY u.id`,
		userID,
	).Scan(
		&detail.ID, &detail.Nickname, &detail.AvatarURL, &detail.Bio, &detail.Title,
		&detail.IsAdmin, &detail.CreatedAt, &detail.AuthProvider,
		&detail.IsActive, &detail.ExpCount,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// Fetch all stats with separate queries for clarity
	// like_received: likes on this user's experiences
	db.QueryRow(ctx, `
		SELECT COUNT(*) FROM likes l
		JOIN experiences e ON e.id = l.experience_id
		WHERE e.author_id = $1`, userID,
	).Scan(&detail.LikeReceived)

	// bookmark_received: bookmarks on this user's experiences
	db.QueryRow(ctx, `
		SELECT COUNT(*) FROM bookmarks b
		JOIN experiences e ON e.id = b.experience_id
		WHERE e.author_id = $1`, userID,
	).Scan(&detail.BookmarkReceived)

	// viewed_count: how many times this user's experiences were viewed
	db.QueryRow(ctx, `
		SELECT COUNT(*) FROM user_views uv
		JOIN experiences e ON e.id = uv.experience_id
		WHERE e.author_id = $1`, userID,
	).Scan(&detail.ViewedCount)

	// liked_count: how many likes this user gave
	db.QueryRow(ctx, `
		SELECT COUNT(*) FROM likes
		WHERE user_id = $1`, userID,
	).Scan(&detail.LikedCount)

	// bookmarked_count: how many bookmarks this user has
	db.QueryRow(ctx, `
		SELECT COUNT(*) FROM bookmarks
		WHERE user_id = $1`, userID,
	).Scan(&detail.BookmarkedCount)

	// chat_count: how many conversations
	db.QueryRow(ctx, `
		SELECT COUNT(*) FROM conversations
		WHERE user_id = $1`, userID,
	).Scan(&detail.ChatCount)

	// msg_count: total messages across all conversations
	db.QueryRow(ctx, `
		SELECT COUNT(*) FROM messages m
		JOIN conversations c ON c.id = m.conversation_id
		WHERE c.user_id = $1`, userID,
	).Scan(&detail.MsgCount)

	c.JSON(http.StatusOK, detail)
}

// ============================================================
// GET /admin/users/:id/experiences
// ============================================================

func adminListUserExperiences(c *gin.Context, db *pgxpool.Pool) {
	ctx := c.Request.Context()
	userID := c.Param("id")

	rows, err := db.Query(ctx, `
		SELECT e.id, e.author_id, e.content, e.interpretation,
			e.domain, e.sub_domain, e.is_private,
			e.review_status, e.review_reason,
			e.quality_score, e.score_details,
			e.is_official, e.source_label,
			e.like_count, e.bookmark_count,
			e.interpretation_generated,
			e.creator_name, e.source_type, e.score_reason,
			e.status, e.created_at, e.updated_at, e.deleted_at
		FROM experiences e
		WHERE e.author_id = $1
		ORDER BY e.created_at DESC
		LIMIT 50`, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("query: %v", err)})
		return
	}
	defer rows.Close()

	var exps []model.Experience
	for rows.Next() {
		var e model.Experience
		if err := rows.Scan(
			&e.ID, &e.AuthorID, &e.Content, &e.Interpretation,
			&e.Domain, &e.SubDomain, &e.IsPrivate,
			&e.ReviewStatus, &e.ReviewReason,
			&e.QualityScore, &e.ScoreDetails,
			&e.IsOfficial, &e.SourceLabel,
			&e.LikeCount, &e.BookmarkCount,
			&e.InterpretationGenerated,
			&e.CreatorName, &e.SourceType, &e.ScoreReason,
			&e.Status, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("scan: %v", err)})
			return
		}
		exps = append(exps, e)
	}

	if exps == nil {
		exps = []model.Experience{}
	}

	c.JSON(http.StatusOK, exps)
}

// ============================================================
// PUT /admin/users/:id/status
// ============================================================

type adminUserStatusRequest struct {
	Active bool    `json:"active"`
	Reason *string `json:"reason"`
}

func adminUpdateUserStatus(c *gin.Context, db *pgxpool.Pool) {
	ctx := c.Request.Context()
	targetUserID := c.Param("id")
	adminID := getAuthUserID(c)
	if adminID == "" {
		return
	}

	var req adminUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	var action string
	var query string

	if req.Active {
		// Enable: clear deleted_at
		action = "启用"
		query = `UPDATE users SET deleted_at = NULL, updated_at = NOW() WHERE id = $1`
	} else {
		// Disable: set deleted_at
		action = "禁用"
		query = `UPDATE users SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1`
	}

	result, err := db.Exec(ctx, query, targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("update: %v", err)})
		return
	}

	if result.RowsAffected() == 0 {
		if req.Active {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在或已处于启用状态"})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在或已处于禁用状态"})
		}
		return
	}

	// Insert admin log
	detail := fmt.Sprintf(`{"target_user":"%s","active":%t}`, targetUserID, req.Active)
	if req.Reason != nil {
		detail = fmt.Sprintf(`{"target_user":"%s","active":%t,"reason":"%s"}`, targetUserID, req.Active, *req.Reason)
	}
	_, logErr := db.Exec(ctx, `
		INSERT INTO admin_logs (admin_id, action_type, target_type, target_id, detail, result)
		VALUES ($1, $2, 'user', $3, $4::jsonb, 'success')`,
		adminID, action, targetUserID, detail,
	)
	if logErr != nil {
		// Log failure is non-fatal — the status update succeeded
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"active":  req.Active,
			"warning": "操作成功，但日志记录失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"active": req.Active,
	})
}

// ============================================================
// POST /admin/users/batch-status — Batch enable/disable users
// ============================================================

type adminBatchUserStatusRequest struct {
	IDs    []string `json:"ids" binding:"required"`
	Active bool     `json:"active"`
	Reason *string  `json:"reason"`
}

func adminBatchUpdateUserStatus(c *gin.Context, db *pgxpool.Pool) {
	ctx := c.Request.Context()
	adminID := getAuthUserID(c)
	if adminID == "" {
		return
	}

	var req adminBatchUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误：需提供 ids 和 active"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids 不能为空"})
		return
	}

	var action string
	var query string
	if req.Active {
		action = "batch_enable"
		query = `UPDATE users SET deleted_at = NULL, updated_at = NOW() WHERE id = ANY($1)`
	} else {
		action = "batch_disable"
		query = `UPDATE users SET deleted_at = NOW(), updated_at = NOW() WHERE id = ANY($1)`
	}

	result, err := db.Exec(ctx, query, req.IDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("batch update: %v", err)})
		return
	}

	// Log
	detail := fmt.Sprintf(`{"ids":%s,"active":%t}`, toJSONArray(req.IDs), req.Active)
	if req.Reason != nil {
		detail = fmt.Sprintf(`{"ids":%s,"active":%t,"reason":"%s"}`, toJSONArray(req.IDs), req.Active, *req.Reason)
	}
	_, _ = db.Exec(ctx,
		`INSERT INTO admin_logs (admin_id, action_type, target_type, detail, result)
		 VALUES ($1, $2, 'user', $3::jsonb, 'success')`,
		adminID, action, detail,
	)

	c.JSON(http.StatusOK, gin.H{
		"status":         "ok",
		"action":         action,
		"affected_count": result.RowsAffected(),
	})
}

func toJSONArray(strs []string) string {
	quoted := make([]string, len(strs))
	for i, s := range strs {
		quoted[i] = `"` + s + `"`
	}
	return "[" + strings.Join(quoted, ",") + "]"
}
