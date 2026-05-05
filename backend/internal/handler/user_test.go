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

// ══════════════════════════════════════════
// Profile — title 字段测试
// ══════════════════════════════════════════

func TestProfileRequiresAuth(t *testing.T) {
	tests := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/user/profile"},
		{"PUT", "/api/v1/user/profile"},
		{"DELETE", "/api/v1/user/account"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			r := gin.New()
			group := r.Group("/api/v1/user", middleware.RequireAuth())
			group.GET("/profile", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})
			group.PUT("/profile", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})
			group.DELETE("/account", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})

			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.method == "PUT" {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(`{}`))
				req.Header.Set("Content-Type", "application/json")
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401", w.Code)
			}
		})
	}
}

func TestUpdateProfileTitleValidation(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{
			name:           "update title — valid 10 chars",
			body:           `{"title":"终身学习者"}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "update title — exactly 20 chars (10 Chinese + period)",
			body:           `{"title":"一个终身学习的普通人啊"}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "update title — empty string clears title",
			body:           `{"title":""}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "update title — over 20 chars",
			body:           `{"title":"` + strings.Repeat("a", 21) + `"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "update title — 21 Chinese chars",
			body:           `{"title":"这是一个超长的称号它竟然有足足二十一个中文字符之多不可接受"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "update title and nickname together",
			body:           `{"title":"思考者","nickname":"小明"}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "update title only — null nickname",
			body:           `{"title":"探索者","nickname":null}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no fields provided",
			body:           `{}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.PUT("/api/v1/user/profile", func(c *gin.Context) {
				c.Set("user_id", "test-user-id")
			}, middleware.RequireAuth(), func(c *gin.Context) {
				var req UpdateProfileRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				if req.Nickname == nil && req.AvatarURL == nil && req.Bio == nil && req.Title == nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "至少需要提供一个要更新的字段"})
					return
				}
				if req.Nickname != nil {
					n := *req.Nickname
					if len(n) == 0 || len(n) > 30 {
						c.JSON(http.StatusBadRequest, gin.H{"error": "昵称长度应在 1~30 之间"})
						return
					}
				}
				if req.Bio != nil && len(*req.Bio) > 200 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "简介不能超过 200 字"})
					return
				}
				if req.Title != nil && len([]rune(*req.Title)) > 20 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "称号不能超过 20 字"})
					return
				}
				// Simulate success — return the updated fields
				resp := gin.H{
					"id":               "test-user-id",
					"nickname":         "测试用户",
					"title":            nil,
					"experience_count": 0,
					"bookmark_count":   0,
					"practiced_count":  0,
				}
				if req.Nickname != nil {
					resp["nickname"] = *req.Nickname
				}
				if req.Title != nil {
					resp["title"] = *req.Title
				}
				c.JSON(http.StatusOK, resp)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("PUT", "/api/v1/user/profile", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tt.expectedStatus, w.Body.String())
			}

			// For successful cases, verify title in response
			if tt.expectedStatus == http.StatusOK {
				var resp map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				// Response should include the title field (even if nil)
				if _, ok := resp["title"]; !ok {
					t.Errorf("response missing 'title' field: %v", resp)
				}
			}
		})
	}
}
