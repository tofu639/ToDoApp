package service

import (
	"context"
	"errors"
	"fmt"

	"todo-api-backend/internal/model"
	"todo-api-backend/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrTodoNotFound      = errors.New("todo not found")
	ErrUnauthorizedAccess = errors.New("unauthorized access to todo")
)

// todoService implements the TodoService interface
type todoService struct {
	todoRepo repository.TodoRepository
	userRepo repository.UserRepository
}

// NewTodoService creates a new todo service
func NewTodoService(todoRepo repository.TodoRepository, userRepo repository.UserRepository) TodoService {
	return &todoService{
		todoRepo: todoRepo,
		userRepo: userRepo,
	}
}

// Create creates a new todo for the authenticated user
func (s *todoService) Create(ctx context.Context, req *model.CreateTodoRequest, userID uint) (*model.Todo, error) {
	// Verify user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}

	// Create new todo
	todo := &model.Todo{
		Title:       req.Title,
		Description: req.Description,
		UserID:      userID,
		Completed:   false, // Default to false for new todos
	}

	if err := s.todoRepo.Create(ctx, todo); err != nil {
		return nil, fmt.Errorf("failed to create todo: %w", err)
	}

	return todo, nil
}

// GetByID retrieves a specific todo by ID, ensuring user ownership
func (s *todoService) GetByID(ctx context.Context, id uint, userID uint) (*model.Todo, error) {
	todo, err := s.todoRepo.GetByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTodoNotFound
		}
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	// Double-check ownership (repository should handle this, but extra safety)
	if todo.UserID != userID {
		return nil, ErrUnauthorizedAccess
	}

	return todo, nil
}

// GetByUserID retrieves all todos belonging to the authenticated user
func (s *todoService) GetByUserID(ctx context.Context, userID uint) ([]*model.Todo, error) {
	// Verify user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}

	todos, err := s.todoRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get todos: %w", err)
	}

	// Return empty slice if no todos found (not an error)
	if todos == nil {
		todos = []*model.Todo{}
	}

	return todos, nil
}

// Update updates an existing todo, ensuring user ownership
func (s *todoService) Update(ctx context.Context, id uint, req *model.UpdateTodoRequest, userID uint) (*model.Todo, error) {
	// Get existing todo to verify ownership
	existingTodo, err := s.todoRepo.GetByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTodoNotFound
		}
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	// Double-check ownership
	if existingTodo.UserID != userID {
		return nil, ErrUnauthorizedAccess
	}

	// Update fields if provided
	if req.Title != nil {
		existingTodo.Title = *req.Title
	}
	if req.Description != nil {
		existingTodo.Description = *req.Description
	}
	if req.Completed != nil {
		existingTodo.Completed = *req.Completed
	}

	// Save updated todo
	if err := s.todoRepo.Update(ctx, existingTodo); err != nil {
		return nil, fmt.Errorf("failed to update todo: %w", err)
	}

	return existingTodo, nil
}

// Delete deletes a todo by ID, ensuring user ownership
func (s *todoService) Delete(ctx context.Context, id uint, userID uint) error {
	// Verify todo exists and belongs to user
	_, err := s.todoRepo.GetByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTodoNotFound
		}
		return fmt.Errorf("failed to verify todo ownership: %w", err)
	}

	// Delete the todo
	if err := s.todoRepo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	return nil
}