package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	
	"todo-api-backend/internal/model"
)

// Register handles user registration
// @Summary Register a new user
// @Description Create a new user account with email and password
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body model.RegisterRequest true "Registration request"
// @Success 201 {object} model.AuthResponse "User successfully registered"
// @Failure 400 {object} model.ErrorResponse "Invalid request data or validation failed"
// @Failure 409 {object} model.ErrorResponse "Email already exists"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/v1/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req model.RegisterRequest
	
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
			case "email":
				details[err.Field()] = "Invalid email format"
			case "min":
				details[err.Field()] = "Password must be at least 8 characters long"
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
	
	// Call service to register user
	response, err := h.services.Auth.Register(c.Request.Context(), &req)
	if err != nil {
		// Handle different types of errors
		switch err.Error() {
		case "email already exists":
			c.JSON(http.StatusConflict, model.ErrorResponse{
				Error:   "email_exists",
				Message: "An account with this email already exists",
			})
		default:
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error:   "registration_failed",
				Message: "Failed to create user account",
			})
		}
		return
	}
	
	c.JSON(http.StatusCreated, response)
}

// Login handles user authentication
// @Summary Login user
// @Description Authenticate user with email and password, returns JWT token
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "Login request"
// @Success 200 {object} model.AuthResponse "User successfully authenticated"
// @Failure 400 {object} model.ErrorResponse "Invalid request data or validation failed"
// @Failure 401 {object} model.ErrorResponse "Invalid credentials"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req model.LoginRequest
	
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
			case "email":
				details[err.Field()] = "Invalid email format"
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
	
	// Call service to authenticate user
	response, err := h.services.Auth.Login(c.Request.Context(), &req)
	if err != nil {
		// Handle different types of errors
		switch err.Error() {
		case "invalid credentials", "user not found":
			c.JSON(http.StatusUnauthorized, model.ErrorResponse{
				Error:   "invalid_credentials",
				Message: "Invalid email or password",
			})
		default:
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error:   "login_failed",
				Message: "Failed to authenticate user",
			})
		}
		return
	}
	
	c.JSON(http.StatusOK, response)
}