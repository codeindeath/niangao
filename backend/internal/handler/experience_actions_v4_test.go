package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/niangao/backend/internal/repository"
)

type fakeV4ExperienceActionStore struct {
	userID        string
	experienceID  string
	eventType     string
	sourceContext string
	metadata      map[string]any
	already       bool
	fail          bool
	unavailable   bool
}

func (f *fakeV4ExperienceActionStore) InspireExperience(ctx context.Context, userID string, experienceID string) (bool, error) {
	f.userID = userID
	f.experienceID = experienceID
	if f.fail {
		return false, errors.New("store failed")
	}
	if f.unavailable {
		return false, repository.ErrExperienceUnavailable
	}
	return f.already, nil
}

func (f *fakeV4ExperienceActionStore) CollectExperience(ctx context.Context, userID string, experienceID string) (bool, error) {
	f.userID = userID
	f.experienceID = experienceID
	if f.fail {
		return false, errors.New("store failed")
	}
	if f.unavailable {
		return false, repository.ErrExperienceUnavailable
	}
	return f.already, nil
}

func (f *fakeV4ExperienceActionStore) UncollectExperience(ctx context.Context, userID string, experienceID string) error {
	f.userID = userID
	f.experienceID = experienceID
	if f.fail {
		return errors.New("store failed")
	}
	if f.unavailable {
		return repository.ErrExperienceUnavailable
	}
	return nil
}

func (f *fakeV4ExperienceActionStore) RecordExperienceEvent(ctx context.Context, userID string, experienceID string, event ExperienceEventRequest) error {
	f.userID = userID
	f.experienceID = experienceID
	f.eventType = event.EventType
	f.sourceContext = event.SourceContext
	f.metadata = event.Metadata
	if f.fail {
		return errors.New("store failed")
	}
	if f.unavailable {
		return repository.ErrExperienceUnavailable
	}
	return nil
}

func TestV4ExperienceActionsRequireAuth(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"inspire", "POST", "/api/v1/experiences/exp-1/inspire"},
		{"collect", "POST", "/api/v1/experiences/exp-1/collect"},
		{"uncollect", "DELETE", "/api/v1/experiences/exp-1/collect"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			RegisterExperienceActionRoutes(r.Group("/api/v1"), &fakeV4ExperienceActionStore{})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			r.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestV4InspireCreatesOneWayFeedback(t *testing.T) {
	r := gin.New()
	store := &fakeV4ExperienceActionStore{}
	v1 := r.Group("/api/v1", func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	RegisterExperienceActionRoutes(v1, store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/experiences/exp-1/inspire", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
	}
	if store.userID != "user-1" || store.experienceID != "exp-1" {
		t.Fatalf("store got user=%q exp=%q", store.userID, store.experienceID)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["inspired"] != true {
		t.Fatalf("inspired = %v, want true", body["inspired"])
	}
}

func TestV4InspireAlreadyInspiredReturnsConflictButKeepsState(t *testing.T) {
	r := gin.New()
	v1 := r.Group("/api/v1", func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	RegisterExperienceActionRoutes(v1, &fakeV4ExperienceActionStore{already: true})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/experiences/exp-1/inspire", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409: %s", w.Code, w.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["inspired"] != true || body["code"] != "already_inspired" {
		t.Fatalf("body = %+v", body)
	}
}

func TestV4CollectAndUncollect(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantKey    string
	}{
		{"collect", "POST", "/api/v1/experiences/exp-1/collect", http.StatusOK, "collected"},
		{"uncollect", "DELETE", "/api/v1/experiences/exp-1/collect", http.StatusOK, "collected"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			store := &fakeV4ExperienceActionStore{}
			v1 := r.Group("/api/v1", func(c *gin.Context) {
				c.Set("user_id", "user-1")
			})
			RegisterExperienceActionRoutes(v1, store)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d: %s", w.Code, tt.wantStatus, w.Body.String())
			}
			if store.userID != "user-1" || store.experienceID != "exp-1" {
				t.Fatalf("store got user=%q exp=%q", store.userID, store.experienceID)
			}

			var body map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if _, ok := body[tt.wantKey]; !ok {
				t.Fatalf("response missing %q: %+v", tt.wantKey, body)
			}
		})
	}
}

func TestV4ExperienceActionFailureReturnsServerError(t *testing.T) {
	r := gin.New()
	v1 := r.Group("/api/v1", func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	RegisterExperienceActionRoutes(v1, &fakeV4ExperienceActionStore{fail: true})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/experiences/exp-1/collect", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500: %s", w.Code, w.Body.String())
	}
}

func TestV4ExperienceActionUnavailableReturnsNotFound(t *testing.T) {
	r := gin.New()
	v1 := r.Group("/api/v1", func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	RegisterExperienceActionRoutes(v1, &fakeV4ExperienceActionStore{unavailable: true})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/experiences/exp-1/inspire", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404: %s", w.Code, w.Body.String())
	}
}

func TestV4ExperienceEventAllowsGuestSearchClick(t *testing.T) {
	r := gin.New()
	store := &fakeV4ExperienceActionStore{}
	RegisterExperienceActionRoutes(r.Group("/api/v1"), store)

	body := []byte(`{"event_type":"search_click","source_context":"search","metadata":{"query":"姜文","rank":0}}`)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/experiences/exp-1/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204: %s", w.Code, w.Body.String())
	}
	if store.userID != "" || store.experienceID != "exp-1" || store.eventType != "search_click" || store.sourceContext != "search" {
		t.Fatalf("store got user=%q exp=%q event=%q source=%q", store.userID, store.experienceID, store.eventType, store.sourceContext)
	}
	if store.metadata["query"] != "姜文" || store.metadata["rank"].(float64) != 0 {
		t.Fatalf("metadata = %+v", store.metadata)
	}
}

func TestV4ExperienceEventRejectsActionEvents(t *testing.T) {
	r := gin.New()
	RegisterExperienceActionRoutes(r.Group("/api/v1"), &fakeV4ExperienceActionStore{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/experiences/exp-1/events", bytes.NewReader([]byte(`{"event_type":"collect"}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", w.Code, w.Body.String())
	}
}
