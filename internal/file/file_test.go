package file

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadHCLFile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "terratag-file-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "valid HCL file",
			content: `resource "aws_instance" "test" {
  ami = "ami-12345"
  instance_type = "t2.micro"
}`,
			wantErr: false,
		},
		{
			name: "valid HCL with locals",
			content: `locals {
  common_tags = {
    Environment = "prod"
  }
}

resource "aws_instance" "test" {
  ami = "ami-12345"
  tags = local.common_tags
}`,
			wantErr: false,
		},
		{
			name:    "empty file",
			content: "",
			wantErr: false,
		},
		{
			name: "invalid HCL syntax",
			content: `resource "aws_instance" "test" {
  ami = "ami-12345"
  // missing closing brace`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write test file
			testFile := filepath.Join(tmpDir, "test.tf")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Read and parse
			hclFile, err := ReadHCLFile(testFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadHCLFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && hclFile == nil {
				t.Error("ReadHCLFile() returned nil file without error")
			}
		})
	}
}

func TestReadHCLFileNonExistent(t *testing.T) {
	_, err := ReadHCLFile("/non/existent/file.tf")
	if err == nil {
		t.Error("ReadHCLFile() expected error for non-existent file")
	}
}

func TestCreateFile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "terratag-file-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		path    string
		content string
		wantErr bool
	}{
		{
			name:    "create simple file",
			path:    filepath.Join(tmpDir, "test.tf"),
			content: "resource \"aws_instance\" \"test\" {}",
			wantErr: false,
		},
		{
			name:    "create file with multi-line content",
			path:    filepath.Join(tmpDir, "multi.tf"),
			content: "resource \"aws_instance\" \"test\" {\n  ami = \"ami-12345\"\n}",
			wantErr: false,
		},
		{
			name:    "overwrite existing file",
			path:    filepath.Join(tmpDir, "existing.tf"),
			content: "new content",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For overwrite test, create the file first
			if tt.name == "overwrite existing file" {
				os.WriteFile(tt.path, []byte("old content"), 0644)
			}

			err := CreateFile(tt.path, tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// Verify file was created
				content, err := os.ReadFile(tt.path)
				if err != nil {
					t.Errorf("Failed to read created file: %v", err)
				}
				if string(content) != tt.content {
					t.Errorf("File content doesn't match. Got %s, want %s", content, tt.content)
				}
			}
		})
	}
}

func TestReplaceWithTerratagFile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "terratag-file-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name          string
		originalFile  string
		content       string
		rename        bool
		expectTagFile bool
	}{
		{
			name:          "replace with rename",
			originalFile:  filepath.Join(tmpDir, "main.tf"),
			content:       "resource \"aws_instance\" \"tagged\" {}",
			rename:        true,
			expectTagFile: true,
		},
		{
			name:          "replace without rename",
			originalFile:  filepath.Join(tmpDir, "module.tf"),
			content:       "resource \"aws_s3_bucket\" \"tagged\" {}",
			rename:        false,
			expectTagFile: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create original file
			originalContent := "resource \"aws_instance\" \"original\" {}"
			if err := os.WriteFile(tt.originalFile, []byte(originalContent), 0644); err != nil {
				t.Fatalf("Failed to create original file: %v", err)
			}

			err := ReplaceWithTerratagFile(tt.originalFile, tt.content, tt.rename)
			if err != nil {
				t.Errorf("ReplaceWithTerratagFile() error = %v", err)
			}

			// Check backup file exists
			backupFile := tt.originalFile + ".bak"
			if _, err := os.Stat(backupFile); os.IsNotExist(err) {
				t.Error("Backup file was not created")
			} else {
				// Verify backup content
				backupContent, _ := os.ReadFile(backupFile)
				if string(backupContent) != originalContent {
					t.Error("Backup file content doesn't match original")
				}
			}

			if tt.rename {
				// Check terratag file exists
				taggedFile := strings.TrimSuffix(tt.originalFile, filepath.Ext(tt.originalFile)) + ".terratag.tf"
				if _, err := os.Stat(taggedFile); os.IsNotExist(err) {
					t.Error("Terratag file was not created")
				} else {
					// Verify content
					taggedContent, _ := os.ReadFile(taggedFile)
					if string(taggedContent) != tt.content {
						t.Error("Terratag file content doesn't match")
					}
				}
				// Original file should not exist
				if _, err := os.Stat(tt.originalFile); !os.IsNotExist(err) {
					t.Error("Original file should not exist when rename=true")
				}
			} else {
				// Original file should have new content
				newContent, _ := os.ReadFile(tt.originalFile)
				if string(newContent) != tt.content {
					t.Error("Original file content was not updated")
				}
			}
		})
	}
}

func TestGetFilename(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "simple filename",
			path:     "/path/to/main.tf",
			expected: "main",
		},
		{
			name:     "filename with dots",
			path:     "/path/to/main.module.tf",
			expected: "main-module",
		},
		{
			name:     "filename with multiple extensions",
			path:     "config.backup.tf",
			expected: "config-backup",
		},
		{
			name:     "just filename",
			path:     "variables.tf",
			expected: "variables",
		},
		{
			name:     "path with dots",
			path:     "./modules/vpc.network/main.tf",
			expected: "main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFilename(tt.path)
			if result != tt.expected {
				t.Errorf("GetFilename(%s) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}