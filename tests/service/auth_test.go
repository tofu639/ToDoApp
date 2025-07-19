package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"todo-api-backend/internal/model"
	"todo-api-backend/internal/service"
	"todo-api-backend/pkg/jwt"
	"todo-api-backend/pkg/password"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

// MockTodoRepository is a mock implementation of TodoRepository
type MockTodoRepository struct {
	mock.Mock
}

func (m *MockTodoRepository) Create(ctx context.Context, todo *model.Todo) error {
	args := m.Called(ctx, todo)
	return args.Error(0)
}

func (m *MockTodoRepository) GetByID(ctx context.Context, id uint, userID uint) (*model.Todo, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Todo), args.Error(1)
}

func (m *MockTodoRepository) GetByUserID(ctx context.Context, userID uint) ([]*model.Todo, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Todo), args.Error(1)
}

func (m *MockTodoRepository) Update(ctx context.Context, todo *model.Todo) error {
	args := m.Called(ctx, todo)
	return args.Error(0)
}

func (m *MockTodoRepository) Delete(ctx context.Context, id uint, userID uint) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func setupAuthService() (service.AuthService, *MockUserRepository, *jwt.TokenManager) {
	mockUserRepo := &MockUserRepository{}
	tokenManager := jwt.NewTokenManager("test-secret", 24)
	authService := service.NewAuthService(mockUserRepo, tokenManager)
	
	return authService, mockUserRepo, tokenManager
}

func TestAuthService_Register_Success(t *testing.T) {
	authService, mockUserRepo, _ := setupAuthService()
	ctx := context.Background()
	
	req := &model.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	
	// Mock user doesn't exist
	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(nil, gorm.ErrRecordNotFound)
	
	// Mock successful user creation
	mockUserRepo.On("Create", ctx, mock.AnythingOfType("*model.User")).Return(nil).Run(func(args mock.Arguments) {
		user := args.Get(1).(*model.User)
		user.ID = 1 // Simulate database setting ID
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
	})
	
	// Call service
	response, err := authService.Register(ctx, req)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, req.Email, response.User.Email)
	assert.Equal(t, uint(1), response.User.ID)
	
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Register_EmailExists(t *testing.T) {
	authService, mockUserRepo, _ := setupAuthService()
	ctx := context.Background()
	
	req := &model.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	
	existingUser := &model.User{
		ID:    1,
		Email: req.Email,
	}
	
	// Mock user already exists
	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(existingUser, nil)
	
	// Call service
	response, err := authService.Register(ctx, req)
	
	// Assertions
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, service.ErrEmailAlreadyExists, err)
	
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Register_WeakPassword(t *testing.T) {
	authService, mockUserRepo, _ := setupAuthService()
	ctx := context.Background()
	
	req := &model.RegisterRequest{
		Email:    "test@example.com",
		Password: "123", // Too short
	}
	
	// Mock user doesn't exist
	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(nil, gorm.ErrRecordNotFound)
	
	// Call service
	response, err := authService.Register(ctx, req)
	
	// Assertions
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "password validation failed")
	
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Register_DatabaseError(t *testing.T) {
	authService, mockUserRepo, _ := setupAuthService()
	ctx := context.Background()
	
	req := &model.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	
	// Mock user doesn't exist
	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(nil, gorm.ErrRecordNotFound)
	
	// Mock database error during creation
	mockUserRepo.On("Create", ctx, mock.AnythingOfType("*model.User")).Return(errors.New("database error"))
	
	// Call service
	response, err := authService.Register(ctx, req)
	
	// Assertions
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to create user")
	
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Login_Success(t *testing.T) {
	authService, mockUserRepo, _ := setupAuthService()
	ctx := context.Background()
	
	req := &model.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	
	// Create a user with hashed password using the actual password package
	hashedPassword, err := password.Hash("password123")
	assert.NoError(t, err)
	
	user := &model.User{
		ID:       1,
		Email:    req.Email,
		Password: hashedPassword,
	}
	
	// Mock user exists
	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(user, nil)
	
	// Call service
	response, err := authService.Login(ctx, req)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, req.Email, response.User.Email)
	assert.Equal(t, uint(1), response.User.ID)
	
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	authService, mockUserRepo, _ := setupAuthService()
	ctx := context.Background()
	
	req := &model.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}
	
	// Mock user doesn't exist
	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(nil, gorm.ErrRecordNotFound)
	
	// Call service
	response, err := authService.Login(ctx, req)
	
	// Assertions
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, service.ErrInvalidCredentials, err)
	
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	authService, mockUserRepo, _ := setupAuthService()
	ctx := context.Background()
	
	req := &model.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	
	// Create a user with different hashed password
	hashedPassword, err := password.Hash("correctpassword")
	assert.NoError(t, err)
	
	user := &model.User{
		ID:       1,
		Email:    req.Email,
		Password: hashedPassword,
	}
	
	// Mock user exists
	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(user, nil)
	
	// Call service
	response, err := authService.Login(ctx, req)
	
	// Assertions
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, service.ErrInvalidCredentials, err)
	
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_ValidateToken_Success(t *testing.T) {
	authService, _, tokenManager := setupAuthService()
	
	// Generate a valid token
	token, err := tokenManager.GenerateToken(1, "test@example.com")
	assert.NoError(t, err)
	
	// Call service
	claims, err := authService.ValidateToken(token)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
}

func TestAuthService_ValidateToken_InvalidToken(t *testing.T) {
	authService, _, _ := setupAuthService()
	
	// Call service with invalid token
	claims, err := authService.ValidateToken("invalid-token")
	
	// Assertions
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "token validation failed")
}