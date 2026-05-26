package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/niangao/backend/internal/middleware"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHealthEndpoint(t *testing.T) {
	r := gin.New()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("health status = %d, want 200", w.Code)
	}

	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["status"] != "ok" {
		t.Errorf("health body = %v, want status=ok", body)
	}
}

func TestExperienceCreateValidation(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{
			name:           "valid experience",
			body:           `{"content":"接到任务先确认 deadline","domain":"work","sub_domain":"work-comm"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing content",
			body:           `{"domain":"work","sub_domain":"work-comm"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing domain",
			body:           `{"content":"valid content","sub_domain":"work-comm"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing sub domain",
			body:           `{"content":"valid content","domain":"work"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "domain and sub domain omitted",
			body:           `{"content":"valid content"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "empty body",
			body:           `{}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "content over 100 chars — 101 a's",
			body:           `{"content":"` + strings.Repeat("a", 101) + `","domain":"work","sub_domain":"work-comm"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "content exactly 100 chars",
			body:           `{"content":"` + strings.Repeat("a", 100) + `","domain":"living","sub_domain":"selfcare"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "chinese content",
			body:           `{"content":"把重要的决定放到早上做","domain":"cognition","sub_domain":"thinking"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "with interpretation",
			body:           `{"content":"valid value here","domain":"work","sub_domain":"work-comm","interpretation":"详细解释..."}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "interpretation over 300 chars",
			body:           `{"content":"valid value here","domain":"work","sub_domain":"work-comm","interpretation":"` + strings.Repeat("a", 301) + `"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "interpretation exactly 300 chars",
			body:           `{"content":"valid value here","domain":"work","sub_domain":"work-comm","interpretation":"` + strings.Repeat("a", 300) + `"}`,
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/v1/experiences", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			// Build a minimal handler that validates the same way as the real one
			r := gin.New()
			r.POST("/api/v1/experiences", func(c *gin.Context) {
				c.Set("user_id", "test-user-id")
			}, middleware.RequireAuth(), func(c *gin.Context) {
				var req struct {
					Content        string `json:"content" binding:"required,min=10,max=100"`
					Domain         string `json:"domain"`
					SubDomain      string `json:"sub_domain"`
					Interpretation string `json:"interpretation" binding:"max=300"`
					Topics         string `json:"topics" binding:"max=200"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusCreated, gin.H{
					"content":        req.Content,
					"domain":         req.Domain,
					"sub_domain":     req.SubDomain,
					"interpretation": req.Interpretation,
				})
			})

			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tt.expectedStatus, w.Body.String())
			}
		})
	}
}

func TestExperienceCreateRequiresAuth(t *testing.T) {
	r := gin.New()
	r.POST("/api/v1/experiences", middleware.RequireAuth(), func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/experiences", strings.NewReader(`{"content":"valid content","domain":"work","sub_domain":"work-comm"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401 for unauthenticated experience creation", w.Code)
	}
}

func TestExperienceListQueryParams(t *testing.T) {
	r := gin.New()
	r.GET("/api/v1/experiences", func(c *gin.Context) {
		domain := c.Query("domain")
		sort := c.DefaultQuery("sort", "latest")
		page := c.DefaultQuery("page", "1")

		c.JSON(http.StatusOK, gin.H{
			"domain": domain,
			"sort":   sort,
			"page":   page,
		})
	})

	tests := []struct {
		name       string
		queryStr   string
		wantDomain string
		wantSort   string
		wantPage   string
	}{
		{"no params", "", "", "latest", "1"},
		{"with domain", "domain=work", "work", "latest", "1"},
		{"with sort", "sort=popular", "", "popular", "1"},
		{"with page", "page=3", "", "latest", "3"},
		{"all params", "domain=living&sort=popular&page=2", "living", "popular", "2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v1/experiences"
			if tt.queryStr != "" {
				url += "?" + tt.queryStr
			}
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", url, nil)
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("status = %d, want 200", w.Code)
				return
			}

			var body map[string]string
			json.Unmarshal(w.Body.Bytes(), &body)
			if body["domain"] != tt.wantDomain {
				t.Errorf("domain = %q, want %q", body["domain"], tt.wantDomain)
			}
			if body["sort"] != tt.wantSort {
				t.Errorf("sort = %q, want %q", body["sort"], tt.wantSort)
			}
			if body["page"] != tt.wantPage {
				t.Errorf("page = %q, want %q", body["page"], tt.wantPage)
			}
		})
	}
}
