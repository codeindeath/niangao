package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

func TestV4MeAccountSoftDeletesAndAnonymizesUser(t *testing.T) {
	source, err := os.ReadFile("me_account_v4.go")
	if err != nil {
		t.Fatalf("read me_account_v4.go: %v", err)
	}
	text := string(source)

	if strings.Contains(text, "DELETE FROM users") {
		t.Fatal("account cancellation should soft-delete/anonymize users instead of hard-deleting user rows")
	}
	for _, want := range []string{
		"UPDATE users",
		"deleted_at=NOW()",
		"apple_user_id='deleted:' || id::text",
		"DELETE FROM refresh_tokens",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("account cancellation should include %q", want)
		}
	}
}

func TestV4MeAccountUsesUserFacingCopy(t *testing.T) {
	source, err := os.ReadFile("me_account_v4.go")
	if err != nil {
		t.Fatalf("read me_account_v4.go: %v", err)
	}
	text := string(source)

	if strings.Contains(text, "删除账号失败") {
		t.Fatal("account cancellation failures should use a softer App-facing retry message")
	}
	for _, want := range []string{
		"暂时注销不了账号，请稍后再试",
		"这个账号已经不存在或已注销",
		"账号已注销",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("account cancellation copy should include %q", want)
		}
	}
}
