package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"todo-api-backend/internal/handler"
	"todo-api-backend/internal/model"
	"todo-api-backend/internal/service"
	"todo-api-backend/pkg/jwt"
)

// MockAuthService is a mock implementation of AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AuthResponse), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AuthResponse), args.Error(1)
}

func (m *MockAuthService) ValidateToken(tokenString string) (*jwt.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Claims), args.Error(1)
}

// MockTodoService is a mock implementation of TodoService
type MockTodoService struct {
	mock.Mock
}

func (m *MockTodoService) Create(ctx context.Context, req *model.CreateTodoRequest, userID uint) (*model.Todo, error) {
	args := m.Called(ctx, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Todo), args.Error(1)
}

func (m *MockTodoService) GetByID(ctx context.Context, id uint, userID uint) (*model.Todo, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Todo), args.Error(1)
}

func (m *MockTodoService) GetByUserID(ctx context.Context, userID uint) ([]*model.Todo, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Todo), args.Error(1)
}

func (m *MockTodoService) Update(ctx context.Context, id uint, req *model.UpdateTodoRequest, userID uint) (*model.Todo, error) {
	args := m.Called(ctx, id, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Todo), args.Error(1)
}

func (m *MockTodoService) Delete(ctx context.Context, id uint, userID uint) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func setupTestHandler() (*handler.Handler, *MockAuthService, *MockTodoService) {
	gin.SetMode(gin.TestMode)
	
	mockAuthService := &MockAuthService{}
	mockTodoService := &MockTodoService{}
	
	services := &service.Services{
		Auth: mockAuthService,
		Todo: mockTodoService,
	}
	
	h := handler.NewHandler(services)
	return h, mockAuthService, mockTodoService
}

func TestRegister_Success(t *testing.T) {
	h, mockAuthService, _ := setupTestHandler()
	
	// Setup request
	reqBody := model.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	
	expectedResponse := &model.AuthResponse{
		Token: "jwt-token",
		User: &model.UserInfo{
			ID:    1,
			Email: "test@example.com",
		},
	}
	
	// Setup mock
	mockAuthService.On("Register", mock.Anything, &reqBody).Return(expectedResponse, nil)
	
	// Create request
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// Call handler
	h.Register(c)
	
	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response model.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse.Token, response.Token)
	assert.Equal(t, expectedResponse.User.Email, response.User.Email)
	
	mockAuthService.AssertExpectations(t)
}

func TestRegister_InvalidJSON(t *testing.T) {
	h, _, _ := setupTestHandler()
	
	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	
	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// Call handler
	h.Register(c)
	
	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response model.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid_request", response.Error)
}

func TestRegister_ValidationError(t *testing.T) {
	h, _, _ := setupTestHandler()
	
	// Setup request with invalid data
	reqBody := model.RegisterRequest{
		Email:    "invalid-email",
		Password: "123", // Too short
	}
	
	// Create request
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// Call handler
	h.Register(c)
	
	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response model.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "validation_failed", response.Error)
	assert.NotEmpty(t, response.Details)
}

func TestRegister_EmailExists(t *testing.T) {
	h, mockAuthService, _ := setupTestHandler()
	
	// Setup request
	reqBody := model.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	
	// Setup mock to return email exists error
	mockAuthService.On("Register", mock.Anything, &reqBody).Return(nil, errors.New("email already exists"))
	
	// Create request
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// Call handler
	h.Register(c)
	
	// Assertions
	assert.Equal(t, http.StatusConflict, w.Code)
	
	var response model.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "email_exists", response.Error)
	
	mockAuthService.AssertExpectations(t)
}

func TestRegister_ServiceError(t *testing.T) {
	h, mockAuthService, _ := setupTestHandler()
	
	// Setup request
	reqBody := model.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	
	// Setup mock to return generic error
	mockAuthService.On("Register", mock.Anything, &reqBody).Return(nil, errors.New("database error"))
	
	// Create request
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// Call handler
	h.Register(c)
	
	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	var response model.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "registration_failed", response.Error)
	
	mockAuthService.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	h, mockAuthService, _ := setupTestHandler()
	
	// Setup request
	reqBody := model.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	
	expectedResponse := &model.AuthResponse{
		Token: "jwt-token",
		User: &model.UserInfo{
			ID:    1,
			Email: "test@example.com",
		},
	}
	
	// Setup mock
	mockAuthService.On("Login", mock.Anything, &reqBody).Return(expectedResponse, nil)
	
	// Create request
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// Call handler
	h.Login(c)
	
	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response model.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse.Token, response.Token)
	assert.Equal(t, expectedResponse.User.Email, response.User.Email)
	
	mockAuthService.AssertExpectations(t)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	h, mockAuthService, _ := setupTestHandler()
	
	// Setup request
	reqBody := model.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	
	// Setup mock to return invalid credentials error
	mockAuthService.On("Login", mock.Anything, &reqBody).Return(nil, errors.New("invalid credentials"))
	
	// Create request
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// Call handler
	h.Login(c)
	
	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response model.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid_credentials", response.Error)
	
	mockAuthService.AssertExpectations(t)
}

func TestLogin_ValidationError(t *testing.T) {
	h, _, _ := setupTestHandler()
	
	// Setup request with invalid data
	reqBody := model.LoginRequest{
		Email:    "invalid-email",
		Password: "", // Empty password
	}
	
	// Create request
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// Call handler
	h.Login(c)
	
	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response model.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "validation_failed", response.Error)
	assert.NotEmpty(t, response.Details)
}