package services

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrationValidator_ValidateMigrations(t *testing.T) {
	// Create temporary directory for test migrations
	tmpDir, err := os.MkdirTemp("", "migration-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test migration files
	createTestMigration(t, tmpDir, "001_initial_schema.up.sql", `
CREATE TABLE test_table (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);
CREATE INDEX idx_test_name ON test_table(name);
`)
	createTestMigration(t, tmpDir, "001_initial_schema.down.sql", `
DROP INDEX idx_test_name;
DROP TABLE test_table;
`)

	createTestMigration(t, tmpDir, "002_add_column.up.sql", `
ALTER TABLE test_table ADD COLUMN description TEXT;
`)
	createTestMigration(t, tmpDir, "002_add_column.down.sql", `
-- SQLite doesn't support DROP COLUMN, so we recreate the table
CREATE TABLE test_table_new (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);
INSERT INTO test_table_new (id, name) SELECT id, name FROM test_table;
DROP TABLE test_table;
ALTER TABLE test_table_new RENAME TO test_table;
CREATE INDEX idx_test_name ON test_table(name);
`)

	// Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create migration validator
	validator := NewMigrationValidator(db, tmpDir)

	// Run validation
	result, err := validator.ValidateMigrations()
	require.NoError(t, err)

	// Check results
	assert.True(t, result.IsValid)
	assert.Equal(t, 2, result.TotalMigrations) // Only up migrations counted
	assert.Equal(t, 0, result.AppliedMigrations) // None applied yet
	assert.Equal(t, 2, result.PendingMigrations)
	assert.Len(t, result.Errors, 0)
}

func TestMigrationValidator_ValidateNaming(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "migration-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create migrations with gaps in numbering
	createTestMigration(t, tmpDir, "001_first.up.sql", "CREATE TABLE test1 (id INTEGER);")
	createTestMigration(t, tmpDir, "001_first.down.sql", "DROP TABLE test1;")
	createTestMigration(t, tmpDir, "003_third.up.sql", "CREATE TABLE test3 (id INTEGER);") // Gap at 2
	createTestMigration(t, tmpDir, "003_third.down.sql", "DROP TABLE test3;")

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	validator := NewMigrationValidator(db, tmpDir)
	result, err := validator.ValidateMigrations()
	require.NoError(t, err)

	// Should have warnings about non-sequential numbering
	assert.Contains(t, result.Warnings, "Migration version 3 is not sequential (expected 2)")
}

func TestMigrationValidator_MissingDownFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "migration-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create up migration without down migration
	createTestMigration(t, tmpDir, "001_test.up.sql", "CREATE TABLE test (id INTEGER);")
	// No down file

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	validator := NewMigrationValidator(db, tmpDir)
	result, err := validator.ValidateMigrations()
	require.NoError(t, err)

	// Should have warning about missing down file
	assert.Contains(t, result.Warnings, "Missing down migration file for version 1 (rollback not possible)")
}

func TestMigrationValidator_DangerousOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "migration-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create migration with dangerous operations
	createTestMigration(t, tmpDir, "001_dangerous.up.sql", `
CREATE TABLE test (id INTEGER);
DROP TABLE old_table;
TRUNCATE some_table;
DELETE FROM users WHERE active = 0;
`)
	createTestMigration(t, tmpDir, "001_dangerous.down.sql", "DROP TABLE test;")

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	validator := NewMigrationValidator(db, tmpDir)
	result, err := validator.ValidateMigrations()
	require.NoError(t, err)

	// Should have warnings about dangerous operations
	foundDangerousWarnings := 0
	for _, warning := range result.Warnings {
		if contains(warning, "dangerous operation") {
			foundDangerousWarnings++
		}
	}
	assert.Greater(t, foundDangerousWarnings, 0)
}

func TestMigrationValidator_SchemaConsistency(t *testing.T) {
	// Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create some tables manually (simulating applied migrations)
	_, err = db.Exec(`
CREATE TABLE tag_standards (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);
CREATE TABLE operations (
    id INTEGER PRIMARY KEY,
    type TEXT NOT NULL
);
`)
	require.NoError(t, err)

	tmpDir, err := os.MkdirTemp("", "migration-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	validator := NewMigrationValidator(db, tmpDir)
	result, err := validator.ValidateMigrations()
	require.NoError(t, err)

	// Should detect missing tables
	foundMissingTables := 0
	for _, error := range result.Errors {
		if contains(error, "does not exist") {
			foundMissingTables++
		}
	}
	assert.Greater(t, foundMissingTables, 0) // Should find missing operation_results and operation_logs
}

func TestMigrationValidator_EmptyMigrations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "migration-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create empty migration
	createTestMigration(t, tmpDir, "001_empty.up.sql", "")
	createTestMigration(t, tmpDir, "001_empty.down.sql", "")

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	validator := NewMigrationValidator(db, tmpDir)
	result, err := validator.ValidateMigrations()
	require.NoError(t, err)

	// Should have warnings about empty migrations
	emptyWarnings := 0
	for _, warning := range result.Warnings {
		if contains(warning, "is empty") {
			emptyWarnings++
		}
	}
	assert.Equal(t, 2, emptyWarnings) // Both up and down files are empty
}

func TestMigrationValidator_RepairMigrations(t *testing.T) {
	// Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create test table
	_, err = db.Exec("CREATE TABLE test_repair (id INTEGER PRIMARY KEY);")
	require.NoError(t, err)

	tmpDir, err := os.MkdirTemp("", "migration-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	validator := NewMigrationValidator(db, tmpDir)
	
	// Should not error even if there are no issues to repair
	err = validator.RepairMigrations()
	assert.NoError(t, err)
}

// Helper functions

func createTestMigration(t *testing.T, dir, filename, content string) {
	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || 
		 (len(s) > len(substr) && 
		  (s[:len(substr)] == substr || 
		   s[len(s)-len(substr):] == substr || 
		   containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}