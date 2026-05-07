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

func RegisterAdminPlatformRoutes(admin *gin.RouterGroup, db *pgxpool.Pool, expRepo *repository.ExperienceRepo) {
	platform := admin.Group("/platform-experiences")
	{
		platform.GET("", func(c *gin.Context) { listPlatformExperiences(c, db) })
		platform.POST("", func(c *gin.Context) { createPlatformExperience(c, db) })
		platform.PUT("/:id", func(c *gin.Context) { updatePlatformExperience(c, db) })
		platform.POST("/batch-ai", func(c *gin.Context) { batchAIScore(c, db) })
		platform.POST("/:id/publish", func(c *gin.Context) { togglePublishPlatformExperience(c, db) })
		platform.POST("/:id/rescore", func(c *gin.Context) { rescorePlatformExperience(c, db) })
		platform.POST("/import-csv", func(c *gin.Context) { importCSVPlatformExperiences(c, db) })
		// Pipeline routes (extract + review + save)
		RegisterAdminPlatformPipelineRoutes(admin, db)
	}
}

func listPlatformExperiences(c *gin.Context, db *pgxpool.Pool) {
	page := 1
	pageSize := 20
	if p, ok := c.GetQuery("page"); ok {
		if n, err := parseInt(p); err == nil && n > 0 {
			page = n
		}
	}
	if ps, ok := c.GetQuery("page_size"); ok {
		if n, err := parseInt(ps); err == nil && n > 0 && n <= 100 {
			pageSize = n
		}
	}

	var conditions []string
	var args []interface{}
	idx := 1

	conditions = append(conditions, "e.source_type='platform' AND e.deleted_at IS NULL")

	if domain := c.Query("domain"); domain != "" {
		conditions = append(conditions, fmt.Sprintf("e.domain=$%d", idx))
		args = append(args, domain)
		idx++
	}
	if hi := c.Query("has_interpretation"); hi == "true" {
		conditions = append(conditions, "e.interpretation IS NOT NULL AND e.interpretation != ''")
	} else if hi == "false" {
		conditions = append(conditions, "(e.interpretation IS NULL OR e.interpretation = '')")
	}
	if search := c.Query("search"); search != "" {
		conditions = append(conditions, fmt.Sprintf("(e.content ILIKE $%d OR e.creator_name ILIKE $%d)", idx, idx+1))
		args = append(args, "%"+search+"%", "%"+search+"%")
		idx += 2
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count
	var total int
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM experiences e WHERE %s", whereClause)
	if err := db.QueryRow(c.Request.Context(), countSQL, args...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// Select
	selectSQL := fmt.Sprintf(`SELECT e.id, e.content, e.domain, e.sub_domain, e.creator_name, e.source_label,
		e.quality_score, e.score_reason, e.interpretation, e.interpretation_generated, e.like_count, e.bookmark_count,
		e.created_at, u.nickname as author_name
		FROM experiences e LEFT JOIN users u ON u.id=e.author_id
		WHERE %s ORDER BY e.created_at DESC LIMIT $%d OFFSET $%d`, whereClause, idx, idx+1)
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := db.Query(c.Request.Context(), selectSQL, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	defer rows.Close()

	type item struct {
		ID                string  `json:"id"`
		Content           string  `json:"content"`
		Domain            string  `json:"domain"`
		SubDomain         *string `json:"sub_domain"`
		CreatorName       *string `json:"creator_name"`
		SourceLabel       *string `json:"source_label"`
		QualityScore      *float64 `json:"quality_score"`
		ScoreReason       *string `json:"score_reason"`
		Interpretation    *string `json:"interpretation"`
		HasInterpretation bool    `json:"has_interpretation"`
		LikeCount         int     `json:"like_count"`
		BookmarkCount     int     `json:"bookmark_count"`
		CreatedAt         string  `json:"created_at"`
		AuthorName        string  `json:"author_name"`
	}

	var items []item
	for rows.Next() {
		var i item
		var createdAt interface{}
		if err := rows.Scan(&i.ID, &i.Content, &i.Domain, &i.SubDomain, &i.CreatorName, &i.SourceLabel,
			&i.QualityScore, &i.ScoreReason, &i.Interpretation, &i.HasInterpretation,
			&i.LikeCount, &i.BookmarkCount, &createdAt, &i.AuthorName); err != nil {
			continue
		}
		if t, ok := createdAt.(interface{ String() string }); ok {
			i.CreatedAt = t.String()
		}
		items = append(items, i)
	}
	if items == nil {
		items = []item{}
	}

	c.JSON(http.StatusOK, model.PaginatedResponse{
		Data:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func createPlatformExperience(c *gin.Context, db *pgxpool.Pool) {
	var req struct {
		Content      string `json:"content"`
		Domain       string `json:"domain"`
		SubDomain    string `json:"sub_domain"`
		CreatorName  string `json:"creator_name"`
		SourceLabel  string `json:"source_label"`
		ScoreReason  string `json:"score_reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Content == "" || req.Domain == "" || req.CreatorName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请填写内容、领域和创作者名称"})
		return
	}
	if !model.IsValidDomain(model.Domain(req.Domain)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的领域"})
		return
	}

	var id string
	officialID := "00000000-0000-0000-0000-000000000000"
	err := db.QueryRow(c.Request.Context(),
		`INSERT INTO experiences (author_id, content, domain, sub_domain, creator_name, source_label,
		 score_reason, source_type, is_official, review_status, status)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,'platform',true,'approved','published') RETURNING id`,
		officialID, req.Content, req.Domain, nilIfEmpty(req.SubDomain),
		req.CreatorName, nilIfEmpty(req.SourceLabel), nilIfEmpty(req.ScoreReason),
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "status": "created"})
}

func updatePlatformExperience(c *gin.Context, db *pgxpool.Pool) {
	id := c.Param("id")
	var req struct {
		Content     *string `json:"content"`
		Domain      *string `json:"domain"`
		SubDomain   *string `json:"sub_domain"`
		CreatorName *string `json:"creator_name"`
		SourceLabel *string `json:"source_label"`
		ScoreReason *string `json:"score_reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求"})
		return
	}

	var sets []string
	var args []interface{}
	idx := 1

	if req.Content != nil {
		sets = append(sets, fmt.Sprintf("content=$%d", idx))
		args = append(args, *req.Content)
		idx++
	}
	if req.Domain != nil {
		if !model.IsValidDomain(model.Domain(*req.Domain)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的领域"})
			return
		}
		sets = append(sets, fmt.Sprintf("domain=$%d", idx))
		args = append(args, *req.Domain)
		idx++
	}
	if req.SubDomain != nil {
		sets = append(sets, fmt.Sprintf("sub_domain=$%d", idx))
		args = append(args, *req.SubDomain)
		idx++
	}
	if req.CreatorName != nil {
		sets = append(sets, fmt.Sprintf("creator_name=$%d", idx))
		args = append(args, *req.CreatorName)
		idx++
	}
	if req.SourceLabel != nil {
		sets = append(sets, fmt.Sprintf("source_label=$%d", idx))
		args = append(args, *req.SourceLabel)
		idx++
	}
	if req.ScoreReason != nil {
		sets = append(sets, fmt.Sprintf("score_reason=$%d", idx))
		args = append(args, *req.ScoreReason)
		idx++
	}

	if len(sets) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有要修改的字段"})
		return
	}

	sets = append(sets, "updated_at=NOW()")
	args = append(args, id)
	sql := fmt.Sprintf("UPDATE experiences SET %s WHERE id=$%d AND source_type='platform'", strings.Join(sets, ","), idx)

	result, err := db.Exec(c.Request.Context(), sql, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "平台经验不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func batchAIScore(c *gin.Context, db *pgxpool.Pool) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择至少一条经验"})
		return
	}

	// Query experiences
	rows, err := db.Query(c.Request.Context(),
		`SELECT id, content, COALESCE(domain::text,''), COALESCE(sub_domain,'') FROM experiences WHERE id = ANY($1) AND source_type='platform'`,
		req.IDs,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询经验失败"})
		return
	}
	defer rows.Close()

	type expRow struct {
		id, content, domain, subDomain string
	}
	var exps []expRow
	for rows.Next() {
		var e expRow
		if err := rows.Scan(&e.id, &e.content, &e.domain, &e.subDomain); err != nil {
			continue
		}
		exps = append(exps, e)
	}

	if len(exps) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到选中的平台经验"})
		return
	}

	// Process each: call AI review
	success, failed := 0, 0
	for _, exp := range exps {
		result, err := callAIReview(ReviewRequest{
			Content:   exp.content,
			Domain:    exp.domain,
			SubDomain: exp.subDomain,
		})
		if err != nil {
			failed++
			continue
		}
		score := 5.0
		if result.Score != nil {
			score = result.Score.Overall
		}
		// Update DB with score
		_, err = db.Exec(c.Request.Context(),
			`UPDATE experiences SET quality_score=$1, score_reason=$2, updated_at=NOW() WHERE id=$3`,
			score, result.Reason, exp.id,
		)
		if err != nil {
			failed++
			continue
		}
		success++
	}

	c.JSON(http.StatusOK, gin.H{
		"total":   len(exps),
		"success": success,
		"failed":  failed,
	})
}

// ============================================================
// POST /platform-experiences/:id/publish — Toggle publish status
// ============================================================

func togglePublishPlatformExperience(c *gin.Context, db *pgxpool.Pool) {
	id := c.Param("id")

	// Check current status
	var currentStatus string
	err := db.QueryRow(c.Request.Context(),
		`SELECT status FROM experiences WHERE id=$1 AND source_type='platform' AND deleted_at IS NULL`,
		id,
	).Scan(&currentStatus)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "平台经验不存在"})
		return
	}

	newStatus := "published"
	if currentStatus == "published" {
		newStatus = "hidden"
	}

	_, err = db.Exec(c.Request.Context(),
		`UPDATE experiences SET status=$1, updated_at=NOW() WHERE id=$2 AND source_type='platform'`,
		newStatus, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": newStatus})
}

// ============================================================
// POST /platform-experiences/:id/rescore — Single rescore via AI
// ============================================================

func rescorePlatformExperience(c *gin.Context, db *pgxpool.Pool) {
	id := c.Param("id")

	var exp struct {
		id, content, domain, subDomain string
	}
	err := db.QueryRow(c.Request.Context(),
		`SELECT id, content, COALESCE(domain::text,''), COALESCE(sub_domain,'')
		 FROM experiences WHERE id=$1 AND source_type='platform' AND deleted_at IS NULL`,
		id,
	).Scan(&exp.id, &exp.content, &exp.domain, &exp.subDomain)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "平台经验不存在"})
		return
	}

	result, err := callAIReview(ReviewRequest{
		Content:   exp.content,
		Domain:    exp.domain,
		SubDomain: exp.subDomain,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI 评分失败"})
		return
	}

	score := 5.0
	if result.Score != nil {
		score = result.Score.Overall
	}

	_, err = db.Exec(c.Request.Context(),
		`UPDATE experiences SET quality_score=$1, score_reason=$2, updated_at=NOW() WHERE id=$3`,
		score, result.Reason, exp.id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新评分失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        exp.id,
		"score":     score,
		"reason":    result.Reason,
		"approved":  result.Approved,
	})
}

// ============================================================
// POST /platform-experiences/import-csv — Batch import from CSV
// ============================================================

func importCSVPlatformExperiences(c *gin.Context, db *pgxpool.Pool) {
	var req struct {
		Data string `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Data == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供 CSV 数据"})
		return
	}

	lines := strings.Split(strings.TrimSpace(req.Data), "\n")
	if len(lines) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV 数据需要包含表头和至少一行数据"})
		return
	}

	// Parse header
	header := strings.Split(lines[0], ",")
	colIndex := map[string]int{}
	for i, h := range header {
		colIndex[strings.TrimSpace(h)] = i
	}

	type csvRow struct {
		Content, Domain, SubDomain, CreatorName, SourceLabel, ScoreReason string
	}

	var rows []csvRow
	for i := 1; i < len(lines); i++ {
		fields := strings.Split(lines[i], ",")
		get := func(col string) string {
			idx, ok := colIndex[col]
			if !ok || idx >= len(fields) {
				return ""
			}
			return strings.TrimSpace(fields[idx])
		}
		row := csvRow{
			Content:     get("content"),
			Domain:      get("domain"),
			SubDomain:   get("sub_domain"),
			CreatorName: get("creator_name"),
			SourceLabel: get("source_label"),
			ScoreReason: get("score_reason"),
		}
		if row.Content != "" && row.Domain != "" && row.CreatorName != "" {
			if model.IsValidDomain(model.Domain(row.Domain)) {
				rows = append(rows, row)
			}
		}
	}

	if len(rows) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV 中没有有效数据行"})
		return
	}

	officialID := "00000000-0000-0000-0000-000000000000"
	inserted := 0
	for _, row := range rows {
		var id string
		err := db.QueryRow(c.Request.Context(),
			`INSERT INTO experiences (author_id, content, domain, sub_domain, creator_name, source_label,
			 score_reason, source_type, is_official, review_status, status)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,'platform',true,'approved','published') RETURNING id`,
			officialID, row.Content, row.Domain, nilIfEmpty(row.SubDomain),
			row.CreatorName, nilIfEmpty(row.SourceLabel), nilIfEmpty(row.ScoreReason),
		).Scan(&id)
		if err == nil {
			inserted++
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"total":    len(rows),
		"inserted": inserted,
	})
}


