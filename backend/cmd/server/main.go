package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/niangao/backend/internal/config"
	"github.com/niangao/backend/internal/handler"
	"github.com/niangao/backend/internal/middleware"
	"github.com/niangao/backend/internal/repository"
)

func main() {
	godotenv.Load()

	cfg := config.Load()
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	// Database
	db, err := repository.NewDB(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()
	log.Println("Database connected")

	// Run migrations (manual)
	// Migrations are run via: psql $DATABASE_URL -f migrations/005_auth_tokens.sql
	_ = "migrations"

	// Repositories
	expRepo := repository.NewExperienceRepo(db)
	likeRepo := repository.NewLikeRepo(db)
	bookmarkRepo := repository.NewBookmarkRepo(db)
	convRepo := repository.NewConversationRepo(db)
	statsRepo := repository.NewStatsRepo(db)

	// Dev mode flag
	devMode := cfg.Env != "production"
	log.Printf("Starting in %s mode (dev=%v)", cfg.Env, devMode)

	// Router
	r := gin.Default()
	r.Use(middleware.CORS())
	r.Use(middleware.AuthMiddleware(cfg.JWTSecret, db))

	// Health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		// 登录（无需认证）
		handler.RegisterAuthRoutes(v1, db, cfg.JWTSecret, cfg.AppleBundleID, devMode)

		// 经验
		handler.RegisterExperienceRoutes(v1, expRepo, likeRepo, bookmarkRepo)
		// 对话（需认证）
		handler.RegisterConversationRoutes(v1, convRepo)
		// 用户
		handler.RegisterUserRoutes(v1, db)
		// 统计
		handler.RegisterStatsRoutes(v1, statsRepo)

		// Admin routes (require admin permission)
		admin := v1.Group("/admin", middleware.RequireAdmin(db))
		{
			// Dashboard
			handler.RegisterAdminDashboardRoutes(admin, db)
			// Review
			handler.RegisterAdminReviewRoutes(admin, db)
			// Content
			handler.RegisterAdminContentRoutes(admin, expRepo, db)
			// Platform
			handler.RegisterAdminPlatformRoutes(admin, db, expRepo)
			// Users
			handler.RegisterAdminUserRoutes(admin, db)
			// Domains
			handler.RegisterAdminDomainRoutes(admin, db)
		// Stats
		handler.RegisterAdminStatsRoutes(admin, db)
		// AI
		handler.RegisterAdminAIRoutes(admin, db)
		// Config
			handler.RegisterAdminConfigRoutes(admin, db)
			// Logs
			handler.RegisterAdminLogRoutes(admin, db)
			// Export
			handler.RegisterAdminExportRoutes(admin, db)
		}
		// Admin auth (login - no RequireAdmin needed)
		handler.RegisterAdminAuthRoutes(v1, db, cfg.JWTSecret, devMode)
	}

	// Server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("年糕 API 已启动 :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Periodic token cleanup
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			middleware.CleanupExpiredTokens(context.Background(), db)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
