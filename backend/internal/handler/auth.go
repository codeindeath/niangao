package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niangao/backend/internal/auth"
)

type AuthHandler struct {
	db          *pgxpool.Pool
	jwtSecret   string
	appleBundle string
}

func RegisterAuthRoutes(r *gin.RouterGroup, db *pgxpool.Pool, jwtSecret string, appleBundle string) {
	h := &AuthHandler{db: db, jwtSecret: jwtSecret, appleBundle: appleBundle}

	auth := r.Group("/auth")
	{
		auth.POST("/apple/login", h.AppleLogin)
		auth.POST("/dev/login", h.DevLogin)
		auth.POST("/refresh", h.RefreshToken)
	}
}

// DevLogin — 开发环境模拟登录，创建测试用户直接返回 JWT
// 仅在开发用途，生产环境应移除

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "登录失败"})
		return
	}

	token, err := auth.GenerateToken(h.jwtSecret, userID, devUserID, nickname)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成 token 失败"})
		return
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		log.Printf("generate refresh token failed: %v", err)
		refreshToken = ""
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

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "refresh token 功能开发中"})
}

// Apple Login

type AppleLoginRequest struct {
	IdentityToken string `json:"identity_token" binding:"required"`
	FullName      string `json:"full_name"`
	Email         string `json:"email"`
}

func (h *AuthHandler) AppleLogin(c *gin.Context) {
	var req AppleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 identity_token 参数"})
		return
	}

	// 1. 验证 Apple identity token
	claims, err := auth.VerifyAppleIDToken(req.IdentityToken, h.appleBundle)
	if err != nil {
		log.Printf("apple login verify error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Apple 登录验证失败，请重试"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "登录失败"})
		return
	}

	// 4. 签发 JWT
	token, err := auth.GenerateToken(h.jwtSecret, userID, claims.Subject, nickname)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成 token 失败"})
		return
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		log.Printf("generate refresh token failed: %v", err)
		refreshToken = ""
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
