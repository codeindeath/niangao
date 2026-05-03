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
	wechatCfg   WechatConfig
	appleBundle string
}

type WechatConfig struct {
	AppID     string
	AppSecret string
}

func RegisterAuthRoutes(r *gin.RouterGroup, db *pgxpool.Pool, jwtSecret string, wechatCfg WechatConfig, appleBundle string) {
	h := &AuthHandler{db: db, jwtSecret: jwtSecret, wechatCfg: wechatCfg, appleBundle: appleBundle}

	auth := r.Group("/auth")
	{
		auth.POST("/wechat/login", h.WechatLogin)
		auth.POST("/apple/login", h.AppleLogin)
		auth.POST("/refresh", h.RefreshToken)
	}
}

type WechatLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

func (h *AuthHandler) WechatLogin(c *gin.Context) {
	var req WechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 code 参数"})
		return
	}

	// 1. 用 code 向微信换取 openid
	tokenResp, err := auth.ExchangeCode(h.wechatCfg.AppID, h.wechatCfg.AppSecret, req.Code)
	if err != nil {
		log.Printf("wechat login error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "微信登录失败，请重试"})
		return
	}

	// 2. 获取微信用户信息
	userInfo, err := auth.GetUserInfo(tokenResp.AccessToken, tokenResp.OpenID)
	if err != nil {
		// 获取用户信息失败不阻塞登录，用默认昵称
		userInfo = &auth.WechatUserInfo{
			OpenID:   tokenResp.OpenID,
			UnionID:  tokenResp.UnionID,
			Nickname: "微信用户",
		}
	}

	// 3. 查找或创建用户
	var userID string
	err = h.db.QueryRow(c.Request.Context(),
		`INSERT INTO users (wechat_openid, wechat_unionid, nickname, avatar_url)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (wechat_openid) DO UPDATE SET
		   wechat_unionid = COALESCE($2, users.wechat_unionid),
		   nickname = CASE WHEN users.nickname = '' OR users.nickname = '微信用户' THEN $3 ELSE users.nickname END,
		   avatar_url = COALESCE($4, users.avatar_url),
		   updated_at = NOW()
		 RETURNING id`,
		userInfo.OpenID, userInfo.UnionID, userInfo.Nickname, userInfo.HeadImgURL,
	).Scan(&userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "登录失败"})
		return
	}

	// 4. 签发 JWT
	token, err := auth.GenerateToken(h.jwtSecret, userID, userInfo.OpenID, userInfo.Nickname)
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
			"nickname": userInfo.Nickname,
			"avatar":   userInfo.HeadImgURL,
		},
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// TODO: 实现 refresh token 逻辑
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
