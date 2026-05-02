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
	// Load env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env")
	}

	cfg := config.Load()

	// Database
	db, err := repository.NewDB(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Repositories
	expRepo := repository.NewExperienceRepo(db)
	likeRepo := repository.NewLikeRepo(db)
	bookmarkRepo := repository.NewBookmarkRepo(db)
	convRepo := repository.NewConversationRepo(db)

	// Router
	r := gin.Default()

	// Middleware
	r.Use(middleware.CORS())
	r.Use(middleware.AuthMiddleware(cfg.SupabaseJWTSecret))

	// Health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		handler.RegisterExperienceRoutes(v1, expRepo, likeRepo, bookmarkRepo)
		handler.RegisterConversationRoutes(v1, convRepo)
		handler.RegisterUserRoutes(v1)
	}

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
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
