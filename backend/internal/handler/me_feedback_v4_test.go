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
)

type fakeV4MeFeedbackStore struct {
	gotUserID     string
	gotType       string
	gotContent    string
	gotAppVersion string
	gotDevice     string
	gotOSVersion  string
	fail          bool
}

func (f *fakeV4MeFeedbackStore) CreateMeFeedback(ctx context.Context, userID, feedbackType, content, appVersion, device, osVersion string) error {
	f.gotUserID = userID
	f.gotType = feedbackType
	f.gotContent = content
	f.gotAppVersion = appVersion
	f.gotDevice = device
	f.gotOSVersion = osVersion
	if f.fail {
		return errors.New("store failed")
	}
	return nil
}

func TestV4MeFeedbackRequiresAuth(t *testing.T) {
	r := gin.New()
	RegisterMeFeedbackRoutes(r.Group("/api/v1"), &fakeV4MeFeedbackStore{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/me/feedback", strings.NewReader(`{"content":"按钮点了没反应"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401: %s", w.Code, w.Body.String())
	}
}

func TestV4MeFeedbackCreatesFeedback(t *testing.T) {
	r := gin.New()
	store := &fakeV4MeFeedbackStore{}
	v1 := r.Group("/api/v1", func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	RegisterMeFeedbackRoutes(v1, store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/me/feedback", strings.NewReader(`{
		"type":"bug",
		"content":"点收藏以后界面没有变化，但重新进来又收藏上了。",
		"app_version":"0.1.0",
		"device":"ios",
		"os_version":"18.0"
	}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", w.Code, w.Body.String())
	}
	if store.gotUserID != "user-1" || store.gotType != "bug" || store.gotContent != "点收藏以后界面没有变化，但重新进来又收藏上了。" {
		t.Fatalf("feedback not propagated: %+v", store)
	}
	if store.gotAppVersion != "0.1.0" || store.gotDevice != "ios" || store.gotOSVersion != "18.0" {
		t.Fatalf("client context not propagated: %+v", store)
	}
}

func TestV4MeFeedbackRejectsEmptyContent(t *testing.T) {
	r := gin.New()
	v1 := r.Group("/api/v1", func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	RegisterMeFeedbackRoutes(v1, &fakeV4MeFeedbackStore{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/me/feedback", strings.NewReader(`{"content":"   "}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", w.Code, w.Body.String())
	}
}

func TestV4MeFeedbackInvalidPayloadUsesUserFacingCopy(t *testing.T) {
	r := gin.New()
	v1 := r.Group("/api/v1", func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	RegisterMeFeedbackRoutes(v1, &fakeV4MeFeedbackStore{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/me/feedback", strings.NewReader(`{"content":`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	errBody, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("error = %+v, want structured error object", body["error"])
	}
	if errBody["message"] != "反馈内容格式不对" {
		t.Fatalf("error.message = %+v, want user-facing payload copy", errBody["message"])
	}
}

func TestV4MeFeedbackStoreFailureUsesUserFacingCopy(t *testing.T) {
	r := gin.New()
	v1 := r.Group("/api/v1", func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	RegisterMeFeedbackRoutes(v1, &fakeV4MeFeedbackStore{fail: true})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/me/feedback", strings.NewReader(`{"content":"这个按钮好像没有反应"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	errBody, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("error = %+v, want structured error object", body["error"])
	}
	if errBody["message"] != "暂时提交不了，请稍后再试" {
		t.Fatalf("error.message = %+v, want user-facing failure copy", errBody["message"])
	}
}
