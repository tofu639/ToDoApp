package service

import (
	"context"

	"todo-api-backend/internal/model"
	"todo-api-backend/internal/repository"
	"todo-api-backend/pkg/jwt"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	// Register creates a new user account with email validation and password hashing
	Register(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error)
	
	// Login authenticates a user with credential verification and JWT generation
	Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error)
	
	// ValidateToken validates a JWT token and returns the claims
	ValidateToken(tokenString string) (*jwt.Claims, error)
}

// TodoService defines the interface for todo business logic operations
type TodoService interface {
	// Create creates a new todo for the authenticated user
	Create(ctx context.Context, req *model.CreateTodoRequest, userID uint) (*model.Todo, error)
	
	// GetByID retrieves a specific todo by ID, ensuring user ownership
	GetByID(ctx context.Context, id uint, userID uint) (*model.Todo, error)
	
	// GetByUserID retrieves all todos belonging to the authenticated user
	GetByUserID(ctx context.Context, userID uint) ([]*model.Todo, error)
	
	// Update updates an existing todo, ensuring user ownership
	Update(ctx context.Context, id uint, req *model.UpdateTodoRequest, userID uint) (*model.Todo, error)
	
	// Delete deletes a todo by ID, ensuring user ownership
	Delete(ctx context.Context, id uint, userID uint) error
}

// Services holds all service interfaces for dependency injection
type Services struct {
	Auth AuthService
	Todo TodoService
}

// NewServices creates a new instance of Services with all implementations
func NewServices(repos *repository.Repositories, tokenManager *jwt.TokenManager) *Services {
	return &Services{
		Auth: NewAuthService(repos.User, tokenManager),
		Todo: NewTodoService(repos.Todo, repos.User),
	}
}