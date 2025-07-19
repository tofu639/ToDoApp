package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	
	"todo-api-backend/internal/middleware"
	"todo-api-backend/internal/model"
)

// CreateTodo handles todo creation
// @Summary Create a new todo
// @Description Create a new todo item for the authenticated user
// @Tags todos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.CreateTodoRequest true "Todo creation request"
// @Success 201 {object} model.Todo "Todo successfully created"
// @Failure 400 {object} model.ErrorResponse "Invalid request data or validation failed"
// @Failure 401 {object} model.ErrorResponse "User not authenticated"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/v1/todos [post]
func (h *Handler) CreateTodo(c *gin.Context) {
	var req model.CreateTodoRequest
	
	// Get user ID from context (set by JWT middleware)
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}
	
	// Bind JSON request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid JSON format",
		})
		return
	}
	
	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		details := make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Tag() {
			case "required":
				details[err.Field()] = "This field is required"
			case "min":
				details[err.Field()] = "Title must be at least 1 character long"
			case "max":
				if err.Field() == "Title" {
					details[err.Field()] = "Title must be at most 255 characters long"
				} else {
					details[err.Field()] = "Description must be at most 1000 characters long"
				}
			default:
				details[err.Field()] = "Invalid value"
			}
		}
		
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_failed",
			Message: "Invalid input data",
			Details: details,
		})
		return
	}
	
	// Call service to create todo
	todo, err := h.services.Todo.Create(c.Request.Context(), &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "creation_failed",
			Message: "Failed to create todo",
		})
		return
	}
	
	c.JSON(http.StatusCreated, todo)
}

// GetTodos handles retrieving all todos for the authenticated user
// @Summary Get all todos
// @Description Retrieve all todos belonging to the authenticated user
// @Tags todos
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.TodoListResponse "List of todos retrieved successfully"
// @Failure 401 {object} model.ErrorResponse "User not authenticated"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/v1/todos [get]
func (h *Handler) GetTodos(c *gin.Context) {
	// Get user ID from context (set by JWT middleware)
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}
	
	// Call service to get todos
	todos, err := h.services.Todo.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "retrieval_failed",
			Message: "Failed to retrieve todos",
		})
		return
	}
	
	// Return todos with count
	response := model.TodoListResponse{
		Todos: todos,
		Count: len(todos),
	}
	
	c.JSON(http.StatusOK, response)
}

// GetTodo handles retrieving a specific todo by ID
// @Summary Get todo by ID
// @Description Retrieve a specific todo by ID, ensuring user ownership
// @Tags todos
// @Produce json
// @Security BearerAuth
// @Param id path int true "Todo ID"
// @Success 200 {object} model.Todo "Todo retrieved successfully"
// @Failure 400 {object} model.ErrorResponse "Invalid todo ID format"
// @Failure 401 {object} model.ErrorResponse "User not authenticated"
// @Failure 404 {object} model.ErrorResponse "Todo not found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/v1/todos/{id} [get]
func (h *Handler) GetTodo(c *gin.Context) {
	// Get user ID from context (set by JWT middleware)
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}
	
	// Parse todo ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid todo ID format",
		})
		return
	}
	
	// Call service to get todo
	todo, err := h.services.Todo.GetByID(c.Request.Context(), uint(id), userID)
	if err != nil {
		switch err.Error() {
		case "todo not found":
			c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "Todo not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error:   "retrieval_failed",
				Message: "Failed to retrieve todo",
			})
		}
		return
	}
	
	c.JSON(http.StatusOK, todo)
}

// UpdateTodo handles updating a specific todo
// @Summary Update todo
// @Description Update a specific todo by ID, ensuring user ownership
// @Tags todos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Todo ID"
// @Param request body model.UpdateTodoRequest true "Todo update request"
// @Success 200 {object} model.Todo "Todo updated successfully"
// @Failure 400 {object} model.ErrorResponse "Invalid request data or validation failed"
// @Failure 401 {object} model.ErrorResponse "User not authenticated"
// @Failure 404 {object} model.ErrorResponse "Todo not found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/v1/todos/{id} [put]
func (h *Handler) UpdateTodo(c *gin.Context) {
	var req model.UpdateTodoRequest
	
	// Get user ID from context (set by JWT middleware)
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}
	
	// Parse todo ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid todo ID format",
		})
		return
	}
	
	// Bind JSON request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid JSON format",
		})
		return
	}
	
	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		details := make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Tag() {
			case "min":
				details[err.Field()] = "Title must be at least 1 character long"
			case "max":
				if err.Field() == "Title" {
					details[err.Field()] = "Title must be at most 255 characters long"
				} else {
					details[err.Field()] = "Description must be at most 1000 characters long"
				}
			default:
				details[err.Field()] = "Invalid value"
			}
		}
		
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_failed",
			Message: "Invalid input data",
			Details: details,
		})
		return
	}
	
	// Call service to update todo
	todo, err := h.services.Todo.Update(c.Request.Context(), uint(id), &req, userID)
	if err != nil {
		switch err.Error() {
		case "todo not found":
			c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "Todo not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error:   "update_failed",
				Message: "Failed to update todo",
			})
		}
		return
	}
	
	c.JSON(http.StatusOK, todo)
}

// DeleteTodo handles deleting a specific todo
// @Summary Delete todo
// @Description Delete a specific todo by ID, ensuring user ownership
// @Tags todos
// @Security BearerAuth
// @Param id path int true "Todo ID"
// @Success 204 "Todo deleted successfully"
// @Failure 400 {object} model.ErrorResponse "Invalid todo ID format"
// @Failure 401 {object} model.ErrorResponse "User not authenticated"
// @Failure 404 {object} model.ErrorResponse "Todo not found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/v1/todos/{id} [delete]
func (h *Handler) DeleteTodo(c *gin.Context) {
	// Get user ID from context (set by JWT middleware)
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}
	
	// Parse todo ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid todo ID format",
		})
		return
	}
	
	// Call service to delete todo
	err = h.services.Todo.Delete(c.Request.Context(), uint(id), userID)
	if err != nil {
		switch err.Error() {
		case "todo not found":
			c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "Todo not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error:   "deletion_failed",
				Message: "Failed to delete todo",
			})
		}
		return
	}
	
	c.Status(http.StatusNoContent)
}