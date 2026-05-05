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

	// Repositories
	expRepo := repository.NewExperienceRepo(db)
	likeRepo := repository.NewLikeRepo(db)
	bookmarkRepo := repository.NewBookmarkRepo(db)
	convRepo := repository.NewConversationRepo(db)

	// Router
	r := gin.Default()
	r.Use(middleware.CORS())
	r.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	// Health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		// 登录（无需认证）
		handler.RegisterAuthRoutes(v1, db, cfg.JWTSecret, cfg.AppleBundleID)

		// 经验
		handler.RegisterExperienceRoutes(v1, expRepo, likeRepo, bookmarkRepo)
		// 对话（需认证）
		handler.RegisterConversationRoutes(v1, convRepo)
		// 用户
		handler.RegisterUserRoutes(v1, db)
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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
