package services

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// MigrationValidator handles database schema migration validation
type MigrationValidator struct {
	db             *sql.DB
	migrationsPath string
}

// MigrationInfo represents information about a migration file
type MigrationInfo struct {
	Version     int
	Name        string
	Direction   string // "up" or "down"
	FilePath    string
	Content     string
	HasDownFile bool
}

// MigrationValidationResult represents the result of migration validation
type MigrationValidationResult struct {
	IsValid              bool                    `json:"is_valid"`
	Errors               []string                `json:"errors,omitempty"`
	Warnings             []string                `json:"warnings,omitempty"`
	TotalMigrations      int                     `json:"total_migrations"`
	AppliedMigrations    int                     `json:"applied_migrations"`
	PendingMigrations    int                     `json:"pending_migrations"`
	MigrationFiles       []MigrationInfo         `json:"migration_files"`
	SchemaVersion        int                     `json:"schema_version"`
	ValidationDetails    map[string]interface{}  `json:"validation_details"`
}

// NewMigrationValidator creates a new migration validator
func NewMigrationValidator(db *sql.DB, migrationsPath string) *MigrationValidator {
	return &MigrationValidator{
		db:             db,
		migrationsPath: migrationsPath,
	}
}

// ValidateMigrations performs comprehensive migration validation
func (mv *MigrationValidator) ValidateMigrations() (*MigrationValidationResult, error) {
	result := &MigrationValidationResult{
		IsValid:           true,
		Errors:            []string{},
		Warnings:          []string{},
		ValidationDetails: make(map[string]interface{}),
	}

	logrus.WithField("component", "migration_validator").Info("Starting migration validation")

	// 1. Validate migration files structure
	migrationFiles, err := mv.collectMigrationFiles()
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to collect migration files: %v", err))
		result.IsValid = false
		return result, nil
	}
	result.MigrationFiles = migrationFiles
	result.TotalMigrations = len(migrationFiles)

	// 2. Validate migration file naming and numbering
	if err := mv.validateMigrationNaming(migrationFiles, result); err != nil {
		logrus.WithError(err).Error("Migration naming validation failed")
	}

	// 3. Validate migration content
	if err := mv.validateMigrationContent(migrationFiles, result); err != nil {
		logrus.WithError(err).Error("Migration content validation failed")
	}

	// 4. Check database schema version
	schemaVersion, err := mv.getCurrentSchemaVersion()
	isFreshDatabase := false
	if err != nil {
		// If schema_migrations table doesn't exist, it's a fresh database
		if strings.Contains(err.Error(), "no such table: schema_migrations") {
			isFreshDatabase = true
			result.SchemaVersion = 0
			logrus.Info("Fresh database detected, skipping schema consistency checks")
		} else {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Could not determine current schema version: %v", err))
		}
	} else {
		result.SchemaVersion = schemaVersion
		if schemaVersion == 0 {
			isFreshDatabase = true
			logrus.Info("Schema version is 0, treating as fresh database")
		}
	}

	// 5. Validate schema consistency (skip for fresh databases)
	if !isFreshDatabase {
		if err := mv.validateSchemaConsistency(result); err != nil {
			logrus.WithError(err).Error("Schema consistency validation failed")
		}
	}

	// 6. Check for orphaned migration records (skip for fresh databases)
	if !isFreshDatabase {
		if err := mv.checkOrphanedMigrations(result); err != nil {
			logrus.WithError(err).Error("Orphaned migrations check failed")
		}
	}

	// 7. Validate foreign key constraints (skip for fresh databases)
	if !isFreshDatabase {
		if err := mv.validateForeignKeyConstraints(result); err != nil {
			logrus.WithError(err).Error("Foreign key validation failed")
		}
	}

	// 8. Set final counts
	result.AppliedMigrations = result.SchemaVersion
	result.PendingMigrations = result.TotalMigrations - result.AppliedMigrations

	if len(result.Errors) > 0 {
		result.IsValid = false
	}

	logrus.WithFields(logrus.Fields{
		"component":           "migration_validator",
		"is_valid":           result.IsValid,
		"total_migrations":   result.TotalMigrations,
		"applied_migrations": result.AppliedMigrations,
		"errors":             len(result.Errors),
		"warnings":           len(result.Warnings),
	}).Info("Migration validation completed")

	return result, nil
}

// collectMigrationFiles finds and parses all migration files
func (mv *MigrationValidator) collectMigrationFiles() ([]MigrationInfo, error) {
	var migrations []MigrationInfo
	migrationFileRegex := regexp.MustCompile(`^(\d+)_(.+)\.(up|down)\.sql$`)

	err := filepath.WalkDir(mv.migrationsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		matches := migrationFileRegex.FindStringSubmatch(d.Name())
		if len(matches) != 4 {
			return nil // Not a migration file
		}

		version, err := strconv.Atoi(matches[1])
		if err != nil {
			return fmt.Errorf("invalid version number in file %s: %w", d.Name(), err)
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", path, err)
		}

		migration := MigrationInfo{
			Version:   version,
			Name:      matches[2],
			Direction: matches[3],
			FilePath:  path,
			Content:   string(content),
		}

		migrations = append(migrations, migration)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		if migrations[i].Version == migrations[j].Version {
			return migrations[i].Direction == "up" // up files first
		}
		return migrations[i].Version < migrations[j].Version
	})

	// Mark which migrations have down files
	downFiles := make(map[int]bool)
	for _, migration := range migrations {
		if migration.Direction == "down" {
			downFiles[migration.Version] = true
		}
	}
	for i := range migrations {
		if migrations[i].Direction == "up" {
			migrations[i].HasDownFile = downFiles[migrations[i].Version]
		}
	}

	return migrations, nil
}

// validateMigrationNaming checks migration file naming conventions
func (mv *MigrationValidator) validateMigrationNaming(migrations []MigrationInfo, result *MigrationValidationResult) error {
	versionMap := make(map[int][]MigrationInfo)
	
	for _, migration := range migrations {
		versionMap[migration.Version] = append(versionMap[migration.Version], migration)
	}

	// Check for sequential numbering
	versions := make([]int, 0, len(versionMap))
	for version := range versionMap {
		versions = append(versions, version)
	}
	sort.Ints(versions)

	for i, version := range versions {
		expectedVersion := i + 1
		if version != expectedVersion {
			result.Warnings = append(result.Warnings, 
				fmt.Sprintf("Migration version %d is not sequential (expected %d)", version, expectedVersion))
		}

		// Check that each version has both up and down files
		upExists := false
		downExists := false
		for _, migration := range versionMap[version] {
			if migration.Direction == "up" {
				upExists = true
			} else if migration.Direction == "down" {
				downExists = true
			}
		}

		if !upExists {
			result.Errors = append(result.Errors, 
				fmt.Sprintf("Missing up migration file for version %d", version))
		}
		if !downExists {
			result.Warnings = append(result.Warnings, 
				fmt.Sprintf("Missing down migration file for version %d (rollback not possible)", version))
		}
	}

	return nil
}

// validateMigrationContent checks migration SQL content
func (mv *MigrationValidator) validateMigrationContent(migrations []MigrationInfo, result *MigrationValidationResult) error {
	for _, migration := range migrations {
		if strings.TrimSpace(migration.Content) == "" {
			result.Warnings = append(result.Warnings, 
				fmt.Sprintf("Migration %d (%s) is empty", migration.Version, migration.Direction))
			continue
		}

		// Check for dangerous operations
		dangerousPatterns := []string{
			`DROP\s+TABLE`,
			`DROP\s+COLUMN`,
			`TRUNCATE`,
			`DELETE\s+FROM.*WHERE.*=.*`,
		}

		for _, pattern := range dangerousPatterns {
			if matched, _ := regexp.MatchString(`(?i)`+pattern, migration.Content); matched {
				result.Warnings = append(result.Warnings, 
					fmt.Sprintf("Migration %d (%s) contains potentially dangerous operation: %s", 
						migration.Version, migration.Direction, pattern))
			}
		}

		// Check for missing transactions in complex migrations
		if migration.Direction == "up" && strings.Count(migration.Content, ";") > 3 {
			if !strings.Contains(strings.ToUpper(migration.Content), "BEGIN") &&
			   !strings.Contains(strings.ToUpper(migration.Content), "TRANSACTION") {
				result.Warnings = append(result.Warnings, 
					fmt.Sprintf("Migration %d contains multiple statements but no explicit transaction", migration.Version))
			}
		}
	}

	return nil
}

// getCurrentSchemaVersion gets the current schema version from the database
func (mv *MigrationValidator) getCurrentSchemaVersion() (int, error) {
	// Check if schema_migrations table exists
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name='schema_migrations'`
	var tableName string
	err := mv.db.QueryRow(query).Scan(&tableName)
	if err == sql.ErrNoRows {
		// No migrations table means version 0
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	// Get the highest version number
	query = `SELECT COALESCE(MAX(version), 0) FROM schema_migrations WHERE dirty = 0`
	var version int
	err = mv.db.QueryRow(query).Scan(&version)
	if err != nil {
		return 0, err
	}

	return version, nil
}

// validateSchemaConsistency checks database schema consistency
func (mv *MigrationValidator) validateSchemaConsistency(result *MigrationValidationResult) error {
	// Check for table existence
	expectedTables := []string{"tag_standards", "operations", "operation_results", "operation_logs"}
	
	for _, table := range expectedTables {
		query := `SELECT name FROM sqlite_master WHERE type='table' AND name=?`
		var tableName string
		err := mv.db.QueryRow(query, table).Scan(&tableName)
		if err == sql.ErrNoRows {
			result.Errors = append(result.Errors, fmt.Sprintf("Expected table '%s' does not exist", table))
		} else if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Error checking table '%s': %v", table, err))
		}
	}

	// Check for required indexes
	expectedIndexes := []string{
		"idx_tag_standards_name",
		"idx_operations_type",
		"idx_operation_results_operation_id",
		"idx_operation_logs_operation_id",
	}

	for _, index := range expectedIndexes {
		query := `SELECT name FROM sqlite_master WHERE type='index' AND name=?`
		var indexName string
		err := mv.db.QueryRow(query, index).Scan(&indexName)
		if err == sql.ErrNoRows {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Expected index '%s' does not exist", index))
		} else if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Error checking index '%s': %v", index, err))
		}
	}

	return nil
}

// checkOrphanedMigrations looks for migration records without corresponding files
func (mv *MigrationValidator) checkOrphanedMigrations(result *MigrationValidationResult) error {
	// Get applied migrations from database
	query := `SELECT version FROM schema_migrations WHERE dirty = 0 ORDER BY version`
	rows, err := mv.db.Query(query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil // No migrations applied
		}
		return err
	}
	defer rows.Close()

	appliedVersions := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return err
		}
		appliedVersions[version] = true
	}

	// Check if all applied migrations have corresponding files
	fileVersions := make(map[int]bool)
	for _, migration := range result.MigrationFiles {
		if migration.Direction == "up" {
			fileVersions[migration.Version] = true
		}
	}

	for version := range appliedVersions {
		if !fileVersions[version] {
			result.Warnings = append(result.Warnings, 
				fmt.Sprintf("Migration version %d is applied in database but migration file is missing", version))
		}
	}

	return nil
}

// validateForeignKeyConstraints checks foreign key constraint integrity
func (mv *MigrationValidator) validateForeignKeyConstraints(result *MigrationValidationResult) error {
	// Enable foreign key constraint checking
	if _, err := mv.db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Could not enable foreign key checks: %v", err))
		return nil
	}

	// Check foreign key constraint violations
	query := `PRAGMA foreign_key_check`
	rows, err := mv.db.Query(query)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Could not check foreign key constraints: %v", err))
		return nil
	}
	defer rows.Close()

	violationCount := 0
	for rows.Next() {
		violationCount++
		// Could extract specific violation details here if needed
	}

	if violationCount > 0 {
		result.Errors = append(result.Errors, 
			fmt.Sprintf("Found %d foreign key constraint violations", violationCount))
	}

	result.ValidationDetails["foreign_key_violations"] = violationCount
	return nil
}

// RepairMigrations attempts to repair common migration issues
func (mv *MigrationValidator) RepairMigrations() error {
	logrus.WithField("component", "migration_validator").Info("Starting migration repair")

	// Enable foreign key constraints
	if _, err := mv.db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		logrus.WithError(err).Warn("Could not enable foreign key constraints")
	}

	// Analyze tables for optimization
	tables := []string{"tag_standards", "operations", "operation_results", "operation_logs"}
	for _, table := range tables {
		if _, err := mv.db.Exec(fmt.Sprintf("ANALYZE %s", table)); err != nil {
			logrus.WithError(err).WithField("table", table).Warn("Could not analyze table")
		}
	}

	// Vacuum database to reclaim space and fix corruption
	if _, err := mv.db.Exec("VACUUM"); err != nil {
		logrus.WithError(err).Warn("Could not vacuum database")
	}

	logrus.WithField("component", "migration_validator").Info("Migration repair completed")
	return nil
}