package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niangao/backend/internal/model"
)

// ============================================================
// AI Extract — call /api/v1/extract on the AI service
// ============================================================

type ExtractRequest struct {
	SourceText  string `json:"source_text"`
	SourceLabel string `json:"source_label"`
	SourceType  string `json:"source_type"` // book/celebrity/ugc
}

type ExtractedItem struct {
	Content             string `json:"content"`
	Domain              string `json:"domain"`
	SubDomain           string `json:"sub_domain"`
	ExpType             string `json:"exp_type"`
	NeedsNewSubdomain   bool   `json:"needs_new_subdomain"`
	SuggestedSubName    string `json:"suggested_sub_name"`
	SuggestedSubLabel   string `json:"suggested_sub_label"`
}

type ExtractResponse struct {
	Items []ExtractedItem `json:"items"`
	Count int             `json:"count"`
}

func callAIExtract(sourceText, sourceLabel, sourceType string) (*ExtractResponse, error) {
	req := ExtractRequest{
		SourceText:  sourceText,
		SourceLabel: sourceLabel,
		SourceType:  sourceType,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal extract request: %w", err)
	}

	aiURL := os.Getenv("AI_SERVICE_URL")
	if aiURL == "" {
		aiURL = "http://localhost:8000"
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Post(aiURL+"/api/v1/extract", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("call AI extract: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI extract returned status %d", resp.StatusCode)
	}

	var result ExtractResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode extract result: %w", err)
	}

	return &result, nil
}

// ============================================================
// POST /admin/platform/pipeline/ingest — Full pipeline
//   extract → review → save
// ============================================================

type PipelineIngestRequest struct {
	SourceText  string `json:"source_text" binding:"required"`
	SourceLabel string `json:"source_label"`  // 书名/人名/来源
	SourceType  string `json:"source_type"`   // book/celebrity/ugc
}

type PipelineIngestResponse struct {
	Extracted   int `json:"extracted"`    // AI 提取出的经验数
	Reviewed    int `json:"reviewed"`     // 通过审核的经验数
	Rejected    int `json:"rejected"`     // 被拒的经验数
	Saved       int `json:"saved"`        // 成功入库数
	NewDomains  int `json:"new_domains"`  // 自动创建的子领域数
	Errors      int `json:"errors"`       // 处理失败的条数
	Total       int `json:"total"`        // 全部处理数
}

func pipelineIngest(c *gin.Context, db *pgxpool.Pool) {
	var req PipelineIngestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供 source_text"})
		return
	}

	if req.SourceType == "" {
		req.SourceType = "celebrity"
	}

	// Step 1: AI 提取
	extractResult, err := callAIExtract(req.SourceText, req.SourceLabel, req.SourceType)
	if err != nil {
		log.Printf("[pipeline] extract failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("AI 提取失败: %v", err)})
		return
	}

	if len(extractResult.Items) == 0 {
		c.JSON(http.StatusOK, PipelineIngestResponse{
			Extracted: 0,
			Total:     0,
		})
		return
	}

	// Step 2: AI 审核 + Step 3: 入库
	result := PipelineIngestResponse{
		Extracted: len(extractResult.Items),
		Total:     len(extractResult.Items),
	}

	officialID := "00000000-0000-0000-0000-000000000000"

	for _, item := range extractResult.Items {
		if item.Content == "" {
			result.Errors++
			continue
		}

		// Domain fallback
		domain := item.Domain
		if !model.IsValidDomain(model.Domain(domain)) {
			domain = "cognition"
		}

		subDomain := item.SubDomain
		if item.NeedsNewSubdomain && item.SuggestedSubName != "" {
			// 尝试自动创建子领域
			if err := autoCreateSubdomain(c, db, domain, item.SuggestedSubName, item.SuggestedSubLabel); err == nil {
				subDomain = item.SuggestedSubName
				result.NewDomains++
			}
		}

		// AI 审核打分
		reviewReq := ReviewRequest{
			Content:   item.Content,
			Domain:    domain,
			SubDomain: subDomain,
		}

		reviewResult, err := callAIReview(reviewReq)
		if err != nil {
			log.Printf("[pipeline] review failed for '%s': %v", truncateStr(item.Content, 30), err)
			result.Errors++
			continue
		}

		// 审核通过才入库
		if !reviewResult.Approved {
			result.Rejected++
			continue
		}

		result.Reviewed++

		score := 5.0
		if reviewResult.Score != nil {
			score = reviewResult.Score.Overall
		}

		scoreReason := reviewResult.Reason
		if len([]rune(scoreReason)) > 100 {
			scoreReason = string([]rune(scoreReason)[:100])
		}

		// 入库
		var subDomainPtr, sourceLabelPtr, creatorNamePtr, scoreReasonPtr *string
		if subDomain != "" {
			subDomainPtr = &subDomain
		}
		if req.SourceLabel != "" {
			sourceLabelPtr = &req.SourceLabel
			creatorNamePtr = &req.SourceLabel
		}

		if scoreReason != "" {
			scoreReasonPtr = &scoreReason
		}

		_, err = db.Exec(c.Request.Context(),
			`INSERT INTO experiences (author_id, content, domain, sub_domain, creator_name, source_label,
			 quality_score, score_reason, source_type, is_official, review_status, status)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'platform',true,'approved','published')`,
			officialID, item.Content, domain, subDomainPtr,
			creatorNamePtr, sourceLabelPtr,
			score, scoreReasonPtr,
		)
		if err != nil {
			log.Printf("[pipeline] insert failed: %v", err)
			result.Errors++
			continue
		}
		result.Saved++
	}

	c.JSON(http.StatusOK, result)
}

// ============================================================
// POST /admin/platform/pipeline/bulk-save — 批量保存已评分经验
//   (供采集脚本使用：脚本先 extract+review，再调此端点入库)
// ============================================================

type BulkSaveItem struct {
	Content     string  `json:"content"`
	Domain      string  `json:"domain"`
	SubDomain   string  `json:"sub_domain"`
	CreatorName string  `json:"creator_name"`
	SourceLabel string  `json:"source_label"`
	ScoreReason string  `json:"score_reason"`
	QualityScore float64 `json:"quality_score"`
	Approved    bool    `json:"approved"`
}

type BulkSaveRequest struct {
	Items []BulkSaveItem `json:"items"`
}

func bulkSavePlatformExperiences(c *gin.Context, db *pgxpool.Pool) {
	var req BulkSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供 items 列表"})
		return
	}

	if len(req.Items) == 0 {
		c.JSON(http.StatusOK, gin.H{"saved": 0})
		return
	}

	officialID := "00000000-0000-0000-0000-000000000000"
	saved := 0
	skipped := 0
	errors := 0

	for _, item := range req.Items {
		if item.Content == "" || !item.Approved {
			skipped++
			continue
		}

		domain := item.Domain
		if !model.IsValidDomain(model.Domain(domain)) {
			domain = "cognition"
		}

		var subDomainPtr, sourceLabelPtr, creatorNamePtr, scoreReasonPtr *string
		if item.SubDomain != "" {
			subDomainPtr = &item.SubDomain
		}
		if item.CreatorName != "" {
			creatorNamePtr = &item.CreatorName
		}
		if item.SourceLabel != "" {
			sourceLabelPtr = &item.SourceLabel
		}
		if item.ScoreReason != "" {
			scoreReasonPtr = &item.ScoreReason
		}

		score := item.QualityScore
		if score <= 0 {
			score = 5.0
		}

		_, err := db.Exec(c.Request.Context(),
			`INSERT INTO experiences (author_id, content, domain, sub_domain, creator_name, source_label,
			 quality_score, score_reason, source_type, is_official, review_status, status)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'platform',true,'approved','published')`,
			officialID, item.Content, domain, subDomainPtr,
			creatorNamePtr, sourceLabelPtr,
			score, scoreReasonPtr,
		)
		if err != nil {
			log.Printf("[bulk-save] insert failed: %v", err)
			errors++
			continue
		}
		saved++
	}

	c.JSON(http.StatusOK, gin.H{
		"saved":   saved,
		"skipped": skipped,
		"errors":  errors,
	})
}

// ============================================================
// Helper: auto-create subdomain if it doesn't exist
// ============================================================

func autoCreateSubdomain(c *gin.Context, db *pgxpool.Pool, parentName, subName, subLabel string) error {
	if subName == "" || parentName == "" {
		return fmt.Errorf("invalid subdomain name")
	}

	// Check if already exists
	var count int
	err := db.QueryRow(c.Request.Context(),
		`SELECT COUNT(*) FROM domains WHERE parent_name=$1 AND name=$2`,
		parentName, subName,
	).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // already exists
	}

	// Check current count for this parent
	var subCount int
	err = db.QueryRow(c.Request.Context(),
		`SELECT COUNT(*) FROM domains WHERE parent_name=$1`,
		parentName,
	).Scan(&subCount)
	if err != nil {
		return err
	}
	if subCount >= 10 {
		return fmt.Errorf("subdomain limit reached for %s", parentName)
	}

	label := subLabel
	if label == "" {
		label = subName
	}

	_, err = db.Exec(c.Request.Context(),
		`INSERT INTO domains (name, parent_name, label, sort_order, active)
		 VALUES ($1, $2, $3, $4, true)`,
		subName, parentName, label, subCount+1,
	)
	return err
}

func truncateStr(s string, maxChars int) string {
	runes := []rune(s)
	if len(runes) <= maxChars {
		return s
	}
	return string(runes[:maxChars])
}

// ============================================================
// Route registration
// ============================================================

func RegisterAdminPlatformPipelineRoutes(admin *gin.RouterGroup, db *pgxpool.Pool) {
	pipeline := admin.Group("/platform/pipeline")
	{
		pipeline.POST("/ingest", func(c *gin.Context) { pipelineIngest(c, db) })
		pipeline.POST("/bulk-save", func(c *gin.Context) { bulkSavePlatformExperiences(c, db) })
	}
}
