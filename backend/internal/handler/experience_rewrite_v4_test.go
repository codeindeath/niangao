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

type fakeExperienceRewriteGateway struct {
	fail       bool
	gotUserID  string
	gotRequest model.ExperienceRewriteGatewayRequest
	response   *model.ExperienceRewriteGatewayResponse
}

func (f *fakeExperienceRewriteGateway) RewriteExperience(ctx context.Context, req model.ExperienceRewriteGatewayRequest) (*model.ExperienceRewriteGatewayResponse, error) {
	f.gotUserID = req.UserID
	f.gotRequest = req
	if f.fail {
		return nil, errors.New("rewrite failed")
	}
	if f.response != nil {
		return f.response, nil
	}
	return &model.ExperienceRewriteGatewayResponse{
		CanRewrite:         true,
		RewrittenContent:   "先把真正刺痛自己的点写清楚，再决定要不要离开。",
		Domain:             string(model.DomainWork),
		SubDomain:          string(model.SubPromotion),
		Topic:              "辞职犹豫",
		RewriteLevel:       "light",
		SourcePreservation: "high",
		Confidence:         0.86,
	}, nil
}

func TestV4ExperienceRewriteRequiresAuth(t *testing.T) {
	r := gin.New()
	RegisterExperienceRoutes(r.Group("/api/v1"), nil, &fakeExperienceRewriteGateway{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/experiences/rewrite", strings.NewReader(`{"content":"我想整理一下"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401: %s", w.Code, w.Body.String())
	}
}

func TestV4ExperienceRewriteCallsGatewayAndReturnsSuggestion(t *testing.T) {
	r := gin.New()
	gateway := &fakeExperienceRewriteGateway{}
	v1 := r.Group("/api/v1", func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	RegisterExperienceRoutes(v1, nil, gateway)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/experiences/rewrite", strings.NewReader(`{"content":"我发现我不是怕换工作，是怕再次证明自己选错了","source":"manual_note","default_visibility":"public","user_selected_domain":"meaning","user_selected_sub_domain":"self"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
	}
	if gateway.gotUserID != "user-1" {
		t.Fatalf("gateway user = %q, want user-1", gateway.gotUserID)
	}
	if gateway.gotRequest.RawText == "" || gateway.gotRequest.Source != "manual_note" {
		t.Fatalf("gateway request = %+v, want raw text and source", gateway.gotRequest)
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["rewritten_content"] == "" || body["can_rewrite"] != true {
		t.Fatalf("unexpected rewrite response: %+v", body)
	}
	if body["domain"] != string(model.DomainWork) {
		t.Fatalf("domain = %v, want work", body["domain"])
	}
}

func TestV4ExperienceRewriteGatewayFailureIsRetryable(t *testing.T) {
	r := gin.New()
	v1 := r.Group("/api/v1", func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	RegisterExperienceRoutes(v1, nil, &fakeExperienceRewriteGateway{fail: true})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/experiences/rewrite", strings.NewReader(`{"content":"我想整理一下"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if retryable, _ := body["retryable"].(bool); !retryable {
		t.Fatalf("retryable = %+v, want true", body["retryable"])
	}
}

func TestV4ExperienceRewriteRejectsOverlongGatewayOutput(t *testing.T) {
	r := gin.New()
	v1 := r.Group("/api/v1", func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	RegisterExperienceRoutes(v1, nil, &fakeExperienceRewriteGateway{
		response: &model.ExperienceRewriteGatewayResponse{
			CanRewrite:       true,
			RewrittenContent: strings.Repeat("年", 101),
		},
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/experiences/rewrite", strings.NewReader(`{"content":"我想整理一下"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502: %s", w.Code, w.Body.String())
	}
}
