package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niangao/backend/internal/middleware"
)

func RegisterMeAccountRoutes(r *gin.RouterGroup, db *pgxpool.Pool) {
	me := r.Group("/me", middleware.RequireAuth())
	{
		me.DELETE("/account", func(c *gin.Context) {
			deleteMeAccount(c, db)
		})
	}
}

const (
	accountDeleteFailedMessage = "暂时注销不了账号，请稍后再试"
	accountNotFoundMessage     = "这个账号已经不存在或已注销"
	accountDeletedMessage      = "账号已注销"
)

func deleteMeAccount(c *gin.Context, db *pgxpool.Pool) {
	userID := getAuthUserID(c)
	if userID == "" {
		return
	}

	tx, err := db.Begin(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "account_delete_failed", accountDeleteFailedMessage)
		return
	}
	defer tx.Rollback(c.Request.Context())

	ct, err := tx.Exec(c.Request.Context(),
		`UPDATE users
		 SET apple_user_id='deleted:' || id::text,
		     nickname='',
		     display_name=NULL,
		     avatar_url=NULL,
		     bio=NULL,
		     title=NULL,
		     user_settings='{}'::jsonb,
		     deleted_at=NOW(),
		     updated_at=NOW()
		 WHERE id = $1 AND deleted_at IS NULL`,
		userID,
	)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "account_delete_failed", accountDeleteFailedMessage)
		return
	}
	if ct.RowsAffected() == 0 {
		respondError(c, http.StatusNotFound, "user_not_found", accountNotFoundMessage)
		return
	}

	if _, err := tx.Exec(c.Request.Context(), `DELETE FROM refresh_tokens WHERE user_id = $1`, userID); err != nil {
		respondError(c, http.StatusInternalServerError, "account_delete_failed", accountDeleteFailedMessage)
		return
	}
	if _, err := tx.Exec(c.Request.Context(), `DELETE FROM token_revocations WHERE user_id = $1`, userID); err != nil {
		respondError(c, http.StatusInternalServerError, "account_delete_failed", accountDeleteFailedMessage)
		return
	}
	if err := tx.Commit(c.Request.Context()); err != nil {
		respondError(c, http.StatusInternalServerError, "account_delete_failed", accountDeleteFailedMessage)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": accountDeletedMessage})
}
