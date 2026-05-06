package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niangao/backend/internal/model"
)

// RegisterAdminLogRoutes registers admin log viewing routes.
func RegisterAdminLogRoutes(admin *gin.RouterGroup, db *pgxpool.Pool) {
	admin.GET("/logs", func(c *gin.Context) {
		listAdminLogs(c, db)
	})
}

// ============================================================
// GET /admin/logs
// ============================================================

func listAdminLogs(c *gin.Context, db *pgxpool.Pool) {
	ctx := c.Request.Context()
	var q model.AdminLogQuery
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

	if q.AdminID != "" {
		conditions = append(conditions, fmt.Sprintf("al.admin_id = $%d", idx))
		args = append(args, q.AdminID)
		idx++
	}
	if q.ActionType != "" {
		conditions = append(conditions, fmt.Sprintf("al.action_type = $%d", idx))
		args = append(args, q.ActionType)
		idx++
	}

	// date_from / date_to (direct query params, not in model)
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("al.created_at >= $%d", idx))
		args = append(args, dateFrom)
		idx++
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		conditions = append(conditions, fmt.Sprintf("al.created_at <= $%d", idx))
		args = append(args, dateTo)
		idx++
	}
	// is_sensitive: filter only sensitive operations
	if c.Query("is_sensitive") == "true" {
		conditions = append(conditions,
			fmt.Sprintf(`al.action_type IN ('user_disable','user_enable','batch_disable','batch_enable','config_update','review_approve','review_reject','hard_delete')`))
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM admin_logs al WHERE %s", whereClause)
	var total int
	if err := db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("count: %v", err)})
		return
	}

	// Select query
	selectQuery := fmt.Sprintf(
		`SELECT al.id, al.admin_id, COALESCE(u.nickname, '') as admin_name,
			al.action_type, al.target_type, al.target_id,
			al.detail, al.result, al.created_at
		 FROM admin_logs al
		 LEFT JOIN users u ON u.id = al.admin_id
		 WHERE %s
		 ORDER BY al.created_at DESC
		 LIMIT $%d OFFSET $%d`,
		whereClause, idx, idx+1,
	)
	args = append(args, q.PageSize, (q.Page-1)*q.PageSize)

	rows, err := db.Query(ctx, selectQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("list: %v", err)})
		return
	}
	defer rows.Close()

	var items []model.AdminLogItem
	for rows.Next() {
		var item model.AdminLogItem
		if err := rows.Scan(
			&item.ID, &item.AdminID, &item.AdminName,
			&item.ActionType, &item.TargetType, &item.TargetID,
			&item.Detail, &item.Result, &item.CreatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("scan: %v", err)})
			return
		}
		items = append(items, item)
	}

	if items == nil {
		items = []model.AdminLogItem{}
	}

	c.JSON(http.StatusOK, model.PaginatedResponse{
		Data:     items,
		Total:    total,
		Page:     q.Page,
		PageSize: q.PageSize,
	})
}
