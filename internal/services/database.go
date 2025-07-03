package services

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/cloudyali/terratag/internal/db"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

type DatabaseService struct {
	DB      *sql.DB
	Queries *db.Queries
}

func NewDatabaseService(dbPath string) (*DatabaseService, error) {
	log.Printf("[DATABASE] Initializing database service: path=%s", dbPath)
	
	// Create database connection
	database, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Printf("[DATABASE] Failed to open database: %v", err)
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	log.Printf("[DATABASE] Database connection opened successfully")

	// Test the connection
	if err := database.Ping(); err != nil {
		log.Printf("[DATABASE] Failed to ping database: %v", err)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	log.Printf("[DATABASE] Database ping successful")

	// Run migrations
	if err := runMigrations(database, dbPath); err != nil {
		log.Printf("[DATABASE] Failed to run migrations: %v", err)
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	log.Printf("[DATABASE] Database migrations completed")

	log.Printf("[DATABASE] Database service initialized successfully")
	return &DatabaseService{
		DB:      database,
		Queries: db.New(database),
	}, nil
}

func runMigrations(database *sql.DB, dbPath string) error {
	// Get the project root directory
	migrationsPath := "file://db/migrations"
	log.Printf("[DATABASE] Starting migrations: path=%s, dbPath=%s", migrationsPath, dbPath)
	
	// Validate migrations before applying
	validator := NewMigrationValidator(database, "db/migrations")
	validationResult, err := validator.ValidateMigrations()
	if err != nil {
		log.Printf("[DATABASE] Migration validation failed: %v", err)
		return fmt.Errorf("migration validation failed: %w", err)
	}
	
	// Log validation results
	if len(validationResult.Errors) > 0 {
		log.Printf("[DATABASE] Migration validation errors: %v", validationResult.Errors)
		return fmt.Errorf("migration validation failed with %d errors", len(validationResult.Errors))
	}
	
	if len(validationResult.Warnings) > 0 {
		log.Printf("[DATABASE] Migration validation warnings: %v", validationResult.Warnings)
	}
	
	log.Printf("[DATABASE] Migration validation passed: %d total migrations, %d applied, %d pending", 
		validationResult.TotalMigrations, validationResult.AppliedMigrations, validationResult.PendingMigrations)
	
	m, err := migrate.New(migrationsPath, "sqlite3://"+dbPath)
	if err != nil {
		log.Printf("[DATABASE] Failed to create migrate instance: %v", err)
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	log.Printf("[DATABASE] Migrate instance created successfully")

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Printf("[DATABASE] Failed to run migrations: %v", err)
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		log.Printf("[DATABASE] No new migrations to apply")
	} else {
		log.Printf("[DATABASE] Migrations applied successfully")
		
		// Re-validate after applying migrations
		postValidation, err := validator.ValidateMigrations()
		if err != nil {
			log.Printf("[DATABASE] Post-migration validation failed: %v", err)
		} else if len(postValidation.Errors) > 0 {
			log.Printf("[DATABASE] Post-migration validation errors: %v", postValidation.Errors)
		} else {
			log.Printf("[DATABASE] Post-migration validation passed successfully")
		}
	}
	return nil
}

func (s *DatabaseService) Close() error {
	return s.DB.Close()
}

func (s *DatabaseService) HealthCheck() error {
	return s.DB.Ping()
}

// ValidateMigrations performs comprehensive migration validation
func (s *DatabaseService) ValidateMigrations() (*MigrationValidationResult, error) {
	validator := NewMigrationValidator(s.DB, "db/migrations")
	return validator.ValidateMigrations()
}

// RepairMigrations attempts to repair common migration issues
func (s *DatabaseService) RepairMigrations() error {
	validator := NewMigrationValidator(s.DB, "db/migrations")
	return validator.RepairMigrations()
}