package providers

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudyali/terratag/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheManager_Basic(t *testing.T) {
	// Create a temporary cache directory
	tmpDir, err := os.MkdirTemp("", "terratag-cache-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a cache manager
	cm := NewCacheManager(tmpDir)

	// Create a test terraform directory
	testDir, err := os.MkdirTemp("", "terraform-test")
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	// Create a simple terraform file
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

	// Test cache miss
	_, _, err = cm.GetCachedSchema(testDir, common.Terraform)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cache miss")

	// Create a temporary terraform directory for caching
	terraformCacheDir, err := os.MkdirTemp("", "terraform-cache")
	require.NoError(t, err)
	defer os.RemoveAll(terraformCacheDir)

	// Test caching schema
	testSchema := `{"provider_schemas": {"aws": {"resource_schemas": {"aws_instance": {}}}}}`
	err = cm.CacheSchema(testDir, common.Terraform, testSchema, terraformCacheDir)
	require.NoError(t, err)

	// Test cache hit
	cachedSchema, terraformDir, err := cm.GetCachedSchema(testDir, common.Terraform)
	require.NoError(t, err)
	assert.Equal(t, testSchema, cachedSchema)
	assert.Equal(t, terraformCacheDir, terraformDir)
}

func TestCacheManager_ExpiredEntries(t *testing.T) {
	// Create a temporary cache directory
	tmpDir, err := os.MkdirTemp("", "terratag-cache-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a cache manager
	cm := NewCacheManager(tmpDir)

	// Create a test terraform directory
	testDir, err := os.MkdirTemp("", "terraform-test")
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	// Create a simple terraform file
	tfContent := `provider "aws" {}`
	err = ioutil.WriteFile(filepath.Join(testDir, "main.tf"), []byte(tfContent), 0644)
	require.NoError(t, err)

	// Cache a schema
	testSchema := `{"provider_schemas": {"aws": {}}}`
	err = cm.CacheSchema(testDir, common.Terraform, testSchema, "/path/to/terraform")
	require.NoError(t, err)

	// Manually create an expired cache entry
	requirements, err := cm.extractProviderRequirements(testDir, common.Terraform)
	require.NoError(t, err)

	cacheKey := cm.generateCacheKey(requirements)
	cacheFile := filepath.Join(tmpDir, cacheKey+".json")

	expiredEntry := CacheEntry{
		Requirements: requirements,
		SchemaData:   testSchema,
		CachedAt:     time.Now().Add(-8 * 24 * time.Hour), // 8 days ago
		TerraformDir: "/path/to/terraform",
	}

	data, err := json.Marshal(expiredEntry)
	require.NoError(t, err)
	err = ioutil.WriteFile(cacheFile, data, 0644)
	require.NoError(t, err)

	// Check that we have a cache file
	files, err := ioutil.ReadDir(tmpDir)
	require.NoError(t, err)
	assert.Len(t, files, 1)

	// Clean up expired entries
	err = cm.CleanupExpiredEntries()
	require.NoError(t, err)

	// Check that the expired file was removed
	files, err = ioutil.ReadDir(tmpDir)
	require.NoError(t, err)
	assert.Len(t, files, 0)
}

func TestCacheManager_ProviderRequirements(t *testing.T) {
	// Create a temporary directory for test
	testDir, err := os.MkdirTemp("", "terraform-test")
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	// Create terraform files with different providers
	tfContent1 := `
provider "aws" {
  region = "us-west-2"
}
`
	tfContent2 := `
provider "google" {
  project = "my-project"
}
`

	err = ioutil.WriteFile(filepath.Join(testDir, "aws.tf"), []byte(tfContent1), 0644)
	require.NoError(t, err)
	err = ioutil.WriteFile(filepath.Join(testDir, "gcp.tf"), []byte(tfContent2), 0644)
	require.NoError(t, err)

	// Create cache manager
	tmpCacheDir, err := os.MkdirTemp("", "terratag-cache-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpCacheDir)

	cm := NewCacheManager(tmpCacheDir)

	// Extract provider requirements
	requirements, err := cm.extractProviderRequirements(testDir, common.Terraform)
	require.NoError(t, err)

	// We should have extracted some provider information
	assert.Greater(t, len(requirements), 0, "Should have extracted provider requirements")

	// Generate cache key
	cacheKey1 := cm.generateCacheKey(requirements)
	assert.NotEmpty(t, cacheKey1)

	// Generate cache key again with same requirements
	cacheKey2 := cm.generateCacheKey(requirements)
	assert.Equal(t, cacheKey1, cacheKey2, "Cache keys should be consistent")

	// Test with different requirements
	emptyRequirements := []ProviderRequirement{}
	cacheKey3 := cm.generateCacheKey(emptyRequirements)
	assert.NotEqual(t, cacheKey1, cacheKey3, "Different requirements should generate different cache keys")
}

func TestCacheManager_GlobalInstance(t *testing.T) {
	// Test that global cache manager is a singleton
	cm1 := GetGlobalCacheManager()
	cm2 := GetGlobalCacheManager()

	assert.Same(t, cm1, cm2, "Global cache manager should be a singleton")
	assert.NotNil(t, cm1.cacheDir, "Cache directory should be set")
}

func TestCacheManager_SharedTerraformDir(t *testing.T) {
	// Create a temporary cache directory
	tmpDir, err := os.MkdirTemp("", "terratag-cache-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a cache manager
	cm := NewCacheManager(tmpDir)

	// Create a test terraform directory
	testDir, err := os.MkdirTemp("", "terraform-test")
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	// Create a simple terraform file
	tfContent := `
provider "aws" {
  region = "us-west-2"
}
`
	err = ioutil.WriteFile(filepath.Join(testDir, "main.tf"), []byte(tfContent), 0644)
	require.NoError(t, err)

	// Get shared terraform directory
	sharedDir1, err := cm.GetOrCreateSharedTerraformDir(testDir, common.Terraform)
	require.NoError(t, err)
	assert.NotEmpty(t, sharedDir1)

	// Check that directory was created
	_, err = os.Stat(sharedDir1)
	assert.NoError(t, err)

	// Get the same shared directory again
	sharedDir2, err := cm.GetOrCreateSharedTerraformDir(testDir, common.Terraform)
	require.NoError(t, err)
	assert.Equal(t, sharedDir1, sharedDir2, "Should return the same shared directory")

	// Check that terraform files were copied
	mainTf := filepath.Join(sharedDir1, "main.tf")
	_, err = os.Stat(mainTf)
	assert.NoError(t, err, "main.tf should be copied to shared directory")
}