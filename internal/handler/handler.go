package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	
	"todo-api-backend/internal/service"
)

// Handler holds all HTTP handlers and their dependencies
type Handler struct {
	services  *service.Services
	validator *validator.Validate
}

// NewHandler creates a new Handler instance with service dependencies
func NewHandler(services *service.Services) *Handler {
	return &Handler{
		services:  services,
		validator: validator.New(),
	}
}

// RegisterRoutes registers all HTTP routes with the Gin router
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	// API v1 routes
	v1 := router.Group("/api/v1")
	
	// Authentication routes (public)
	auth := v1.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
	}
	
	// Todo routes (protected - will be implemented with JWT middleware)
	todos := v1.Group("/todos")
	// Note: JWT middleware will be applied to these routes in the main server setup
	{
		todos.POST("", h.CreateTodo)
		todos.GET("", h.GetTodos)
		todos.GET("/:id", h.GetTodo)
		todos.PUT("/:id", h.UpdateTodo)
		todos.DELETE("/:id", h.DeleteTodo)
	}
	
	// Health check route
	router.GET("/health", h.HealthCheck)
}