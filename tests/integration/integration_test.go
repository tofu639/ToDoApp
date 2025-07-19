package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"todo-api-backend/internal/handler"
	"todo-api-backend/internal/middleware"
	"todo-api-backend/internal/model"
	"todo-api-backend/internal/repository"
	"todo-api-backend/internal/service"
	"todo-api-backend/pkg/jwt"
)

// IntegrationTestSuite defines the test suite for integration tests
type IntegrationTestSuite struct {
	suite.Suite
	db           *gorm.DB
	router       *gin.Engine
	tokenManager *jwt.TokenManager
	testUser     *model.User
	testToken    string
}

// SetupSuite runs once before all tests in the suite
func (suite *IntegrationTestSuite) SetupSuite() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Setup test database connection
	suite.setupTestDatabase()

	// Setup JWT token manager
	suite.tokenManager = jwt.NewTokenManager("test-secret-key", 24)

	// Setup repositories
	repos := repository.NewRepositories(suite.db)

	// Setup services
	services := service.NewServices(repos, suite.tokenManager)

	// Setup handlers
	h := handler.NewHandler(services)

	// Setup router with middleware
	suite.router = gin.New()
	suite.router.Use(gin.Recovery())
	suite.router.Use(middleware.CORSMiddleware(nil))

	// Auth routes (no middleware)
	auth := suite.router.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
	}

	// Protected routes (with JWT middleware)
	api := suite.router.Group("/api")
	api.Use(middleware.AuthMiddleware(suite.tokenManager))
	{
		todos := api.Group("/todos")
		{
			todos.POST("", h.CreateTodo)
			todos.GET("", h.GetTodos)
			todos.GET("/:id", h.GetTodo)
			todos.PUT("/:id", h.UpdateTodo)
			todos.DELETE("/:id", h.DeleteTodo)
		}
	}

	// Create test user and token
	suite.createTestUser()
}

// setupTestDatabase initializes the test database connection
func (suite *IntegrationTestSuite) setupTestDatabase() {
	// Use test database URL from environment or skip tests
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		suite.T().Skip("TEST_DATABASE_URL not set. To run integration tests, set TEST_DATABASE_URL environment variable. Example: TEST_DATABASE_URL=postgres://postgres:password@localhost:5432/todoapi_test?sslmode=disable")
	}

	var err error
	suite.db, err = gorm.Open(postgres.Open(testDBURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(suite.T(), err, "Failed to connect to test database")

	// Auto-migrate the schema
	err = suite.db.AutoMigrate(&model.User{}, &model.Todo{})
	require.NoError(suite.T(), err, "Failed to migrate test database")
}

// createTestUser creates a test user and generates a JWT token
func (suite *IntegrationTestSuite) createTestUser() {
	suite.testUser = &model.User{
		Email:    "test@example.com",
		Password: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/RK.PmvlmO", // "password123"
	}

	err := suite.db.Create(suite.testUser).Error
	require.NoError(suite.T(), err, "Failed to create test user")

	// Generate JWT token for the test user
	suite.testToken, err = suite.tokenManager.GenerateToken(suite.testUser.ID, suite.testUser.Email)
	require.NoError(suite.T(), err, "Failed to generate test token")
}

// TearDownSuite runs once after all tests in the suite
func (suite *IntegrationTestSuite) TearDownSuite() {
	// Clean up test data
	suite.db.Exec("DELETE FROM todos")
	suite.db.Exec("DELETE FROM users")

	// Close database connection
	sqlDB, err := suite.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

// SetupTest runs before each test
func (suite *IntegrationTestSuite) SetupTest() {
	// Clean todos table before each test (keep test user)
	suite.db.Where("user_id = ?", suite.testUser.ID).Delete(&model.Todo{})
}

// TestAuthenticationFlow tests the complete authentication flow
func (suite *IntegrationTestSuite) TestAuthenticationFlow() {
	suite.Run("Register new user", func() {
		registerReq := map[string]string{
			"email":    "newuser@example.com",
			"password": "newpassword123",
		}

		body, _ := json.Marshal(registerReq)
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		assert.Contains(suite.T(), response, "token")
		assert.Contains(suite.T(), response, "user")
	})

	suite.Run("Register with duplicate email", func() {
		registerReq := map[string]string{
			"email":    "test@example.com", // Already exists
			"password": "password123",
		}

		body, _ := json.Marshal(registerReq)
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusConflict, w.Code)
	})

	suite.Run("Login with valid credentials", func() {
		loginReq := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		assert.Contains(suite.T(), response, "token")
		assert.Contains(suite.T(), response, "user")
	})

	suite.Run("Login with invalid credentials", func() {
		loginReq := map[string]string{
			"email":    "test@example.com",
			"password": "wrongpassword",
		}

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	})
}

// TestTodoCRUDOperations tests complete CRUD operations for todos
func (suite *IntegrationTestSuite) TestTodoCRUDOperations() {
	var createdTodoID uint

	suite.Run("Create todo", func() {
		createReq := map[string]interface{}{
			"title":       "Test Todo",
			"description": "This is a test todo",
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/todos", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.testToken)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusCreated, w.Code)

		var response model.Todo
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "Test Todo", response.Title)
		assert.Equal(suite.T(), "This is a test todo", response.Description)
		assert.False(suite.T(), response.Completed)
		assert.Equal(suite.T(), suite.testUser.ID, response.UserID)

		createdTodoID = response.ID
	})

	suite.Run("Create todo without authentication", func() {
		createReq := map[string]interface{}{
			"title": "Unauthorized Todo",
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/todos", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	})

	suite.Run("Get all todos", func() {
		// Create another todo first
		todo2 := &model.Todo{
			Title:       "Second Todo",
			Description: "Another test todo",
			UserID:      suite.testUser.ID,
		}
		suite.db.Create(todo2)

		req := httptest.NewRequest("GET", "/api/todos", nil)
		req.Header.Set("Authorization", "Bearer "+suite.testToken)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var response []model.Todo
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		assert.Len(suite.T(), response, 2)
	})

	suite.Run("Get todo by ID", func() {
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/todos/%d", createdTodoID), nil)
		req.Header.Set("Authorization", "Bearer "+suite.testToken)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var response model.Todo
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), createdTodoID, response.ID)
		assert.Equal(suite.T(), "Test Todo", response.Title)
	})

	suite.Run("Get non-existent todo", func() {
		req := httptest.NewRequest("GET", "/api/todos/99999", nil)
		req.Header.Set("Authorization", "Bearer "+suite.testToken)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	})

	suite.Run("Update todo", func() {
		updateReq := map[string]interface{}{
			"title":       "Updated Todo Title",
			"description": "Updated description",
			"completed":   true,
		}

		body, _ := json.Marshal(updateReq)
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/todos/%d", createdTodoID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.testToken)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var response model.Todo
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "Updated Todo Title", response.Title)
		assert.Equal(suite.T(), "Updated description", response.Description)
		assert.True(suite.T(), response.Completed)
	})

	suite.Run("Update non-existent todo", func() {
		updateReq := map[string]interface{}{
			"title": "Should not work",
		}

		body, _ := json.Marshal(updateReq)
		req := httptest.NewRequest("PUT", "/api/todos/99999", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.testToken)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	})

	suite.Run("Delete todo", func() {
		req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/todos/%d", createdTodoID), nil)
		req.Header.Set("Authorization", "Bearer "+suite.testToken)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusNoContent, w.Code)

		// Verify todo is deleted
		var count int64
		suite.db.Model(&model.Todo{}).Where("id = ?", createdTodoID).Count(&count)
		assert.Equal(suite.T(), int64(0), count)
	})

	suite.Run("Delete non-existent todo", func() {
		req := httptest.NewRequest("DELETE", "/api/todos/99999", nil)
		req.Header.Set("Authorization", "Bearer "+suite.testToken)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	})
}

// TestUserIsolation tests that users can only access their own todos
func (suite *IntegrationTestSuite) TestUserIsolation() {
	// Create another user
	anotherUser := &model.User{
		Email:    "another@example.com",
		Password: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/RK.PmvlmO",
	}
	suite.db.Create(anotherUser)

	// Create a todo for the other user
	otherUserTodo := &model.Todo{
		Title:  "Other User's Todo",
		UserID: anotherUser.ID,
	}
	suite.db.Create(otherUserTodo)

	// Generate token for the other user
	otherUserToken, err := suite.tokenManager.GenerateToken(anotherUser.ID, anotherUser.Email)
	require.NoError(suite.T(), err)

	suite.Run("User cannot access other user's todo", func() {
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/todos/%d", otherUserTodo.ID), nil)
		req.Header.Set("Authorization", "Bearer "+suite.testToken)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	})

	suite.Run("User cannot update other user's todo", func() {
		updateReq := map[string]interface{}{
			"title": "Hacked Todo",
		}

		body, _ := json.Marshal(updateReq)
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/todos/%d", otherUserTodo.ID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.testToken)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	})

	suite.Run("User cannot delete other user's todo", func() {
		req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/todos/%d", otherUserTodo.ID), nil)
		req.Header.Set("Authorization", "Bearer "+suite.testToken)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	})

	suite.Run("Users only see their own todos in list", func() {
		// Create a todo for the test user
		testUserTodo := &model.Todo{
			Title:  "Test User's Todo",
			UserID: suite.testUser.ID,
		}
		suite.db.Create(testUserTodo)

		// Test user should only see their own todo
		req := httptest.NewRequest("GET", "/api/todos", nil)
		req.Header.Set("Authorization", "Bearer "+suite.testToken)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var response []model.Todo
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		assert.Len(suite.T(), response, 1)
		assert.Equal(suite.T(), "Test User's Todo", response[0].Title)

		// Other user should only see their own todo
		req = httptest.NewRequest("GET", "/api/todos", nil)
		req.Header.Set("Authorization", "Bearer "+otherUserToken)
		w = httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)

		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		assert.Len(suite.T(), response, 1)
		assert.Equal(suite.T(), "Other User's Todo", response[0].Title)
	})
}

// TestInputValidation tests input validation for API endpoints
func (suite *IntegrationTestSuite) TestInputValidation() {
	suite.Run("Register with invalid email", func() {
		registerReq := map[string]string{
			"email":    "invalid-email",
			"password": "password123",
		}

		body, _ := json.Marshal(registerReq)
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	})

	suite.Run("Register with short password", func() {
		registerReq := map[string]string{
			"email":    "valid@example.com",
			"password": "short",
		}

		body, _ := json.Marshal(registerReq)
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	})

	suite.Run("Create todo with empty title", func() {
		createReq := map[string]interface{}{
			"title":       "",
			"description": "Valid description",
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/todos", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.testToken)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	})

	suite.Run("Create todo with missing title", func() {
		createReq := map[string]interface{}{
			"description": "Valid description",
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/todos", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.testToken)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	})
}

// TestJWTAuthentication tests JWT token validation
func (suite *IntegrationTestSuite) TestJWTAuthentication() {
	suite.Run("Access protected endpoint without token", func() {
		req := httptest.NewRequest("GET", "/api/todos", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	})

	suite.Run("Access protected endpoint with invalid token", func() {
		req := httptest.NewRequest("GET", "/api/todos", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	})

	suite.Run("Access protected endpoint with expired token", func() {
		// Create a token manager with very short expiration (1 nanosecond in hours)
		shortTokenManager := jwt.NewTokenManager("test-secret", 0) // 0 hours = immediate expiration
		expiredToken, err := shortTokenManager.GenerateToken(suite.testUser.ID, suite.testUser.Email)
		require.NoError(suite.T(), err)

		// Wait for token to expire
		time.Sleep(1 * time.Millisecond)

		req := httptest.NewRequest("GET", "/api/todos", nil)
		req.Header.Set("Authorization", "Bearer "+expiredToken)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	})

	suite.Run("Access protected endpoint with malformed authorization header", func() {
		req := httptest.NewRequest("GET", "/api/todos", nil)
		req.Header.Set("Authorization", "InvalidFormat "+suite.testToken)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	})
}

// TestDatabaseInteractions tests database-specific behaviors
func (suite *IntegrationTestSuite) TestDatabaseInteractions() {
	suite.Run("Database constraints and relationships", func() {
		// Test foreign key constraint
		todo := &model.Todo{
			Title:  "Orphaned Todo",
			UserID: 99999, // Non-existent user ID
		}
		
		err := suite.db.Create(todo).Error
		// This should fail due to foreign key constraint
		assert.Error(suite.T(), err)
	})

	suite.Run("Database transaction rollback", func() {
		// Start a transaction
		tx := suite.db.Begin()
		
		// Create a user in transaction
		user := &model.User{
			Email:    "transaction@example.com",
			Password: "hashedpassword",
		}
		tx.Create(user)
		
		// Rollback the transaction
		tx.Rollback()
		
		// Verify user was not created
		var count int64
		suite.db.Model(&model.User{}).Where("email = ?", "transaction@example.com").Count(&count)
		assert.Equal(suite.T(), int64(0), count)
	})

	suite.Run("Concurrent access handling", func() {
		// Create a todo
		todo := &model.Todo{
			Title:  "Concurrent Todo",
			UserID: suite.testUser.ID,
		}
		suite.db.Create(todo)

		// Simulate concurrent updates
		var todo1, todo2 model.Todo
		suite.db.First(&todo1, todo.ID)
		suite.db.First(&todo2, todo.ID)

		// Update from first "session"
		todo1.Title = "Updated by Session 1"
		suite.db.Save(&todo1)

		// Update from second "session"
		todo2.Title = "Updated by Session 2"
		suite.db.Save(&todo2)

		// Verify final state
		var finalTodo model.Todo
		suite.db.First(&finalTodo, todo.ID)
		assert.Equal(suite.T(), "Updated by Session 2", finalTodo.Title)
	})
}

// TestAPIErrorHandling tests comprehensive error scenarios
func (suite *IntegrationTestSuite) TestAPIErrorHandling() {
	suite.Run("Malformed JSON requests", func() {
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBufferString("{invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	})

	suite.Run("Missing Content-Type header", func() {
		registerReq := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}

		body, _ := json.Marshal(registerReq)
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
		// Intentionally not setting Content-Type
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		// Should still work or return appropriate error
		assert.True(suite.T(), w.Code == http.StatusBadRequest || w.Code == http.StatusConflict)
	})

	suite.Run("Very large request body", func() {
		// Create a very large description
		largeDescription := make([]byte, 10000) // 10KB
		for i := range largeDescription {
			largeDescription[i] = 'a'
		}

		createReq := map[string]interface{}{
			"title":       "Large Todo",
			"description": string(largeDescription),
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/todos", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.testToken)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		// Should handle large requests appropriately
		assert.True(suite.T(), w.Code == http.StatusBadRequest || w.Code == http.StatusCreated)
	})
}

// TestCompleteWorkflows tests end-to-end user workflows
func (suite *IntegrationTestSuite) TestCompleteWorkflows() {
	suite.Run("Complete user journey", func() {
		// 1. Register a new user
		registerReq := map[string]string{
			"email":    "journey@example.com",
			"password": "password123",
		}

		body, _ := json.Marshal(registerReq)
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)
		assert.Equal(suite.T(), http.StatusCreated, w.Code)

		var registerResponse map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &registerResponse)
		userToken := registerResponse["token"].(string)

		// 2. Create multiple todos
		todoTitles := []string{"First Todo", "Second Todo", "Third Todo"}
		var todoIDs []uint

		for _, title := range todoTitles {
			createReq := map[string]interface{}{
				"title":       title,
				"description": "Description for " + title,
			}

			body, _ := json.Marshal(createReq)
			req := httptest.NewRequest("POST", "/api/todos", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+userToken)
			w := httptest.NewRecorder()

			suite.router.ServeHTTP(w, req)
			assert.Equal(suite.T(), http.StatusCreated, w.Code)

			var todo model.Todo
			json.Unmarshal(w.Body.Bytes(), &todo)
			todoIDs = append(todoIDs, todo.ID)
		}

		// 3. List all todos
		req = httptest.NewRequest("GET", "/api/todos", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w = httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)
		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var todos []model.Todo
		json.Unmarshal(w.Body.Bytes(), &todos)
		assert.Len(suite.T(), todos, 3)

		// 4. Update one todo
		updateReq := map[string]interface{}{
			"title":     "Updated First Todo",
			"completed": true,
		}

		body, _ = json.Marshal(updateReq)
		req = httptest.NewRequest("PUT", fmt.Sprintf("/api/todos/%d", todoIDs[0]), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+userToken)
		w = httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)
		assert.Equal(suite.T(), http.StatusOK, w.Code)

		// 5. Delete one todo
		req = httptest.NewRequest("DELETE", fmt.Sprintf("/api/todos/%d", todoIDs[1]), nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w = httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)
		assert.Equal(suite.T(), http.StatusNoContent, w.Code)

		// 6. Verify final state
		req = httptest.NewRequest("GET", "/api/todos", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w = httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)
		assert.Equal(suite.T(), http.StatusOK, w.Code)

		json.Unmarshal(w.Body.Bytes(), &todos)
		assert.Len(suite.T(), todos, 2) // One deleted
		
		// Find the updated todo
		var updatedTodo *model.Todo
		for _, todo := range todos {
			if todo.ID == todoIDs[0] {
				updatedTodo = &todo
				break
			}
		}
		assert.NotNil(suite.T(), updatedTodo)
		assert.Equal(suite.T(), "Updated First Todo", updatedTodo.Title)
		assert.True(suite.T(), updatedTodo.Completed)
	})
}

// TestIntegrationTestSuite runs the integration test suite
func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}