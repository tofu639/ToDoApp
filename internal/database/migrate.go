package database

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"
	"todo-api-backend/internal/model"
)

// Migration represents a database migration
type Migration struct {
	Version string
	Name    string
	SQL     string
}

// AutoMigrate runs GORM auto-migration for all models
func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&model.User{},
		&model.Todo{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto-migrate database: %w", err)
	}
	return nil
}

// RunSQLMigrations executes SQL migration files from the migrations directory
func RunSQLMigrations(db *gorm.DB, migrationsPath string) error {
	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of migration files
	migrations, err := loadMigrations(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Execute migrations in order
	for _, migration := range migrations {
		if err := executeMigration(db, migration); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", migration.Version, err)
		}
	}

	return nil
}

// createMigrationsTable creates the migrations tracking table
func createMigrationsTable(db *gorm.DB) error {
	sql := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`
	return db.Exec(sql).Error
}

// loadMigrations loads all migration files from the specified directory
func loadMigrations(migrationsPath string) ([]Migration, error) {
	files, err := ioutil.ReadDir(migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []Migration
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		// Extract version from filename (e.g., "001_init.sql" -> "001")
		parts := strings.Split(file.Name(), "_")
		if len(parts) < 2 {
			continue
		}
		version := parts[0]
		name := strings.TrimSuffix(strings.Join(parts[1:], "_"), ".sql")

		// Read migration file content
		filePath := filepath.Join(migrationsPath, file.Name())
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		migrations = append(migrations, Migration{
			Version: version,
			Name:    name,
			SQL:     string(content),
		})
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// executeMigration executes a single migration if it hasn't been applied yet
func executeMigration(db *gorm.DB, migration Migration) error {
	// Check if migration has already been applied
	var count int64
	err := db.Raw("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", migration.Version).Scan(&count).Error
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if count > 0 {
		// Migration already applied, skip
		return nil
	}

	// Execute migration in a transaction
	tx := db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Execute the migration SQL
	if err := tx.Exec(migration.SQL).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record the migration as applied
	if err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", migration.Version).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit migration transaction: %w", err)
	}

	fmt.Printf("Applied migration: %s_%s\n", migration.Version, migration.Name)
	return nil
}