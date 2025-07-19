package service

import (
	"context"
	"errors"
	"fmt"

	"todo-api-backend/internal/model"
	"todo-api-backend/internal/repository"
	"todo-api-backend/pkg/jwt"
	"todo-api-backend/pkg/password"
	"gorm.io/gorm"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
)

// authService implements the AuthService interface
type authService struct {
	userRepo     repository.UserRepository
	tokenManager *jwt.TokenManager
	hasher       *password.Hasher
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo repository.UserRepository, tokenManager *jwt.TokenManager) AuthService {
	return &authService{
		userRepo:     userRepo,
		tokenManager: tokenManager,
		hasher:       password.NewHasher(),
	}
}

// Register creates a new user account with email validation and password hashing
func (s *authService) Register(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error) {
	// Check if user with email already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	
	if existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Validate password strength
	if err := password.ValidatePasswordStrength(req.Password); err != nil {
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

	// Hash the password
	hashedPassword, err := s.hasher.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create new user
	user := &model.User{
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token
	token, err := s.tokenManager.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Return auth response
	return &model.AuthResponse{
		Token: token,
		User:  user.ToUserInfo(),
	}, nil
}

// Login authenticates a user with credential verification and JWT generation
func (s *authService) Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Verify password
	if err := s.hasher.VerifyPassword(user.Password, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.tokenManager.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Return auth response
	return &model.AuthResponse{
		Token: token,
		User:  user.ToUserInfo(),
	}, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *authService) ValidateToken(tokenString string) (*jwt.Claims, error) {
	claims, err := s.tokenManager.ValidateToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	return claims, nil
}