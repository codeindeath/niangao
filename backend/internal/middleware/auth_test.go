package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCORSMiddleware(t *testing.T) {
	r := gin.New()
	r.Use(CORS())
	r.Any("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		checkHeaders   bool
	}{
		{"GET request passes", "GET", http.StatusOK, true},
		{"OPTIONS request returns 204", "OPTIONS", http.StatusNoContent, true},
		{"POST request passes", "POST", http.StatusOK, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, "/test", nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}

			if tt.checkHeaders {
				if w.Header().Get("Access-Control-Allow-Origin") != "*" {
					t.Error("missing CORS Allow-Origin header")
				}
				if w.Header().Get("Access-Control-Allow-Methods") == "" {
					t.Error("missing CORS Allow-Methods header")
				}
			}
		})
	}
}

func TestAuthMiddlewareNoToken(t *testing.T) {
	r := gin.New()
	r.Use(AuthMiddleware("test-secret"))
	r.GET("/test", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 (unauthenticated requests pass through)", w.Code)
	}

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["user_id"] != nil && body["user_id"] != "" {
		t.Error("user_id should be empty for unauthenticated request")
	}
}

func TestAuthMiddlewareInvalidToken(t *testing.T) {
	r := gin.New()
	r.Use(AuthMiddleware("test-secret"))
	r.GET("/test", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-here")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 (invalid token passes through unauthenticated)", w.Code)
	}

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["user_id"] != nil && body["user_id"] != "" {
		t.Error("user_id should be empty for invalid token")
	}
}

func TestAuthMiddlewareMalformedHeader(t *testing.T) {
	r := gin.New()
	r.Use(AuthMiddleware("test-secret"))
	r.GET("/test", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	tests := []struct {
		name  string
		value string
	}{
		{"no Bearer prefix", "token-here"},
		{"empty value", ""},
		{"basic auth instead of bearer", "Basic dXNlcjpwYXNz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.value != "" {
				req.Header.Set("Authorization", tt.value)
			}
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("status = %d, want 200", w.Code)
			}
		})
	}
}

func TestRequireAuthMiddleware(t *testing.T) {
	r := gin.New()
	r.GET("/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"public": true})
	})
	r.GET("/private", RequireAuth(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"private": true})
	})

	t.Run("public endpoint accessible", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/public", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("public status = %d, want 200", w.Code)
		}
	})

	t.Run("private endpoint returns 401 without auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/private", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("private status = %d, want 401", w.Code)
		}

		var body map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &body)
		if body["error"] == nil {
			t.Error("error message should be present")
		}
	})

	t.Run("private endpoint returns 401 with empty user_id", func(t *testing.T) {
		// Manually set empty user_id in context
		r2 := gin.New()
		r2.GET("/private", func(c *gin.Context) {
			c.Set("user_id", "")
			c.Next()
		}, RequireAuth(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/private", nil)
		r2.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", w.Code)
		}
	})
}
