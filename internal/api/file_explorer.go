package api

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// DirectoryItem represents a file or directory in the file system
type DirectoryItem struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	IsDirectory bool   `json:"is_directory"`
	Size        int64  `json:"size,omitempty"`
	ModTime     string `json:"mod_time,omitempty"`
	HasTerraform bool  `json:"has_terraform,omitempty"`
}

// DirectoryListing represents the response for directory browsing
type DirectoryListing struct {
	CurrentPath string          `json:"current_path"`
	ParentPath  string          `json:"parent_path,omitempty"`
	Items       []DirectoryItem `json:"items"`
	IsRoot      bool            `json:"is_root"`
}

// BrowseDirectory handles directory browsing requests
func (h *Handlers) BrowseDirectory(c *gin.Context) {
	logger := logrus.WithFields(logrus.Fields{
		"component": "api",
		"action":    "browseDirectory",
	})

	// Get the path parameter
	requestedPath := c.Query("path")
	if requestedPath == "" {
		requestedPath = "/"
	}

	logger.WithField("requestedPath", requestedPath).Info("Browsing directory")

	// Security: Clean and validate the path
	cleanPath := filepath.Clean(requestedPath)
	
	// For Docker environment, restrict to certain paths
	allowedPaths := []string{
		"/workspace",
		"/demo-deployment",
		"/standards",
		"/tmp",
	}
	
	isAllowed := false
	for _, allowed := range allowedPaths {
		if strings.HasPrefix(cleanPath, allowed) || cleanPath == "/" {
			isAllowed = true
			break
		}
	}
	
	if !isAllowed {
		logger.WithField("path", cleanPath).Warn("Access denied to path")
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied to this path",
			"code":  403,
		})
		return
	}

	// Handle root path special case
	if cleanPath == "/" {
		listing := DirectoryListing{
			CurrentPath: "/",
			Items: []DirectoryItem{
				{Name: "workspace", Path: "/workspace", IsDirectory: true},
				{Name: "demo-deployment", Path: "/demo-deployment", IsDirectory: true},
				{Name: "standards", Path: "/standards", IsDirectory: true},
			},
			IsRoot: true,
		}
		c.JSON(http.StatusOK, listing)
		return
	}

	// Check if the path exists
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		logger.WithFields(logrus.Fields{
			"path":  cleanPath,
			"error": err.Error(),
		}).Warn("Path does not exist")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Path does not exist",
			"code":  404,
		})
		return
	}

	// Read directory contents
	entries, err := os.ReadDir(cleanPath)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"path":  cleanPath,
			"error": err.Error(),
		}).Error("Failed to read directory")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read directory",
			"code":  500,
		})
		return
	}

	// Build the response
	var items []DirectoryItem
	
	for _, entry := range entries {
		// Skip hidden files and certain directories
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		
		entryPath := filepath.Join(cleanPath, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		item := DirectoryItem{
			Name:        entry.Name(),
			Path:        entryPath,
			IsDirectory: entry.IsDir(),
			Size:        info.Size(),
			ModTime:     info.ModTime().Format("2006-01-02 15:04:05"),
		}

		// Check if directory contains Terraform files
		if entry.IsDir() {
			item.HasTerraform = hasTerraformFiles(entryPath)
		}

		items = append(items, item)
	}

	// Sort items: directories first, then by name
	sort.Slice(items, func(i, j int) bool {
		if items[i].IsDirectory != items[j].IsDirectory {
			return items[i].IsDirectory
		}
		return items[i].Name < items[j].Name
	})

	// Determine parent path
	parentPath := ""
	if cleanPath != "/" {
		parentPath = filepath.Dir(cleanPath)
		if parentPath == "." {
			parentPath = "/"
		}
	}

	listing := DirectoryListing{
		CurrentPath: cleanPath,
		ParentPath:  parentPath,
		Items:       items,
		IsRoot:      cleanPath == "/",
	}

	logger.WithFields(logrus.Fields{
		"path":      cleanPath,
		"itemCount": len(items),
	}).Info("Directory browsed successfully")

	c.JSON(http.StatusOK, listing)
}

// hasTerraformFiles checks if a directory contains .tf files
func hasTerraformFiles(dirPath string) bool {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tf") {
			return true
		}
	}
	return false
}

// GetDirectoryInfo returns information about a specific directory
func (h *Handlers) GetDirectoryInfo(c *gin.Context) {
	logger := logrus.WithFields(logrus.Fields{
		"component": "api",
		"action":    "getDirectoryInfo",
	})

	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Path parameter is required",
			"code":  400,
		})
		return
	}

	cleanPath := filepath.Clean(path)
	logger.WithField("path", cleanPath).Info("Getting directory info")

	// Check if path exists and is a directory
	info, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Path does not exist",
				"code":  404,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to access path",
				"code":  500,
			})
		}
		return
	}

	if !info.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Path is not a directory",
			"code":  400,
		})
		return
	}

	// Count Terraform files
	terraformCount := 0
	subdirCount := 0
	
	entries, err := os.ReadDir(cleanPath)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				subdirCount++
			} else if strings.HasSuffix(entry.Name(), ".tf") {
				terraformCount++
			}
		}
	}

	// Check for terraform initialization
	isInitialized := false
	initPaths := []string{
		filepath.Join(cleanPath, ".terraform"),
		filepath.Join(cleanPath, ".terraform.lock.hcl"),
	}
	
	for _, initPath := range initPaths {
		if _, err := os.Stat(initPath); err == nil {
			isInitialized = true
			break
		}
	}

	response := gin.H{
		"path":             cleanPath,
		"exists":           true,
		"is_directory":     true,
		"terraform_files":  terraformCount,
		"subdirectories":   subdirCount,
		"is_initialized":   isInitialized,
		"has_terraform":    terraformCount > 0,
	}

	logger.WithFields(logrus.Fields{
		"path":           cleanPath,
		"terraformFiles": terraformCount,
		"subdirs":        subdirCount,
		"initialized":    isInitialized,
	}).Info("Directory info retrieved")

	c.JSON(http.StatusOK, response)
}