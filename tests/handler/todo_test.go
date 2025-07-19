package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"todo-api-backend/internal/middleware"
	"todo-api-backend/internal/model"
)

func setupTodoTestContext(userID uint) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	// Set user ID in context (simulating JWT middleware)
	c.Set(middleware.UserIDKey, userID)
	
	return c
}

func TestCreateTodo_Success(t *testing.T) {
	h, _, mockTodoService := setupTestHandler()
	
	// Setup request
	reqBody := model.CreateTodoRequest{
		Title:       "Test Todo",
		Description: "Test Description",
	}
	
	expectedTodo := &model.Todo{
		ID:          1,
		Title:       "Test Todo",
		Description: "Test Description",
		Completed:   false,
		UserID:      1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// Setup mock
	mockTodoService.On("Create", mock.Anything, &reqBody, uint(1)).Return(expectedTodo, nil)
	
	// Create request
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Create context with user ID
	c := setupTodoTestContext(1)
	c.Request = req
	
	// Call handler
	h.CreateTodo(c)
	
	// Assertions
	assert.Equal(t, http.StatusCreated, c.Writer.Status())
	
	mockTodoService.AssertExpectations(t)
}

func TestCreateTodo_NoUserID(t *testing.T) {
	h, _, _ := setupTestHandler()
	
	// Setup request
	reqBody := model.CreateTodoRequest{
		Title:       "Test Todo",
		Description: "Test Description",
	}
	
	// Create request
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Create context without user ID
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// Call handler
	h.CreateTodo(c)
	
	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response model.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", response.Error)
}

func TestCreateTodo_InvalidJSON(t *testing.T) {
	h, _, _ := setupTestHandler()
	
	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	
	// Create context with user ID
	c := setupTodoTestContext(1)
	c.Request = req
	
	// Call handler
	h.CreateTodo(c)
	
	// Assertions
	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
}

func TestCreateTodo_ValidationError(t *testing.T) {
	h, _, _ := setupTestHandler()
	
	// Setup request with invalid data (empty title)
	reqBody := model.CreateTodoRequest{
		Title:       "", // Empty title should fail validation
		Description: "Test Description",
	}
	
	// Create request
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Create context with user ID
	c := setupTodoTestContext(1)
	c.Request = req
	
	// Call handler
	h.CreateTodo(c)
	
	// Assertions
	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
}

func TestCreateTodo_ServiceError(t *testing.T) {
	h, _, mockTodoService := setupTestHandler()
	
	// Setup request
	reqBody := model.CreateTodoRequest{
		Title:       "Test Todo",
		Description: "Test Description",
	}
	
	// Setup mock to return error
	mockTodoService.On("Create", mock.Anything, &reqBody, uint(1)).Return(nil, errors.New("database error"))
	
	// Create request
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Create context with user ID
	c := setupTodoTestContext(1)
	c.Request = req
	
	// Call handler
	h.CreateTodo(c)
	
	// Assertions
	assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())
	
	mockTodoService.AssertExpectations(t)
}

func TestGetTodos_Success(t *testing.T) {
	h, _, mockTodoService := setupTestHandler()
	
	expectedTodos := []*model.Todo{
		{
			ID:          1,
			Title:       "Todo 1",
			Description: "Description 1",
			Completed:   false,
			UserID:      1,
		},
		{
			ID:          2,
			Title:       "Todo 2",
			Description: "Description 2",
			Completed:   true,
			UserID:      1,
		},
	}
	
	// Setup mock
	mockTodoService.On("GetByUserID", mock.Anything, uint(1)).Return(expectedTodos, nil)
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/todos", nil)
	
	// Create context with user ID
	c := setupTodoTestContext(1)
	c.Request = req
	
	// Call handler
	h.GetTodos(c)
	
	// Assertions
	assert.Equal(t, http.StatusOK, c.Writer.Status())
	
	mockTodoService.AssertExpectations(t)
}

func TestGetTodos_NoUserID(t *testing.T) {
	h, _, _ := setupTestHandler()
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/todos", nil)
	
	// Create context without user ID
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// Call handler
	h.GetTodos(c)
	
	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetTodos_ServiceError(t *testing.T) {
	h, _, mockTodoService := setupTestHandler()
	
	// Setup mock to return error
	mockTodoService.On("GetByUserID", mock.Anything, uint(1)).Return(nil, errors.New("database error"))
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/todos", nil)
	
	// Create context with user ID
	c := setupTodoTestContext(1)
	c.Request = req
	
	// Call handler
	h.GetTodos(c)
	
	// Assertions
	assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())
	
	mockTodoService.AssertExpectations(t)
}

func TestGetTodo_Success(t *testing.T) {
	h, _, mockTodoService := setupTestHandler()
	
	expectedTodo := &model.Todo{
		ID:          1,
		Title:       "Test Todo",
		Description: "Test Description",
		Completed:   false,
		UserID:      1,
	}
	
	// Setup mock
	mockTodoService.On("GetByID", mock.Anything, uint(1), uint(1)).Return(expectedTodo, nil)
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/todos/1", nil)
	
	// Create context with user ID and URL param
	c := setupTodoTestContext(1)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	
	// Call handler
	h.GetTodo(c)
	
	// Assertions
	assert.Equal(t, http.StatusOK, c.Writer.Status())
	
	mockTodoService.AssertExpectations(t)
}

func TestGetTodo_InvalidID(t *testing.T) {
	h, _, _ := setupTestHandler()
	
	// Create request with invalid ID
	req := httptest.NewRequest(http.MethodGet, "/todos/invalid", nil)
	
	// Create context with user ID and invalid URL param
	c := setupTodoTestContext(1)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	
	// Call handler
	h.GetTodo(c)
	
	// Assertions
	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
}

func TestGetTodo_NotFound(t *testing.T) {
	h, _, mockTodoService := setupTestHandler()
	
	// Setup mock to return not found error
	mockTodoService.On("GetByID", mock.Anything, uint(1), uint(1)).Return(nil, errors.New("todo not found"))
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/todos/1", nil)
	
	// Create context with user ID and URL param
	c := setupTodoTestContext(1)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	
	// Call handler
	h.GetTodo(c)
	
	// Assertions
	assert.Equal(t, http.StatusNotFound, c.Writer.Status())
	
	mockTodoService.AssertExpectations(t)
}

func TestUpdateTodo_Success(t *testing.T) {
	h, _, mockTodoService := setupTestHandler()
	
	// Setup request
	title := "Updated Todo"
	completed := true
	reqBody := model.UpdateTodoRequest{
		Title:     &title,
		Completed: &completed,
	}
	
	expectedTodo := &model.Todo{
		ID:          1,
		Title:       "Updated Todo",
		Description: "Test Description",
		Completed:   true,
		UserID:      1,
	}
	
	// Setup mock
	mockTodoService.On("Update", mock.Anything, uint(1), &reqBody, uint(1)).Return(expectedTodo, nil)
	
	// Create request
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/todos/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Create context with user ID and URL param
	c := setupTodoTestContext(1)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	
	// Call handler
	h.UpdateTodo(c)
	
	// Assertions
	assert.Equal(t, http.StatusOK, c.Writer.Status())
	
	mockTodoService.AssertExpectations(t)
}

func TestUpdateTodo_InvalidID(t *testing.T) {
	h, _, _ := setupTestHandler()
	
	// Setup request
	title := "Updated Todo"
	reqBody := model.UpdateTodoRequest{
		Title: &title,
	}
	
	// Create request
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/todos/invalid", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Create context with user ID and invalid URL param
	c := setupTodoTestContext(1)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	
	// Call handler
	h.UpdateTodo(c)
	
	// Assertions
	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
}

func TestUpdateTodo_NotFound(t *testing.T) {
	h, _, mockTodoService := setupTestHandler()
	
	// Setup request
	title := "Updated Todo"
	reqBody := model.UpdateTodoRequest{
		Title: &title,
	}
	
	// Setup mock to return not found error
	mockTodoService.On("Update", mock.Anything, uint(1), &reqBody, uint(1)).Return(nil, errors.New("todo not found"))
	
	// Create request
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/todos/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Create context with user ID and URL param
	c := setupTodoTestContext(1)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	
	// Call handler
	h.UpdateTodo(c)
	
	// Assertions
	assert.Equal(t, http.StatusNotFound, c.Writer.Status())
	
	mockTodoService.AssertExpectations(t)
}

func TestDeleteTodo_Success(t *testing.T) {
	h, _, mockTodoService := setupTestHandler()
	
	// Setup mock
	mockTodoService.On("Delete", mock.Anything, uint(1), uint(1)).Return(nil)
	
	// Create request
	req := httptest.NewRequest(http.MethodDelete, "/todos/1", nil)
	
	// Create context with user ID and URL param
	c := setupTodoTestContext(1)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	
	// Call handler
	h.DeleteTodo(c)
	
	// Assertions
	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	
	mockTodoService.AssertExpectations(t)
}

func TestDeleteTodo_InvalidID(t *testing.T) {
	h, _, _ := setupTestHandler()
	
	// Create request with invalid ID
	req := httptest.NewRequest(http.MethodDelete, "/todos/invalid", nil)
	
	// Create context with user ID and invalid URL param
	c := setupTodoTestContext(1)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	
	// Call handler
	h.DeleteTodo(c)
	
	// Assertions
	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
}

func TestDeleteTodo_NotFound(t *testing.T) {
	h, _, mockTodoService := setupTestHandler()
	
	// Setup mock to return not found error
	mockTodoService.On("Delete", mock.Anything, uint(1), uint(1)).Return(errors.New("todo not found"))
	
	// Create request
	req := httptest.NewRequest(http.MethodDelete, "/todos/1", nil)
	
	// Create context with user ID and URL param
	c := setupTodoTestContext(1)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	
	// Call handler
	h.DeleteTodo(c)
	
	// Assertions
	assert.Equal(t, http.StatusNotFound, c.Writer.Status())
	
	mockTodoService.AssertExpectations(t)
}