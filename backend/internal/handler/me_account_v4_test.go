package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestV4MeAccountRequiresAuth(t *testing.T) {
	r := gin.New()
	v1 := r.Group("/api/v1")
	RegisterMeAccountRoutes(v1, nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/me/account", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
