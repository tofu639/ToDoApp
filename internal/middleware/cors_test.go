package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		config         *CORSConfig
		method         string
		origin         string
		expectedStatus int
		expectedHeaders map[string]string
	}{
		{
			name:   "Default config with wildcard origin",
			config: DefaultCORSConfig(),
			method: http.MethodGet,
			origin: "https://example.com",
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin": "*",
			},
		},
		{
			name: "Specific allowed origin",
			config: &CORSConfig{
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
			config: &CORSConfig{
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
			config: &CORSConfig{
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
			config: &CORSConfig{
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
			router.Use(CORSMiddleware(tt.config))
			
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
	config := DefaultCORSConfig()

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

func TestIsOriginAllowed(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		allowedOrigins []string
		expected       bool
	}{
		{
			name:           "Wildcard allows all",
			origin:         "https://example.com",
			allowedOrigins: []string{"*"},
			expected:       true,
		},
		{
			name:           "Exact match",
			origin:         "https://example.com",
			allowedOrigins: []string{"https://example.com", "https://test.com"},
			expected:       true,
		},
		{
			name:           "No match",
			origin:         "https://notallowed.com",
			allowedOrigins: []string{"https://example.com", "https://test.com"},
			expected:       false,
		},
		{
			name:           "Empty allowed origins",
			origin:         "https://example.com",
			allowedOrigins: []string{},
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isOriginAllowed(tt.origin, tt.allowedOrigins)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJoinHeaders(t *testing.T) {
	tests := []struct {
		name     string
		headers  []string
		expected string
	}{
		{
			name:     "Multiple headers",
			headers:  []string{"Content-Type", "Authorization", "X-Custom"},
			expected: "Content-Type, Authorization, X-Custom",
		},
		{
			name:     "Single header",
			headers:  []string{"Content-Type"},
			expected: "Content-Type",
		},
		{
			name:     "Empty headers",
			headers:  []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinHeaders(tt.headers)
			assert.Equal(t, tt.expected, result)
		})
	}
}