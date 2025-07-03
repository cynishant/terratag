package providers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cloudyali/terratag/internal/common"
)

// CacheManager manages a centralized provider cache to avoid storage bloat
type CacheManager struct {
	cacheDir string
	mu       sync.RWMutex
}

// ProviderRequirement represents a provider requirement for caching
type ProviderRequirement struct {
	Source  string `json:"source"`
	Version string `json:"version,omitempty"`
}

// CacheEntry represents a cached provider schema entry
type CacheEntry struct {
	Requirements []ProviderRequirement `json:"requirements"`
	SchemaData   string                 `json:"schema_data"`
	CachedAt     time.Time              `json:"cached_at"`
	TerraformDir string                 `json:"terraform_dir"`
}

var (
	globalCacheManager *CacheManager
	cacheInitOnce      sync.Once
)

// GetGlobalCacheManager returns the singleton cache manager
func GetGlobalCacheManager() *CacheManager {
	cacheInitOnce.Do(func() {
		cacheDir := filepath.Join(os.TempDir(), "terratag-provider-cache")
		globalCacheManager = NewCacheManager(cacheDir)
	})
	return globalCacheManager
}

// NewCacheManager creates a new provider cache manager
func NewCacheManager(cacheDir string) *CacheManager {
	return &CacheManager{
		cacheDir: cacheDir,
	}
}

// GetCachedSchema retrieves cached provider schema if available
func (cm *CacheManager) GetCachedSchema(dir string, iacType common.IACType) (string, string, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	requirements, err := cm.extractProviderRequirements(dir, iacType)
	if err != nil {
		return "", "", fmt.Errorf("failed to extract provider requirements: %w", err)
	}

	cacheKey := cm.generateCacheKey(requirements)
	cacheFile := filepath.Join(cm.cacheDir, cacheKey+".json")

	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return "", "", fmt.Errorf("cache miss")
	}

	data, err := ioutil.ReadFile(cacheFile)
	if err != nil {
		return "", "", fmt.Errorf("failed to read cache file: %w", err)
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return "", "", fmt.Errorf("failed to unmarshal cache entry: %w", err)
	}

	// Check if cache is still valid (24 hours)
	if time.Since(entry.CachedAt) > 24*time.Hour {
		return "", "", fmt.Errorf("cache expired")
	}

	// Verify that the cached .terraform directory still exists
	if _, err := os.Stat(entry.TerraformDir); os.IsNotExist(err) {
		// Cache entry exists but .terraform directory was removed, invalidate cache
		os.Remove(cacheFile)
		return "", "", fmt.Errorf("cached terraform directory no longer exists")
	}

	log.Printf("[INFO] Using cached provider schema for directory: %s", dir)
	return entry.SchemaData, entry.TerraformDir, nil
}

// CacheSchema stores provider schema in cache
func (cm *CacheManager) CacheSchema(dir string, iacType common.IACType, schemaData string, terraformDir string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if err := os.MkdirAll(cm.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	requirements, err := cm.extractProviderRequirements(dir, iacType)
	if err != nil {
		return fmt.Errorf("failed to extract provider requirements: %w", err)
	}

	cacheKey := cm.generateCacheKey(requirements)
	cacheFile := filepath.Join(cm.cacheDir, cacheKey+".json")

	entry := CacheEntry{
		Requirements: requirements,
		SchemaData:   schemaData,
		CachedAt:     time.Now(),
		TerraformDir: terraformDir,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	if err := ioutil.WriteFile(cacheFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	log.Printf("[INFO] Cached provider schema for directory: %s (key: %s)", dir, cacheKey)
	return nil
}

// GetOrCreateSharedTerraformDir returns a shared .terraform directory for the given requirements
func (cm *CacheManager) GetOrCreateSharedTerraformDir(dir string, iacType common.IACType) (string, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	requirements, err := cm.extractProviderRequirements(dir, iacType)
	if err != nil {
		return "", fmt.Errorf("failed to extract provider requirements: %w", err)
	}

	cacheKey := cm.generateCacheKey(requirements)
	sharedTerraformDir := filepath.Join(cm.cacheDir, "terraform-"+cacheKey)

	// Check if shared .terraform directory already exists
	if _, err := os.Stat(sharedTerraformDir); err == nil {
		log.Printf("[INFO] Using existing shared terraform directory: %s", sharedTerraformDir)
		return sharedTerraformDir, nil
	}

	// Create shared .terraform directory
	if err := os.MkdirAll(sharedTerraformDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create shared terraform directory: %w", err)
	}

	// Copy provider configuration to shared directory
	if err := cm.copyProviderConfig(dir, sharedTerraformDir, iacType); err != nil {
		return "", fmt.Errorf("failed to copy provider config: %w", err)
	}

	log.Printf("[INFO] Created new shared terraform directory: %s", sharedTerraformDir)
	return sharedTerraformDir, nil
}

// CleanupExpiredEntries removes expired cache entries
func (cm *CacheManager) CleanupExpiredEntries() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, err := os.Stat(cm.cacheDir); os.IsNotExist(err) {
		return nil
	}

	entries, err := ioutil.ReadDir(cm.cacheDir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	cleaned := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		cacheFile := filepath.Join(cm.cacheDir, entry.Name())
		data, err := ioutil.ReadFile(cacheFile)
		if err != nil {
			continue
		}

		var cacheEntry CacheEntry
		if err := json.Unmarshal(data, &cacheEntry); err != nil {
			continue
		}

		// Remove entries older than 7 days
		if time.Since(cacheEntry.CachedAt) > 7*24*time.Hour {
			os.Remove(cacheFile)
			// Also remove associated terraform directory if it exists
			if cacheEntry.TerraformDir != "" {
				os.RemoveAll(cacheEntry.TerraformDir)
			}
			cleaned++
		}
	}

	if cleaned > 0 {
		log.Printf("[INFO] Cleaned up %d expired cache entries", cleaned)
	}

	return nil
}

// extractProviderRequirements extracts provider requirements from terraform configuration
func (cm *CacheManager) extractProviderRequirements(dir string, iacType common.IACType) ([]ProviderRequirement, error) {
	var requirements []ProviderRequirement

	// For now, create a simple hash based on provider configuration files
	// This could be enhanced to actually parse terraform configuration
	configFiles := []string{}
	
	// Look for terraform configuration files
	patterns := []string{"*.tf", "*.tf.json"}
	if iacType == common.Terragrunt || iacType == common.TerragruntRunAll {
		patterns = append(patterns, "*.hcl")
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			continue
		}
		configFiles = append(configFiles, matches...)
	}

	// Simple approach: create requirements based on file content hashes
	// This ensures that directories with identical provider configurations
	// will share the same cache entry
	for _, file := range configFiles {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			continue
		}

		// Extract provider information from content (simplified)
		// This could be enhanced with proper HCL parsing
		content_str := string(content)
		
		// Look for provider blocks
		if strings.Contains(content_str, "provider \"") {
			lines := strings.Split(content_str, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "provider \"") && strings.Contains(line, "\"") {
					// Extract provider name from provider "name" syntax
					start := strings.Index(line, "\"") + 1
					end := strings.Index(line[start:], "\"")
					if end > 0 {
						providerName := line[start : start+end]
						requirements = append(requirements, ProviderRequirement{
							Source: providerName,
						})
					}
				}
			}
		}
	}

	// Sort requirements for consistent cache keys
	sort.Slice(requirements, func(i, j int) bool {
		return requirements[i].Source < requirements[j].Source
	})

	return requirements, nil
}

// generateCacheKey generates a cache key from provider requirements
func (cm *CacheManager) generateCacheKey(requirements []ProviderRequirement) string {
	data, _ := json.Marshal(requirements)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])[:16] // Use first 16 chars for shorter keys
}

// copyProviderConfig copies provider configuration to shared directory
func (cm *CacheManager) copyProviderConfig(sourceDir, targetDir string, iacType common.IACType) error {
	// Look for terraform configuration files to copy
	patterns := []string{"*.tf", "*.tf.json", ".terraform.lock.hcl"}
	if iacType == common.Terragrunt || iacType == common.TerragruntRunAll {
		patterns = append(patterns, "*.hcl")
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(sourceDir, pattern))
		if err != nil {
			continue
		}

		for _, match := range matches {
			fileName := filepath.Base(match)
			targetFile := filepath.Join(targetDir, fileName)

			sourceContent, err := ioutil.ReadFile(match)
			if err != nil {
				continue
			}

			if err := ioutil.WriteFile(targetFile, sourceContent, 0644); err != nil {
				return fmt.Errorf("failed to copy %s: %w", fileName, err)
			}
		}
	}

	return nil
}

// InitProviders initializes providers in the shared directory
func (cm *CacheManager) InitProviders(terraformDir string, iacType common.IACType, defaultToTerraform bool) error {
	name := "terraform"
	if iacType == common.Terragrunt || iacType == common.TerragruntRunAll {
		name = "terragrunt"
	} else if _, err := exec.LookPath("tofu"); !defaultToTerraform && err == nil {
		name = "tofu"
	}

	log.Printf("[INFO] Initializing providers in shared directory: %s", terraformDir)

	cmd := exec.Command(name, "init", "-backend=false")
	cmd.Dir = terraformDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[ERROR] Failed to initialize providers: %s", string(output))
		return fmt.Errorf("failed to run %s init: %w", name, err)
	}

	log.Printf("[INFO] Successfully initialized providers in shared directory")
	return nil
}