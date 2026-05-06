package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RegisterAdminExportRoutes registers admin CSV export routes.
func RegisterAdminExportRoutes(admin *gin.RouterGroup, db *pgxpool.Pool) {
	export := admin.Group("/export")
	{
		export.GET("/users", func(c *gin.Context) {
			exportUsersCSV(c, db)
		})
		export.GET("/experiences", func(c *gin.Context) {
			exportExperiencesCSV(c, db)
		})
	}
}

// ============================================================
// GET /admin/export/users?format=csv
// ============================================================

func exportUsersCSV(c *gin.Context, db *pgxpool.Pool) {
	ctx := c.Request.Context()

	rows, err := db.Query(ctx,
		`SELECT u.id, u.nickname, u.created_at,
		        COALESCE((u.deleted_at IS NULL), true) AS is_active,
		        COALESCE(u.is_admin, false) AS is_admin,
		        COUNT(DISTINCT e.id) AS exp_count
		 FROM users u
		 LEFT JOIN experiences e ON e.author_id = u.id AND e.deleted_at IS NULL
		 WHERE u.deleted_at IS NULL
		 GROUP BY u.id
		 ORDER BY u.created_at DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("查询用户失败: %v", err)})
		return
	}
	defer rows.Close()

	filename := fmt.Sprintf("users_export_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// Write BOM for Excel compatibility
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(c.Writer)
	w.Write([]string{"ID", "昵称", "注册时间", "是否活跃", "是否管理员", "经验数"})

	for rows.Next() {
		var id, nickname string
		var createdAt time.Time
		var isActive, isAdmin bool
		var expCount int
		if err := rows.Scan(&id, &nickname, &createdAt, &isActive, &isAdmin, &expCount); err != nil {
			continue
		}
		w.Write([]string{
			id,
			nickname,
			createdAt.Format("2006-01-02 15:04:05"),
			boolToStr(isActive),
			boolToStr(isAdmin),
			fmt.Sprintf("%d", expCount),
		})
	}
	w.Flush()
}

// ============================================================
// GET /admin/export/experiences?format=csv
// ============================================================

func exportExperiencesCSV(c *gin.Context, db *pgxpool.Pool) {
	ctx := c.Request.Context()

	rows, err := db.Query(ctx,
		`SELECT e.id, e.content, COALESCE(u.nickname, '') AS author_name,
		        e.domain::text, e.sub_domain, e.source_type, e.review_status,
		        e.quality_score, e.created_at
		 FROM experiences e
		 LEFT JOIN users u ON u.id = e.author_id
		 WHERE e.deleted_at IS NULL
		 ORDER BY e.created_at DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("查询经验失败: %v", err)})
		return
	}
	defer rows.Close()

	filename := fmt.Sprintf("experiences_export_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// Write BOM for Excel compatibility
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(c.Writer)
	w.Write([]string{"ID", "内容", "作者", "领域", "子领域", "来源", "审核状态", "质量分", "创建时间"})

	for rows.Next() {
		var id, content, authorName, domain, subDomain, sourceType, reviewStatus string
		var qualityScore *float64
		var createdAt time.Time
		if err := rows.Scan(&id, &content, &authorName, &domain, &subDomain, &sourceType, &reviewStatus, &qualityScore, &createdAt); err != nil {
			continue
		}
		scoreStr := ""
		if qualityScore != nil {
			scoreStr = fmt.Sprintf("%.1f", *qualityScore)
		}
		w.Write([]string{
			id,
			content,
			authorName,
			domain,
			subDomain,
			sourceType,
			reviewStatus,
			scoreStr,
			createdAt.Format("2006-01-02 15:04:05"),
		})
	}
	w.Flush()
}

func boolToStr(b bool) string {
	if b {
		return "是"
	}
	return "否"
}
