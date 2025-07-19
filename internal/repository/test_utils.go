package repository

import (
	"time"

	"todo-api-backend/internal/model"
)

// createTestUser creates a test user with mock data
func createTestUser(email string) *model.User {
	return &model.User{
		ID:        1,
		Email:     email,
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// createTestTodo creates a test todo with mock data
func createTestTodo(id uint, userID uint, title string) *model.Todo {
	return &model.Todo{
		ID:          id,
		Title:       title,
		Description: "Test description",
		UserID:      userID,
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}