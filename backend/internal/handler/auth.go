package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niangao/backend/internal/auth"
)

type AuthHandler struct {
	db        *pgxpool.Pool
	jwtSecret string
	wechatCfg WechatConfig
}

type WechatConfig struct {
	AppID     string
	AppSecret string
}

func RegisterAuthRoutes(r *gin.RouterGroup, db *pgxpool.Pool, jwtSecret string, wechatCfg WechatConfig) {
	h := &AuthHandler{db: db, jwtSecret: jwtSecret, wechatCfg: wechatCfg}

	auth := r.Group("/auth")
	{
		auth.POST("/wechat/login", h.WechatLogin)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "微信登录失败: " + err.Error()})
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
	err = h.db.QueryRow(context.Background(),
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

	refreshToken, _ := auth.GenerateRefreshToken()

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
	c.JSON(http.StatusOK, gin.H{"status": "not implemented yet"})
}
