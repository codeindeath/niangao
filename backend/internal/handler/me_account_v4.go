package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niangao/backend/internal/middleware"
)

func RegisterMeAccountRoutes(r *gin.RouterGroup, db *pgxpool.Pool) {
	me := r.Group("/me", middleware.RequireAuth())
	{
		me.DELETE("/account", func(c *gin.Context) {
			deleteAccount(c, db)
		})
	}
}
