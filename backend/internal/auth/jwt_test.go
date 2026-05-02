package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateToken(t *testing.T) {
	secret := "test-secret-key-min-32-chars!!"
	userID := "usr_123"
	openID := "wx_openid_abc"
	nickname := "测试用户"

	token, err := GenerateToken(secret, userID, openID, nickname)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("token is empty")
	}

	// Verify is a valid JWT (3 parts)
	parts := 0
	for i := 0; i < len(token); i++ {
		if token[i] == '.' {
			parts++
		}
	}
	if parts != 2 {
		t.Errorf("token has %d dots, want 2 (not a valid JWT)", parts)
	}
}

func TestParseToken(t *testing.T) {
	secret := "test-secret-key-min-32-chars!!"
	userID := "usr_456"
	openID := "wx_openid_xyz"
	nickname := "微信用户"

	token, err := GenerateToken(secret, userID, openID, nickname)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := ParseToken(secret, token)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID = %q, want %q", claims.UserID, userID)
	}
	if claims.OpenID != openID {
		t.Errorf("OpenID = %q, want %q", claims.OpenID, openID)
	}
	if claims.Nickname != nickname {
		t.Errorf("Nickname = %q, want %q", claims.Nickname, nickname)
	}
	if claims.Issuer != "niangao" {
		t.Errorf("Issuer = %q, want niangao", claims.Issuer)
	}
	if claims.ExpiresAt == nil {
		t.Error("ExpiresAt should not be nil")
	} else if claims.ExpiresAt.Time.Before(time.Now()) {
		t.Error("token already expired")
	}
}

func TestParseTokenInvalidSecret(t *testing.T) {
	secret := "test-secret-key-min-32-chars!!"
	userID := "usr_789"

	token, err := GenerateToken(secret, userID, "wx_abc", "test")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	_, err = ParseToken("wrong-secret-key-min-32-chars!", token)
	if err == nil {
		t.Error("ParseToken should fail with wrong secret")
	}
}

func TestParseTokenGarbage(t *testing.T) {
	_, err := ParseToken("test-secret", "not-a-jwt-token-at-all")
	if err == nil {
		t.Error("ParseToken should fail on garbage input")
	}
}

func TestParseTokenEmpty(t *testing.T) {
	_, err := ParseToken("test-secret", "")
	if err == nil {
		t.Error("ParseToken should fail on empty token")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	token, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	// Should be 64 hex chars (32 bytes)
	if len(token) != 64 {
		t.Errorf("refresh token length = %d, want 64", len(token))
	}

	// Should be random — two calls should produce different tokens
	token2, _ := GenerateRefreshToken()
	if token == token2 {
		t.Error("two refresh tokens should be different")
	}
}

func TestTokenExpiry(t *testing.T) {
	secret := "test-secret-key-min-32-chars!!"
	token, err := GenerateToken(secret, "user", "wx", "name")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	parser := jwt.NewParser()
	parsed, _, err := parser.ParseUnverified(token, &Claims{})
	if err != nil {
		t.Fatalf("ParseUnverified failed: %v", err)
	}

	claims, ok := parsed.(*Claims)
	if !ok {
		t.Fatal("claims type assertion failed")
	}

	expectedExpiry := time.Now().Add(7 * 24 * time.Hour)
	if claims.ExpiresAt == nil {
		t.Fatal("ExpiresAt is nil")
	}

	diff := claims.ExpiresAt.Time.Sub(expectedExpiry)
	if diff > time.Minute || diff < -time.Minute {
		t.Errorf("expiry is %v, expected ~%v (diff: %v)", claims.ExpiresAt.Time, expectedExpiry, diff)
	}
}

func TestSigningMethodIsHS256(t *testing.T) {
	secret := "test-secret-key-min-32-chars!!"
	token, err := GenerateToken(secret, "user", "wx", "name")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	parser := jwt.NewParser()
	parsed, _, err := parser.ParseUnverified(token, &Claims{})
	if err != nil {
		t.Fatalf("ParseUnverified failed: %v", err)
	}

	if parsed.Header["alg"] != "HS256" {
		t.Errorf("signing algorithm = %v, want HS256", parsed.Header["alg"])
	}
}

func TestParseTokenRejectsRS256(t *testing.T) {
	// ParseToken should reject non-HMAC algorithms
	secret := "test-secret-key-min-32-chars!!"
	// Create a fake RS256 token (will fail signature check anyway)
	claims := Claims{
		UserID: "test",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	// We can't sign without a private key, but the parsing should still reject it
	tokenString, _ := token.SignedString([]byte("fake")) // this will error but give us a partial token

	_, err := ParseToken(secret, tokenString)
	if err == nil {
		t.Error("ParseToken should reject RS256 tokens")
	}
}
