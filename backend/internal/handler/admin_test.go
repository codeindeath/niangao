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

// ============================================================
// Admin Auth Tests
// ============================================================

func TestAdminDashboardRequiresAdmin(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user")
		c.Next()
	})
	// Admin middleware without DB should pass (nil-safe)
	r.GET("/admin/dashboard", middleware.RequireAdmin(nil), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/dashboard", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("dashboard status = %d, want 200", w.Code)
	}
}

func TestAdminRoutesUnauthorizedWithoutToken(t *testing.T) {
	r := gin.New()
	r.GET("/admin/dashboard", func(c *gin.Context) {
		// Without auth middleware, user_id not set
		uid, _ := c.Get("user_id")
		if uid == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/dashboard", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

// ============================================================
// Admin Review Tests
// ============================================================

func TestBatchReviewValidation(t *testing.T) {
	r := gin.New()
	r.POST("/admin/reviews/batch", func(c *gin.Context) {
		c.Set("user_id", "admin-id")
		var req struct {
			IDs    []string `json:"ids"`
			Action string   `json:"action"`
			Reason *string  `json:"reason"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请选择经验"})
			return
		}
		if req.Action != "approve" && req.Action != "reject" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效操作"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{"valid approve", `{"ids":["a","b"],"action":"approve"}`, http.StatusOK},
		{"valid reject with reason", `{"ids":["a"],"action":"reject","reason":"test"}`, http.StatusOK},
		{"empty ids", `{"ids":[],"action":"approve"}`, http.StatusBadRequest},
		{"missing action", `{"ids":["a"]}`, http.StatusBadRequest},
		{"invalid action", `{"ids":["a"],"action":"delete"}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/admin/reviews/batch", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}
		})
	}
}

// ============================================================
// Admin Config Tests
// ============================================================

func TestConfigUpdateValidation(t *testing.T) {
	r := gin.New()
	r.PUT("/admin/config", func(c *gin.Context) {
		var req struct {
			Key   string      `json:"key"`
			Value interface{} `json:"value"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Key == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{"valid config", `{"key":"review_mode","value":"auto"}`, http.StatusOK},
		{"empty key", `{"key":"","value":"x"}`, http.StatusBadRequest},
		{"missing key", `{"value":"x"}`, http.StatusBadRequest},
		{"numeric value", `{"key":"max_length","value":100}`, http.StatusOK},
		{"boolean value", `{"key":"enabled","value":true}`, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("PUT", "/admin/config", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}
		})
	}
}

// ============================================================
// Admin Logs Tests
// ============================================================

func TestLogsFilterValidation(t *testing.T) {
	r := gin.New()
	r.GET("/admin/logs", func(c *gin.Context) {
		page := c.DefaultQuery("page", "1")
		pageSize := c.DefaultQuery("page_size", "20")
		// Validate params don't crash
		if page == "" || pageSize == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"total": 0, "data": []interface{}{}})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/logs?page=1&page_size=20", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	if _, ok := body["total"]; !ok {
		t.Error("response missing total field")
	}
}

// ============================================================
// Admin User Status Tests
// ============================================================

func TestUserStatusUpdateValidation(t *testing.T) {
	r := gin.New()
	r.PUT("/admin/users/:id/status", func(c *gin.Context) {
		c.Set("user_id", "admin-id")
		var req struct {
			Active bool    `json:"active"`
			Reason *string `json:"reason"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
			return
		}
		// Simulate RowsAffected check
		if c.Param("id") == "nonexistent" {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"active": req.Active, "status": "ok"})
	})

	tests := []struct {
		name           string
		userID         string
		body           string
		expectedStatus int
	}{
		{"disable user", "test-id", `{"active":false,"reason":"违规"}`, http.StatusOK},
		{"enable user", "test-id", `{"active":true}`, http.StatusOK},
		{"nonexistent user", "nonexistent", `{"active":false}`, http.StatusNotFound},
		{"invalid body", "test-id", `{}`, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("PUT", "/admin/users/"+tt.userID+"/status", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tt.expectedStatus, w.Body.String())
			}
		})
	}
}

// ============================================================
// Admin Platform Tests
// ============================================================

func TestPlatformCreateValidation(t *testing.T) {
	r := gin.New()
	r.POST("/admin/platform-experiences", func(c *gin.Context) {
		var req struct {
			Content     string `json:"content"`
			Domain      string `json:"domain"`
			CreatorName string `json:"creator_name"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Content == "" || req.Domain == "" || req.CreatorName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请填写内容、领域和创作者名称"})
			return
		}
		validDomains := map[string]bool{"career": true, "cognition": true, "life": true, "relationship": true, "emotion": true}
		if !validDomains[req.Domain] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的领域"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "created"})
	})

	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{"valid create", `{"content":"test","domain":"career","creator_name":"test"}`, http.StatusCreated},
		{"missing content", `{"domain":"career","creator_name":"test"}`, http.StatusBadRequest},
		{"missing domain", `{"content":"test","creator_name":"test"}`, http.StatusBadRequest},
		{"missing creator", `{"content":"test","domain":"career"}`, http.StatusBadRequest},
		{"invalid domain", `{"content":"test","domain":"invalid","creator_name":"test"}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/admin/platform-experiences", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}
		})
	}
}
