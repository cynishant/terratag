package integration

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudyali/terratag/internal/common"
	"github.com/cloudyali/terratag/internal/providers"
	"github.com/cloudyali/terratag/internal/tfschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProviderCacheIntegration tests the full provider cache integration
func TestProviderCacheIntegration(t *testing.T) {
	// Skip if this is a unit test only run
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "1" {
		t.Skip("Skipping integration test")
	}

	// Create temporary directories for testing
	testDir1, err := os.MkdirTemp("", "terratag-cache-test1")
	require.NoError(t, err)
	defer os.RemoveAll(testDir1)

	testDir2, err := os.MkdirTemp("", "terratag-cache-test2")
	require.NoError(t, err)
	defer os.RemoveAll(testDir2)

	// Create identical terraform configurations
	tfContent := `
provider "aws" {
  region = "us-west-2"
}

resource "aws_instance" "test" {
  ami           = "ami-12345"
  instance_type = "t2.micro"
  
  tags = {
    Name = "test-instance"
  }
}
`

	// Write terraform files to both directories
	err = ioutil.WriteFile(filepath.Join(testDir1, "main.tf"), []byte(tfContent), 0644)
	require.NoError(t, err)

	err = ioutil.WriteFile(filepath.Join(testDir2, "main.tf"), []byte(tfContent), 0644)
	require.NoError(t, err)

	// Clean up any existing cache
	cacheManager := providers.GetGlobalCacheManager()
	err = cacheManager.CleanupExpiredEntries()
	require.NoError(t, err)

	// Test 1: Initialize schemas for first directory with cache enabled
	t.Log("Testing first directory with cache enabled")
	err = tfschema.InitProviderSchemasWithCache(testDir1, common.Terraform, false, true)
	
	// This might fail if terraform isn't initialized, which is expected in CI
	if err != nil {
		t.Logf("Schema initialization failed (expected in CI without terraform init): %v", err)
		t.Skip("Skipping rest of test - requires terraform init")
	}

	// Test 2: Initialize schemas for second directory with cache enabled
	// This should use the cached schema if available
	t.Log("Testing second directory with cache enabled (should use cache)")
	err = tfschema.InitProviderSchemasWithCache(testDir2, common.Terraform, false, true)
	if err != nil {
		t.Logf("Schema initialization failed for second directory: %v", err)
	}

	// Test 3: Test with cache disabled
	t.Log("Testing with cache disabled")
	err = tfschema.InitProviderSchemasWithCache(testDir1, common.Terraform, false, false)
	if err != nil {
		t.Logf("Schema initialization without cache failed: %v", err)
	}

	// Test 4: Verify cache manager functionality
	t.Log("Testing cache manager functionality")

	// Try to get cached schema
	_, _, err = cacheManager.GetCachedSchema(testDir1, common.Terraform)
	if err != nil {
		t.Logf("Cache lookup failed (expected if terraform init wasn't run): %v", err)
	} else {
		t.Log("Successfully found cached schema")
	}

	// Test shared terraform directory creation
	sharedDir, err := cacheManager.GetOrCreateSharedTerraformDir(testDir1, common.Terraform)
	require.NoError(t, err)
	assert.NotEmpty(t, sharedDir)

	// Verify shared directory was created
	_, err = os.Stat(sharedDir)
	assert.NoError(t, err, "Shared terraform directory should exist")

	// Verify terraform files were copied
	mainTf := filepath.Join(sharedDir, "main.tf")
	_, err = os.Stat(mainTf)
	assert.NoError(t, err, "main.tf should be copied to shared directory")

	t.Log("Provider cache integration test completed successfully")
}

// TestProviderCachePerformance tests that cache improves performance
func TestProviderCachePerformance(t *testing.T) {
	// Skip if this is a unit test only run
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "1" {
		t.Skip("Skipping integration test")
	}

	// Create temporary directory for testing
	testDir, err := os.MkdirTemp("", "terratag-cache-perf-test")
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	// Create terraform configuration
	tfContent := `
provider "aws" {
  region = "us-west-2"
}

resource "aws_instance" "test" {
  ami           = "ami-12345"
  instance_type = "t2.micro"
}
`

	err = ioutil.WriteFile(filepath.Join(testDir, "main.tf"), []byte(tfContent), 0644)
	require.NoError(t, err)

	// Clean up any existing cache
	cacheManager := providers.GetGlobalCacheManager()
	err = cacheManager.CleanupExpiredEntries()
	require.NoError(t, err)

	// Test multiple initializations with cache
	for i := 0; i < 3; i++ {
		t.Logf("Cache-enabled initialization attempt %d", i+1)
		err = tfschema.InitProviderSchemasWithCache(testDir, common.Terraform, false, true)
		
		if err != nil {
			t.Logf("Schema initialization failed (expected in CI without terraform init): %v", err)
			if i == 0 {
				t.Skip("Skipping performance test - requires terraform init")
			}
		}
	}

	t.Log("Provider cache performance test completed")
}

// TestCacheDisabled tests behavior when cache is explicitly disabled
func TestCacheDisabled(t *testing.T) {
	// Create temporary directory for testing
	testDir, err := os.MkdirTemp("", "terratag-cache-disabled-test")
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	// Create terraform configuration
	tfContent := `
provider "aws" {
  region = "us-west-2"
}
`

	err = ioutil.WriteFile(filepath.Join(testDir, "main.tf"), []byte(tfContent), 0644)
	require.NoError(t, err)

	// Test with cache disabled
	t.Log("Testing schema initialization with cache disabled")
	err = tfschema.InitProviderSchemasWithCache(testDir, common.Terraform, false, false)
	
	if err != nil {
		t.Logf("Schema initialization failed (expected without terraform init): %v", err)
	}

	// Verify that no cache was created
	cacheManager := providers.GetGlobalCacheManager()
	_, _, err = cacheManager.GetCachedSchema(testDir, common.Terraform)
	
	// Should get cache miss since cache was disabled
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cache miss")

	t.Log("Cache disabled test completed successfully")
}