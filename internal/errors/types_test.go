package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTerratagError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *TerratagError
		expected string
	}{
		{
			name: "error with message only",
			err: &TerratagError{
				Code:    ErrCodeFileNotFound,
				Message: "file not found",
			},
			expected: "[FILE_NOT_FOUND] file not found",
		},
		{
			name: "error with message and details",
			err: &TerratagError{
				Code:    ErrCodeValidationFailed,
				Message: "validation failed",
				Details: "missing required tag 'Environment'",
			},
			expected: "[VALIDATION_FAILED] validation failed: missing required tag 'Environment'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestTerratagError_WithContext(t *testing.T) {
	err := New(ErrCodeFileNotFound, "file not found")
	err = err.WithContext("file_path", "/path/to/file.tf")
	err = err.WithContext("operation", "read")

	assert.Equal(t, "/path/to/file.tf", err.Context["file_path"])
	assert.Equal(t, "read", err.Context["operation"])
}

func TestTerratagError_WithComponent(t *testing.T) {
	err := New(ErrCodeHCLParseError, "parse error").WithComponent("hcl_parser")
	assert.Equal(t, "hcl_parser", err.Component)
}

func TestTerratagError_WithFile(t *testing.T) {
	err := New(ErrCodeHCLParseError, "parse error").WithFile("/path/to/file.tf", 42)
	assert.Equal(t, "/path/to/file.tf", err.FilePath)
	assert.Equal(t, 42, err.LineNumber)
}

func TestTerratagError_IsCode(t *testing.T) {
	err := New(ErrCodeFileNotFound, "file not found")
	
	assert.True(t, err.IsCode(ErrCodeFileNotFound))
	assert.False(t, err.IsCode(ErrCodeValidationFailed))
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := Wrap(originalErr, ErrCodeFileNotFound, "file not found")
	
	assert.Equal(t, ErrCodeFileNotFound, wrappedErr.Code)
	assert.Equal(t, "file not found", wrappedErr.Message)
	assert.Equal(t, originalErr, wrappedErr.Cause)
	assert.Equal(t, originalErr, wrappedErr.Unwrap())
}

func TestWrapf(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := Wrapf(originalErr, ErrCodeFileNotFound, "file %s not found", "test.tf")
	
	assert.Equal(t, "file test.tf not found", wrappedErr.Message)
	assert.Equal(t, originalErr, wrappedErr.Cause)
}

func TestErrorCollection(t *testing.T) {
	collection := &ErrorCollection{}
	
	// Initially empty
	assert.False(t, collection.HasErrors())
	assert.Equal(t, 0, collection.Count())
	assert.Nil(t, collection.First())
	assert.Equal(t, "no errors", collection.Error())

	// Add one error
	err1 := New(ErrCodeFileNotFound, "file1 not found")
	collection.Add(err1)
	
	assert.True(t, collection.HasErrors())
	assert.Equal(t, 1, collection.Count())
	assert.Equal(t, err1, collection.First())
	assert.Equal(t, err1.Error(), collection.Error())

	// Add another error
	err2 := New(ErrCodeValidationFailed, "validation failed")
	collection.Add(err2)
	
	assert.Equal(t, 2, collection.Count())
	assert.Contains(t, collection.Error(), "multiple errors occurred")
	assert.Contains(t, collection.Error(), err1.Error())
	assert.Contains(t, collection.Error(), err2.Error())

	// Adding nil should not increase count
	collection.Add(nil)
	assert.Equal(t, 2, collection.Count())
}

func TestErrorToTerratagError(t *testing.T) {
	tests := []struct {
		name     string
		input    error
		expected *TerratagError
	}{
		{
			name:     "nil error",
			input:    nil,
			expected: nil,
		},
		{
			name:  "already TerratagError",
			input: New(ErrCodeFileNotFound, "file not found"),
			expected: &TerratagError{
				Code:    ErrCodeFileNotFound,
				Message: "file not found",
			},
		},
		{
			name:  "standard error",
			input: errors.New("standard error"),
			expected: &TerratagError{
				Code:    ErrCodeInternalError,
				Message: "standard error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ErrorToTerratagError(tt.input)
			
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected.Code, result.Code)
				assert.Equal(t, tt.expected.Message, result.Message)
			}
		})
	}
}

func TestIsErrorCode(t *testing.T) {
	// Create wrapped error
	originalErr := errors.New("original")
	terratagErr := Wrap(originalErr, ErrCodeFileNotFound, "file not found")
	wrappedAgain := Wrap(terratagErr, ErrCodeValidationFailed, "validation failed")

	// Should find the code in the chain
	assert.True(t, IsErrorCode(wrappedAgain, ErrCodeValidationFailed))
	assert.True(t, IsErrorCode(wrappedAgain, ErrCodeFileNotFound))
	assert.False(t, IsErrorCode(wrappedAgain, ErrCodeHCLParseError))

	// Test with standard error
	assert.False(t, IsErrorCode(originalErr, ErrCodeFileNotFound))
}

func TestGetErrorCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorCode
	}{
		{
			name:     "TerratagError",
			err:      New(ErrCodeFileNotFound, "file not found"),
			expected: ErrCodeFileNotFound,
		},
		{
			name:     "wrapped TerratagError",
			err:      Wrap(New(ErrCodeFileNotFound, "file not found"), ErrCodeValidationFailed, "validation failed"),
			expected: ErrCodeValidationFailed,
		},
		{
			name:     "standard error",
			err:      errors.New("standard error"),
			expected: ErrCodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetErrorCode(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInferErrorCode(t *testing.T) {
	tests := []struct {
		message  string
		expected ErrorCode
	}{
		{"file not found", ErrCodeFileNotFound},
		{"no such file or directory", ErrCodeFileNotFound},
		{"permission denied", ErrCodeFileAccessDenied},
		{"access denied", ErrCodeFileAccessDenied},
		{"parse error in file", ErrCodeHCLParseError},
		{"syntax error", ErrCodeHCLParseError},
		{"terraform not found", ErrCodeTerraformNotFound},
		{"validation failed", ErrCodeValidationFailed},
		{"invalid configuration", ErrCodeValidationFailed},
		{"database connection error", ErrCodeDatabaseConnection},
		{"sql error", ErrCodeDatabaseConnection},
		{"timeout occurred", ErrCodeTimeoutError},
		{"operation cancelled", ErrCodeCancellationError},
		{"unauthorized access", ErrCodeUnauthorized},
		{"forbidden operation", ErrCodeForbidden},
		{"unknown error", ErrCodeInternalError},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			result := inferErrorCode(tt.message)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommonErrorConstructors(t *testing.T) {
	t.Run("FileNotFound", func(t *testing.T) {
		err := FileNotFound("/path/to/file.tf")
		assert.Equal(t, ErrCodeFileNotFound, err.Code)
		assert.Equal(t, "/path/to/file.tf", err.FilePath)
		assert.Equal(t, "/path/to/file.tf", err.Context["file_path"])
	})

	t.Run("FileAccessDenied", func(t *testing.T) {
		cause := errors.New("permission denied")
		err := FileAccessDenied("/path/to/file.tf", cause)
		assert.Equal(t, ErrCodeFileAccessDenied, err.Code)
		assert.Equal(t, cause, err.Cause)
		assert.Equal(t, "/path/to/file.tf", err.FilePath)
	})

	t.Run("HCLParseError", func(t *testing.T) {
		cause := errors.New("syntax error")
		err := HCLParseError("/path/to/file.tf", 42, cause)
		assert.Equal(t, ErrCodeHCLParseError, err.Code)
		assert.Equal(t, "/path/to/file.tf", err.FilePath)
		assert.Equal(t, 42, err.LineNumber)
		assert.Equal(t, "hcl_parser", err.Component)
		assert.Equal(t, cause, err.Cause)
	})

	t.Run("ValidationFailed", func(t *testing.T) {
		err := ValidationFailed("missing required tags")
		assert.Equal(t, ErrCodeValidationFailed, err.Code)
		assert.Equal(t, "validator", err.Component)
	})

	t.Run("InternalError", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := InternalError("something went wrong", cause)
		assert.Equal(t, ErrCodeInternalError, err.Code)
		assert.Equal(t, cause, err.Cause)
		assert.NotEmpty(t, err.StackTrace) // Should have stack trace
	})
}