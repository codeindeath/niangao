package handler

import (
	"encoding/json"
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
			RegisterExperienceRoutes(r.Group("/api/v1"), nil, nil, nil)

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
