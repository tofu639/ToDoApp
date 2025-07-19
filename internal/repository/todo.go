package repository

import (
	"context"
	"errors"

	"todo-api-backend/internal/model"
	"gorm.io/gorm"
)

// todoRepository implements the TodoRepository interface
type todoRepository struct {
	db *gorm.DB
}

// NewTodoRepository creates a new todo repository instance
func NewTodoRepository(db *gorm.DB) TodoRepository {
	return &todoRepository{
		db: db,
	}
}

// Create creates a new todo in the database
func (r *todoRepository) Create(ctx context.Context, todo *model.Todo) error {
	if err := r.db.WithContext(ctx).Create(todo).Error; err != nil {
		return err
	}
	return nil
}

// GetByID retrieves a todo by ID, ensuring it belongs to the specified user
func (r *todoRepository) GetByID(ctx context.Context, id uint, userID uint) (*model.Todo, error) {
	var todo model.Todo
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&todo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &todo, nil
}

// GetByUserID retrieves all todos belonging to a specific user
func (r *todoRepository) GetByUserID(ctx context.Context, userID uint) ([]*model.Todo, error) {
	var todos []*model.Todo
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&todos).Error
	if err != nil {
		return nil, err
	}
	return todos, nil
}

// Update updates an existing todo
func (r *todoRepository) Update(ctx context.Context, todo *model.Todo) error {
	err := r.db.WithContext(ctx).Save(todo).Error
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a todo by ID, ensuring it belongs to the specified user
func (r *todoRepository) Delete(ctx context.Context, id uint, userID uint) error {
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&model.Todo{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}