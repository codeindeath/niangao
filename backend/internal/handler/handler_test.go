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
			body:           `{"content":"接到任务先确认 deadline","domain":"career"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing content",
			body:           `{"domain":"career"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing domain",
			body:           `{"content":"valid content"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body",
			body:           `{}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "content over 100 chars — 101 a's",
			body:           `{"content":"` + strings.Repeat("a", 101) + `","domain":"career"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "content exactly 100 chars",
			body:           `{"content":"` + strings.Repeat("a", 100) + `","domain":"life"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "chinese content",
			body:           `{"content":"把重要的决定放到早上做","domain":"cognition"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "with interpretation",
			body:           `{"content":"valid","domain":"career","interpretation":"详细解释..."}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "interpretation over 500 chars",
			body:           `{"content":"valid","domain":"career","interpretation":"` + strings.Repeat("a", 501) + `"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "interpretation exactly 500 chars",
			body:           `{"content":"valid","domain":"career","interpretation":"` + strings.Repeat("a", 500) + `"}`,
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
					Content        string `json:"content" binding:"required,max=100"`
					Domain         string `json:"domain" binding:"required"`
					Interpretation string `json:"interpretation" binding:"max=500"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusCreated, gin.H{
					"content":        req.Content,
					"domain":         req.Domain,
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
	req := httptest.NewRequest("POST", "/api/v1/experiences", strings.NewReader(`{"content":"test","domain":"career"}`))
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
		{"with domain", "domain=career", "career", "latest", "1"},
		{"with sort", "sort=popular", "", "popular", "1"},
		{"with page", "page=3", "", "latest", "3"},
		{"all params", "domain=life&sort=popular&page=2", "life", "popular", "2"},
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

func TestMyExperiencesRequiresAuth(t *testing.T) {
	r := gin.New()
	r.GET("/api/v1/me/experiences", middleware.RequireAuth(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/me/experiences", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestMyBookmarksRequiresAuth(t *testing.T) {
	r := gin.New()
	r.GET("/api/v1/me/bookmarks", middleware.RequireAuth(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/me/bookmarks", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestMyExperiencesQueryParams(t *testing.T) {
	r := gin.New()
	r.GET("/api/v1/me/experiences", func(c *gin.Context) {
		page := c.DefaultQuery("page", "1")
		pageSize := c.DefaultQuery("page_size", "20")
		c.JSON(http.StatusOK, gin.H{"page": page, "page_size": pageSize})
	})
	tests := []struct {
		name, queryStr, wantPage, wantSize string
	}{
		{"defaults", "", "1", "20"},
		{"custom page", "page=3", "3", "20"},
		{"custom size", "page_size=10", "1", "10"},
		{"both", "page=2&page_size=30", "2", "30"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v1/me/experiences"
			if tt.queryStr != "" {
				url += "?" + tt.queryStr
			}
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", url, nil)
			r.ServeHTTP(w, req)
			var body map[string]string
			json.Unmarshal(w.Body.Bytes(), &body)
			if body["page"] != tt.wantPage {
				t.Errorf("page = %q, want %q", body["page"], tt.wantPage)
			}
			if body["page_size"] != tt.wantSize {
				t.Errorf("page_size = %q, want %q", body["page_size"], tt.wantSize)
			}
		})
	}
}

func TestGetRecommendationsRequiresAuth(t *testing.T) {
	r := gin.New()
	r.GET("/api/v1/experiences/recommend", middleware.RequireAuth(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/experiences/recommend", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401 for unauthenticated recommend", w.Code)
	}
}

func TestGetRecommendationsRouteNotCapturedById(t *testing.T) {
	// Verify /recommend is NOT captured by /:id — it must be a separate route
	r := gin.New()
	recommendHit := false
	idHit := false

	r.GET("/api/v1/experiences/recommend", func(c *gin.Context) {
		recommendHit = true
		c.JSON(http.StatusOK, gin.H{"route": "recommend"})
	})
	r.GET("/api/v1/experiences/:id", func(c *gin.Context) {
		idHit = true
		c.JSON(http.StatusOK, gin.H{"route": "id", "id": c.Param("id")})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/experiences/recommend", nil)
	r.ServeHTTP(w, req)

	if !recommendHit {
		t.Error("recommend route was not hit — /:id captured it instead")
	}
	if idHit {
		t.Error("id route was hit when recommend was expected — route ordering issue")
	}
}
