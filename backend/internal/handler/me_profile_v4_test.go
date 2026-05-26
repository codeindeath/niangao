package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/niangao/backend/internal/model"
)

type fakeV4MeProfileStore struct {
	gotUserID string
	gotPatch  model.MeProfilePatch
	fail      bool
}

func (f *fakeV4MeProfileStore) MeProfile(ctx context.Context, userID string) (*model.MeProfile, error) {
	f.gotUserID = userID
	if f.fail {
		return nil, errors.New("store failed")
	}
	return &model.MeProfile{DisplayName: "阿树", CareerStage: "稳定工作", CommonIssues: []string{"自我怀疑"}}, nil
}

func (f *fakeV4MeProfileStore) UpdateMeProfile(ctx context.Context, userID string, patch model.MeProfilePatch) (*model.MeProfile, error) {
	f.gotUserID = userID
	f.gotPatch = patch
	if f.fail {
		return nil, errors.New("store failed")
	}
	return &model.MeProfile{DisplayName: valueOr(patch.DisplayName, "阿树")}, nil
}

func TestV4MeProfileRequiresAuth(t *testing.T) {
	tests := []struct {
		method string
		path   string
		body   string
	}{
		{"GET", "/api/v1/me/profile", ""},
		{"PATCH", "/api/v1/me/profile", `{"display_name":"阿树"}`},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			r := gin.New()
			RegisterMeProfileRoutes(r.Group("/api/v1"), &fakeV4MeProfileStore{})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestV4MeProfileGetAndPatch(t *testing.T) {
	r := gin.New()
	store := &fakeV4MeProfileStore{}
	v1 := r.Group("/api/v1", func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	RegisterMeProfileRoutes(v1, store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/me/profile", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want 200: %s", w.Code, w.Body.String())
	}
	var profile model.MeProfile
	if err := json.Unmarshal(w.Body.Bytes(), &profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if profile.DisplayName != "阿树" || store.gotUserID != "user-1" {
		t.Fatalf("profile=%+v user=%q", profile, store.gotUserID)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest("PATCH", "/api/v1/me/profile", strings.NewReader(`{"display_name":"新名字","common_issues":["未来迷茫"]}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("PATCH status = %d, want 200: %s", w.Code, w.Body.String())
	}
	if store.gotPatch.DisplayName == nil || *store.gotPatch.DisplayName != "新名字" {
		t.Fatalf("display_name patch = %+v", store.gotPatch.DisplayName)
	}
	if store.gotPatch.CommonIssues == nil || len(*store.gotPatch.CommonIssues) != 1 {
		t.Fatalf("common_issues patch = %+v", store.gotPatch.CommonIssues)
	}
}

func TestV4MeProfileRejectsLongDisplayName(t *testing.T) {
	r := gin.New()
	v1 := r.Group("/api/v1", func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	RegisterMeProfileRoutes(v1, &fakeV4MeProfileStore{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/api/v1/me/profile", strings.NewReader(`{"display_name":"`+strings.Repeat("名", 31)+`"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", w.Code, w.Body.String())
	}
}

func valueOr(v *string, fallback string) string {
	if v == nil {
		return fallback
	}
	return *v
}
