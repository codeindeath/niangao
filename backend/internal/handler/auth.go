package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niangao/backend/internal/auth"
	"github.com/niangao/backend/internal/middleware"
)

type AuthHandler struct {
	db          *pgxpool.Pool
	jwtSecret   string
	appleBundle string
	devMode     bool
}

func RegisterAuthRoutes(r *gin.RouterGroup, db *pgxpool.Pool, jwtSecret string, appleBundle string, devMode bool) {
	h := &AuthHandler{db: db, jwtSecret: jwtSecret, appleBundle: appleBundle, devMode: devMode}

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/apple/login", h.AppleLogin)
		if devMode {
			authGroup.POST("/dev/login", h.DevLogin)
		}
		authGroup.POST("/refresh", h.RefreshToken)
	}

	// Logout (needs auth to identify the token)
	r.POST("/auth/logout", middleware.RequireAuth(), h.Logout)
}

// ============================================================
// Apple Login
// ============================================================

type AppleLoginRequest struct {
	IdentityToken string `json:"identity_token" binding:"required"`
	FullName      string `json:"full_name"`
	Email         string `json:"email"`
}

func (h *AuthHandler) AppleLogin(c *gin.Context) {
	var req AppleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "missing_identity_token", "缺少 identity_token 参数")
		return
	}

	// 1. 验证 Apple identity token
	claims, err := auth.VerifyAppleIDToken(req.IdentityToken, h.appleBundle)
	if err != nil {
		log.Printf("apple login verify error: %v", err)
		respondError(c, http.StatusBadRequest, "apple_login_verify_failed", "Apple 登录验证失败，请重试")
		return
	}

	// 2. 决定昵称
	nickname := "年糕用户"
	if req.FullName != "" {
		nickname = req.FullName
	}
	if claims.Email != "" {
		email := claims.Email
		_ = email // 预留：可保存 email
	}

	// 3. 查找或创建用户（apple_user_id = claims.Subject）
	var userID string
	err = h.db.QueryRow(c.Request.Context(),
		`INSERT INTO users (apple_user_id, nickname)
		 VALUES ($1, $2)
		 ON CONFLICT (apple_user_id) DO UPDATE SET
		   nickname = CASE WHEN users.nickname = '' OR users.nickname = '年糕用户' THEN $2 ELSE users.nickname END,
		   updated_at = NOW()
		 RETURNING id`,
		claims.Subject, nickname,
	).Scan(&userID)

	if err != nil {
		respondError(c, http.StatusInternalServerError, "login_failed", "登录失败")
		return
	}

	// 4. 签发 JWT
	token, err := auth.GenerateToken(h.jwtSecret, userID, claims.Subject, nickname, false)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "token_generate_failed", "生成 token 失败")
		return
	}

	// 5. 生成并存储 refresh token
	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		log.Printf("generate refresh token failed: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user": gin.H{
				"id":       userID,
				"nickname": nickname,
				"avatar":   nil,
			},
		})
		return
	}

	if err := auth.StoreRefreshToken(c.Request.Context(), h.db, userID, refreshToken); err != nil {
		log.Printf("store refresh token failed: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"token":         token,
		"refresh_token": refreshToken,
		"user": gin.H{
			"id":       userID,
			"nickname": nickname,
			"avatar":   nil,
		},
	})
}

// ============================================================
// Dev Login (仅开发模式)
// ============================================================

type DevLoginRequest struct {
	Nickname string `json:"nickname"`
}

func (h *AuthHandler) DevLogin(c *gin.Context) {
	var req DevLoginRequest
	_ = c.ShouldBindJSON(&req)

	nickname := req.Nickname
	if nickname == "" {
		nickname = "开发者"
	}

	devUserID := "dev-" + nickname

	var userID string
	err := h.db.QueryRow(c.Request.Context(),
		`INSERT INTO users (apple_user_id, nickname)
		 VALUES ($1, $2)
		 ON CONFLICT (apple_user_id) DO UPDATE SET
		   nickname = CASE WHEN users.nickname = '' THEN $2 ELSE users.nickname END,
		   updated_at = NOW()
		 RETURNING id`,
		devUserID, nickname,
	).Scan(&userID)

	if err != nil {
		log.Printf("dev login db error: %v", err)
		respondError(c, http.StatusInternalServerError, "login_failed", "登录失败")
		return
	}

	token, err := auth.GenerateToken(h.jwtSecret, userID, devUserID, nickname, false)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "token_generate_failed", "生成 token 失败")
		return
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		log.Printf("generate refresh token failed: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user": gin.H{
				"id":       userID,
				"nickname": nickname,
				"avatar":   nil,
			},
		})
		return
	}

	if err := auth.StoreRefreshToken(c.Request.Context(), h.db, userID, refreshToken); err != nil {
		log.Printf("store refresh token failed: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"token":         token,
		"refresh_token": refreshToken,
		"user": gin.H{
			"id":       userID,
			"nickname": nickname,
			"avatar":   nil,
		},
	})
}

// ============================================================
// Refresh Token
// ============================================================

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "missing_refresh_token", "缺少 refresh_token 参数")
		return
	}

	// 验证并轮换 refresh token
	userID, err := auth.ValidateAndRotateRefreshToken(c.Request.Context(), h.db, req.RefreshToken)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "refresh_token_invalid", "refresh token 无效或已过期，请重新登录")
		return
	}

	// 查询用户信息
	var nickname, appleUserID string
	err = h.db.QueryRow(c.Request.Context(),
		`SELECT nickname, COALESCE(apple_user_id, '') FROM users WHERE id = $1`,
		userID,
	).Scan(&nickname, &appleUserID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "user_not_found", "用户不存在")
		return
	}

	// 签发新 JWT
	token, err := auth.GenerateToken(h.jwtSecret, userID, appleUserID, nickname, false)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "token_generate_failed", "生成 token 失败")
		return
	}

	// 生成新 refresh token（轮换）
	newRefreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		log.Printf("generate refresh token failed: %v", err)
		c.JSON(http.StatusOK, gin.H{"token": token})
		return
	}

	if err := auth.StoreRefreshToken(c.Request.Context(), h.db, userID, newRefreshToken); err != nil {
		log.Printf("store refresh token failed: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"token":         token,
		"refresh_token": newRefreshToken,
	})
}

// ============================================================
// Logout — 吊销当前 JWT + 所有 refresh token
// ============================================================

func (h *AuthHandler) Logout(c *gin.Context) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}

	jti := getAuthJTI(c)

	// 吊销当前 JWT
	if jti != "" {
		// JWT 在 7 天后自然过期，吊销表存到相同时间
		expiresAt := time.Now().Add(7 * 24 * time.Hour)
		if err := auth.RevokeToken(c.Request.Context(), h.db, jti, userID, expiresAt); err != nil {
			log.Printf("revoke token failed: %v", err)
		}
	}

	// 吊销所有 refresh token
	if err := auth.RevokeAllRefreshTokens(c.Request.Context(), h.db, userID); err != nil {
		log.Printf("revoke refresh tokens failed: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "已登出"})
}
