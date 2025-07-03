package cleanup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ResourceType represents different types of resources that need cleanup
type ResourceType string

const (
	ResourceTypeTempFile      ResourceType = "temp_file"
	ResourceTypeTempDir       ResourceType = "temp_dir"
	ResourceTypeBackupFile    ResourceType = "backup_file"
	ResourceTypeLockFile      ResourceType = "lock_file"
	ResourceTypeSchemaCache   ResourceType = "schema_cache"
	ResourceTypeProviderCache ResourceType = "provider_cache"
)

// Resource represents a resource that needs cleanup
type Resource struct {
	Type        ResourceType  `json:"type"`
	Path        string        `json:"path"`
	Description string        `json:"description"`
	CreatedAt   time.Time     `json:"created_at"`
	TTL         time.Duration `json:"ttl,omitempty"` // Time to live
	Priority    int           `json:"priority"`      // Cleanup priority (higher = clean first)
	OnFailure   string        `json:"on_failure"`    // Action on cleanup failure: "ignore", "warn", "error"
}

// CleanupManager manages resource cleanup
type CleanupManager struct {
	resources    map[string]*Resource
	mu           sync.RWMutex
	logger       *logrus.Logger
	cleanupHooks []func() error
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewCleanupManager creates a new cleanup manager
func NewCleanupManager(logger *logrus.Logger) *CleanupManager {
	if logger == nil {
		logger = logrus.New()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	cm := &CleanupManager{
		resources:    make(map[string]*Resource),
		logger:       logger,
		cleanupHooks: make([]func() error, 0),
		ctx:          ctx,
		cancel:       cancel,
	}
	
	// Start background cleanup routine
	go cm.backgroundCleanup()
	
	return cm
}

// Register registers a resource for cleanup
func (cm *CleanupManager) Register(resource *Resource) error {
	if resource == nil {
		return fmt.Errorf("resource cannot be nil")
	}
	
	if resource.Path == "" {
		return fmt.Errorf("resource path cannot be empty")
	}
	
	// Set defaults
	if resource.CreatedAt.IsZero() {
		resource.CreatedAt = time.Now()
	}
	if resource.OnFailure == "" {
		resource.OnFailure = "warn"
	}
	
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	cm.resources[resource.Path] = resource
	
	cm.logger.WithFields(logrus.Fields{
		"type":        resource.Type,
		"path":        resource.Path,
		"description": resource.Description,
		"ttl":         resource.TTL,
	}).Debug("Registered resource for cleanup")
	
	return nil
}

// RegisterTempFile registers a temporary file for cleanup
func (cm *CleanupManager) RegisterTempFile(path, description string) error {
	return cm.Register(&Resource{
		Type:        ResourceTypeTempFile,
		Path:        path,
		Description: description,
		TTL:         24 * time.Hour, // Default TTL for temp files
		Priority:    1,
		OnFailure:   "warn",
	})
}

// RegisterTempDir registers a temporary directory for cleanup
func (cm *CleanupManager) RegisterTempDir(path, description string) error {
	return cm.Register(&Resource{
		Type:        ResourceTypeTempDir,
		Path:        path,
		Description: description,
		TTL:         24 * time.Hour,
		Priority:    2, // Directories cleaned after files
		OnFailure:   "warn",
	})
}

// RegisterBackupFile registers a backup file for cleanup
func (cm *CleanupManager) RegisterBackupFile(path, description string, ttl time.Duration) error {
	return cm.Register(&Resource{
		Type:        ResourceTypeBackupFile,
		Path:        path,
		Description: description,
		TTL:         ttl,
		Priority:    0, // Lower priority for backup files
		OnFailure:   "ignore", // Don't fail if backup cleanup fails
	})
}

// Unregister removes a resource from cleanup
func (cm *CleanupManager) Unregister(path string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	if resource, exists := cm.resources[path]; exists {
		delete(cm.resources, path)
		cm.logger.WithFields(logrus.Fields{
			"type": resource.Type,
			"path": path,
		}).Debug("Unregistered resource from cleanup")
	}
}

// AddCleanupHook adds a function to be called during cleanup
func (cm *CleanupManager) AddCleanupHook(hook func() error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.cleanupHooks = append(cm.cleanupHooks, hook)
}

// CleanupAll performs cleanup of all registered resources
func (cm *CleanupManager) CleanupAll() error {
	cm.logger.Info("Starting cleanup of all resources")
	
	var allErrors []error
	
	// Execute cleanup hooks first
	cm.mu.RLock()
	hooks := make([]func() error, len(cm.cleanupHooks))
	copy(hooks, cm.cleanupHooks)
	cm.mu.RUnlock()
	
	for _, hook := range hooks {
		if err := hook(); err != nil {
			cm.logger.WithError(err).Warn("Cleanup hook failed")
			allErrors = append(allErrors, err)
		}
	}
	
	// Get all resources and sort by priority
	cm.mu.RLock()
	resources := make([]*Resource, 0, len(cm.resources))
	for _, resource := range cm.resources {
		resources = append(resources, resource)
	}
	cm.mu.RUnlock()
	
	// Sort by priority (higher priority first)
	for i := 0; i < len(resources); i++ {
		for j := i + 1; j < len(resources); j++ {
			if resources[i].Priority < resources[j].Priority {
				resources[i], resources[j] = resources[j], resources[i]
			}
		}
	}
	
	// Cleanup resources
	for _, resource := range resources {
		if err := cm.cleanupResource(resource); err != nil {
			allErrors = append(allErrors, err)
		}
	}
	
	// Clear all registered resources
	cm.mu.Lock()
	cm.resources = make(map[string]*Resource)
	cm.mu.Unlock()
	
	if len(allErrors) > 0 {
		cm.logger.WithField("error_count", len(allErrors)).Warn("Some cleanup operations failed")
		return fmt.Errorf("cleanup failed with %d errors", len(allErrors))
	}
	
	cm.logger.Info("All resources cleaned up successfully")
	return nil
}

// CleanupExpired cleans up resources that have exceeded their TTL
func (cm *CleanupManager) CleanupExpired() error {
	cm.logger.Debug("Cleaning up expired resources")
	
	now := time.Now()
	var expiredResources []*Resource
	
	cm.mu.RLock()
	for _, resource := range cm.resources {
		if resource.TTL > 0 && now.Sub(resource.CreatedAt) > resource.TTL {
			expiredResources = append(expiredResources, resource)
		}
	}
	cm.mu.RUnlock()
	
	var allErrors []error
	for _, resource := range expiredResources {
		if err := cm.cleanupResource(resource); err != nil {
			allErrors = append(allErrors, err)
		} else {
			// Remove from registered resources
			cm.Unregister(resource.Path)
		}
	}
	
	if len(expiredResources) > 0 {
		cm.logger.WithField("cleaned_count", len(expiredResources)).Info("Cleaned up expired resources")
	}
	
	if len(allErrors) > 0 {
		return fmt.Errorf("failed to cleanup %d expired resources", len(allErrors))
	}
	
	return nil
}

// cleanupResource performs cleanup of a single resource
func (cm *CleanupManager) cleanupResource(resource *Resource) error {
	logger := cm.logger.WithFields(logrus.Fields{
		"type":        resource.Type,
		"path":        resource.Path,
		"description": resource.Description,
	})
	
	logger.Debug("Cleaning up resource")
	
	// Check if resource exists
	if _, err := os.Stat(resource.Path); os.IsNotExist(err) {
		logger.Debug("Resource does not exist, skipping cleanup")
		return nil
	}
	
	var err error
	switch resource.Type {
	case ResourceTypeTempFile, ResourceTypeBackupFile, ResourceTypeLockFile:
		err = os.Remove(resource.Path)
	case ResourceTypeTempDir, ResourceTypeSchemaCache, ResourceTypeProviderCache:
		err = os.RemoveAll(resource.Path)
	default:
		// Default to file removal
		err = os.Remove(resource.Path)
	}
	
	if err != nil {
		switch resource.OnFailure {
		case "ignore":
			logger.WithError(err).Debug("Resource cleanup failed (ignored)")
			return nil
		case "warn":
			logger.WithError(err).Warn("Resource cleanup failed")
			return nil
		case "error":
			logger.WithError(err).Error("Resource cleanup failed")
			return err
		default:
			logger.WithError(err).Warn("Resource cleanup failed")
			return nil
		}
	}
	
	logger.Info("Resource cleaned up successfully")
	return nil
}

// backgroundCleanup runs periodic cleanup of expired resources
func (cm *CleanupManager) backgroundCleanup() {
	ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if err := cm.CleanupExpired(); err != nil {
				cm.logger.WithError(err).Debug("Background cleanup failed")
			}
		case <-cm.ctx.Done():
			cm.logger.Debug("Background cleanup stopped")
			return
		}
	}
}

// Stop stops the cleanup manager and performs final cleanup
func (cm *CleanupManager) Stop() error {
	cm.cancel() // Stop background cleanup
	return cm.CleanupAll()
}

// GetRegisteredResources returns a copy of all registered resources
func (cm *CleanupManager) GetRegisteredResources() map[string]*Resource {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	result := make(map[string]*Resource)
	for k, v := range cm.resources {
		resourceCopy := *v
		result[k] = &resourceCopy
	}
	return result
}

// GetResourceCount returns the number of registered resources
func (cm *CleanupManager) GetResourceCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.resources)
}

// CleanupByType cleans up all resources of a specific type
func (cm *CleanupManager) CleanupByType(resourceType ResourceType) error {
	cm.mu.RLock()
	var targetResources []*Resource
	for _, resource := range cm.resources {
		if resource.Type == resourceType {
			targetResources = append(targetResources, resource)
		}
	}
	cm.mu.RUnlock()
	
	var allErrors []error
	for _, resource := range targetResources {
		if err := cm.cleanupResource(resource); err != nil {
			allErrors = append(allErrors, err)
		} else {
			cm.Unregister(resource.Path)
		}
	}
	
	if len(allErrors) > 0 {
		return fmt.Errorf("failed to cleanup %d resources of type %s", len(allErrors), resourceType)
	}
	
	return nil
}

// CreateTempFile creates a temporary file and registers it for cleanup
func (cm *CleanupManager) CreateTempFile(pattern, description string) (*os.File, error) {
	file, err := os.CreateTemp("", pattern)
	if err != nil {
		return nil, err
	}
	
	if err := cm.RegisterTempFile(file.Name(), description); err != nil {
		file.Close()
		os.Remove(file.Name())
		return nil, err
	}
	
	return file, nil
}

// CreateTempDir creates a temporary directory and registers it for cleanup
func (cm *CleanupManager) CreateTempDir(pattern, description string) (string, error) {
	dir, err := os.MkdirTemp("", pattern)
	if err != nil {
		return "", err
	}
	
	if err := cm.RegisterTempDir(dir, description); err != nil {
		os.RemoveAll(dir)
		return "", err
	}
	
	return dir, nil
}

// CleanupValidationFiles specifically cleans up files created during validation
func (cm *CleanupManager) CleanupValidationFiles(baseDir string) error {
	cm.logger.WithField("base_dir", baseDir).Info("Cleaning up validation-specific files")
	
	// Patterns for validation-related temporary files
	patterns := []string{
		"*.terratag.tf",
		"*.tf.bak",
		".terraform.lock.hcl",
		".terratag_*",
		"terraform_*",
	}
	
	var allErrors []error
	
	// Walk through the directory and find matching files
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			// Skip .terraform directories
			if strings.HasSuffix(path, ".terraform") {
				return filepath.SkipDir
			}
			return nil
		}
		
		// Check if file matches any pattern
		for _, pattern := range patterns {
			if matched, _ := filepath.Match(pattern, info.Name()); matched {
				cm.logger.WithField("file", path).Debug("Removing validation file")
				if err := os.Remove(path); err != nil {
					cm.logger.WithError(err).WithField("file", path).Warn("Failed to remove validation file")
					allErrors = append(allErrors, err)
				}
				break
			}
		}
		
		return nil
	})
	
	if err != nil {
		allErrors = append(allErrors, err)
	}
	
	if len(allErrors) > 0 {
		return fmt.Errorf("failed to cleanup %d validation files", len(allErrors))
	}
	
	return nil
}

// Global cleanup manager instance
var globalCleanupManager *CleanupManager
var globalCleanupOnce sync.Once

// GetGlobalCleanupManager returns the global cleanup manager instance
func GetGlobalCleanupManager() *CleanupManager {
	globalCleanupOnce.Do(func() {
		globalCleanupManager = NewCleanupManager(logrus.StandardLogger())
	})
	return globalCleanupManager
}

// RegisterGlobalCleanup registers a resource with the global cleanup manager
func RegisterGlobalCleanup(resource *Resource) error {
	return GetGlobalCleanupManager().Register(resource)
}

// GlobalCleanup performs global cleanup
func GlobalCleanup() error {
	if globalCleanupManager != nil {
		return globalCleanupManager.CleanupAll()
	}
	return nil
}