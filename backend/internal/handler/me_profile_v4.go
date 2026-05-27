package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/niangao/backend/internal/middleware"
	"github.com/niangao/backend/internal/model"
)

type V4MeProfileStore interface {
	MeProfile(ctx context.Context, userID string) (*model.MeProfile, error)
	UpdateMeProfile(ctx context.Context, userID string, patch model.MeProfilePatch) (*model.MeProfile, error)
}

type MeProfileHandler struct {
	store V4MeProfileStore
}

func RegisterMeProfileRoutes(r *gin.RouterGroup, store V4MeProfileStore) {
	h := &MeProfileHandler{store: store}
	me := r.Group("/me", middleware.RequireAuth())
	{
		me.GET("/profile", h.Get)
		me.PATCH("/profile", h.Update)
	}
}

func (h *MeProfileHandler) Get(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}
	profile, err := h.store.MeProfile(c.Request.Context(), userID)
	if err != nil {
		log.Printf("v4 me profile failed user=%s: %v", userID, err)
		respondError(c, http.StatusInternalServerError, "profile_load_failed", "failed to load profile")
		return
	}
	c.JSON(http.StatusOK, profile)
}

func (h *MeProfileHandler) Update(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}

	var patch model.MeProfilePatch
	if err := c.ShouldBindJSON(&patch); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_profile_payload", "invalid profile payload")
		return
	}
	if err := normalizeMeProfilePatch(&patch); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_profile", err.Error())
		return
	}

	profile, err := h.store.UpdateMeProfile(c.Request.Context(), userID, patch)
	if err != nil {
		log.Printf("v4 me profile update failed user=%s: %v", userID, err)
		respondError(c, http.StatusInternalServerError, "profile_update_failed", "failed to update profile")
		return
	}
	c.JSON(http.StatusOK, profile)
}

func normalizeMeProfilePatch(patch *model.MeProfilePatch) error {
	if patch.DisplayName != nil {
		v := strings.TrimSpace(*patch.DisplayName)
		if len([]rune(v)) > 30 {
			return errors.New("展示名不超过 30 字")
		}
		patch.DisplayName = &v
	}
	if patch.FreeDescription != nil {
		v := strings.TrimSpace(*patch.FreeDescription)
		if len([]rune(v)) > 200 {
			return errors.New("个人描述不超过 200 字")
		}
		patch.FreeDescription = &v
	}
	if patch.CommonIssues != nil {
		issues := make([]string, 0, len(*patch.CommonIssues))
		for _, raw := range *patch.CommonIssues {
			issue := strings.TrimSpace(raw)
			if issue == "" {
				continue
			}
			if len([]rune(issue)) > 20 {
				return errors.New("常见困扰单项不超过 20 字")
			}
			issues = append(issues, issue)
		}
		patch.CommonIssues = &issues
	}
	return nil
}
