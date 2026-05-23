package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niangao/backend/internal/model"
)

// RegisterAdminStatsRoutes registers admin statistics routes on the given
// admin group (which should already have RequireAdmin middleware applied).
func RegisterAdminStatsRoutes(admin *gin.RouterGroup, db *pgxpool.Pool) {
	r := admin.Group("/stats")
	{
		r.GET("/users", func(c *gin.Context) {
			getAdminUserStats(c, db)
		})
		r.GET("/experiences", func(c *gin.Context) {
			getAdminExperienceStats(c, db)
		})
		r.GET("/interactions", func(c *gin.Context) {
			getAdminInteractionStats(c, db)
		})
		r.GET("/reviews", func(c *gin.Context) {
			getAdminReviewStats(c, db)
		})
		r.GET("/domains", func(c *gin.Context) {
			getAdminDomainStats(c, db)
		})
		r.GET("/ai", func(c *gin.Context) {
			getAdminAIStats(c, db)
		})
		r.GET("/retention", func(c *gin.Context) {
			getAdminRetentionStats(c, db)
		})
	}
}

// ============================================================
// helpers
// ============================================================

// parseDays extracts the "days" query parameter, clamped to 1-90 (default 7).
func parseDays(c *gin.Context) int {
	days := 7
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed >= 1 && parsed <= 90 {
			days = parsed
		}
	}
	return days
}

// queryTrend executes a generate_series-based daily count query and returns []model.Trend.
func queryTrend(ctx context.Context, db *pgxpool.Pool, days int, query string, args ...interface{}) ([]model.Trend, error) {
	allArgs := []interface{}{days}
	allArgs = append(allArgs, args...)
	rows, err := db.Query(ctx, query, allArgs...)
	if err != nil {
		log.Printf("queryTrend error: %v", err)
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

// ============================================================
// GET /admin/stats/users?days=7
// ============================================================

func getAdminUserStats(c *gin.Context, db *pgxpool.Pool) {
	ctx := c.Request.Context()
	days := parseDays(c)

	data, err := queryTrend(ctx, db, days,
		`SELECT d::date, COALESCE(COUNT(u.id), 0)
		 FROM generate_series(CURRENT_DATE - ($1 - 1), CURRENT_DATE, '1 day'::interval) AS d
		 LEFT JOIN users u ON u.created_at::date = d
		 GROUP BY d ORDER BY d`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户增长数据失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"days": days,
		"data": data,
	})
}

// ============================================================
// GET /admin/stats/experiences?days=7&source_type=all
// ============================================================

func getAdminExperienceStats(c *gin.Context, db *pgxpool.Pool) {
	ctx := c.Request.Context()
	days := parseDays(c)
	sourceType := c.Query("source_type")

	var query string
	var extraArgs []interface{}

	if sourceType == "" || sourceType == "all" {
		query = `SELECT d::date, COALESCE(COUNT(e.id), 0)
			 FROM generate_series(CURRENT_DATE - ($1 - 1), CURRENT_DATE, '1 day'::interval) AS d
			 LEFT JOIN experiences e ON e.created_at::date = d
			 GROUP BY d ORDER BY d`
	} else {
		query = `SELECT d::date, COALESCE(COUNT(e.id), 0)
			 FROM generate_series(CURRENT_DATE - ($1 - 1), CURRENT_DATE, '1 day'::interval) AS d
			 LEFT JOIN experiences e ON e.created_at::date = d AND e.source_type = $2
			 GROUP BY d ORDER BY d`
		extraArgs = append(extraArgs, sourceType)
	}

	data, err := queryTrend(ctx, db, days, query, extraArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取经验增长数据失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"days":        days,
		"source_type": sourceType,
		"data":        data,
	})
}

// ============================================================
// GET /admin/stats/interactions?days=7
// ============================================================

func getAdminInteractionStats(c *gin.Context, db *pgxpool.Pool) {
	ctx := c.Request.Context()
	days := parseDays(c)

	// Daily likes
	likes, err := queryTrend(ctx, db, days,
		`SELECT d::date, COALESCE(COUNT(l.user_id), 0)
		 FROM generate_series(CURRENT_DATE - ($1 - 1), CURRENT_DATE, '1 day'::interval) AS d
		 LEFT JOIN likes l ON l.created_at::date = d
		 GROUP BY d ORDER BY d`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取互动数据失败"})
		return
	}

	// Daily bookmarks
	bookmarks, err := queryTrend(ctx, db, days,
		`SELECT d::date, COALESCE(COUNT(b.user_id), 0)
		 FROM generate_series(CURRENT_DATE - ($1 - 1), CURRENT_DATE, '1 day'::interval) AS d
		 LEFT JOIN bookmarks b ON b.created_at::date = d
		 GROUP BY d ORDER BY d`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取互动数据失败"})
		return
	}

	// Daily views
	views, err := queryTrend(ctx, db, days,
		`SELECT d::date, COALESCE(COUNT(uv.user_id), 0)
		 FROM generate_series(CURRENT_DATE - ($1 - 1), CURRENT_DATE, '1 day'::interval) AS d
		 LEFT JOIN user_views uv ON uv.viewed_at::date = d
		 GROUP BY d ORDER BY d`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取互动数据失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"days":      days,
		"likes":     likes,
		"bookmarks": bookmarks,
		"views":     views,
	})
}

// ============================================================
// GET /admin/stats/reviews?days=7
// ============================================================

func getAdminReviewStats(c *gin.Context, db *pgxpool.Pool) {
	ctx := c.Request.Context()
	days := parseDays(c)

	// Daily approved
	approved, err := queryTrend(ctx, db, days,
		`SELECT d::date, COALESCE(COUNT(e.id), 0)
		 FROM generate_series(CURRENT_DATE - ($1 - 1), CURRENT_DATE, '1 day'::interval) AS d
		 LEFT JOIN experiences e ON e.updated_at::date = d AND e.review_status = 'approved'
		 GROUP BY d ORDER BY d`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取审核数据失败"})
		return
	}

	// Daily rejected
	rejected, err := queryTrend(ctx, db, days,
		`SELECT d::date, COALESCE(COUNT(e.id), 0)
		 FROM generate_series(CURRENT_DATE - ($1 - 1), CURRENT_DATE, '1 day'::interval) AS d
		 LEFT JOIN experiences e ON e.updated_at::date = d AND e.review_status = 'rejected'
		 GROUP BY d ORDER BY d`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取审核数据失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"days":     days,
		"approved": approved,
		"rejected": rejected,
	})
}

// ============================================================
// GET /admin/stats/domains
// ============================================================

type domainCountResponse struct {
	Domain string `json:"domain"`
	Count  int64  `json:"count"`
}

func getAdminDomainStats(c *gin.Context, db *pgxpool.Pool) {
	ctx := c.Request.Context()

	rows, err := db.Query(ctx,
		`SELECT e.domain::text, COUNT(*)
		 FROM experiences e
		 WHERE e.deleted_at IS NULL
		 GROUP BY e.domain
		 ORDER BY COUNT(*) DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取领域统计失败"})
		return
	}
	defer rows.Close()

	var data []domainCountResponse
	for rows.Next() {
		var d domainCountResponse
		if err := rows.Scan(&d.Domain, &d.Count); err != nil {
			continue
		}
		data = append(data, d)
	}

	if data == nil {
		data = []domainCountResponse{}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": data,
	})
}

// ============================================================
// GET /admin/stats/ai?days=7
// ============================================================

func getAdminAIStats(c *gin.Context, db *pgxpool.Pool) {
	ctx := c.Request.Context()
	days := parseDays(c)

	// Daily chat (conversation) count
	chats, err := queryTrend(ctx, db, days,
		`SELECT d::date, COALESCE(COUNT(c.id), 0)
		 FROM generate_series(CURRENT_DATE - ($1 - 1), CURRENT_DATE, '1 day'::interval) AS d
		 LEFT JOIN conversations c ON c.created_at::date = d
		 GROUP BY d ORDER BY d`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取AI使用数据失败"})
		return
	}

	// Daily message count
	messages, err := queryTrend(ctx, db, days,
		`SELECT d::date, COALESCE(COUNT(m.id), 0)
		 FROM generate_series(CURRENT_DATE - ($1 - 1), CURRENT_DATE, '1 day'::interval) AS d
		 LEFT JOIN messages m ON m.created_at::date = d
		 GROUP BY d ORDER BY d`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取AI使用数据失败"})
		return
	}

	// Daily interpretations generated
	interpretations, err := queryTrend(ctx, db, days,
		`SELECT d::date, COALESCE(COUNT(e.id), 0)
		 FROM generate_series(CURRENT_DATE - ($1 - 1), CURRENT_DATE, '1 day'::interval) AS d
		 LEFT JOIN experiences e ON e.updated_at::date = d
		   AND e.interpretation_generated = true
		   AND e.interpretation IS NOT NULL
		 GROUP BY d ORDER BY d`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取AI使用数据失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"days":            days,
		"chats":           chats,
		"messages":        messages,
		"interpretations": interpretations,
	})
}

// ============================================================
// GET /admin/stats/retention?days=30
// Returns day1/day7/day30 retention rates.
// Retention = users who performed any action (view/like/bookmark/chat)
// within N days of registration / total new users in the lookback window.
// ============================================================

type retentionResponse struct {
	Days  int           `json:"days"`
	Day1  []model.Trend `json:"day1"`
	Day7  []model.Trend `json:"day7"`
	Day30 []model.Trend `json:"day30"`
}

func getAdminRetentionStats(c *gin.Context, db *pgxpool.Pool) {
	ctx := c.Request.Context()
	days := parseDays(c)

	// For each day in the period, compute:
	//   total = users registered on that day
	//   returned_1  = users from that cohort who had activity within 1 day
	//   returned_7  = users from that cohort who had activity within 7 days
	//   returned_30 = users from that cohort who had activity within 30 days

	// We use a union of activity tables as "user returned":
	//   user_views, likes, bookmarks, conversations

	retentionQuery := func(windowDays int) ([]model.Trend, error) {
		query := fmt.Sprintf(`
			SELECT d::date,
				CASE WHEN registered = 0 THEN 0
				     ELSE ROUND(returned::numeric / registered * 100, 1)
				END
			FROM (
				SELECT d,
					COUNT(DISTINCT u.id) AS registered,
					COUNT(DISTINCT act.user_id) AS returned
				FROM generate_series(
					CURRENT_DATE - ($1 - 1), CURRENT_DATE, '1 day'::interval
				) AS d
				LEFT JOIN users u ON u.created_at::date = d AND u.deleted_at IS NULL
				LEFT JOIN (
					SELECT user_id, viewed_at::date AS act_date FROM user_views
					UNION ALL
					SELECT user_id, created_at::date AS act_date FROM likes
					UNION ALL
					SELECT user_id, created_at::date AS act_date FROM bookmarks
					UNION ALL
					SELECT user_id, created_at::date AS act_date FROM conversations
				) act ON act.user_id = u.id
					AND act.act_date >= u.created_at::date
					AND act.act_date <= u.created_at::date + INTERVAL '%d days'
				GROUP BY d
			) sub
			ORDER BY d
		`, windowDays)
		rows, err := db.Query(ctx, query, days)
		if err != nil {
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

	day1, err := retentionQuery(1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取Day1留存率失败"})
		return
	}
	day7, err := retentionQuery(7)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取Day7留存率失败"})
		return
	}
	day30, err := retentionQuery(30)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取Day30留存率失败"})
		return
	}

	c.JSON(http.StatusOK, retentionResponse{
		Days:  days,
		Day1:  day1,
		Day7:  day7,
		Day30: day30,
	})
}
