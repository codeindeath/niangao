package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AppleJWK 表示 Apple 的 JSON Web Key
type AppleJWK struct {
	KTY string `json:"kty"`
	KID string `json:"kid"`
	Use string `json:"use"`
	ALG string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// AppleJWKS Apple JWKS 响应
type AppleJWKS struct {
	Keys []AppleJWK `json:"keys"`
}

// AppleIDTokenClaims Apple identity token 的 claims
type AppleIDTokenClaims struct {
	Email        string `json:"email"`
	EmailVerified string `json:"email_verified"`
	jwt.RegisteredClaims
}

var (
	appleKeysCache     *AppleJWKS
	appleKeysCacheLock sync.RWMutex
	appleKeysFetchedAt time.Time
	httpClient         = &http.Client{Timeout: 10 * time.Second}
)

const (
	appleJWKSURL = "https://appleid.apple.com/auth/keys"
	appleIssuer  = "https://appleid.apple.com"
	keyCacheTTL  = 1 * time.Hour
)

// VerifyAppleIDToken 验证 Apple identity token
func VerifyAppleIDToken(tokenString, bundleID string) (*AppleIDTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AppleIDTokenClaims{}, func(token *jwt.Token) (any, error) {
		// 1. 提取 kid
		kid, ok := token.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, errors.New("missing kid in token header")
		}

		// 2. 获取 Apple 公钥
		key, err := getApplePublicKey(kid)
		if err != nil {
			return nil, fmt.Errorf("get apple public key: %w", err)
		}

		return key, nil
	})

	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*AppleIDTokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid apple identity token")
	}

	// 3. 验证 issuer
	if claims.Issuer != appleIssuer {
		return nil, fmt.Errorf("invalid issuer: %s", claims.Issuer)
	}

	// 4. 验证 audience (bundle ID)
	audOK := false
	for _, aud := range claims.Audience {
		if aud == bundleID {
			audOK = true
			break
		}
	}
	if !audOK {
		return nil, fmt.Errorf("invalid audience: bundle ID mismatch")
	}

	return claims, nil
}

// getApplePublicKey 从 Apple JWKS 获取指定 kid 的公钥
func getApplePublicKey(kid string) (*rsa.PublicKey, error) {
	keys, err := fetchAppleKeys()
	if err != nil {
		return nil, err
	}

	for _, jwk := range keys.Keys {
		if jwk.KID == kid {
			return jwkToRSA(jwk)
		}
	}

	return nil, fmt.Errorf("key with kid %s not found", kid)
}

// fetchAppleKeys 获取 Apple 公钥（带缓存）
func fetchAppleKeys() (*AppleJWKS, error) {
	appleKeysCacheLock.RLock()
	if appleKeysCache != nil && time.Since(appleKeysFetchedAt) < keyCacheTTL {
		cache := appleKeysCache
		appleKeysCacheLock.RUnlock()
		return cache, nil
	}
	appleKeysCacheLock.RUnlock()

	appleKeysCacheLock.Lock()
	defer appleKeysCacheLock.Unlock()

	// Double-check after acquiring write lock
	if appleKeysCache != nil && time.Since(appleKeysFetchedAt) < keyCacheTTL {
		return appleKeysCache, nil
	}

	resp, err := httpClient.Get(appleJWKSURL)
	if err != nil {
		return nil, fmt.Errorf("fetch apple jwks: %w", err)
	}
	defer resp.Body.Close()

	var jwks AppleJWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("decode apple jwks: %w", err)
	}

	appleKeysCache = &jwks
	appleKeysFetchedAt = time.Now()

	return &jwks, nil
}

// jwkToRSA 将 Apple JWK 转换为 RSA 公钥
func jwkToRSA(jwk AppleJWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("decode N: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("decode E: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)

	// 将 eBytes 转为 int
	e := 0
	for _, b := range eBytes {
		e = e<<8 | int(b)
	}
	if e == 0 {
		return nil, errors.New("invalid exponent")
	}

	return &rsa.PublicKey{
		N: n,
		E: e,
	}, nil
}
