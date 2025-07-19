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
)

func setupTodoService() (service.TodoService, *MockTodoRepository, *MockUserRepository) {
	mockTodoRepo := &MockTodoRepository{}
	mockUserRepo := &MockUserRepository{}
	todoService := service.NewTodoService(mockTodoRepo, mockUserRepo)
	
	return todoService, mockTodoRepo, mockUserRepo
}

func TestTodoService_Create_Success(t *testing.T) {
	todoService, mockTodoRepo, mockUserRepo := setupTodoService()
	ctx := context.Background()
	
	req := &model.CreateTodoRequest{
		Title:       "Test Todo",
		Description: "Test Description",
	}
	userID := uint(1)
	
	user := &model.User{
		ID:    userID,
		Email: "test@example.com",
	}
	
	// Mock user exists
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	
	// Mock successful todo creation
	mockTodoRepo.On("Create", ctx, mock.AnythingOfType("*model.Todo")).Return(nil).Run(func(args mock.Arguments) {
		todo := args.Get(1).(*model.Todo)
		todo.ID = 1 // Simulate database setting ID
		todo.CreatedAt = time.Now()
		todo.UpdatedAt = time.Now()
	})
	
	// Call service
	todo, err := todoService.Create(ctx, req, userID)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, todo)
	assert.Equal(t, req.Title, todo.Title)
	assert.Equal(t, req.Description, todo.Description)
	assert.Equal(t, userID, todo.UserID)
	assert.False(t, todo.Completed) // Should default to false
	assert.Equal(t, uint(1), todo.ID)
	
	mockUserRepo.AssertExpectations(t)
	mockTodoRepo.AssertExpectations(t)
}

func TestTodoService_Create_UserNotFound(t *testing.T) {
	todoService, _, mockUserRepo := setupTodoService()
	ctx := context.Background()
	
	req := &model.CreateTodoRequest{
		Title:       "Test Todo",
		Description: "Test Description",
	}
	userID := uint(1)
	
	// Mock user doesn't exist
	mockUserRepo.On("GetByID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)
	
	// Call service
	todo, err := todoService.Create(ctx, req, userID)
	
	// Assertions
	assert.Error(t, err)
	assert.Nil(t, todo)
	assert.Equal(t, service.ErrUserNotFound, err)
	
	mockUserRepo.AssertExpectations(t)
}

func TestTodoService_Create_DatabaseError(t *testing.T) {
	todoService, mockTodoRepo, mockUserRepo := setupTodoService()
	ctx := context.Background()
	
	req := &model.CreateTodoRequest{
		Title:       "Test Todo",
		Description: "Test Description",
	}
	userID := uint(1)
	
	user := &model.User{
		ID:    userID,
		Email: "test@example.com",
	}
	
	// Mock user exists
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	
	// Mock database error during creation
	mockTodoRepo.On("Create", ctx, mock.AnythingOfType("*model.Todo")).Return(errors.New("database error"))
	
	// Call service
	todo, err := todoService.Create(ctx, req, userID)
	
	// Assertions
	assert.Error(t, err)
	assert.Nil(t, todo)
	assert.Contains(t, err.Error(), "failed to create todo")
	
	mockUserRepo.AssertExpectations(t)
	mockTodoRepo.AssertExpectations(t)
}

func TestTodoService_GetByID_Success(t *testing.T) {
	todoService, mockTodoRepo, _ := setupTodoService()
	ctx := context.Background()
	
	todoID := uint(1)
	userID := uint(1)
	
	expectedTodo := &model.Todo{
		ID:          todoID,
		Title:       "Test Todo",
		Description: "Test Description",
		Completed:   false,
		UserID:      userID,
	}
	
	// Mock todo exists and belongs to user
	mockTodoRepo.On("GetByID", ctx, todoID, userID).Return(expectedTodo, nil)
	
	// Call service
	todo, err := todoService.GetByID(ctx, todoID, userID)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, todo)
	assert.Equal(t, expectedTodo.ID, todo.ID)
	assert.Equal(t, expectedTodo.Title, todo.Title)
	assert.Equal(t, expectedTodo.UserID, todo.UserID)
	
	mockTodoRepo.AssertExpectations(t)
}

func TestTodoService_GetByID_NotFound(t *testing.T) {
	todoService, mockTodoRepo, _ := setupTodoService()
	ctx := context.Background()
	
	todoID := uint(1)
	userID := uint(1)
	
	// Mock todo doesn't exist
	mockTodoRepo.On("GetByID", ctx, todoID, userID).Return(nil, gorm.ErrRecordNotFound)
	
	// Call service
	todo, err := todoService.GetByID(ctx, todoID, userID)
	
	// Assertions
	assert.Error(t, err)
	assert.Nil(t, todo)
	assert.Equal(t, service.ErrTodoNotFound, err)
	
	mockTodoRepo.AssertExpectations(t)
}

func TestTodoService_GetByID_UnauthorizedAccess(t *testing.T) {
	todoService, mockTodoRepo, _ := setupTodoService()
	ctx := context.Background()
	
	todoID := uint(1)
	userID := uint(1)
	otherUserID := uint(2)
	
	// Todo belongs to different user
	todoFromOtherUser := &model.Todo{
		ID:          todoID,
		Title:       "Test Todo",
		Description: "Test Description",
		Completed:   false,
		UserID:      otherUserID, // Different user
	}
	
	// Mock todo exists but belongs to different user
	mockTodoRepo.On("GetByID", ctx, todoID, userID).Return(todoFromOtherUser, nil)
	
	// Call service
	todo, err := todoService.GetByID(ctx, todoID, userID)
	
	// Assertions
	assert.Error(t, err)
	assert.Nil(t, todo)
	assert.Equal(t, service.ErrUnauthorizedAccess, err)
	
	mockTodoRepo.AssertExpectations(t)
}

func TestTodoService_GetByUserID_Success(t *testing.T) {
	todoService, mockTodoRepo, mockUserRepo := setupTodoService()
	ctx := context.Background()
	
	userID := uint(1)
	
	user := &model.User{
		ID:    userID,
		Email: "test@example.com",
	}
	
	expectedTodos := []*model.Todo{
		{
			ID:          1,
			Title:       "Todo 1",
			Description: "Description 1",
			Completed:   false,
			UserID:      userID,
		},
		{
			ID:          2,
			Title:       "Todo 2",
			Description: "Description 2",
			Completed:   true,
			UserID:      userID,
		},
	}
	
	// Mock user exists
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	
	// Mock todos exist
	mockTodoRepo.On("GetByUserID", ctx, userID).Return(expectedTodos, nil)
	
	// Call service
	todos, err := todoService.GetByUserID(ctx, userID)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, todos)
	assert.Len(t, todos, 2)
	assert.Equal(t, expectedTodos[0].Title, todos[0].Title)
	assert.Equal(t, expectedTodos[1].Title, todos[1].Title)
	
	mockUserRepo.AssertExpectations(t)
	mockTodoRepo.AssertExpectations(t)
}

func TestTodoService_GetByUserID_EmptyResult(t *testing.T) {
	todoService, mockTodoRepo, mockUserRepo := setupTodoService()
	ctx := context.Background()
	
	userID := uint(1)
	
	user := &model.User{
		ID:    userID,
		Email: "test@example.com",
	}
	
	// Mock user exists
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	
	// Mock no todos found (nil result)
	mockTodoRepo.On("GetByUserID", ctx, userID).Return(nil, nil)
	
	// Call service
	todos, err := todoService.GetByUserID(ctx, userID)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, todos)
	assert.Len(t, todos, 0) // Should return empty slice, not nil
	
	mockUserRepo.AssertExpectations(t)
	mockTodoRepo.AssertExpectations(t)
}

func TestTodoService_GetByUserID_UserNotFound(t *testing.T) {
	todoService, _, mockUserRepo := setupTodoService()
	ctx := context.Background()
	
	userID := uint(1)
	
	// Mock user doesn't exist
	mockUserRepo.On("GetByID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)
	
	// Call service
	todos, err := todoService.GetByUserID(ctx, userID)
	
	// Assertions
	assert.Error(t, err)
	assert.Nil(t, todos)
	assert.Equal(t, service.ErrUserNotFound, err)
	
	mockUserRepo.AssertExpectations(t)
}

func TestTodoService_Update_Success(t *testing.T) {
	todoService, mockTodoRepo, _ := setupTodoService()
	ctx := context.Background()
	
	todoID := uint(1)
	userID := uint(1)
	
	newTitle := "Updated Title"
	newCompleted := true
	req := &model.UpdateTodoRequest{
		Title:     &newTitle,
		Completed: &newCompleted,
	}
	
	existingTodo := &model.Todo{
		ID:          todoID,
		Title:       "Original Title",
		Description: "Original Description",
		Completed:   false,
		UserID:      userID,
	}
	
	// Mock existing todo
	mockTodoRepo.On("GetByID", ctx, todoID, userID).Return(existingTodo, nil)
	
	// Mock successful update
	mockTodoRepo.On("Update", ctx, mock.AnythingOfType("*model.Todo")).Return(nil)
	
	// Call service
	updatedTodo, err := todoService.Update(ctx, todoID, req, userID)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, updatedTodo)
	assert.Equal(t, newTitle, updatedTodo.Title)
	assert.Equal(t, "Original Description", updatedTodo.Description) // Should remain unchanged
	assert.Equal(t, newCompleted, updatedTodo.Completed)
	assert.Equal(t, userID, updatedTodo.UserID)
	
	mockTodoRepo.AssertExpectations(t)
}

func TestTodoService_Update_PartialUpdate(t *testing.T) {
	todoService, mockTodoRepo, _ := setupTodoService()
	ctx := context.Background()
	
	todoID := uint(1)
	userID := uint(1)
	
	newDescription := "Updated Description"
	req := &model.UpdateTodoRequest{
		Description: &newDescription,
		// Title and Completed not provided
	}
	
	existingTodo := &model.Todo{
		ID:          todoID,
		Title:       "Original Title",
		Description: "Original Description",
		Completed:   false,
		UserID:      userID,
	}
	
	// Mock existing todo
	mockTodoRepo.On("GetByID", ctx, todoID, userID).Return(existingTodo, nil)
	
	// Mock successful update
	mockTodoRepo.On("Update", ctx, mock.AnythingOfType("*model.Todo")).Return(nil)
	
	// Call service
	updatedTodo, err := todoService.Update(ctx, todoID, req, userID)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, updatedTodo)
	assert.Equal(t, "Original Title", updatedTodo.Title) // Should remain unchanged
	assert.Equal(t, newDescription, updatedTodo.Description)
	assert.False(t, updatedTodo.Completed) // Should remain unchanged
	
	mockTodoRepo.AssertExpectations(t)
}

func TestTodoService_Update_NotFound(t *testing.T) {
	todoService, mockTodoRepo, _ := setupTodoService()
	ctx := context.Background()
	
	todoID := uint(1)
	userID := uint(1)
	
	newTitle := "Updated Title"
	req := &model.UpdateTodoRequest{
		Title: &newTitle,
	}
	
	// Mock todo doesn't exist
	mockTodoRepo.On("GetByID", ctx, todoID, userID).Return(nil, gorm.ErrRecordNotFound)
	
	// Call service
	updatedTodo, err := todoService.Update(ctx, todoID, req, userID)
	
	// Assertions
	assert.Error(t, err)
	assert.Nil(t, updatedTodo)
	assert.Equal(t, service.ErrTodoNotFound, err)
	
	mockTodoRepo.AssertExpectations(t)
}

func TestTodoService_Update_UnauthorizedAccess(t *testing.T) {
	todoService, mockTodoRepo, _ := setupTodoService()
	ctx := context.Background()
	
	todoID := uint(1)
	userID := uint(1)
	otherUserID := uint(2)
	
	newTitle := "Updated Title"
	req := &model.UpdateTodoRequest{
		Title: &newTitle,
	}
	
	// Todo belongs to different user
	todoFromOtherUser := &model.Todo{
		ID:          todoID,
		Title:       "Original Title",
		Description: "Original Description",
		Completed:   false,
		UserID:      otherUserID, // Different user
	}
	
	// Mock todo exists but belongs to different user
	mockTodoRepo.On("GetByID", ctx, todoID, userID).Return(todoFromOtherUser, nil)
	
	// Call service
	updatedTodo, err := todoService.Update(ctx, todoID, req, userID)
	
	// Assertions
	assert.Error(t, err)
	assert.Nil(t, updatedTodo)
	assert.Equal(t, service.ErrUnauthorizedAccess, err)
	
	mockTodoRepo.AssertExpectations(t)
}

func TestTodoService_Delete_Success(t *testing.T) {
	todoService, mockTodoRepo, _ := setupTodoService()
	ctx := context.Background()
	
	todoID := uint(1)
	userID := uint(1)
	
	existingTodo := &model.Todo{
		ID:          todoID,
		Title:       "Test Todo",
		Description: "Test Description",
		Completed:   false,
		UserID:      userID,
	}
	
	// Mock todo exists and belongs to user
	mockTodoRepo.On("GetByID", ctx, todoID, userID).Return(existingTodo, nil)
	
	// Mock successful deletion
	mockTodoRepo.On("Delete", ctx, todoID, userID).Return(nil)
	
	// Call service
	err := todoService.Delete(ctx, todoID, userID)
	
	// Assertions
	assert.NoError(t, err)
	
	mockTodoRepo.AssertExpectations(t)
}

func TestTodoService_Delete_NotFound(t *testing.T) {
	todoService, mockTodoRepo, _ := setupTodoService()
	ctx := context.Background()
	
	todoID := uint(1)
	userID := uint(1)
	
	// Mock todo doesn't exist
	mockTodoRepo.On("GetByID", ctx, todoID, userID).Return(nil, gorm.ErrRecordNotFound)
	
	// Call service
	err := todoService.Delete(ctx, todoID, userID)
	
	// Assertions
	assert.Error(t, err)
	assert.Equal(t, service.ErrTodoNotFound, err)
	
	mockTodoRepo.AssertExpectations(t)
}

func TestTodoService_Delete_DatabaseError(t *testing.T) {
	todoService, mockTodoRepo, _ := setupTodoService()
	ctx := context.Background()
	
	todoID := uint(1)
	userID := uint(1)
	
	existingTodo := &model.Todo{
		ID:          todoID,
		Title:       "Test Todo",
		Description: "Test Description",
		Completed:   false,
		UserID:      userID,
	}
	
	// Mock todo exists and belongs to user
	mockTodoRepo.On("GetByID", ctx, todoID, userID).Return(existingTodo, nil)
	
	// Mock database error during deletion
	mockTodoRepo.On("Delete", ctx, todoID, userID).Return(errors.New("database error"))
	
	// Call service
	err := todoService.Delete(ctx, todoID, userID)
	
	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete todo")
	
	mockTodoRepo.AssertExpectations(t)
}