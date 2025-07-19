package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	
	"todo-api-backend/internal/database"
	"todo-api-backend/internal/model"
)

// HealthCheck handles health check requests
// @Summary Health check
// @Description Check the health status of the API and database connection
// @Tags health
// @Produce json
// @Success 200 {object} model.HealthResponse "Service is healthy"
// @Failure 503 {object} model.ErrorResponse "Service is unhealthy"
// @Router /health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Check database connectivity
	dbStatus := "connected"
	if err := database.HealthCheck(ctx); err != nil {
		dbStatus = "disconnected"
		// Return 503 Service Unavailable if database is not healthy
		c.JSON(http.StatusServiceUnavailable, model.ErrorResponse{
			Error:   "service_unavailable",
			Message: "Database health check failed",
			Details: map[string]string{
				"database_error": err.Error(),
			},
		})
		return
	}

	response := model.HealthResponse{
		Status:   "ok",
		Database: dbStatus,
		Time:     time.Now().UTC().Format(time.RFC3339),
	}
	
	c.JSON(http.StatusOK, response)
}

// ReadinessCheck handles readiness check requests
// @Summary Readiness check
// @Description Check if the API is ready to serve requests
// @Tags health
// @Produce json
// @Success 200 {object} model.HealthResponse "Service is ready"
// @Failure 503 {object} model.ErrorResponse "Service is not ready"
// @Router /ready [get]
func (h *Handler) ReadinessCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	// Check if all critical dependencies are ready
	dbStatus := "ready"
	if err := database.HealthCheck(ctx); err != nil {
		dbStatus = "not_ready"
		c.JSON(http.StatusServiceUnavailable, model.ErrorResponse{
			Error:   "service_unavailable",
			Message: "Service is not ready to serve requests",
			Details: map[string]string{
				"database_status": "not_ready",
				"database_error":  err.Error(),
			},
		})
		return
	}

	response := model.HealthResponse{
		Status:   "ready",
		Database: dbStatus,
		Time:     time.Now().UTC().Format(time.RFC3339),
	}
	
	c.JSON(http.StatusOK, response)
}