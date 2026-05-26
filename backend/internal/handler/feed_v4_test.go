package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/niangao/backend/internal/middleware"
	"github.com/niangao/backend/internal/model"
)

type fakeV4FeedStore struct {
	wantUserID string
	gotUserID  string
	gotLimit   int
	gotCursor  string
	fail       bool
}

func (f *fakeV4FeedStore) RecommendFeed(ctx context.Context, userID string, limit int, cursor string) (*model.FeedPage, error) {
	return f.feedPage(userID, limit, cursor)
}

func (f *fakeV4FeedStore) CollectionsFeed(ctx context.Context, userID string, limit int, cursor string) (*model.FeedPage, error) {
	return f.feedPage(userID, limit, cursor)
}

func (f *fakeV4FeedStore) MineFeed(ctx context.Context, userID string, limit int, cursor string) (*model.FeedPage, error) {
	return f.feedPage(userID, limit, cursor)
}

func (f *fakeV4FeedStore) feedPage(userID string, limit int, cursor string) (*model.FeedPage, error) {
	f.gotUserID = userID
	f.gotLimit = limit
	f.gotCursor = cursor
	if f.fail {
		return nil, errors.New("store failed")
	}
	if f.wantUserID != "" && userID != f.wantUserID {
		return nil, errors.New("wrong user id")
	}
	return &model.FeedPage{
		Data: []model.ExperienceCard{
			{
				ID:                             "exp-1",
				Content:                        "重要决定放到早上做",
				ExperienceType:                 string(model.ExperienceTypeUserOriginal),
				Visibility:                     string(model.VisibilityPublic),
				LifecycleStatus:                string(model.LifecycleActive),
				Domain:                         string(model.DomainCognition),
				SubDomain:                      string(model.SubThinking),
				Topic:                          "#判断",
				CreatorDisplayName:             "阿树",
				InterpretationStatus:           string(model.InterpretationReady),
				InterpretationSummaryAvailable: true,
				QualityTier:                    string(model.QualityTierAICitable),
				StarRating:                     4,
				InspirationCount:               12,
				CollectionCount:                8,
				IsCollected:                    false,
				IsInspired:                     false,
			},
		},
		NextCursor: "cursor-2",
		SessionID:  "session-1",
		HasMore:    true,
	}, nil
}

func TestV4RecommendFeedAllowsGuestAndReturnsEnvelope(t *testing.T) {
	r := gin.New()
	store := &fakeV4FeedStore{}
	RegisterFeedRoutes(r.Group("/api/v1"), store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/feed/recommend?limit=15&cursor=abc", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
	}
	if store.gotUserID != "" {
		t.Fatalf("guest recommend userID = %q, want empty", store.gotUserID)
	}
	if store.gotLimit != 15 {
		t.Fatalf("limit = %d, want 15", store.gotLimit)
	}
	if store.gotCursor != "abc" {
		t.Fatalf("cursor = %q, want abc", store.gotCursor)
	}

	var body model.FeedPage
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Data) != 1 {
		t.Fatalf("data length = %d, want 1", len(body.Data))
	}
	if body.NextCursor != "cursor-2" || body.SessionID != "session-1" || !body.HasMore {
		t.Fatalf("feed envelope = %+v", body)
	}
}

func TestV4PrivateFeedsRequireAuth(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"collections", "/api/v1/feed/collections"},
		{"mine", "/api/v1/feed/mine"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			RegisterFeedRoutes(r.Group("/api/v1"), &fakeV4FeedStore{})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", tt.path, nil)
			r.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestV4PrivateFeedsUseAuthenticatedUser(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"collections", "/api/v1/feed/collections"},
		{"mine", "/api/v1/feed/mine"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			store := &fakeV4FeedStore{wantUserID: "user-1"}
			v1 := r.Group("/api/v1", func(c *gin.Context) {
				c.Set("user_id", "user-1")
			})
			RegisterFeedRoutes(v1, store)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", tt.path+"?limit=1000", nil)
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
			}
			if store.gotUserID != "user-1" {
				t.Fatalf("userID = %q, want user-1", store.gotUserID)
			}
			if store.gotLimit != maxFeedLimit {
				t.Fatalf("limit = %d, want maxFeedLimit %d", store.gotLimit, maxFeedLimit)
			}
		})
	}
}

func TestV4FeedStoreFailureReturnsServerError(t *testing.T) {
	r := gin.New()
	v1 := r.Group("/api/v1", func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	RegisterFeedRoutes(v1, &fakeV4FeedStore{fail: true})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/feed/mine", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500: %s", w.Code, w.Body.String())
	}
}

func TestV4FeedRoutesDoNotRequireAuthForRecommend(t *testing.T) {
	r := gin.New()
	v1 := r.Group("/api/v1")
	v1.Use(middleware.AuthMiddleware("unused", nil))
	RegisterFeedRoutes(v1, &fakeV4FeedStore{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/feed/recommend", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 for guest recommend", w.Code)
	}
}
