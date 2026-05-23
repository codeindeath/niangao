package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestVerifyAppleIDToken_InvalidToken(t *testing.T) {
	// 完全无效的 token
	_, err := VerifyAppleIDToken("not.a.valid.token", "com.test.app")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestVerifyAppleIDToken_ExpiredToken(t *testing.T) {
	// 构造一个已过期的 token（用任意 key 签名）
	claims := AppleIDTokenClaims{
		Email: "test@privaterelay.appleid.com",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "https://appleid.apple.com",
			Subject:   "001234.abcdef",
			Audience:  jwt.ClaimStrings{"com.test.app"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret"))

	_, err := VerifyAppleIDToken(tokenString, "com.test.app")
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestVerifyAppleIDToken_WrongIssuer(t *testing.T) {
	claims := AppleIDTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "https://evil.example.com",
			Subject:   "001234.abcdef",
			Audience:  jwt.ClaimStrings{"com.test.app"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret"))

	_, err := VerifyAppleIDToken(tokenString, "com.test.app")
	if err == nil {
		t.Error("expected error for wrong issuer")
	}
}

func TestVerifyAppleIDToken_WrongAudience(t *testing.T) {
	claims := AppleIDTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "https://appleid.apple.com",
			Subject:   "001234.abcdef",
			Audience:  jwt.ClaimStrings{"com.other.app"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret"))

	_, err := VerifyAppleIDToken(tokenString, "com.test.app")
	if err == nil {
		t.Error("expected error for wrong audience")
	}
}
