package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"enterprise-agent/backend/handlers"
	"enterprise-agent/backend/middleware"
	"enterprise-agent/backend/models"
	"enterprise-agent/backend/routes"
	"enterprise-agent/backend/services"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
		log.Println("Running in production mode (ReleaseMode)")
	}

	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("FATAL: JWT_SECRET environment variable is required")
	}
	if os.Getenv("DB_PASSWORD") == "" {
		log.Fatal("FATAL: DB_PASSWORD environment variable is required")
	}

	if err := models.ValidateConfig(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	db := services.NewDatabase()
	
	// ============================================================
	// ✅ 注释掉 Redis（暂时不使用）
	// ============================================================
	// redis := services.NewRedis()
	// if redis != nil {
	// 	pingCtx, pingCancel := context.WithTimeout(context.Background(), 3*time.Second)
	// 	if err := redis.Ping(pingCtx).Err(); err != nil {
	// 		log.Printf("Warning: Redis ping failed at startup: %v", err)
	// 	} else {
	// 		log.Println("Redis connection verified")
	// 	}
	// 	pingCancel()
	// }
	
	// ✅ 使用 nil 替代 Redis
	var redis *redis.Client = nil
	log.Println("Redis is disabled (set to nil)")

	userRepo := services.NewUserRepository(db)

	authHandler := handlers.NewAuthHandler(db, redis)
	userHandler := handlers.NewUserHandler(userRepo)
	adminHandler := handlers.NewAdminHandler(db)
	operationLogHandler := handlers.NewOperationLogHandler(db)
	handlers.GlobalLogger = operationLogHandler

	router := gin.New()
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())
	router.Use(handlers.OperationLogMiddleware(db))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/ready", func(c *gin.Context) {
		if err := db.PingContext(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	routeHandlers := &routes.Handlers{
		Auth:         authHandler,
		User:         userHandler,
		Admin:        adminHandler,
		OperationLog: operationLogHandler,
	}
	routes.SetupRoutes(router, routeHandlers, redis)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	listener, err := net.Listen("tcp4", ":"+port)
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}

	srv := &http.Server{Handler: router}

	go func() {
		log.Printf("Server starting on port %s (IPv4)", port)
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	if db != nil {
		db.Close()
	}
	if redis != nil {
		if err := redis.Close(); err != nil {
			log.Printf("Error closing redis: %v", err)
		}
	}
	log.Println("Server stopped")
}
