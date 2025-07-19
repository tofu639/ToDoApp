package config_test

import (
	"fmt"
	"log"
	"os"

	"todo-api-backend/internal/config"
)

func ExampleLoad() {
	// Set some environment variables for the example
	os.Setenv("JWT_SECRET", "example-secret-key-for-testing")
	os.Setenv("JWT_EXPIRATION", "24")
	os.Setenv("PORT", "3000")
	os.Setenv("ENVIRONMENT", "development")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Server will run on port: %s\n", cfg.Port)
	fmt.Printf("Environment: %s\n", cfg.Environment)
	fmt.Printf("Is Development: %t\n", cfg.IsDevelopment())
	fmt.Printf("JWT Expiration: %d hours\n", cfg.JWTExpiration)

	// Clean up
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("JWT_EXPIRATION")
	os.Unsetenv("PORT")
	os.Unsetenv("ENVIRONMENT")

	// Output:
	// Server will run on port: 3000
	// Environment: development
	// Is Development: true
	// JWT Expiration: 24 hours
}