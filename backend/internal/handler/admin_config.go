package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niangao/backend/internal/model"
)

// RegisterAdminConfigRoutes registers admin config routes on the given
// admin group (which should already have RequireAdmin middleware applied).
func RegisterAdminConfigRoutes(admin *gin.RouterGroup, db *pgxpool.Pool) {
	cfg := admin.Group("/config")
	{
		cfg.GET("", func(c *gin.Context) {
			getConfig(c, db)
		})
		cfg.PUT("", func(c *gin.Context) {
			updateConfig(c, db)
		})
		cfg.GET("/defaults", func(c *gin.Context) {
			getConfigDefaults(c)
		})
		cfg.GET("/sensitive-words", func(c *gin.Context) {
			listSensitiveWords(c, db)
		})
		cfg.POST("/sensitive-words", func(c *gin.Context) {
			addSensitiveWord(c, db)
		})
		cfg.DELETE("/sensitive-words/:id", func(c *gin.Context) {
			deleteSensitiveWord(c, db)
		})
	}
}

// ============================================================
// GET /admin/config
// ============================================================

func getConfig(c *gin.Context, db *pgxpool.Pool) {
	ctx := c.Request.Context()

	rows, err := db.Query(ctx, `SELECT key, value FROM system_config`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取配置失败"})
		return
	}
	defer rows.Close()

	result := make(map[string]interface{})
	for rows.Next() {
		var key string
		var value []byte
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}
		var v interface{}
		if err := json.Unmarshal(value, &v); err != nil {
			result[key] = string(value)
		} else {
			result[key] = v
		}
	}

	c.JSON(http.StatusOK, result)
}

// ============================================================
// PUT /admin/config
// ============================================================

func updateConfig(c *gin.Context, db *pgxpool.Pool) {
	ctx := c.Request.Context()
	adminID := getAuthUserID(c)
	if adminID == "" {
		return // getAuthUserID already wrote the error response
	}

	var req model.ConfigUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	if req.Key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key 不能为空"})
		return
	}

	// Marshal value to JSON for JSONB column
	valueJSON, err := json.Marshal(req.Value)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "value 格式错误"})
		return
	}

	tag, err := db.Exec(ctx,
		`UPDATE system_config SET value = $2, updated_at = NOW() WHERE key = $1`,
		req.Key, valueJSON,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("更新配置失败: %v", err)})
		return
	}

	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "配置项不存在: " + req.Key})
		return
	}

	// Log to admin_logs
	detail := map[string]interface{}{
		"key":   req.Key,
		"value": req.Value,
	}
	detailJSON, _ := json.Marshal(detail)
	_, _ = db.Exec(ctx,
		`INSERT INTO admin_logs (admin_id, action_type, target_type, detail, result) VALUES ($1, $2, $3, $4, $5)`,
		adminID, "config_update", "system_config", detailJSON, "success",
	)

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ============================================================
// GET /admin/config/sensitive-words
// ============================================================

type sensitiveWordItem struct {
	ID   int    `json:"id"`
	Word string `json:"word"`
}

func listSensitiveWords(c *gin.Context, db *pgxpool.Pool) {
	ctx := c.Request.Context()

	rows, err := db.Query(ctx, `SELECT id, word FROM sensitive_words ORDER BY word`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取敏感词失败"})
		return
	}
	defer rows.Close()

	var words []sensitiveWordItem
	for rows.Next() {
		var w sensitiveWordItem
		if err := rows.Scan(&w.ID, &w.Word); err != nil {
			continue
		}
		words = append(words, w)
	}

	if words == nil {
		words = []sensitiveWordItem{}
	}
	c.JSON(http.StatusOK, words)
}

// ============================================================
// POST /admin/config/sensitive-words
// ============================================================

type addSensitiveWordRequest struct {
	Word string `json:"word"`
}

func addSensitiveWord(c *gin.Context, db *pgxpool.Pool) {
	ctx := c.Request.Context()
	adminID := getAuthUserID(c)
	if adminID == "" {
		return
	}

	var req addSensitiveWordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	if req.Word == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "word 不能为空"})
		return
	}

	_, err := db.Exec(ctx,
		`INSERT INTO sensitive_words (word) VALUES ($1) ON CONFLICT (word) DO NOTHING`,
		req.Word,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("添加敏感词失败: %v", err)})
		return
	}

	// Log to admin_logs
	detail := map[string]interface{}{"word": req.Word}
	detailJSON, _ := json.Marshal(detail)
	_, _ = db.Exec(ctx,
		`INSERT INTO admin_logs (admin_id, action_type, target_type, detail, result) VALUES ($1, $2, $3, $4, $5)`,
		adminID, "sensitive_word_add", "sensitive_words", detailJSON, "success",
	)

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ============================================================
// DELETE /admin/config/sensitive-words/:id
// ============================================================

func deleteSensitiveWord(c *gin.Context, db *pgxpool.Pool) {
	ctx := c.Request.Context()
	adminID := getAuthUserID(c)
	if adminID == "" {
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}

	tag, err := db.Exec(ctx, `DELETE FROM sensitive_words WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("删除敏感词失败: %v", err)})
		return
	}

	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "敏感词不存在"})
		return
	}

	// Log to admin_logs
	detail := map[string]interface{}{"id": id}
	detailJSON, _ := json.Marshal(detail)
	_, _ = db.Exec(ctx,
		`INSERT INTO admin_logs (admin_id, action_type, target_type, detail, result) VALUES ($1, $2, $3, $4, $5)`,
		adminID, "sensitive_word_delete", "sensitive_words", detailJSON, "success",
	)

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ============================================================
// GET /admin/config/defaults — Hardcoded default config
// ============================================================

func getConfigDefaults(c *gin.Context) {
	defaults := gin.H{
		"features": gin.H{
			"registration":     true,
			"ai_interpretation": true,
			"search":           true,
			"ai_chat":          true,
			"comments":         true,
		},
		"limits": gin.H{
			"content_min_rune":        10,
			"content_max_rune":        200,
			"interpretation_max_rune": 300,
		},
		"rate_limits": gin.H{
			"create_per_hour": 5,
			"chat_per_hour":   20,
		},
		"review_mode": gin.H{
			"auto_approve_platform":   false,
			"ai_first_human_confirm":  true,
		},
	}
	c.JSON(http.StatusOK, defaults)
}
