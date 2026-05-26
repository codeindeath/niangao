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

type fakeV4MeStatsStore struct {
	gotUserID string
	gotRange  string
	gotLimit  int
	fail      bool
}

func (f *fakeV4MeStatsStore) AssetStats(ctx context.Context, userID string) (*model.AssetStats, error) {
	f.gotUserID = userID
	if f.fail {
		return nil, errors.New("store failed")
	}
	return &model.AssetStats{MyExperiences: 7, Collections: 3, MonthAdded: 2}, nil
}

func (f *fakeV4MeStatsStore) ContributionStats(ctx context.Context, userID string) (*model.ContributionStats, error) {
	f.gotUserID = userID
	if f.fail {
		return nil, errors.New("store failed")
	}
	return &model.ContributionStats{InspiredUsers: 12, CollectedCount: 5, MonthInspiredUsers: 4, MonthCollected: 2}, nil
}

func (f *fakeV4MeStatsStore) ChangeStats(ctx context.Context, userID string) (*model.ChangeStats, error) {
	f.gotUserID = userID
	if f.fail {
		return nil, errors.New("store failed")
	}
	return &model.ChangeStats{ChatTopics: 6, ClearerCount: 1, MonthChatExperiences: 2}, nil
}

func (f *fakeV4MeStatsStore) RecentHarvestStats(ctx context.Context, userID string, rangeKey string) (*model.RecentHarvestStats, error) {
	f.gotUserID = userID
	f.gotRange = rangeKey
	if f.fail {
		return nil, errors.New("store failed")
	}
	return &model.RecentHarvestStats{Range: rangeKey, NoteAdded: 3, ChatExperiences: 2, InspiredUsers: 9, CollectedCount: 4}, nil
}

func (f *fakeV4MeStatsStore) RecentRespondedExperiences(ctx context.Context, userID string, limit int) ([]model.RespondedExperienceCard, error) {
	f.gotUserID = userID
	f.gotLimit = limit
	if f.fail {
		return nil, errors.New("store failed")
	}
	return []model.RespondedExperienceCard{{
		ID:               "exp-1",
		Content:          "把想证明自己没错，换成先把关系说清楚。",
		Domain:           string(model.DomainRelationship),
		SubDomain:        string(model.SubRomance),
		StarRating:       4,
		InspirationCount: 9,
		CollectionCount:  3,
	}}, nil
}

func TestV4MeStatsRequireAuth(t *testing.T) {
	tests := []string{
		"/api/v1/me/stats/assets",
		"/api/v1/me/stats/contribution",
		"/api/v1/me/stats/change",
		"/api/v1/me/stats/recent-harvest",
		"/api/v1/me/recent-responded-experiences",
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			r := gin.New()
			RegisterMeStatsRoutes(r.Group("/api/v1"), &fakeV4MeStatsStore{})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", path, nil)
			r.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestV4MeStatsUseAuthenticatedUserAndReturnShapes(t *testing.T) {
	tests := []struct {
		name string
		path string
		key  string
	}{
		{"assets", "/api/v1/me/stats/assets", "my_experiences"},
		{"contribution", "/api/v1/me/stats/contribution", "inspired_users"},
		{"change", "/api/v1/me/stats/change", "chat_topics"},
		{"recent harvest", "/api/v1/me/stats/recent-harvest?range=7d", "note_added"},
		{"recent responded", "/api/v1/me/recent-responded-experiences?limit=2", "data"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			store := &fakeV4MeStatsStore{}
			v1 := r.Group("/api/v1", func(c *gin.Context) {
				c.Set("user_id", "user-1")
			})
			RegisterMeStatsRoutes(v1, store)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", tt.path, nil)
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
			}
			if store.gotUserID != "user-1" {
				t.Fatalf("userID = %q, want user-1", store.gotUserID)
			}

			var body map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if _, ok := body[tt.key]; !ok {
				t.Fatalf("response missing %q: %+v", tt.key, body)
			}
		})
	}
}

func TestV4MeStatsRecentHarvestValidatesRange(t *testing.T) {
	r := gin.New()
	v1 := r.Group("/api/v1", func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	RegisterMeStatsRoutes(v1, &fakeV4MeStatsStore{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/me/stats/recent-harvest?range=14d", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", w.Code, w.Body.String())
	}
}

func TestV4MeStatsRecentRespondedClampsLimit(t *testing.T) {
	r := gin.New()
	store := &fakeV4MeStatsStore{}
	v1 := r.Group("/api/v1", func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	RegisterMeStatsRoutes(v1, store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/me/recent-responded-experiences?limit=99", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
	}
	if store.gotLimit != 10 {
		t.Fatalf("limit = %d, want 10", store.gotLimit)
	}
}

func TestV4MeStatsFailureReturnsServerError(t *testing.T) {
	r := gin.New()
	v1 := r.Group("/api/v1", func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	RegisterMeStatsRoutes(v1, &fakeV4MeStatsStore{fail: true})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/me/stats/assets", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500: %s", w.Code, w.Body.String())
	}
}
