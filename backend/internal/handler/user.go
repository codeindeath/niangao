package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niangao/backend/internal/middleware"
)

func RegisterUserRoutes(r *gin.RouterGroup, db *pgxpool.Pool) {
	user := r.Group("/user", middleware.RequireAuth())
	{
		user.GET("/profile", func(c *gin.Context) {
			getProfile(c, db)
		})
		user.PUT("/profile", func(c *gin.Context) {
			updateProfile(c, db)
		})
		user.DELETE("/account", func(c *gin.Context) {
			deleteAccount(c, db)
		})
	}
}

// ============================================================
// GET /user/profile
// ============================================================

func getProfile(c *gin.Context, db *pgxpool.Pool) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}

	var nickname string
	var avatarURL *string
	var bio *string
	var expCount, bmCount, practicedCount int

	err := db.QueryRow(c.Request.Context(),
		`SELECT nickname, avatar_url, bio, experience_count, bookmark_count, practiced_count
		 FROM users WHERE id = $1`, userID,
	).Scan(&nickname, &avatarURL, &bio, &expCount, &bmCount, &practicedCount)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":               userID,
		"nickname":         nickname,
		"avatar_url":       avatarURL,
		"bio":              bio,
		"experience_count": expCount,
		"bookmark_count":   bmCount,
		"practiced_count":  practicedCount,
	})
}

// ============================================================
// PUT /user/profile
// ============================================================

type UpdateProfileRequest struct {
	Nickname  *string `json:"nickname"`
	AvatarURL *string `json:"avatar_url"`
	Bio       *string `json:"bio"`
}

func updateProfile(c *gin.Context, db *pgxpool.Pool) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	// 至少提供一个字段
	if req.Nickname == nil && req.AvatarURL == nil && req.Bio == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "至少需要提供一个要更新的字段"})
		return
	}

	// 校验昵称
	if req.Nickname != nil {
		nickname := *req.Nickname
		if len(nickname) == 0 || len(nickname) > 30 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "昵称长度应在 1~30 之间"})
			return
		}
	}

	// 校验简介
	if req.Bio != nil && len(*req.Bio) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "简介不能超过 200 字"})
		return
	}

	// 动态构建 UPDATE
	var (
		nickname  string
		avatarURL *string
		bio       *string
		expCount  int
		bmCount   int
		practicedCount int
	)

	// 先查当前值
	err := db.QueryRow(c.Request.Context(),
		`UPDATE users SET
		   nickname  = COALESCE($2, nickname),
		   avatar_url = COALESCE($3, avatar_url),
		   bio       = COALESCE($4, bio),
		   updated_at = NOW()
		 WHERE id = $1
		 RETURNING nickname, avatar_url, bio, experience_count, bookmark_count, practiced_count`,
		userID, req.Nickname, req.AvatarURL, req.Bio,
	).Scan(&nickname, &avatarURL, &bio, &expCount, &bmCount, &practicedCount)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":               userID,
		"nickname":         nickname,
		"avatar_url":       avatarURL,
		"bio":              bio,
		"experience_count": expCount,
		"bookmark_count":   bmCount,
		"practiced_count":  practicedCount,
	})
}

// ============================================================
// DELETE /user/account — 删除账号及关联数据
// ============================================================

func deleteAccount(c *gin.Context, db *pgxpool.Pool) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}

	// CASCADE: users 表的删除会级联到 experiences, likes, bookmarks,
	// refresh_tokens, token_revocations（通过 ON DELETE CASCADE）
	ct, err := db.Exec(c.Request.Context(),
		`DELETE FROM users WHERE id = $1`,
		userID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除账号失败"})
		return
	}

	if ct.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "账号已删除"})
}
