package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	Port        string `env:"PORT"`
	Environment string `env:"ENVIRONMENT"`
	LogLevel    string `env:"LOG_LEVEL"`

	// Database configuration
	DatabaseURL string `env:"DATABASE_URL"`

	// JWT configuration
	JWTSecret     string `env:"JWT_SECRET"`
	JWTExpiration int    `env:"JWT_EXPIRATION"`

	// CORS configuration
	AllowedOrigins []string `env:"ALLOWED_ORIGINS"`
}

// Load loads configuration from environment variables with defaults
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if file doesn't exist)
	_ = godotenv.Load()

	config := &Config{
		// Set defaults
		Port:           getEnvWithDefault("PORT", "8080"),
		Environment:    getEnvWithDefault("ENVIRONMENT", "development"),
		LogLevel:       getEnvWithDefault("LOG_LEVEL", "info"),
		DatabaseURL:    getEnvWithDefault("DATABASE_URL", "postgres://user:password@localhost/todoapi?sslmode=disable"),
		JWTSecret:      os.Getenv("JWT_SECRET"), // No default for JWT_SECRET - must be explicitly set
		JWTExpiration:  getEnvIntWithDefault("JWT_EXPIRATION", 24),
		AllowedOrigins: getEnvSliceWithDefault("ALLOWED_ORIGINS", []string{"*"}),
	}

	// Validate required configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate ensures all required configuration is present and valid
func (c *Config) Validate() error {
	var errors []string

	// Validate required fields
	if c.DatabaseURL == "" {
		errors = append(errors, "DATABASE_URL is required")
	}

	if c.JWTSecret == "" {
		errors = append(errors, "JWT_SECRET is required")
	}

	// Validate JWT secret strength in production
	if c.Environment == "production" && len(c.JWTSecret) < 32 {
		errors = append(errors, "JWT_SECRET must be at least 32 characters in production")
	}

	// Validate JWT expiration
	if c.JWTExpiration <= 0 {
		errors = append(errors, "JWT_EXPIRATION must be greater than 0")
	}

	// Validate port
	if c.Port == "" {
		errors = append(errors, "PORT is required")
	}

	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLogLevels, strings.ToLower(c.LogLevel)) {
		errors = append(errors, "LOG_LEVEL must be one of: debug, info, warn, error")
	}

	// Validate environment
	validEnvironments := []string{"development", "staging", "production"}
	if !contains(validEnvironments, strings.ToLower(c.Environment)) {
		errors = append(errors, "ENVIRONMENT must be one of: development, staging, production")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.Environment) == "development"
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return strings.ToLower(c.Environment) == "production"
}

// getEnvWithDefault gets an environment variable with a default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntWithDefault gets an environment variable as int with a default value
func getEnvIntWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvSliceWithDefault gets an environment variable as slice with a default value
func getEnvSliceWithDefault(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}