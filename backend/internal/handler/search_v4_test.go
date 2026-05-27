package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/niangao/backend/internal/model"
)

type fakeV4SearchStore struct {
	gotUserID string
	gotQuery  string
	gotLimit  int
	gotCursor string
	fail      bool
}

func (f *fakeV4SearchStore) SearchExperiences(ctx context.Context, userID string, query string, limit int, cursor string) (*model.FeedPage, error) {
	f.gotUserID = userID
	f.gotQuery = query
	f.gotLimit = limit
	f.gotCursor = cursor
	if f.fail {
		return nil, errors.New("store failed")
	}
	return &model.FeedPage{
		Data: []model.ExperienceCard{{
			ID:                 "exp-1",
			Content:            "先完成最小版本",
			ExperienceType:     string(model.ExperienceTypePlatformSelected),
			Visibility:         string(model.VisibilityPublic),
			LifecycleStatus:    string(model.LifecycleActive),
			Domain:             string(model.DomainWork),
			SubDomain:          string(model.SubStartup),
			Topic:              "#行动",
			CreatorDisplayName: "Paul Graham",
			QualityTier:        string(model.QualityTierAICitable),
			StarRating:         4,
		}},
		NextCursor: "20",
		HasMore:    true,
	}, nil
}

func TestV4SearchAllowsGuestAndReturnsEnvelope(t *testing.T) {
	r := gin.New()
	store := &fakeV4SearchStore{}
	RegisterSearchRoutes(r.Group("/api/v1"), store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/search/experiences?q=%20行动%20&limit=12&cursor=8", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
	}
	if store.gotUserID != "" || store.gotQuery != "行动" || store.gotLimit != 12 || store.gotCursor != "8" {
		t.Fatalf("store got user=%q query=%q limit=%d cursor=%q", store.gotUserID, store.gotQuery, store.gotLimit, store.gotCursor)
	}

	var body model.FeedPage
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Data) != 1 || body.NextCursor != "20" || !body.HasMore {
		t.Fatalf("body = %+v", body)
	}
}

func TestV4SearchEmptyQueryReturnsEmptyPage(t *testing.T) {
	r := gin.New()
	store := &fakeV4SearchStore{}
	RegisterSearchRoutes(r.Group("/api/v1"), store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/search/experiences?q=%20%20", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
	}
	if store.gotQuery != "" {
		t.Fatalf("empty query should not call store, got query %q", store.gotQuery)
	}

	var body model.FeedPage
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Data) != 0 || body.HasMore || body.NextCursor != "" {
		t.Fatalf("body = %+v", body)
	}
}

func TestV4SearchUsesAuthenticatedUserWhenPresent(t *testing.T) {
	r := gin.New()
	store := &fakeV4SearchStore{}
	v1 := r.Group("/api/v1", func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	RegisterSearchRoutes(v1, store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/search/experiences?q=self", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
	}
	if store.gotUserID != "user-1" {
		t.Fatalf("userID = %q, want user-1", store.gotUserID)
	}
}

func TestV4SearchFailureReturnsServerError(t *testing.T) {
	r := gin.New()
	RegisterSearchRoutes(r.Group("/api/v1"), &fakeV4SearchStore{fail: true})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/search/experiences?q=anything", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500: %s", w.Code, w.Body.String())
	}
	assertStructuredErrorMessage(t, w, "暂时搜索不了，请稍后再试")
}
