package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		expected    *Config
	}{
		{
			name: "default configuration",
			envVars: map[string]string{
				"JWT_SECRET": "test-secret-key-that-is-long-enough",
			},
			expectError: false,
			expected: &Config{
				Port:           "8080",
				Environment:    "development",
				LogLevel:       "info",
				DatabaseURL:    "postgres://user:password@localhost/todoapi?sslmode=disable",
				JWTSecret:      "test-secret-key-that-is-long-enough",
				JWTExpiration:  24,
				AllowedOrigins: []string{"*"},
			},
		},
		{
			name: "custom configuration",
			envVars: map[string]string{
				"PORT":            "3000",
				"ENVIRONMENT":     "production",
				"LOG_LEVEL":       "error",
				"DATABASE_URL":    "postgres://prod:pass@db:5432/todoapi",
				"JWT_SECRET":      "super-secret-production-key-that-is-very-long",
				"JWT_EXPIRATION":  "48",
				"ALLOWED_ORIGINS": "https://example.com,https://app.example.com",
			},
			expectError: false,
			expected: &Config{
				Port:           "3000",
				Environment:    "production",
				LogLevel:       "error",
				DatabaseURL:    "postgres://prod:pass@db:5432/todoapi",
				JWTSecret:      "super-secret-production-key-that-is-very-long",
				JWTExpiration:  48,
				AllowedOrigins: []string{"https://example.com", "https://app.example.com"},
			},
		},
		{
			name: "missing JWT secret",
			envVars: map[string]string{
				"JWT_SECRET": "",
			},
			expectError: true,
		},
		{
			name: "invalid JWT expiration",
			envVars: map[string]string{
				"JWT_SECRET":     "test-secret-key",
				"JWT_EXPIRATION": "0",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			clearEnv()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Load configuration
			config, err := Load()

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected.Port, config.Port)
			assert.Equal(t, tt.expected.Environment, config.Environment)
			assert.Equal(t, tt.expected.LogLevel, config.LogLevel)
			assert.Equal(t, tt.expected.DatabaseURL, config.DatabaseURL)
			assert.Equal(t, tt.expected.JWTSecret, config.JWTSecret)
			assert.Equal(t, tt.expected.JWTExpiration, config.JWTExpiration)
			assert.Equal(t, tt.expected.AllowedOrigins, config.AllowedOrigins)

			// Clean up
			clearEnv()
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid development config",
			config: &Config{
				Port:          "8080",
				Environment:   "development",
				LogLevel:      "info",
				DatabaseURL:   "postgres://localhost/test",
				JWTSecret:     "test-secret",
				JWTExpiration: 24,
			},
			expectError: false,
		},
		{
			name: "valid production config",
			config: &Config{
				Port:          "8080",
				Environment:   "production",
				LogLevel:      "error",
				DatabaseURL:   "postgres://localhost/test",
				JWTSecret:     "super-long-production-secret-key-that-meets-requirements",
				JWTExpiration: 24,
			},
			expectError: false,
		},
		{
			name: "missing database URL",
			config: &Config{
				Port:          "8080",
				Environment:   "development",
				LogLevel:      "info",
				DatabaseURL:   "",
				JWTSecret:     "test-secret",
				JWTExpiration: 24,
			},
			expectError: true,
			errorMsg:    "DATABASE_URL is required",
		},
		{
			name: "missing JWT secret",
			config: &Config{
				Port:          "8080",
				Environment:   "development",
				LogLevel:      "info",
				DatabaseURL:   "postgres://localhost/test",
				JWTSecret:     "",
				JWTExpiration: 24,
			},
			expectError: true,
			errorMsg:    "JWT_SECRET is required",
		},
		{
			name: "short JWT secret in production",
			config: &Config{
				Port:          "8080",
				Environment:   "production",
				LogLevel:      "info",
				DatabaseURL:   "postgres://localhost/test",
				JWTSecret:     "short",
				JWTExpiration: 24,
			},
			expectError: true,
			errorMsg:    "JWT_SECRET must be at least 32 characters in production",
		},
		{
			name: "invalid JWT expiration",
			config: &Config{
				Port:          "8080",
				Environment:   "development",
				LogLevel:      "info",
				DatabaseURL:   "postgres://localhost/test",
				JWTSecret:     "test-secret",
				JWTExpiration: 0,
			},
			expectError: true,
			errorMsg:    "JWT_EXPIRATION must be greater than 0",
		},
		{
			name: "invalid log level",
			config: &Config{
				Port:          "8080",
				Environment:   "development",
				LogLevel:      "invalid",
				DatabaseURL:   "postgres://localhost/test",
				JWTSecret:     "test-secret",
				JWTExpiration: 24,
			},
			expectError: true,
			errorMsg:    "LOG_LEVEL must be one of: debug, info, warn, error",
		},
		{
			name: "invalid environment",
			config: &Config{
				Port:          "8080",
				Environment:   "invalid",
				LogLevel:      "info",
				DatabaseURL:   "postgres://localhost/test",
				JWTSecret:     "test-secret",
				JWTExpiration: 24,
			},
			expectError: true,
			errorMsg:    "ENVIRONMENT must be one of: development, staging, production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_IsDevelopment(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		expected    bool
	}{
		{"development", "development", true},
		{"Development", "Development", true},
		{"DEVELOPMENT", "DEVELOPMENT", true},
		{"production", "production", false},
		{"staging", "staging", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{Environment: tt.environment}
			assert.Equal(t, tt.expected, config.IsDevelopment())
		})
	}
}

func TestConfig_IsProduction(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		expected    bool
	}{
		{"production", "production", true},
		{"Production", "Production", true},
		{"PRODUCTION", "PRODUCTION", true},
		{"development", "development", false},
		{"staging", "staging", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{Environment: tt.environment}
			assert.Equal(t, tt.expected, config.IsProduction())
		})
	}
}

// clearEnv clears relevant environment variables for testing
func clearEnv() {
	envVars := []string{
		"PORT", "ENVIRONMENT", "LOG_LEVEL", "DATABASE_URL",
		"JWT_SECRET", "JWT_EXPIRATION", "ALLOWED_ORIGINS",
	}
	for _, env := range envVars {
		os.Unsetenv(env)
	}
}