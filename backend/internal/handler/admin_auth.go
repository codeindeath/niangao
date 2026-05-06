package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niangao/backend/internal/auth"
)

// ============================================================
// Admin Auth Handler
// ============================================================

type AdminAuthHandler struct {
	db        *pgxpool.Pool
	jwtSecret string
	devMode   bool
}

// RegisterAdminAuthRoutes registers admin auth endpoints on v1 group.
func RegisterAdminAuthRoutes(v1 *gin.RouterGroup, db *pgxpool.Pool, jwtSecret string, devMode bool) {
	h := &AdminAuthHandler{db: db, jwtSecret: jwtSecret, devMode: devMode}

	adminAuth := v1.Group("/auth/admin")
	{
		adminAuth.POST("/login", h.Login)
		if devMode {
			adminAuth.POST("/dev/login", h.DevLogin)
		}
	}
}

// ============================================================
// Admin Login
// ============================================================

type AdminLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AdminAuthHandler) Login(c *gin.Context) {
	var req AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入用户名和密码"})
		return
	}

	// Look up user by nickname and verify password
	var userID, appleUserID, nickname, passwordHash string
	var isAdmin bool
	err := h.db.QueryRow(c.Request.Context(),
		`SELECT id, COALESCE(apple_user_id, ''), nickname, COALESCE(password_hash, ''), is_admin
		 FROM users WHERE nickname = $1`,
		req.Username,
	).Scan(&userID, &appleUserID, &nickname, &passwordHash, &isAdmin)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	if passwordHash == "" || !auth.CheckPassword(req.Password, passwordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	// Issue JWT with admin flag
	token, err := auth.GenerateToken(h.jwtSecret, userID, appleUserID, nickname, true)
	if err != nil {
		log.Printf("admin login generate token error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成 token 失败"})
		return
	}

	// Generate and store refresh token
	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		log.Printf("generate refresh token failed: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user": gin.H{
				"id":       userID,
				"nickname": nickname,
				"is_admin": true,
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
			"is_admin": true,
		},
	})
}

// ============================================================
// Dev Admin Login (仅开发模式)
// ============================================================

type AdminDevLoginRequest struct {
	Nickname string `json:"nickname"`
}

func (h *AdminAuthHandler) DevLogin(c *gin.Context) {
	var req AdminDevLoginRequest
	_ = c.ShouldBindJSON(&req)

	nickname := req.Nickname
	if nickname == "" {
		nickname = "管理员"
	}

	devUserID := "dev-admin-" + nickname

	// Upsert user and set is_admin = true
	var userID string
	err := h.db.QueryRow(c.Request.Context(),
		`INSERT INTO users (apple_user_id, nickname, is_admin)
		 VALUES ($1, $2, TRUE)
		 ON CONFLICT (apple_user_id) DO UPDATE SET
		   nickname = CASE WHEN users.nickname = '' THEN $2 ELSE users.nickname END,
		   is_admin = TRUE,
		   updated_at = NOW()
		 RETURNING id`,
		devUserID, nickname,
	).Scan(&userID)

	if err != nil {
		log.Printf("dev admin login db error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "登录失败"})
		return
	}

	// Issue JWT with admin flag
	token, err := auth.GenerateToken(h.jwtSecret, userID, devUserID, nickname, true)
	if err != nil {
		log.Printf("dev admin login generate token error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成 token 失败"})
		return
	}

	// Generate and store refresh token
	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		log.Printf("generate refresh token failed: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user": gin.H{
				"id":       userID,
				"nickname": nickname,
				"is_admin": true,
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
			"is_admin": true,
		},
	})
}
