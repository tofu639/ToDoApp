package repository

import (
	"context"

	"todo-api-backend/internal/model"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Create creates a new user in the database
	Create(ctx context.Context, user *model.User) error
	
	// GetByEmail retrieves a user by email address
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	
	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uint) (*model.User, error)
}

// TodoRepository defines the interface for todo data operations
type TodoRepository interface {
	// Create creates a new todo in the database
	Create(ctx context.Context, todo *model.Todo) error
	
	// GetByID retrieves a todo by ID, ensuring it belongs to the specified user
	GetByID(ctx context.Context, id uint, userID uint) (*model.Todo, error)
	
	// GetByUserID retrieves all todos belonging to a specific user
	GetByUserID(ctx context.Context, userID uint) ([]*model.Todo, error)
	
	// Update updates an existing todo
	Update(ctx context.Context, todo *model.Todo) error
	
	// Delete deletes a todo by ID, ensuring it belongs to the specified user
	Delete(ctx context.Context, id uint, userID uint) error
}

// Repositories holds all repository interfaces for dependency injection
type Repositories struct {
	User UserRepository
	Todo TodoRepository
}

// NewRepositories creates a new instance of Repositories with all implementations
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User: NewUserRepository(db),
		Todo: NewTodoRepository(db),
	}
}