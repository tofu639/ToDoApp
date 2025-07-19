package model

// RegisterRequest represents the request payload for user registration
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=8" example:"password123"`
}

// LoginRequest represents the request payload for user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required" example:"password123"`
}

// CreateTodoRequest represents the request payload for creating a todo
type CreateTodoRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=255" example:"Complete project"`
	Description string `json:"description" validate:"max=1000" example:"Finish the todo API backend project"`
}

// UpdateTodoRequest represents the request payload for updating a todo
type UpdateTodoRequest struct {
	Title       *string `json:"title,omitempty" validate:"omitempty,min=1,max=255" example:"Updated task title"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000" example:"Updated description"`
	Completed   *bool   `json:"completed,omitempty" example:"true"`
}