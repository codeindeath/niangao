package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDeprecatedExperienceAppRoutesReturnGone(t *testing.T) {
	tests := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/experiences/recommend"},
		{http.MethodPost, "/api/v1/experiences/exp-1/like"},
		{http.MethodPost, "/api/v1/experiences/exp-1/bookmark"},
		{http.MethodGet, "/api/v1/me/experiences"},
		{http.MethodGet, "/api/v1/me/bookmarks"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			r := gin.New()
			RegisterExperienceRoutes(r.Group("/api/v1"), nil)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			r.ServeHTTP(w, req)

			if w.Code != http.StatusGone {
				t.Fatalf("status = %d, want 410: %s", w.Code, w.Body.String())
			}
			var body map[string]map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["error"]["code"] != "deprecated_endpoint" {
				t.Fatalf("error code = %q, want deprecated_endpoint", body["error"]["code"])
			}
		})
	}
}

func TestProductionMainDoesNotRegisterDeprecatedMobileRouteGroups(t *testing.T) {
	source, err := os.ReadFile("../../cmd/server/main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	mainSource := string(source)

	for _, call := range []string{
		"RegisterChatRoutes(",
		"RegisterUserRoutes(",
		"RegisterStatsRoutes(",
	} {
		if strings.Contains(mainSource, call) {
			t.Fatalf("production main should not register deprecated mobile route group %s", call)
		}
	}
}

func TestDeprecatedMobileHandlerSourcesAreRemoved(t *testing.T) {
	for _, path := range []string{
		"chat.go",
		"conversation.go",
		"stats.go",
		"user.go",
	} {
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("deprecated mobile handler source %s should be removed, stat err=%v", path, err)
		}
	}
}

func TestDeprecatedMobileInteractionRepositorySourcesAreRemoved(t *testing.T) {
	for _, path := range []string{
		"../repository/bookmark.go",
		"../repository/like.go",
		"../repository/stats.go",
	} {
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("deprecated mobile interaction repository source %s should be removed, stat err=%v", path, err)
		}
	}
}
