package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RegisterAdminAIRoutes registers AI service admin routes.
func RegisterAdminAIRoutes(admin *gin.RouterGroup, db *pgxpool.Pool) {
	admin.GET("/ai-status", func(c *gin.Context) {
		getAIStatus(c, db)
	})
}

// ============================================================
// Tier stats sub-structs
// ============================================================

type tierCounts struct {
	Today int64 `json:"today"`
	Total int64 `json:"total"`
}

type tierStats struct {
	Review         tierCounts `json:"review"`
	Chat           tierCounts `json:"chat"`
	Interpretation tierCounts `json:"interpretation"`
}

type costEstimate struct {
	TodayEstimated float64 `json:"today_estimated"`
	MonthEstimated float64 `json:"month_estimated"`
}

type promptConfig struct {
	ReviewPromptLength     int `json:"review_prompt_length"`
	ChatSystemPromptLength int `json:"chat_system_prompt_length"`
}

type batchTaskItem struct {
	ID         string    `json:"id"`
	ActionType string    `json:"action_type"`
	TargetID   string    `json:"target_id"`
	Result     string    `json:"result"`
	CreatedAt  time.Time `json:"created_at"`
}

type AIStatusResponse struct {
	Model         string          `json:"model"`
	Healthy       bool            `json:"healthy"`
	LatencyMs     float64         `json:"latency_ms"`
	SuccessRate   float64         `json:"success_rate"`
	LastHourCalls int64           `json:"last_hour_calls"`
	ErrorMsg      string          `json:"error_msg,omitempty"`
	TierStats     *tierStats      `json:"tier_stats,omitempty"`
	DailyCost     *costEstimate   `json:"daily_cost,omitempty"`
	PromptConfig  *promptConfig   `json:"prompt_config,omitempty"`
	BatchTasks    []batchTaskItem `json:"batch_tasks,omitempty"`
}

func getAIStatus(c *gin.Context, db *pgxpool.Pool) {
	aiURL := os.Getenv("AI_SERVICE_URL")
	if aiURL == "" {
		aiURL = "http://localhost:8000"
	}

	status := AIStatusResponse{
		Model:   "deepseek-chat",
		Healthy: false,
	}

	// Health check
	client := &http.Client{Timeout: 5 * time.Second}
	start := time.Now()
	resp, err := client.Get(aiURL + "/health")
	elapsed := time.Since(start).Seconds() * 1000
	status.LatencyMs = elapsed

	if err != nil {
		status.ErrorMsg = err.Error()
		c.JSON(http.StatusOK, status)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		status.ErrorMsg = "AI service returned " + resp.Status
		c.JSON(http.StatusOK, status)
		return
	}

	var health struct {
		Status string `json:"status"`
		Model  string `json:"model"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&health); err == nil {
		if health.Model != "" {
			status.Model = health.Model
		}
	}

	status.Healthy = true

	// Only fetch DB stats if db is available
	if db != nil {
		ctx := c.Request.Context()
		status.TierStats = fetchTierStats(ctx, db)
		status.DailyCost = fetchDailyCost(ctx, db)
		status.PromptConfig = fetchPromptConfig(ctx, db)
		status.BatchTasks = fetchBatchTasks(ctx, db)
	}

	c.JSON(http.StatusOK, status)
}

// ============================================================
// tier_stats helpers
// ============================================================

func fetchTierStats(ctx context.Context, db *pgxpool.Pool) *tierStats {
	ts := &tierStats{}

	// Review: experiences reviewed today vs total (use admin_logs for approve/reject)
	db.QueryRow(ctx, `SELECT COUNT(*) FROM admin_logs WHERE action_type IN ('review_approve','review_reject') AND created_at >= CURRENT_DATE`).Scan(&ts.Review.Today)
	db.QueryRow(ctx, `SELECT COUNT(*) FROM admin_logs WHERE action_type IN ('review_approve','review_reject')`).Scan(&ts.Review.Total)

	// Chat: conversations today vs total
	db.QueryRow(ctx, `SELECT COUNT(*) FROM conversations WHERE created_at >= CURRENT_DATE`).Scan(&ts.Chat.Today)
	db.QueryRow(ctx, `SELECT COUNT(*) FROM conversations`).Scan(&ts.Chat.Total)

	// Interpretation: experiences with interpretation today vs total
	db.QueryRow(ctx, `SELECT COUNT(*) FROM experiences WHERE interpretation_generated = true AND updated_at >= CURRENT_DATE`).Scan(&ts.Interpretation.Today)
	db.QueryRow(ctx, `SELECT COUNT(*) FROM experiences WHERE interpretation_generated = true`).Scan(&ts.Interpretation.Total)

	return ts
}

// ============================================================
// daily_cost helper (estimated: $0.002/message for review, $0.001/message for chat)
// ============================================================

func fetchDailyCost(ctx context.Context, db *pgxpool.Pool) *costEstimate {
	const reviewCostPerMsg = 0.002
	const chatCostPerMsg = 0.001

	// Today: count messages and interpretations generated today
	var todayMsgs int64
	var todayInters int64
	db.QueryRow(ctx, `SELECT COUNT(*) FROM messages WHERE created_at >= CURRENT_DATE`).Scan(&todayMsgs)
	db.QueryRow(ctx, `SELECT COUNT(*) FROM experiences WHERE interpretation_generated = true AND updated_at >= CURRENT_DATE`).Scan(&todayInters)

	todayCost := float64(todayMsgs)*chatCostPerMsg + float64(todayInters)*reviewCostPerMsg

	// This month
	var monthMsgs int64
	var monthInters int64
	db.QueryRow(ctx, `SELECT COUNT(*) FROM messages WHERE created_at >= date_trunc('month', CURRENT_DATE)`).Scan(&monthMsgs)
	db.QueryRow(ctx, `SELECT COUNT(*) FROM experiences WHERE interpretation_generated = true AND updated_at >= date_trunc('month', CURRENT_DATE)`).Scan(&monthInters)

	monthCost := float64(monthMsgs)*chatCostPerMsg + float64(monthInters)*reviewCostPerMsg

	return &costEstimate{
		TodayEstimated: float64(int(todayCost*100)) / 100,
		MonthEstimated: float64(int(monthCost*100)) / 100,
	}
}

// ============================================================
// prompt_config helper
// ============================================================

func fetchPromptConfig(ctx context.Context, db *pgxpool.Pool) *promptConfig {
	pc := &promptConfig{}

	var val []byte
	err := db.QueryRow(ctx, `SELECT value FROM system_config WHERE key = 'review_prompt'`).Scan(&val)
	if err == nil && val != nil {
		pc.ReviewPromptLength = len(val)
	}

	err = db.QueryRow(ctx, `SELECT value FROM system_config WHERE key = 'chat_system_prompt'`).Scan(&val)
	if err == nil && val != nil {
		pc.ChatSystemPromptLength = len(val)
	}

	return pc
}

// ============================================================
// batch_tasks helper
// ============================================================

func fetchBatchTasks(ctx context.Context, db *pgxpool.Pool) []batchTaskItem {
	rows, err := db.Query(ctx,
		`SELECT id, action_type, COALESCE(target_id,''), COALESCE(result,''), created_at
		 FROM admin_logs
		 WHERE action_type LIKE 'batch%' OR action_type LIKE 'review_batch%'
		 ORDER BY created_at DESC
		 LIMIT 5`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var tasks []batchTaskItem
	for rows.Next() {
		var t batchTaskItem
		if err := rows.Scan(&t.ID, &t.ActionType, &t.TargetID, &t.Result, &t.CreatedAt); err != nil {
			continue
		}
		tasks = append(tasks, t)
	}
	return tasks
}
