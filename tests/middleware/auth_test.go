package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"todo-api-backend/internal/middleware"
	"todo-api-backend/pkg/jwt"
)

func TestAuthMiddleware(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a token manager for testing
	tokenManager := jwt.NewTokenManager("test-secret-key", 24)

	// Generate a valid token for testing
	validToken, err := tokenManager.GenerateToken(1, "test@example.com")
	require.NoError(t, err)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
		shouldSetUser  bool
	}{
		{
			name:           "Valid token",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			shouldSetUser:  true,
		},
		{
			name:           "Missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"unauthorized","message":"Authorization header is required"}`,
			shouldSetUser:  false,
		},
		{
			name:           "Invalid bearer prefix",
			authHeader:     "Basic " + validToken,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"unauthorized","message":"Authorization header must start with 'Bearer '"}`,
			shouldSetUser:  false,
		},
		{
			name:           "Empty token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"unauthorized","message":"Token is required"}`,
			shouldSetUser:  false,
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"unauthorized","message":"Invalid token"}`,
			shouldSetUser:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new Gin router
			router := gin.New()
			
			// Add the auth middleware
			router.Use(middleware.AuthMiddleware(tokenManager))
			
			// Add a test route
			router.GET("/test", func(c *gin.Context) {
				userID, exists := middleware.GetUserID(c)
				if exists {
					c.JSON(http.StatusOK, gin.H{"user_id": userID})
				} else {
					c.JSON(http.StatusOK, gin.H{"message": "no user"})
				}
			})

			// Create a test request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create a response recorder
			w := httptest.NewRecorder()

			// Perform the request
			router.ServeHTTP(w, req)

			// Assert the response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}

			if tt.shouldSetUser && w.Code == http.StatusOK {
				assert.Contains(t, w.Body.String(), `"user_id":1`)
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		setupFunc   func(*gin.Context)
		expectedID  uint
		expectedOK  bool
	}{
		{
			name: "User ID exists",
			setupFunc: func(c *gin.Context) {
				c.Set(middleware.UserIDKey, uint(123))
			},
			expectedID: 123,
			expectedOK: true,
		},
		{
			name: "User ID does not exist",
			setupFunc: func(c *gin.Context) {
				// Don't set anything
			},
			expectedID: 0,
			expectedOK: false,
		},
		{
			name: "User ID has wrong type",
			setupFunc: func(c *gin.Context) {
				c.Set(middleware.UserIDKey, "not-a-uint")
			},
			expectedID: 0,
			expectedOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a Gin context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup the context
			tt.setupFunc(c)

			// Test GetUserID
			userID, ok := middleware.GetUserID(c)

			assert.Equal(t, tt.expectedID, userID)
			assert.Equal(t, tt.expectedOK, ok)
		})
	}
}

func TestGetUserEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		setupFunc     func(*gin.Context)
		expectedEmail string
		expectedOK    bool
	}{
		{
			name: "User email exists",
			setupFunc: func(c *gin.Context) {
				c.Set(middleware.UserEmailKey, "test@example.com")
			},
			expectedEmail: "test@example.com",
			expectedOK:    true,
		},
		{
			name: "User email does not exist",
			setupFunc: func(c *gin.Context) {
				// Don't set anything
			},
			expectedEmail: "",
			expectedOK:    false,
		},
		{
			name: "User email has wrong type",
			setupFunc: func(c *gin.Context) {
				c.Set(middleware.UserEmailKey, 123)
			},
			expectedEmail: "",
			expectedOK:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a Gin context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup the context
			tt.setupFunc(c)

			// Test GetUserEmail
			email, ok := middleware.GetUserEmail(c)

			assert.Equal(t, tt.expectedEmail, email)
			assert.Equal(t, tt.expectedOK, ok)
		})
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a token manager with very short expiration for testing
	tokenManager := jwt.NewTokenManager("test-secret-key", 0) // 0 hours = immediate expiration
	
	// Generate a token that will be expired
	expiredToken, err := tokenManager.GenerateToken(1, "test@example.com")
	require.NoError(t, err)

	// Create a new Gin router
	router := gin.New()
	
	// Add the auth middleware
	router.Use(middleware.AuthMiddleware(tokenManager))
	
	// Add a test route
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create a test request with expired token
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Token has expired")
}

func TestAuthMiddleware_TokenValidationScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokenManager := jwt.NewTokenManager("test-secret-key", 24)

	tests := []struct {
		name           string
		setupToken     func() string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Token with wrong signing method",
			setupToken: func() string {
				// This would require creating a token with wrong signing method
				// For simplicity, we'll use a malformed token
				return "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature"
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid token",
		},
		{
			name: "Malformed token",
			setupToken: func() string {
				return "not.a.valid.jwt.token"
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid token",
		},
		{
			name: "Token with invalid claims",
			setupToken: func() string {
				// Create a token with different secret to make it invalid
				wrongTokenManager := jwt.NewTokenManager("wrong-secret", 24)
				token, _ := wrongTokenManager.GenerateToken(1, "test@example.com")
				return token
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new Gin router
			router := gin.New()
			
			// Add the auth middleware
			router.Use(middleware.AuthMiddleware(tokenManager))
			
			// Add a test route
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Create a test request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", "Bearer "+tt.setupToken())

			// Create a response recorder
			w := httptest.NewRecorder()

			// Perform the request
			router.ServeHTTP(w, req)

			// Assert the response
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedError)
		})
	}
}