package cleanup

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanupManager_Register(t *testing.T) {
	cm := NewCleanupManager(nil)
	defer cm.Stop()

	resource := &Resource{
		Type:        ResourceTypeTempFile,
		Path:        "/tmp/test.txt",
		Description: "test file",
	}

	err := cm.Register(resource)
	assert.NoError(t, err)

	resources := cm.GetRegisteredResources()
	assert.Len(t, resources, 1)
	assert.Equal(t, resource.Path, resources["/tmp/test.txt"].Path)
}

func TestCleanupManager_RegisterValidation(t *testing.T) {
	cm := NewCleanupManager(nil)
	defer cm.Stop()

	// Test nil resource
	err := cm.Register(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")

	// Test empty path
	err = cm.Register(&Resource{Type: ResourceTypeTempFile})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestCleanupManager_RegisterTempFile(t *testing.T) {
	cm := NewCleanupManager(nil)
	defer cm.Stop()

	err := cm.RegisterTempFile("/tmp/test.txt", "test file")
	assert.NoError(t, err)

	resources := cm.GetRegisteredResources()
	resource := resources["/tmp/test.txt"]
	assert.Equal(t, ResourceTypeTempFile, resource.Type)
	assert.Equal(t, 24*time.Hour, resource.TTL)
	assert.Equal(t, "warn", resource.OnFailure)
}

func TestCleanupManager_Unregister(t *testing.T) {
	cm := NewCleanupManager(nil)
	defer cm.Stop()

	// Register a resource
	err := cm.RegisterTempFile("/tmp/test.txt", "test file")
	require.NoError(t, err)
	assert.Equal(t, 1, cm.GetResourceCount())

	// Unregister it
	cm.Unregister("/tmp/test.txt")
	assert.Equal(t, 0, cm.GetResourceCount())

	// Unregister non-existent (should not error)
	cm.Unregister("/tmp/nonexistent.txt")
}

func TestCleanupManager_CleanupTempFiles(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "cleanup-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	cm := NewCleanupManager(nil)
	defer cm.Stop()

	// Create test files
	testFile1 := filepath.Join(tmpDir, "test1.txt")
	testFile2 := filepath.Join(tmpDir, "test2.txt")
	
	err = os.WriteFile(testFile1, []byte("test1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(testFile2, []byte("test2"), 0644)
	require.NoError(t, err)

	// Register files for cleanup
	err = cm.RegisterTempFile(testFile1, "test file 1")
	require.NoError(t, err)
	err = cm.RegisterTempFile(testFile2, "test file 2")
	require.NoError(t, err)

	// Verify files exist
	assert.FileExists(t, testFile1)
	assert.FileExists(t, testFile2)

	// Cleanup all
	err = cm.CleanupAll()
	assert.NoError(t, err)

	// Verify files are gone
	assert.NoFileExists(t, testFile1)
	assert.NoFileExists(t, testFile2)

	// Verify resources are unregistered
	assert.Equal(t, 0, cm.GetResourceCount())
}

func TestCleanupManager_CleanupTempDir(t *testing.T) {
	// Create temporary directory for test
	parentDir, err := os.MkdirTemp("", "cleanup-test")
	require.NoError(t, err)
	defer os.RemoveAll(parentDir)

	cm := NewCleanupManager(nil)
	defer cm.Stop()

	// Create test directory with files
	testDir := filepath.Join(parentDir, "testdir")
	err = os.Mkdir(testDir, 0755)
	require.NoError(t, err)
	
	testFile := filepath.Join(testDir, "file.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	// Register directory for cleanup
	err = cm.RegisterTempDir(testDir, "test directory")
	require.NoError(t, err)

	// Verify directory exists
	assert.DirExists(t, testDir)
	assert.FileExists(t, testFile)

	// Cleanup all
	err = cm.CleanupAll()
	assert.NoError(t, err)

	// Verify directory and contents are gone
	assert.NoDirExists(t, testDir)
	assert.NoFileExists(t, testFile)
}

func TestCleanupManager_CleanupExpired(t *testing.T) {
	cm := NewCleanupManager(nil)
	defer cm.Stop()

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "cleanup-test")
	require.NoError(t, err)
	tmpFile.Close()

	// Register with very short TTL
	resource := &Resource{
		Type:      ResourceTypeTempFile,
		Path:      tmpFile.Name(),
		CreatedAt: time.Now().Add(-2 * time.Second), // Created 2 seconds ago
		TTL:       1 * time.Second,                  // TTL of 1 second (expired)
		OnFailure: "warn",
	}
	err = cm.Register(resource)
	require.NoError(t, err)

	// Register another file that's not expired
	tmpFile2, err := os.CreateTemp("", "cleanup-test")
	require.NoError(t, err)
	tmpFile2.Close()

	resource2 := &Resource{
		Type:      ResourceTypeTempFile,
		Path:      tmpFile2.Name(),
		CreatedAt: time.Now(),              // Created now
		TTL:       1 * time.Hour,           // TTL of 1 hour (not expired)
		OnFailure: "warn",
	}
	err = cm.Register(resource2)
	require.NoError(t, err)

	// Cleanup expired
	err = cm.CleanupExpired()
	assert.NoError(t, err)

	// First file should be gone, second should remain
	assert.NoFileExists(t, tmpFile.Name())
	assert.FileExists(t, tmpFile2.Name())

	// Clean up remaining file
	os.Remove(tmpFile2.Name())
}

func TestCleanupManager_CleanupByType(t *testing.T) {
	cm := NewCleanupManager(nil)
	defer cm.Stop()

	// Create test files
	tmpFile1, err := os.CreateTemp("", "cleanup-test")
	require.NoError(t, err)
	tmpFile1.Close()

	tmpFile2, err := os.CreateTemp("", "cleanup-test")
	require.NoError(t, err)
	tmpFile2.Close()

	// Register different types
	err = cm.RegisterTempFile(tmpFile1.Name(), "temp file")
	require.NoError(t, err)

	err = cm.RegisterBackupFile(tmpFile2.Name(), "backup file", time.Hour)
	require.NoError(t, err)

	// Cleanup only temp files
	err = cm.CleanupByType(ResourceTypeTempFile)
	assert.NoError(t, err)

	// Temp file should be gone, backup file should remain
	assert.NoFileExists(t, tmpFile1.Name())
	assert.FileExists(t, tmpFile2.Name())

	// Clean up remaining file
	os.Remove(tmpFile2.Name())
}

func TestCleanupManager_CreateTempFile(t *testing.T) {
	cm := NewCleanupManager(nil)
	defer cm.Stop()

	file, err := cm.CreateTempFile("test-*", "test temp file")
	require.NoError(t, err)
	defer file.Close()

	// File should exist
	assert.FileExists(t, file.Name())

	// Should be registered for cleanup
	assert.Equal(t, 1, cm.GetResourceCount())

	// Cleanup should remove it
	err = cm.CleanupAll()
	assert.NoError(t, err)
	assert.NoFileExists(t, file.Name())
}

func TestCleanupManager_CreateTempDir(t *testing.T) {
	cm := NewCleanupManager(nil)
	defer cm.Stop()

	dir, err := cm.CreateTempDir("test-*", "test temp directory")
	require.NoError(t, err)

	// Directory should exist
	assert.DirExists(t, dir)

	// Should be registered for cleanup
	assert.Equal(t, 1, cm.GetResourceCount())

	// Cleanup should remove it
	err = cm.CleanupAll()
	assert.NoError(t, err)
	assert.NoDirExists(t, dir)
}

func TestCleanupManager_CleanupHooks(t *testing.T) {
	cm := NewCleanupManager(nil)
	defer cm.Stop()

	hookCalled := false
	cm.AddCleanupHook(func() error {
		hookCalled = true
		return nil
	})

	err := cm.CleanupAll()
	assert.NoError(t, err)
	assert.True(t, hookCalled)
}

func TestCleanupManager_CleanupValidationFiles(t *testing.T) {
	// Create temporary directory structure
	tmpDir, err := os.MkdirTemp("", "validation-cleanup-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	cm := NewCleanupManager(nil)
	defer cm.Stop()

	// Create various files including validation files
	files := map[string]bool{
		"main.tf":                  false, // Should not be cleaned
		"variables.tf":             false, // Should not be cleaned
		"main.terratag.tf":         true,  // Should be cleaned
		"test.tf.bak":              true,  // Should be cleaned
		".terraform.lock.hcl":      true,  // Should be cleaned
		".terratag_temp":           true,  // Should be cleaned
		"terraform_state.tmp":      true,  // Should be cleaned
		"regular_file.txt":         false, // Should not be cleaned
	}

	for filename, _ := range files {
		filePath := filepath.Join(tmpDir, filename)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		require.NoError(t, err)
	}

	// Create .terraform directory (should be skipped)
	terraformDir := filepath.Join(tmpDir, ".terraform")
	err = os.Mkdir(terraformDir, 0755)
	require.NoError(t, err)
	
	terraformFile := filepath.Join(terraformDir, "some_file")
	err = os.WriteFile(terraformFile, []byte("terraform content"), 0644)
	require.NoError(t, err)

	// Run validation cleanup
	err = cm.CleanupValidationFiles(tmpDir)
	assert.NoError(t, err)

	// Check which files were cleaned
	for filename, shouldClean := range files {
		filePath := filepath.Join(tmpDir, filename)
		if shouldClean {
			assert.NoFileExists(t, filePath, "File %s should have been cleaned", filename)
		} else {
			assert.FileExists(t, filePath, "File %s should not have been cleaned", filename)
		}
	}

	// .terraform directory should still exist
	assert.DirExists(t, terraformDir)
	assert.FileExists(t, terraformFile)
}

func TestCleanupManager_NonExistentFile(t *testing.T) {
	cm := NewCleanupManager(nil)
	defer cm.Stop()

	// Register non-existent file
	err := cm.RegisterTempFile("/tmp/nonexistent.txt", "non-existent file")
	require.NoError(t, err)

	// Cleanup should not error
	err = cm.CleanupAll()
	assert.NoError(t, err)
}

func TestGlobalCleanupManager(t *testing.T) {
	// Test global cleanup manager
	globalMgr := GetGlobalCleanupManager()
	assert.NotNil(t, globalMgr)

	// Should return the same instance
	globalMgr2 := GetGlobalCleanupManager()
	assert.Same(t, globalMgr, globalMgr2)

	// Test global registration
	resource := &Resource{
		Type:        ResourceTypeTempFile,
		Path:        "/tmp/global-test.txt",
		Description: "global test file",
	}
	
	err := RegisterGlobalCleanup(resource)
	assert.NoError(t, err)

	// Should be registered in global manager
	assert.Equal(t, 1, globalMgr.GetResourceCount())

	// Clean up
	err = GlobalCleanup()
	assert.NoError(t, err)
}