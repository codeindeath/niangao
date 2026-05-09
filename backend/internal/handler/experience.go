package handler

import (
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niangao/backend/internal/middleware"
	"github.com/niangao/backend/internal/model"
	"github.com/niangao/backend/internal/repository"
)

type ExperienceHandler struct {
	repo      *repository.ExperienceRepo
	likeRepo  *repository.LikeRepo
	bookRepo  *repository.BookmarkRepo
}

func RegisterExperienceRoutes(r *gin.RouterGroup, expRepo *repository.ExperienceRepo, likeRepo *repository.LikeRepo, bookRepo *repository.BookmarkRepo) {
	h := &ExperienceHandler{repo: expRepo, likeRepo: likeRepo, bookRepo: bookRepo}

	exp := r.Group("/experiences")
	{
		exp.GET("", h.List)
		exp.GET("/recommend", middleware.RequireAuth(), h.GetRecommendations)
		exp.GET("/:id", h.Get)
		exp.POST("", middleware.RequireAuth(), h.Create)
		exp.PUT("/:id", middleware.RequireAuth(), h.Update)
		exp.DELETE("/:id", middleware.RequireAuth(), h.Delete)
		exp.POST("/:id/like", middleware.RequireAuth(), h.ToggleLike)
		exp.POST("/:id/bookmark", middleware.RequireAuth(), h.ToggleBookmark)
	}

	// 个人维度 API — 直接在 v1 下注册，不走子 Group
	r.GET("/me/experiences", middleware.RequireAuth(), h.MyExperiences)
	r.GET("/me/bookmarks", middleware.RequireAuth(), h.MyBookmarks)
}

func (h *ExperienceHandler) List(c *gin.Context) {
	var query model.ExperienceListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "查询参数错误"})
		return
	}

	viewerStr := getOptionalUserID(c)

	experiences, total, err := h.repo.List(c.Request.Context(), query, viewerStr)
	if err != nil {
		log.Printf("ERROR List: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list experiences"})
		return
	}

	// Round-robin to break creator clustering (only for default/latest sort)
	if query.Sort == "" || query.Sort == "latest" {
		experiences = interleaveByCreator(experiences)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  experiences,
		"total": total,
		"page":  query.Page,
	})
}

func (h *ExperienceHandler) Get(c *gin.Context) {
	id := c.Param("id")
	viewerStr := getOptionalUserID(c)

	exp, err := h.repo.GetByID(c.Request.Context(), id, viewerStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "experience not found"})
		return
	}

	c.JSON(http.StatusOK, exp)
}

func (h *ExperienceHandler) Create(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}

	var req model.CreateExperienceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请填写完整：领域和子领域"})
		return
	}

	// Content validation with rune count (Chinese chars)
	contentRunes := len([]rune(req.Content))
	if contentRunes < 10 || contentRunes > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "经验内容需 10-100 字"})
		return
	}

	// Interpretation validation with rune count
	if req.Interpretation != "" && len([]rune(req.Interpretation)) > 300 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "经验解读不超过 300 字"})
		return
	}

	if !model.IsValidDomain(req.Domain) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain"})
		return
	}

	if !model.IsValidSubDomain(req.SubDomain) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sub_domain"})
		return
	}

	if !model.SubDomainBelongsToParent(req.Domain, req.SubDomain) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sub_domain does not belong to domain"})
		return
	}

	// 私密经验：跳过审核，直接保存
	if req.IsPrivate {
		exp, err := h.repo.Create(c.Request.Context(), userID, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create experience"})
			return
		}
		if bookmarked, err := h.bookRepo.Toggle(c.Request.Context(), userID, exp.ID); err != nil {
			log.Printf("auto-bookmark failed for exp=%s user=%s: %v", exp.ID, userID, err)
		} else if bookmarked {
			exp.BookmarkCount = 1
			exp.IsBookmarked = true
		}
		c.JSON(http.StatusCreated, exp)
		return
	}

	// 公开经验：硬策略检查
	if result := CheckHardPolicy(req.Content); !result.Passed {
		reason := result.Reason
		exp, err := h.repo.CreateWithReview(c.Request.Context(), userID, req,
			string(model.ReviewRejected), &reason, nil, nil, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save experience"})
			return
		}
		if bookmarked, err := h.bookRepo.Toggle(c.Request.Context(), userID, exp.ID); err != nil {
			log.Printf("auto-bookmark failed for exp=%s user=%s: %v", exp.ID, userID, err)
		} else if bookmarked {
			exp.BookmarkCount = 1
			exp.IsBookmarked = true
		}
		c.JSON(http.StatusCreated, gin.H{
			"experience": exp,
			"review": gin.H{
				"status":  "rejected",
				"reason":  result.Reason,
				"message": "经验已保存到你的个人经验，但因内容不符合准入规则，未进入平台经验池",
			},
		})
		return
	}

	// AI 审核
	aiResult, err := callAIReview(ReviewRequest{
		Content:   req.Content,
		Domain:    string(req.Domain),
		SubDomain: string(req.SubDomain),
	})
	if err != nil {
		log.Printf("AI review failed: %v — saving as pending", err)
		exp, err := h.repo.CreateWithReview(c.Request.Context(), userID, req,
			string(model.ReviewPending), nil, nil, nil, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save experience"})
			return
		}
		if bookmarked, err := h.bookRepo.Toggle(c.Request.Context(), userID, exp.ID); err != nil {
			log.Printf("auto-bookmark failed for exp=%s user=%s: %v", exp.ID, userID, err)
		} else if bookmarked {
			exp.BookmarkCount = 1
			exp.IsBookmarked = true
		}
		c.JSON(http.StatusCreated, gin.H{
			"experience": exp,
			"review": gin.H{
				"status":  "pending",
				"message": "经验已保存，审核中，稍后自动进入平台经验池",
			},
		})
		return
	}

	score, details := qualityScoreToDB(aiResult.Score)

	if !aiResult.Approved {
		exp, err := h.repo.CreateWithReview(c.Request.Context(), userID, req,
			string(model.ReviewRejected), &aiResult.Reason, score, details, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save experience"})
			return
		}
		if bookmarked, err := h.bookRepo.Toggle(c.Request.Context(), userID, exp.ID); err != nil {
			log.Printf("auto-bookmark failed for exp=%s user=%s: %v", exp.ID, userID, err)
		} else if bookmarked {
			exp.BookmarkCount = 1
			exp.IsBookmarked = true
		}
		c.JSON(http.StatusCreated, gin.H{
			"experience": exp,
			"review": gin.H{
				"status":  "rejected",
				"reason":  aiResult.Reason,
				"message": "经验已保存到你的个人经验，但未通过 AI 审核，未进入平台经验池",
			},
		})
		return
	}

	// 审核通过 — 检测古文并翻译
	var originalText *string
	if translateResult := callAITranslate(req.Content); translateResult != nil && translateResult.IsClassical {
		req.Content = translateResult.ModernText
		originalText = &translateResult.OriginalText
		log.Printf("Translation applied (lang=%s): orig=%s", translateResult.DetectedLang, (*originalText)[:min(len(*originalText), 30)])
	}

	exp, err := h.repo.CreateWithReview(c.Request.Context(), userID, req,
		string(model.ReviewApproved), nil, score, details, originalText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save experience"})
		return
	}

	if bookmarked, err := h.bookRepo.Toggle(c.Request.Context(), userID, exp.ID); err != nil {
		log.Printf("auto-bookmark failed for exp=%s user=%s: %v", exp.ID, userID, err)
	} else if bookmarked {
		exp.BookmarkCount = 1
		exp.IsBookmarked = true
	}

	c.JSON(http.StatusCreated, gin.H{
		"experience": exp,
		"review": gin.H{
			"status":  "approved",
			"score":   aiResult.Score,
			"message": "经验已发布并进入平台经验池",
		},
	})
}


func (h *ExperienceHandler) Update(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	id := c.Param("id")

	var req model.CreateExperienceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请填写完整：领域和子领域"})
		return
	}

	contentRunes := len([]rune(req.Content))
	if contentRunes < 10 || contentRunes > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "经验内容需 10-100 字"})
		return
	}

	// Interpretation validation with rune count
	if req.Interpretation != "" && len([]rune(req.Interpretation)) > 300 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "经验解读不超过 300 字"})
		return
	}

	if !model.IsValidDomain(req.Domain) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain"})
		return
	}

	if err := h.repo.Update(c.Request.Context(), id, userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *ExperienceHandler) Delete(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	id := c.Param("id")

	if err := h.repo.Delete(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *ExperienceHandler) ToggleLike(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	id := c.Param("id")

	liked, err := h.likeRepo.Toggle(c.Request.Context(), userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to toggle like"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"liked": liked})
}

func (h *ExperienceHandler) ToggleBookmark(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	id := c.Param("id")

	bookmarked, err := h.bookRepo.Toggle(c.Request.Context(), userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to toggle bookmark"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"bookmarked": bookmarked})
}

func parseIntParam(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

// MyExperiences — 用户自己发布的经验
func (h *ExperienceHandler) MyExperiences(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}

	page := parseIntParam(c.Query("page"), 1)
	pageSize := parseIntParam(c.Query("page_size"), 20)

	experiences, total, err := h.repo.ListByAuthor(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list experiences"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  experiences,
		"total": total,
		"page":  page,
	})
}

// MyBookmarks — 用户收藏的经验
func (h *ExperienceHandler) MyBookmarks(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}

	page := parseIntParam(c.Query("page"), 1)
	pageSize := parseIntParam(c.Query("page_size"), 20)

	experiences, total, err := h.repo.ListBookmarked(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list bookmarks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  experiences,
		"total": total,
		"page":  page,
	})
}

// GetRecommendations returns personalized experience recommendations.
// Ranks by domain preference (publish×2 + bookmark×1) × hotness.
func (h *ExperienceHandler) GetRecommendations(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}

	limit := parseIntParam(c.Query("limit"), 20)
	offset := parseIntParam(c.Query("offset"), 0)

	experiences, err := h.repo.Recommend(c.Request.Context(), userID, limit, offset)
	if err != nil {
		log.Printf("ERROR Recommend: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get recommendations"})
		return
	}

	// Round-robin by creator to prevent consecutive same creator
	experiences = interleaveByCreator(experiences)

	c.JSON(http.StatusOK, gin.H{
		"data":  experiences,
		"total": len(experiences),
	})
}

// interleaveByCreator distributes experiences round-robin by creator.
// Items are grouped by creator, then taken one from each bucket in random order.
// Within each bucket, items keep their original order (already sorted by score).
func interleaveByCreator(experiences []model.Experience) []model.Experience {
	if len(experiences) <= 1 {
		return experiences
	}

	// 1. Group by creator
	buckets := make(map[string][]model.Experience)
	var creatorOrder []string
	for _, e := range experiences {
		creator := ""
		if e.CreatorName != nil {
			creator = *e.CreatorName
		}
		if _, ok := buckets[creator]; !ok {
			creatorOrder = append(creatorOrder, creator)
		}
		buckets[creator] = append(buckets[creator], e)
	}

	if len(buckets) <= 1 {
		return experiences
	}

	// 2. Shuffle creator order for variety
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(creatorOrder), func(i, j int) {
		creatorOrder[i], creatorOrder[j] = creatorOrder[j], creatorOrder[i]
	})

	// 3. Round-robin: take one from each bucket
	result := make([]model.Experience, 0, len(experiences))
	indices := make(map[string]int)
	for {
		added := false
		for _, creator := range creatorOrder {
			bucket := buckets[creator]
			if indices[creator] < len(bucket) {
				result = append(result, bucket[indices[creator]])
				indices[creator]++
				added = true
			}
		}
		if !added {
			break
		}
	}

	return result
}

