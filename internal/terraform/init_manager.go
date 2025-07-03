package terraform

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cloudyali/terratag/internal/common"
	"github.com/cloudyali/terratag/internal/providers"
)

// InitManager handles intelligent terraform initialization with error detection and retry logic
type InitManager struct {
	workingDir         string
	iacType           common.IACType
	defaultToTerraform bool
	useCache          bool
	logger            *log.Logger
}

// InitError represents different types of initialization errors
type InitError struct {
	Type    InitErrorType
	Message string
	Cause   error
}

type InitErrorType int

const (
	InitErrorPlugin InitErrorType = iota
	InitErrorBackend
	InitErrorDependency
	InitErrorModule
	InitErrorCloud
	InitErrorGeneric
)

func (e *InitError) Error() string {
	return fmt.Sprintf("terraform init error (%v): %s", e.Type, e.Message)
}

// NewInitManager creates a new initialization manager
func NewInitManager(workingDir string, iacType common.IACType, defaultToTerraform bool, useCache bool) *InitManager {
	return &InitManager{
		workingDir:         workingDir,
		iacType:           iacType,
		defaultToTerraform: defaultToTerraform,
		useCache:          useCache,
		logger:            log.New(os.Stdout, "[TERRAFORM-INIT] ", log.LstdFlags),
	}
}

// EnsureInitialized ensures that terraform is properly initialized in the working directory
// This implements the improved algorithm with automatic init on failure
func (im *InitManager) EnsureInitialized() error {
	// Step 1: Check if already initialized
	if im.isAlreadyInitialized() {
		im.logger.Printf("Directory %s is already initialized", im.workingDir)
		return nil
	}

	// Step 2: Try to run a simple terraform validate to detect init needs
	err := im.tryValidateCommand()
	if err == nil {
		im.logger.Printf("Terraform validate succeeded, initialization appears complete")
		return nil
	}

	// Step 3: Detect if this is an init-related error
	initErr := im.detectInitError(err)
	if initErr == nil {
		// Not an init error, return original error
		return fmt.Errorf("terraform validation failed with non-init error: %w", err)
	}

	im.logger.Printf("Detected init error: %s", initErr.Message)

	// Step 4: Run initialization
	if err := im.runSmartInit(initErr.Type); err != nil {
		return fmt.Errorf("failed to initialize terraform: %w", err)
	}

	// Step 5: Verify initialization succeeded
	if err := im.tryValidateCommand(); err != nil {
		return fmt.Errorf("terraform still failing after init: %w", err)
	}

	im.logger.Printf("Successfully initialized terraform in %s", im.workingDir)
	return nil
}

// isAlreadyInitialized checks if terraform is already initialized
func (im *InitManager) isAlreadyInitialized() bool {
	initPaths := []string{
		filepath.Join(im.workingDir, ".terraform"),
		filepath.Join(im.workingDir, ".terraform.lock.hcl"),
	}

	if im.iacType == common.Terragrunt || im.iacType == common.TerragruntRunAll {
		initPaths = append(initPaths, filepath.Join(im.workingDir, ".terragrunt-cache"))
	}

	for _, path := range initPaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return false
}

// tryValidateCommand runs a simple terraform validate to detect initialization status
func (im *InitManager) tryValidateCommand() error {
	cmdName := im.getCommandName()
	
	var cmd *exec.Cmd
	if im.iacType == common.TerragruntRunAll {
		cmd = exec.Command(cmdName, "run-all", "validate", "-no-color")
	} else {
		cmd = exec.Command(cmdName, "validate", "-no-color")
	}
	
	cmd.Dir = im.workingDir
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("validate command failed: %s", string(output))
	}
	
	return nil
}

// detectInitError analyzes error output to determine if it's an initialization error
func (im *InitManager) detectInitError(err error) *InitError {
	errorStr := strings.ToLower(err.Error())
	
	// Plugin-related errors
	pluginPatterns := []string{
		"error: could not load plugin",
		"error: required plugins are not installed",
		"error: provider requirements cannot be satisfied by locked dependencies",
		"error: inconsistent dependency lock file",
		"required provider",
		"plugin not found",
	}
	
	for _, pattern := range pluginPatterns {
		if strings.Contains(errorStr, pattern) {
			return &InitError{
				Type:    InitErrorPlugin,
				Message: "Provider plugin initialization required",
				Cause:   err,
			}
		}
	}
	
	// Backend-related errors
	backendPatterns := []string{
		"error: initialization required",
		"error: backend initialization required",
		"backend configuration changed",
		"terraform has not been initialized",
	}
	
	for _, pattern := range backendPatterns {
		if strings.Contains(errorStr, pattern) {
			return &InitError{
				Type:    InitErrorBackend,
				Message: "Backend initialization required",
				Cause:   err,
			}
		}
	}
	
	// Module-related errors
	modulePatterns := []string{
		"error: module not installed",
		"module source has changed",
		"module not found",
	}
	
	for _, pattern := range modulePatterns {
		if strings.Contains(errorStr, pattern) {
			return &InitError{
				Type:    InitErrorModule,
				Message: "Module installation required",
				Cause:   err,
			}
		}
	}
	
	// Terraform Cloud errors
	cloudPatterns := []string{
		"error: terraform cloud initialization required",
		"terraform cloud workspace",
		"remote backend configuration",
	}
	
	for _, pattern := range cloudPatterns {
		if strings.Contains(errorStr, pattern) {
			return &InitError{
				Type:    InitErrorCloud,
				Message: "Terraform Cloud initialization required",
				Cause:   err,
			}
		}
	}
	
	// Generic init prompt
	if strings.Contains(errorStr, "please run \"terraform init\"") ||
		strings.Contains(errorStr, "run 'terraform init'") {
		return &InitError{
			Type:    InitErrorGeneric,
			Message: "Terraform initialization required",
			Cause:   err,
		}
	}
	
	return nil
}

// runSmartInit runs terraform init with appropriate flags based on error type and cache settings
func (im *InitManager) runSmartInit(errorType InitErrorType) error {
	// If cache is enabled, try to use cached initialization first
	if im.useCache {
		if err := im.trySmartCacheInit(); err == nil {
			return nil
		}
		im.logger.Printf("Cache-based init failed, falling back to local init")
	}

	// Fall back to local initialization
	return im.runLocalInit(errorType)
}

// trySmartCacheInit attempts to initialize using the provider cache system
func (im *InitManager) trySmartCacheInit() error {
	cacheManager := providers.GetGlobalCacheManager()
	
	// Get or create shared terraform directory
	sharedTerraformDir, err := cacheManager.GetOrCreateSharedTerraformDir(im.workingDir, im.iacType)
	if err != nil {
		return fmt.Errorf("failed to create shared terraform directory: %w", err)
	}

	// Initialize providers in shared directory if needed
	if !im.isDirectoryInitialized(sharedTerraformDir) {
		if err := cacheManager.InitProviders(sharedTerraformDir, im.iacType, im.defaultToTerraform); err != nil {
			return fmt.Errorf("failed to initialize shared providers: %w", err)
		}
	}

	// Create symlink or copy .terraform directory to working directory
	localTerraformDir := filepath.Join(im.workingDir, ".terraform")
	sharedTerraformPluginDir := filepath.Join(sharedTerraformDir, ".terraform")
	
	// Remove existing .terraform directory if it exists
	if _, err := os.Stat(localTerraformDir); err == nil {
		os.RemoveAll(localTerraformDir)
	}

	// Try to create symlink first (faster), fall back to copy
	if err := os.Symlink(sharedTerraformPluginDir, localTerraformDir); err != nil {
		im.logger.Printf("Symlink failed, copying terraform directory: %v", err)
		if err := im.copyTerraformDir(sharedTerraformPluginDir, localTerraformDir); err != nil {
			return fmt.Errorf("failed to copy terraform directory: %w", err)
		}
	}

	// Copy lock file if it exists
	sharedLockFile := filepath.Join(sharedTerraformDir, ".terraform.lock.hcl")
	localLockFile := filepath.Join(im.workingDir, ".terraform.lock.hcl")
	
	if _, err := os.Stat(sharedLockFile); err == nil {
		if err := im.copyFile(sharedLockFile, localLockFile); err != nil {
			im.logger.Printf("Warning: failed to copy lock file: %v", err)
		}
	}

	im.logger.Printf("Successfully linked cached terraform providers to %s", im.workingDir)
	return nil
}

// runLocalInit runs terraform init locally with appropriate flags
func (im *InitManager) runLocalInit(errorType InitErrorType) error {
	cmdName := im.getCommandName()
	args := []string{"init", "-input=false", "-no-color"}

	// Add specific flags based on error type
	switch errorType {
	case InitErrorPlugin, InitErrorDependency:
		args = append(args, "-upgrade")
	case InitErrorBackend:
		args = append(args, "-reconfigure")
	case InitErrorModule:
		args = append(args, "-get=true")
	}

	// Terragrunt-specific handling
	if im.iacType == common.TerragruntRunAll {
		args = append([]string{"run-all"}, args...)
		args = append(args, "--terragrunt-ignore-external-dependencies")
	}

	im.logger.Printf("Running: %s %s", cmdName, strings.Join(args, " "))

	cmd := exec.Command(cmdName, args...)
	cmd.Dir = im.workingDir

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("init command failed: %s\nOutput: %s", err.Error(), string(output))
	}

	im.logger.Printf("Init completed successfully")
	return nil
}

// getCommandName returns the appropriate command name based on IaC type and preferences
func (im *InitManager) getCommandName() string {
	if im.iacType == common.Terragrunt || im.iacType == common.TerragruntRunAll {
		return "terragrunt"
	}

	// Check for OpenTofu unless defaulting to Terraform
	if !im.defaultToTerraform {
		if _, err := exec.LookPath("tofu"); err == nil {
			return "tofu"
		}
	}

	return "terraform"
}

// isDirectoryInitialized checks if a specific directory is initialized
func (im *InitManager) isDirectoryInitialized(dir string) bool {
	terraformDir := filepath.Join(dir, ".terraform")
	_, err := os.Stat(terraformDir)
	return err == nil
}

// copyTerraformDir copies a terraform directory from source to destination
func (im *InitManager) copyTerraformDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return im.copyFile(path, dstPath)
	})
}

// copyFile copies a single file from source to destination
func (im *InitManager) copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0644)
}

// GetInitStatus returns the current initialization status
func (im *InitManager) GetInitStatus() (bool, error) {
	if im.isAlreadyInitialized() {
		// Quick check passed, but verify with validate command
		if err := im.tryValidateCommand(); err != nil {
			// Initialized but has issues
			return false, fmt.Errorf("initialized but validation failed: %w", err)
		}
		return true, nil
	}

	return false, nil
}

// ForceReinit forces a complete reinitialization
func (im *InitManager) ForceReinit() error {
	im.logger.Printf("Forcing reinitialization of %s", im.workingDir)

	// Remove existing initialization
	initPaths := []string{
		filepath.Join(im.workingDir, ".terraform"),
		filepath.Join(im.workingDir, ".terraform.lock.hcl"),
	}

	if im.iacType == common.Terragrunt || im.iacType == common.TerragruntRunAll {
		initPaths = append(initPaths, filepath.Join(im.workingDir, ".terragrunt-cache"))
	}

	for _, path := range initPaths {
		if err := os.RemoveAll(path); err != nil {
			im.logger.Printf("Warning: failed to remove %s: %v", path, err)
		}
	}

	// Run initialization
	return im.runLocalInit(InitErrorGeneric)
}