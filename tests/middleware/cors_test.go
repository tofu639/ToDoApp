package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	
	"todo-api-backend/internal/middleware"
)

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		config         *middleware.CORSConfig
		method         string
		origin         string
		expectedStatus int
		expectedHeaders map[string]string
	}{
		{
			name:   "Default config with wildcard origin",
			config: middleware.DefaultCORSConfig(),
			method: http.MethodGet,
			origin: "https://example.com",
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin": "*",
			},
		},
		{
			name: "Specific allowed origin",
			config: &middleware.CORSConfig{
				AllowOrigins:     []string{"https://example.com", "https://test.com"},
				AllowMethods:     []string{http.MethodGet, http.MethodPost},
				AllowHeaders:     []string{"Content-Type", "Authorization"},
				AllowCredentials: true,
			},
			method: http.MethodGet,
			origin: "https://example.com",
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "https://example.com",
				"Access-Control-Allow-Credentials": "true",
			},
		},
		{
			name: "Origin not allowed",
			config: &middleware.CORSConfig{
				AllowOrigins: []string{"https://allowed.com"},
				AllowMethods: []string{http.MethodGet},
			},
			method: http.MethodGet,
			origin: "https://notallowed.com",
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin": "", // Should not be set
			},
		},
		{
			name: "Preflight OPTIONS request",
			config: &middleware.CORSConfig{
				AllowOrigins: []string{"*"},
				AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut},
				AllowHeaders: []string{"Content-Type", "Authorization"},
				MaxAge:       3600,
			},
			method: http.MethodOptions,
			origin: "https://example.com",
			expectedStatus: http.StatusNoContent,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT",
				"Access-Control-Allow-Headers": "Content-Type, Authorization",
				"Access-Control-Max-Age":       "3600",
			},
		},
		{
			name: "With expose headers",
			config: &middleware.CORSConfig{
				AllowOrigins:  []string{"*"},
				ExposeHeaders: []string{"Content-Length", "X-Custom-Header"},
			},
			method: http.MethodGet,
			origin: "https://example.com",
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":   "*",
				"Access-Control-Expose-Headers": "Content-Length, X-Custom-Header",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new Gin router
			router := gin.New()
			
			// Add the CORS middleware
			router.Use(middleware.CORSMiddleware(tt.config))
			
			// Add a test route
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Create a test request
			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			// Create a response recorder
			w := httptest.NewRecorder()

			// Perform the request
			router.ServeHTTP(w, req)

			// Assert the response status
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Assert the headers
			for headerName, expectedValue := range tt.expectedHeaders {
				if expectedValue == "" {
					// Check that header is not set or is empty
					actualValue := w.Header().Get(headerName)
					assert.Empty(t, actualValue, "Header %s should not be set", headerName)
				} else {
					actualValue := w.Header().Get(headerName)
					assert.Equal(t, expectedValue, actualValue, "Header %s mismatch", headerName)
				}
			}
		})
	}
}

func TestDefaultCORSConfig(t *testing.T) {
	config := middleware.DefaultCORSConfig()

	assert.NotNil(t, config)
	assert.Equal(t, []string{"*"}, config.AllowOrigins)
	assert.Contains(t, config.AllowMethods, http.MethodGet)
	assert.Contains(t, config.AllowMethods, http.MethodPost)
	assert.Contains(t, config.AllowMethods, http.MethodPut)
	assert.Contains(t, config.AllowMethods, http.MethodDelete)
	assert.Contains(t, config.AllowMethods, http.MethodOptions)
	assert.Contains(t, config.AllowHeaders, "Authorization")
	assert.Contains(t, config.AllowHeaders, "Content-Type")
	assert.False(t, config.AllowCredentials)
	assert.Equal(t, 12*3600, config.MaxAge)
}

func TestCORSMiddleware_NilConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a new Gin router
	router := gin.New()
	
	// Add the CORS middleware with nil config (should use default)
	router.Use(middleware.CORSMiddleware(nil))
	
	// Add a test route
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://example.com")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSMiddleware_ComplexPreflightRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &middleware.CORSConfig{
		AllowOrigins:     []string{"https://app.example.com"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"X-Total-Count", "X-Page-Count"},
		AllowCredentials: true,
		MaxAge:           7200,
	}

	// Create a new Gin router
	router := gin.New()
	
	// Add the CORS middleware
	router.Use(middleware.CORSMiddleware(config))
	
	// Add a test route
	router.POST("/api/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "success"})
	})

	// Create a preflight OPTIONS request
	req := httptest.NewRequest(http.MethodOptions, "/api/data", nil)
	req.Header.Set("Origin", "https://app.example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "https://app.example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "GET, POST, PUT, DELETE", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type, Authorization, X-Requested-With", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "7200", w.Header().Get("Access-Control-Max-Age"))
}

func TestCORSMiddleware_ActualRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &middleware.CORSConfig{
		AllowOrigins:     []string{"https://app.example.com"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"X-Total-Count"},
	}

	// Create a new Gin router
	router := gin.New()
	
	// Add the CORS middleware
	router.Use(middleware.CORSMiddleware(config))
	
	// Add a test route
	router.GET("/api/data", func(c *gin.Context) {
		c.Header("X-Total-Count", "100")
		c.JSON(http.StatusOK, gin.H{"data": "success"})
	})

	// Create an actual GET request
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	req.Header.Set("Origin", "https://app.example.com")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "https://app.example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "X-Total-Count", w.Header().Get("Access-Control-Expose-Headers"))
	assert.Equal(t, "100", w.Header().Get("X-Total-Count"))
}

func TestCORSMiddleware_MultipleOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &middleware.CORSConfig{
		AllowOrigins: []string{"https://app1.example.com", "https://app2.example.com", "https://app3.example.com"},
	}

	tests := []struct {
		name           string
		origin         string
		expectedOrigin string
	}{
		{
			name:           "First allowed origin",
			origin:         "https://app1.example.com",
			expectedOrigin: "https://app1.example.com",
		},
		{
			name:           "Second allowed origin",
			origin:         "https://app2.example.com",
			expectedOrigin: "https://app2.example.com",
		},
		{
			name:           "Third allowed origin",
			origin:         "https://app3.example.com",
			expectedOrigin: "https://app3.example.com",
		},
		{
			name:           "Not allowed origin",
			origin:         "https://malicious.com",
			expectedOrigin: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new Gin router
			router := gin.New()
			
			// Add the CORS middleware
			router.Use(middleware.CORSMiddleware(config))
			
			// Add a test route
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Create a test request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Origin", tt.origin)

			// Create a response recorder
			w := httptest.NewRecorder()

			// Perform the request
			router.ServeHTTP(w, req)

			// Assert the response
			assert.Equal(t, http.StatusOK, w.Code)
			
			actualOrigin := w.Header().Get("Access-Control-Allow-Origin")
			if tt.expectedOrigin == "" {
				assert.Empty(t, actualOrigin)
			} else {
				assert.Equal(t, tt.expectedOrigin, actualOrigin)
			}
		})
	}
}