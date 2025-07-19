package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSConfig holds CORS configuration options
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"Accept",
			"X-Requested-With",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
		},
		AllowCredentials: false,
		MaxAge:           12 * 3600, // 12 hours
	}
}

// CORSMiddleware creates a CORS middleware with the given configuration
func CORSMiddleware(config *CORSConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultCORSConfig()
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Set Access-Control-Allow-Origin
		if len(config.AllowOrigins) == 1 && config.AllowOrigins[0] == "*" {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if origin != "" && isOriginAllowed(origin, config.AllowOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		// Set Access-Control-Allow-Credentials
		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// Set Access-Control-Expose-Headers
		if len(config.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", joinHeaders(config.ExposeHeaders))
		}

		// Handle preflight OPTIONS request
		if c.Request.Method == http.MethodOptions {
			// Set Access-Control-Allow-Methods
			if len(config.AllowMethods) > 0 {
				c.Header("Access-Control-Allow-Methods", joinHeaders(config.AllowMethods))
			}

			// Set Access-Control-Allow-Headers
			if len(config.AllowHeaders) > 0 {
				c.Header("Access-Control-Allow-Headers", joinHeaders(config.AllowHeaders))
			}

			// Set Access-Control-Max-Age
			if config.MaxAge > 0 {
				c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
			}

			// Return 204 No Content for preflight requests
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// Continue to the next handler
		c.Next()
	}
}

// isOriginAllowed checks if the given origin is allowed
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}
	}
	return false
}

// joinHeaders joins a slice of strings with commas
func joinHeaders(headers []string) string {
	if len(headers) == 0 {
		return ""
	}
	
	result := headers[0]
	for i := 1; i < len(headers); i++ {
		result += ", " + headers[i]
	}
	return result
}