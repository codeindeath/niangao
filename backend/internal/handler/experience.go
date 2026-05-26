package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niangao/backend/internal/middleware"
	"github.com/niangao/backend/internal/model"
	"github.com/niangao/backend/internal/repository"
)

type ExperienceHandler struct {
	repo      *repository.ExperienceRepo
	aiGateway ExperienceAIGateway
}

type ExperienceAIGateway interface {
	RewriteExperience(ctx context.Context, req model.ExperienceRewriteGatewayRequest) (*model.ExperienceRewriteGatewayResponse, error)
}

func RegisterExperienceRoutes(r *gin.RouterGroup, expRepo *repository.ExperienceRepo, gateways ...ExperienceAIGateway) {
	var aiGateway ExperienceAIGateway
	if len(gateways) > 0 {
		aiGateway = gateways[0]
	}
	h := &ExperienceHandler{repo: expRepo, aiGateway: aiGateway}

	exp := r.Group("/experiences")
	{
		exp.GET("", deprecatedMobileEndpoint)
		exp.GET("/recommend", deprecatedMobileEndpoint)
		exp.GET("/:id", h.Get)
		exp.POST("/rewrite", middleware.RequireAuth(), h.Rewrite)
		exp.POST("", middleware.RequireAuth(), h.Create)
		exp.PUT("/:id", middleware.RequireAuth(), h.Update)
		exp.DELETE("/:id", middleware.RequireAuth(), h.Delete)
		exp.POST("/:id/like", deprecatedMobileEndpoint)
		exp.POST("/:id/bookmark", deprecatedMobileEndpoint)
	}

	// 个人维度 API — 直接在 v1 下注册，不走子 Group
	r.GET("/me/experiences", deprecatedMobileEndpoint)
	r.GET("/me/bookmarks", deprecatedMobileEndpoint)
}

func (h *ExperienceHandler) Rewrite(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	var req model.ExperienceRewriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rewrite payload"})
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
		return
	}
	if len([]rune(req.Content)) > 2000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is too long"})
		return
	}
	if req.Source == "" {
		req.Source = "manual_note"
	}
	if req.Source != "manual_note" && req.Source != "chat_note" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid source"})
		return
	}
	if req.DefaultVisibility == "" {
		if req.Source == "chat_note" {
			req.DefaultVisibility = model.VisibilityPrivate
		} else {
			req.DefaultVisibility = model.VisibilityPublic
		}
	}
	if !model.IsValidVisibility(req.DefaultVisibility) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid default_visibility"})
		return
	}
	if req.UserSelectedDomain != "" && !model.IsValidDomain(req.UserSelectedDomain) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_selected_domain"})
		return
	}
	if req.UserSelectedSubDomain != "" && !model.IsValidSubDomain(req.UserSelectedSubDomain) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_selected_sub_domain"})
		return
	}
	if req.UserSelectedDomain != "" && req.UserSelectedSubDomain != "" &&
		!model.SubDomainBelongsToParent(req.UserSelectedDomain, req.UserSelectedSubDomain) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sub_domain does not belong to domain"})
		return
	}
	if h.aiGateway == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "rewrite service unavailable", "retryable": true})
		return
	}

	result, err := h.aiGateway.RewriteExperience(c.Request.Context(), model.ExperienceRewriteGatewayRequest{
		UserID:                userID,
		Source:                req.Source,
		RawText:               req.Content,
		SourceMessageIDs:      req.SourceMessageIDs,
		DefaultVisibility:     req.DefaultVisibility,
		UserSelectedDomain:    req.UserSelectedDomain,
		UserSelectedSubDomain: req.UserSelectedSubDomain,
		TopicContext:          strings.TrimSpace(req.TopicContext),
	})
	if err != nil || result == nil {
		log.Printf("v4 rewrite gateway failed user=%s: %v", userID, err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "rewrite service unavailable", "retryable": true})
		return
	}
	result.RewrittenContent = strings.TrimSpace(result.RewrittenContent)
	if result.CanRewrite && (result.RewrittenContent == "" || len([]rune(result.RewrittenContent)) > 100) {
		c.JSON(http.StatusBadGateway, gin.H{"error": "invalid rewrite output", "retryable": true})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"can_rewrite":         result.CanRewrite,
		"rewritten_content":   result.RewrittenContent,
		"domain":              result.Domain,
		"sub_domain":          result.SubDomain,
		"topic":               result.Topic,
		"rewrite_level":       result.RewriteLevel,
		"source_preservation": result.SourcePreservation,
		"needs_user_edit":     result.NeedsUserEdit,
		"reason":              result.Reason,
		"confidence":          result.Confidence,
		"warnings":            result.Warnings,
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

	c.JSON(http.StatusOK, toExperienceDetailResponse(exp))
}

type experienceDetailResponse struct {
	ID                      string     `json:"id"`
	OwnerUserID             *string    `json:"owner_user_id,omitempty"`
	Content                 string     `json:"content"`
	Interpretation          *string    `json:"interpretation,omitempty"`
	Domain                  string     `json:"domain"`
	SubDomain               *string    `json:"sub_domain,omitempty"`
	Topic                   string     `json:"topic,omitempty"`
	ExperienceType          string     `json:"experience_type,omitempty"`
	Visibility              string     `json:"visibility,omitempty"`
	LifecycleStatus         string     `json:"lifecycle_status,omitempty"`
	SourceLabel             *string    `json:"source_label,omitempty"`
	CreatorDisplayName      string     `json:"creator_display_name,omitempty"`
	ScoreReason             *string    `json:"score_reason,omitempty"`
	InspirationCount        int        `json:"inspiration_count"`
	CollectionCount         int        `json:"collection_count"`
	AuthorAvatar            *string    `json:"author_avatar,omitempty"`
	AuthorTitle             *string    `json:"author_title,omitempty"`
	IsInspired              bool       `json:"is_inspired"`
	IsCollected             bool       `json:"is_collected"`
	QualityTier             string     `json:"quality_tier,omitempty"`
	QualityScore            *float64   `json:"quality_score,omitempty"`
	ScoreDetails            *string    `json:"score_details,omitempty"`
	OriginalText            *string    `json:"original_text,omitempty"`
	InterpretationGenerated bool       `json:"interpretation_generated"`
	InterpretationStatus    string     `json:"interpretation_status,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               *time.Time `json:"updated_at,omitempty"`
}

func toExperienceDetailResponse(exp *model.Experience) experienceDetailResponse {
	topic := exp.Topic
	if topic == "" {
		topic = exp.Topics
	}
	creatorDisplayName := firstNonEmptyPtr(exp.CreatorDisplayName, exp.CreatorName, exp.AuthorName)

	return experienceDetailResponse{
		ID:                      exp.ID,
		OwnerUserID:             exp.OwnerUserID,
		Content:                 exp.Content,
		Interpretation:          exp.Interpretation,
		Domain:                  string(exp.Domain),
		SubDomain:               exp.SubDomain,
		Topic:                   topic,
		ExperienceType:          exp.ExperienceType,
		Visibility:              exp.Visibility,
		LifecycleStatus:         exp.LifecycleStatus,
		SourceLabel:             exp.SourceLabel,
		CreatorDisplayName:      creatorDisplayName,
		ScoreReason:             exp.ScoreReason,
		InspirationCount:        exp.InspirationCount,
		CollectionCount:         exp.CollectionCount,
		AuthorAvatar:            exp.AuthorAvatar,
		AuthorTitle:             exp.AuthorTitle,
		IsInspired:              exp.IsLiked,
		IsCollected:             exp.IsBookmarked,
		QualityTier:             exp.QualityTier,
		QualityScore:            exp.QualityScore,
		ScoreDetails:            exp.ScoreDetails,
		OriginalText:            exp.OriginalText,
		InterpretationGenerated: exp.InterpretationGenerated,
		InterpretationStatus:    exp.InterpretationStatus,
		CreatedAt:               exp.CreatedAt,
		UpdatedAt:               &exp.UpdatedAt,
	}
}

func firstNonEmptyPtr(values ...*string) string {
	for _, value := range values {
		if value == nil {
			continue
		}
		trimmed := strings.TrimSpace(*value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func (h *ExperienceHandler) Create(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}

	var req model.CreateExperienceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入经验内容"})
		return
	}

	if err := normalizeCreateExperienceRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Visibility == model.VisibilityPublic {
		displayName, err := h.repo.GetUserDisplayName(c.Request.Context(), userID)
		if err != nil {
			log.Printf("display name gate failed user=%s: %v", userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check display name"})
			return
		}
		if strings.TrimSpace(displayName) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
				"code":    "display_name_required",
				"message": "需要先设置展示名",
			}})
			return
		}
	}

	reviewStatus := string(model.ReviewPending)
	if req.IsPrivate {
		reviewStatus = string(model.ReviewPrivate)
	}

	exp, err := h.repo.CreateWithReview(c.Request.Context(), userID, req,
		reviewStatus, nil, nil, nil, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save experience"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"experience": toExperienceDetailResponse(exp)})
}

func normalizeCreateExperienceRequest(req *model.CreateExperienceRequest) error {
	req.Content = strings.TrimSpace(req.Content)
	contentRunes := len([]rune(req.Content))
	if contentRunes < 1 || contentRunes > 100 {
		return errors.New("经验内容需 1-100 字")
	}

	if req.Visibility == "" {
		if req.IsPrivate {
			req.Visibility = model.VisibilityPrivate
		} else {
			req.Visibility = model.VisibilityPublic
		}
	}
	if !model.IsValidVisibility(req.Visibility) {
		return errors.New("invalid visibility")
	}
	req.IsPrivate = req.Visibility == model.VisibilityPrivate

	if req.SourceScene == "" {
		req.SourceScene = string(model.SourceSceneNote)
	}
	if req.SourceScene != string(model.SourceSceneNote) && req.SourceScene != string(model.SourceSceneChat) {
		return errors.New("invalid source_scene")
	}
	req.SourceChatTopicID = strings.TrimSpace(req.SourceChatTopicID)
	req.SourceChatMessageID = strings.TrimSpace(req.SourceChatMessageID)
	req.SourceChatMessageSnapshot = strings.TrimSpace(req.SourceChatMessageSnapshot)
	req.SourceMessageIDs = compactSourceMessageIDs(req.SourceMessageIDs)
	if req.SourceChatTopicID != "" && !isUUIDLike(req.SourceChatTopicID) {
		return errors.New("invalid source_chat_topic_id")
	}
	if req.SourceChatMessageID != "" && !isUUIDLike(req.SourceChatMessageID) {
		return errors.New("invalid source_chat_message_id")
	}
	if req.SourceScene == string(model.SourceSceneChat) && req.SourceChatMessageID == "" {
		for i := len(req.SourceMessageIDs) - 1; i >= 0; i-- {
			if isUUIDLike(req.SourceMessageIDs[i]) {
				req.SourceChatMessageID = req.SourceMessageIDs[i]
				break
			}
		}
	}
	if req.SourceScene == string(model.SourceSceneChat) &&
		req.SourceChatMessageSnapshot == "" &&
		len(req.SourceMessageIDs) > 0 {
		req.SourceChatMessageSnapshot = strings.Join(req.SourceMessageIDs, ",")
	}

	if req.Topic == "" && req.Topics != "" {
		req.Topic = req.Topics
	}
	if req.Topics == "" && req.Topic != "" {
		req.Topics = req.Topic
	}
	if len([]rune(req.Topic)) > 200 || len([]rune(req.Topics)) > 200 {
		return errors.New("话题不超过 200 字")
	}

	if req.Interpretation != "" && len([]rune(req.Interpretation)) > 300 {
		return errors.New("经验解读不超过 300 字")
	}
	if req.Domain != "" && !model.IsValidDomain(req.Domain) {
		return errors.New("invalid domain")
	}
	if req.SubDomain != "" && !model.IsValidSubDomain(req.SubDomain) {
		return errors.New("invalid sub_domain")
	}
	if req.Domain == "" && req.SubDomain != "" {
		return errors.New("sub_domain requires domain")
	}
	if req.Domain != "" && req.SubDomain != "" && !model.SubDomainBelongsToParent(req.Domain, req.SubDomain) {
		return errors.New("sub_domain does not belong to domain")
	}
	return nil
}

func compactSourceMessageIDs(ids []string) []string {
	if len(ids) == 0 {
		return nil
	}
	compacted := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" {
			compacted = append(compacted, id)
		}
	}
	return compacted
}

func isUUIDLike(value string) bool {
	if len(value) != 36 {
		return false
	}
	for i, r := range value {
		switch i {
		case 8, 13, 18, 23:
			if r != '-' {
				return false
			}
		default:
			if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
				return false
			}
		}
	}
	return true
}

func (h *ExperienceHandler) Update(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	id := c.Param("id")

	var req model.CreateExperienceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入经验内容"})
		return
	}

	if err := normalizeCreateExperienceRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
