package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"todo-api-backend/pkg/jwt"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	UserIDKey          = "user_id"
	UserEmailKey       = "user_email"
)

// AuthMiddleware creates a JWT authentication middleware
func AuthMiddleware(tokenManager *jwt.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, BearerPrefix) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authorization header must start with 'Bearer '",
			})
			c.Abort()
			return
		}

		// Extract the token part
		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Token is required",
			})
			c.Abort()
			return
		}

		// Validate the token
		claims, err := tokenManager.ValidateToken(tokenString)
		if err != nil {
			var message string
			switch err {
			case jwt.ErrExpiredToken:
				message = "Token has expired"
			case jwt.ErrInvalidToken:
				message = "Invalid token"
			case jwt.ErrTokenClaims:
				message = "Invalid token claims"
			default:
				message = "Token validation failed"
			}

			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": message,
			})
			c.Abort()
			return
		}

		// Add user information to the context
		c.Set(UserIDKey, claims.UserID)
		c.Set(UserEmailKey, claims.Email)

		// Continue to the next handler
		c.Next()
	}
}

// GetUserID extracts the user ID from the Gin context
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return 0, false
	}

	id, ok := userID.(uint)
	return id, ok
}

// GetUserEmail extracts the user email from the Gin context
func GetUserEmail(c *gin.Context) (string, bool) {
	userEmail, exists := c.Get(UserEmailKey)
	if !exists {
		return "", false
	}

	email, ok := userEmail.(string)
	return email, ok
}