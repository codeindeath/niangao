package handler

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niangao/backend/internal/model"
)

// RegisterAdminDashboardRoutes registers admin dashboard routes on the given
// admin group (which should already have RequireAdmin middleware applied).
func RegisterAdminDashboardRoutes(admin *gin.RouterGroup, db *pgxpool.Pool) {
	admin.GET("/dashboard", func(c *gin.Context) {
		getDashboard(c, db)
	})
	admin.GET("/trends", func(c *gin.Context) {
		getTrends(c, db)
	})
}

// ============================================================
// GET /admin/dashboard
// ============================================================

func getDashboard(c *gin.Context, db *pgxpool.Pool) {
	ctx := c.Request.Context()
	d := model.AdminDashboard{}

	// Total counts
	db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`).Scan(&d.TotalUsers)
	db.QueryRow(ctx, `SELECT COUNT(*) FROM experiences WHERE status='published' AND deleted_at IS NULL`).Scan(&d.TotalExperiences)

	// Today's counts
	db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE created_at >= CURRENT_DATE`).Scan(&d.TodayNewUsers)
	db.QueryRow(ctx, `SELECT COUNT(*) FROM experiences WHERE created_at >= CURRENT_DATE`).Scan(&d.TodayNewExps)
	db.QueryRow(ctx, `SELECT COUNT(DISTINCT user_id) FROM user_views WHERE viewed_at >= CURRENT_DATE`).Scan(&d.TodayActiveUsers)
	db.QueryRow(ctx, `SELECT COUNT(*) FROM conversations WHERE created_at >= CURRENT_DATE`).Scan(&d.TodayAIChats)

	// Review counts
	db.QueryRow(ctx, `SELECT COUNT(*) FROM experiences WHERE review_status='pending'`).Scan(&d.PendingReviews)
	db.QueryRow(ctx, `SELECT COUNT(*) FROM admin_logs WHERE action_type='review_approve' AND created_at >= CURRENT_DATE`).Scan(&d.TodayApproved)
	db.QueryRow(ctx, `SELECT COUNT(*) FROM admin_logs WHERE action_type='review_reject' AND created_at >= CURRENT_DATE`).Scan(&d.TodayRejected)


	// Yesterday's counts
	var yesterdayNewUsers, yesterdayNewExps int64
	db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE created_at >= CURRENT_DATE - INTERVAL '1 day' AND created_at < CURRENT_DATE`).Scan(&yesterdayNewUsers)
	db.QueryRow(ctx, `SELECT COUNT(*) FROM experiences WHERE created_at >= CURRENT_DATE - INTERVAL '1 day' AND created_at < CURRENT_DATE`).Scan(&yesterdayNewExps)

	// Review preview: latest 5 pending reviews
	type reviewPreviewItem struct {
		ID          string `json:"id"`
		Content     string `json:"content"`
		Domain      string `json:"domain"`
		SubmittedAt string `json:"submitted_at"`
	}
	var reviewPreview []reviewPreviewItem
	rows, err := db.Query(ctx,
		`SELECT id, content, COALESCE(domain::text,''), created_at
		 FROM experiences WHERE review_status='pending' AND deleted_at IS NULL
		 ORDER BY created_at DESC LIMIT 5`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var item reviewPreviewItem
			var submittedAt interface{}
			if err := rows.Scan(&item.ID, &item.Content, &item.Domain, &submittedAt); err != nil {
				continue
			}
			if t, ok := submittedAt.(interface{ String() string }); ok {
				item.SubmittedAt = t.String()
			}
			reviewPreview = append(reviewPreview, item)
		}
	}
	if reviewPreview == nil {
		reviewPreview = []reviewPreviewItem{}
	}

	// Platform production stats
	var platformTotal, platformToday, platformUnscored int64
	db.QueryRow(ctx, `SELECT COUNT(*) FROM experiences WHERE source_type='platform' AND status='published' AND deleted_at IS NULL`).Scan(&platformTotal)
	db.QueryRow(ctx, `SELECT COUNT(*) FROM experiences WHERE source_type='platform' AND created_at >= CURRENT_DATE`).Scan(&platformToday)
	db.QueryRow(ctx, `SELECT COUNT(*) FROM experiences WHERE source_type='platform' AND quality_score IS NULL AND deleted_at IS NULL`).Scan(&platformUnscored)

	c.JSON(http.StatusOK, gin.H{
		"total_users":        d.TotalUsers,
		"total_experiences":  d.TotalExperiences,
		"today_new_users":    d.TodayNewUsers,
		"today_new_exps":     d.TodayNewExps,
		"today_active_users": d.TodayActiveUsers,
		"today_ai_chats":     d.TodayAIChats,
		"pending_reviews":    d.PendingReviews,
		"today_approved":     d.TodayApproved,
		"today_rejected":     d.TodayRejected,
		"yesterday_new_users": yesterdayNewUsers,
		"yesterday_new_exps":  yesterdayNewExps,
		"platform_total":      platformTotal,
		"platform_today":      platformToday,
		"platform_unscored":   platformUnscored,
		"review_preview":     reviewPreview,
	})
}

// ============================================================
// GET /admin/trends?days=7
// ============================================================

func getTrends(c *gin.Context, db *pgxpool.Pool) {
	ctx := c.Request.Context()

	days := 7
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 90 {
			days = parsed
		}
	}

	type TrendResponse struct {
		Days        int           `json:"days"`
		Users       []model.Trend `json:"users"`
		Experiences []model.Trend `json:"experiences"`
	}

	// Generate daily series with LEFT JOIN to get counts per day.
	users, err := queryDashboardTrend(ctx, db, days,
		`SELECT d::date, COALESCE(COUNT(u.id), 0)
		 FROM generate_series(CURRENT_DATE - ($1 - 1), CURRENT_DATE, '1 day'::interval) AS d
		 LEFT JOIN users u ON u.created_at::date = d
		 GROUP BY d ORDER BY d`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户趋势失败"})
		return
	}

	experiences, err := queryDashboardTrend(ctx, db, days,
		`SELECT d::date, COALESCE(COUNT(e.id), 0)
		 FROM generate_series(CURRENT_DATE - ($1 - 1), CURRENT_DATE, '1 day'::interval) AS d
		 LEFT JOIN experiences e ON e.created_at::date = d
		 GROUP BY d ORDER BY d`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取经验趋势失败"})
		return
	}

	c.JSON(http.StatusOK, TrendResponse{
		Days:        days,
		Users:       users,
		Experiences: experiences,
	})
}

func queryDashboardTrend(ctx context.Context, db *pgxpool.Pool, days int, query string) ([]model.Trend, error) {
	rows, err := db.Query(ctx, query, days)
	if err != nil {
		log.Printf("queryDashboardTrend error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var trends []model.Trend
	for rows.Next() {
		var t model.Trend
		if err := rows.Scan(&t.Date, &t.Count); err != nil {
			continue
		}
		trends = append(trends, t)
	}
	return trends, nil
}
