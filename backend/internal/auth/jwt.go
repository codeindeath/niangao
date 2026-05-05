package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Claims struct {
	UserID   string `json:"user_id"`
	OpenID   string `json:"open_id"`
	Nickname string `json:"nickname"`
	jwt.RegisteredClaims
}

// GenerateToken 签发 JWT（7 天有效），含唯一 jti
func GenerateToken(secret string, userID, openID, nickname string) (string, error) {
	jti, err := generateJTI()
	if err != nil {
		return "", err
	}
	claims := Claims{
		UserID:   userID,
		OpenID:   openID,
		Nickname: nickname,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "niangao",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 验证并解析 JWT
func ParseToken(secret, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// GenerateRefreshToken 生成随机 refresh token
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HashToken SHA-256 hashes a token for DB storage
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// StoreRefreshToken 将 refresh token hash 存入数据库
func StoreRefreshToken(ctx context.Context, db *pgxpool.Pool, userID, token string) error {
	tokenHash := HashToken(token)
	_, err := db.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, tokenHash, time.Now().Add(30*24*time.Hour),
	)
	return err
}

// ValidateAndRotateRefreshToken 验证 refresh token 并轮换（删除旧的，返回新的 user_id）
func ValidateAndRotateRefreshToken(ctx context.Context, db *pgxpool.Pool, token string) (string, error) {
	tokenHash := HashToken(token)

	var userID string
	err := db.QueryRow(ctx,
		`DELETE FROM refresh_tokens
		 WHERE token_hash = $1 AND expires_at > NOW()
		 RETURNING user_id`,
		tokenHash,
	).Scan(&userID)

	if err != nil {
		return "", fmt.Errorf("invalid or expired refresh token")
	}
	return userID, nil
}

// RevokeAllRefreshTokens 吊销用户所有 refresh token
func RevokeAllRefreshTokens(ctx context.Context, db *pgxpool.Pool, userID string) error {
	_, err := db.Exec(ctx,
		`DELETE FROM refresh_tokens WHERE user_id = $1`,
		userID,
	)
	return err
}

// IsTokenRevoked 检查 JWT 是否已被吊销（通过 jti）
func IsTokenRevoked(ctx context.Context, db *pgxpool.Pool, jti string) (bool, error) {
	var exists bool
	err := db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM token_revocations WHERE jti = $1 AND expires_at > NOW())`,
		jti,
	).Scan(&exists)
	return exists, err
}

// RevokeToken 将 JWT jti 加入吊销表
func RevokeToken(ctx context.Context, db *pgxpool.Pool, jti, userID string, expiresAt time.Time) error {
	_, err := db.Exec(ctx,
		`INSERT INTO token_revocations (jti, user_id, expires_at)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (jti) DO NOTHING`,
		jti, userID, expiresAt,
	)
	return err
}

func generateJTI() (string, error) {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
