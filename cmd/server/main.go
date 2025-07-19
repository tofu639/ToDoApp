package main

// @title Todo API Backend
// @version 1.0
// @description A comprehensive Todo API backend built with Go and Gin framework
// @description This API provides user authentication, JWT-based authorization, and full CRUD operations for todo items.
// @description The backend uses PostgreSQL for data persistence and follows clean architecture principles.

// @contact.name API Support
// @contact.email support@todoapi.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @tag.name authentication
// @tag.description User registration and authentication endpoints

// @tag.name todos
// @tag.description Todo CRUD operations (requires authentication)

// @tag.name health
// @tag.description Health check and readiness endpoints

// @schemes http https
// @produce application/json

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "todo-api-backend/docs" // Import generated docs
	"todo-api-backend/internal/config"
	"todo-api-backend/internal/database"
	"todo-api-backend/internal/handler"
	"todo-api-backend/internal/middleware"
	"todo-api-backend/internal/repository"
	"todo-api-backend/internal/service"
	"todo-api-backend/pkg/jwt"
)

func main() {
	log.Println("Todo API Backend starting...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode based on environment
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to database
	db, err := database.ConnectWithDSN(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run database migrations
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Initialize JWT token manager
	tokenManager := jwt.NewTokenManager(cfg.JWTSecret, cfg.JWTExpiration)

	// Initialize repositories
	repos := repository.NewRepositories(db)

	// Initialize services
	services := service.NewServices(repos, tokenManager)

	// Initialize handlers
	h := handler.NewHandler(services)

	// Create Gin router
	router := gin.New()

	// Add recovery middleware
	router.Use(gin.Recovery())

	// Add logging middleware
	if cfg.IsDevelopment() {
		router.Use(gin.Logger())
	}

	// Add CORS middleware
	corsConfig := &middleware.CORSConfig{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           12 * 3600, // 12 hours
	}
	router.Use(middleware.CORSMiddleware(corsConfig))

	// Register public routes (health check, auth)
	registerPublicRoutes(router, h)

	// Register protected routes with JWT middleware
	registerProtectedRoutes(router, h, tokenManager)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Close database connection
	if err := database.Close(); err != nil {
		log.Printf("Failed to close database connection: %v", err)
	}

	log.Println("Server exited")
}

// registerPublicRoutes registers routes that don't require authentication
func registerPublicRoutes(router *gin.Engine, h *handler.Handler) {
	// Health check endpoint
	router.GET("/health", h.HealthCheck)
	router.GET("/ready", h.ReadinessCheck)

	// Swagger documentation endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := router.Group("/api/v1")

	// Authentication routes (public)
	auth := v1.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
	}
}

// registerProtectedRoutes registers routes that require JWT authentication
func registerProtectedRoutes(router *gin.Engine, h *handler.Handler, tokenManager *jwt.TokenManager) {
	// API v1 routes
	v1 := router.Group("/api/v1")

	// Apply JWT middleware to protected routes
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(tokenManager))

	// Todo routes (protected)
	todos := protected.Group("/todos")
	{
		todos.POST("", h.CreateTodo)
		todos.GET("", h.GetTodos)
		todos.GET("/:id", h.GetTodo)
		todos.PUT("/:id", h.UpdateTodo)
		todos.DELETE("/:id", h.DeleteTodo)
	}
}