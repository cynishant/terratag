package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// ErrorCode represents different types of errors that can occur
type ErrorCode string

const (
	// File and I/O errors
	ErrCodeFileNotFound     ErrorCode = "FILE_NOT_FOUND"
	ErrCodeFileAccessDenied ErrorCode = "FILE_ACCESS_DENIED"
	ErrCodeFileCorrupted    ErrorCode = "FILE_CORRUPTED"
	ErrCodeDirectoryInvalid ErrorCode = "DIRECTORY_INVALID"

	// Configuration errors
	ErrCodeConfigInvalid     ErrorCode = "CONFIG_INVALID"
	ErrCodeConfigMissing     ErrorCode = "CONFIG_MISSING"
	ErrCodeTagStandardInvalid ErrorCode = "TAG_STANDARD_INVALID"

	// Terraform/HCL errors
	ErrCodeHCLParseError      ErrorCode = "HCL_PARSE_ERROR"
	ErrCodeTerraformNotFound  ErrorCode = "TERRAFORM_NOT_FOUND"
	ErrCodeTerraformInitError ErrorCode = "TERRAFORM_INIT_ERROR"
	ErrCodeSchemaLoadError    ErrorCode = "SCHEMA_LOAD_ERROR"

	// Validation errors
	ErrCodeValidationFailed   ErrorCode = "VALIDATION_FAILED"
	ErrCodeTagViolation      ErrorCode = "TAG_VIOLATION"
	ErrCodeResourceInvalid   ErrorCode = "RESOURCE_INVALID"

	// Database errors
	ErrCodeDatabaseConnection ErrorCode = "DATABASE_CONNECTION"
	ErrCodeDatabaseMigration  ErrorCode = "DATABASE_MIGRATION"
	ErrCodeDatabaseQuery      ErrorCode = "DATABASE_QUERY"

	// API and security errors
	ErrCodeUnauthorized    ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden       ErrorCode = "FORBIDDEN"
	ErrCodeRateLimited     ErrorCode = "RATE_LIMITED"
	ErrCodeInvalidRequest  ErrorCode = "INVALID_REQUEST"

	// Internal system errors
	ErrCodeInternalError     ErrorCode = "INTERNAL_ERROR"
	ErrCodeTimeoutError      ErrorCode = "TIMEOUT_ERROR"
	ErrCodeCancellationError ErrorCode = "CANCELLATION_ERROR"
)

// TerratagError represents a structured error with additional context
type TerratagError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Cause      error                  `json:"-"`
	StackTrace []string               `json:"stack_trace,omitempty"`
	Component  string                 `json:"component,omitempty"`
	FilePath   string                 `json:"file_path,omitempty"`
	LineNumber int                    `json:"line_number,omitempty"`
}

// Error implements the error interface
func (e *TerratagError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *TerratagError) Unwrap() error {
	return e.Cause
}

// WithContext adds context information to the error
func (e *TerratagError) WithContext(key string, value interface{}) *TerratagError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithComponent adds component information
func (e *TerratagError) WithComponent(component string) *TerratagError {
	e.Component = component
	return e
}

// WithFile adds file location information
func (e *TerratagError) WithFile(filePath string, lineNumber int) *TerratagError {
	e.FilePath = filePath
	e.LineNumber = lineNumber
	return e
}

// WithStackTrace adds stack trace information
func (e *TerratagError) WithStackTrace() *TerratagError {
	e.StackTrace = getStackTrace()
	return e
}

// IsCode checks if the error has a specific error code
func (e *TerratagError) IsCode(code ErrorCode) bool {
	return e.Code == code
}

// New creates a new TerratagError
func New(code ErrorCode, message string) *TerratagError {
	return &TerratagError{
		Code:    code,
		Message: message,
	}
}

// Newf creates a new TerratagError with formatted message
func Newf(code ErrorCode, format string, args ...interface{}) *TerratagError {
	return &TerratagError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap wraps an existing error with TerratagError context
func Wrap(cause error, code ErrorCode, message string) *TerratagError {
	return &TerratagError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// Wrapf wraps an existing error with formatted message
func Wrapf(cause error, code ErrorCode, format string, args ...interface{}) *TerratagError {
	return &TerratagError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Cause:   cause,
	}
}

// ErrorCollection represents a collection of errors
type ErrorCollection struct {
	Errors []error `json:"errors"`
}

// Error implements the error interface
func (ec *ErrorCollection) Error() string {
	if len(ec.Errors) == 0 {
		return "no errors"
	}
	if len(ec.Errors) == 1 {
		return ec.Errors[0].Error()
	}
	
	messages := make([]string, len(ec.Errors))
	for i, err := range ec.Errors {
		messages[i] = err.Error()
	}
	return fmt.Sprintf("multiple errors occurred: %s", strings.Join(messages, "; "))
}

// Add adds an error to the collection
func (ec *ErrorCollection) Add(err error) {
	if err != nil {
		ec.Errors = append(ec.Errors, err)
	}
}

// HasErrors returns true if the collection has any errors
func (ec *ErrorCollection) HasErrors() bool {
	return len(ec.Errors) > 0
}

// Count returns the number of errors
func (ec *ErrorCollection) Count() int {
	return len(ec.Errors)
}

// First returns the first error or nil
func (ec *ErrorCollection) First() error {
	if len(ec.Errors) > 0 {
		return ec.Errors[0]
	}
	return nil
}

// ErrorToTerratagError converts any error to TerratagError
func ErrorToTerratagError(err error) *TerratagError {
	if err == nil {
		return nil
	}
	
	// If it's already a TerratagError, return it
	if terratagErr, ok := err.(*TerratagError); ok {
		return terratagErr
	}
	
	// Try to infer error code from message
	code := inferErrorCode(err.Error())
	
	return &TerratagError{
		Code:    code,
		Message: err.Error(),
		Cause:   err,
	}
}

// IsErrorCode checks if any error in the chain has the specified code
func IsErrorCode(err error, code ErrorCode) bool {
	for err != nil {
		if terratagErr, ok := err.(*TerratagError); ok {
			if terratagErr.Code == code {
				return true
			}
		}
		
		// Check wrapped errors
		if unwrapper, ok := err.(interface{ Unwrap() error }); ok {
			err = unwrapper.Unwrap()
		} else {
			break
		}
	}
	return false
}

// GetErrorCode extracts the error code from any error in the chain
func GetErrorCode(err error) ErrorCode {
	for err != nil {
		if terratagErr, ok := err.(*TerratagError); ok {
			return terratagErr.Code
		}
		
		// Check wrapped errors
		if unwrapper, ok := err.(interface{ Unwrap() error }); ok {
			err = unwrapper.Unwrap()
		} else {
			break
		}
	}
	return ErrCodeInternalError
}

// Helper functions

func getStackTrace() []string {
	var stack []string
	for i := 2; i < 10; i++ { // Skip first 2 frames (this function and caller)
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		
		// Only include frames from our package
		if strings.Contains(file, "terratag") {
			stack = append(stack, fmt.Sprintf("%s:%d %s", file, line, fn.Name()))
		}
	}
	return stack
}

func inferErrorCode(message string) ErrorCode {
	lowerMsg := strings.ToLower(message)
	
	switch {
	case strings.Contains(lowerMsg, "file not found") || strings.Contains(lowerMsg, "no such file"):
		return ErrCodeFileNotFound
	case strings.Contains(lowerMsg, "permission denied") || strings.Contains(lowerMsg, "access denied"):
		return ErrCodeFileAccessDenied
	case strings.Contains(lowerMsg, "parse") || strings.Contains(lowerMsg, "syntax"):
		return ErrCodeHCLParseError
	case strings.Contains(lowerMsg, "terraform") && strings.Contains(lowerMsg, "not found"):
		return ErrCodeTerraformNotFound
	case strings.Contains(lowerMsg, "validation") || strings.Contains(lowerMsg, "invalid"):
		return ErrCodeValidationFailed
	case strings.Contains(lowerMsg, "database") || strings.Contains(lowerMsg, "sql"):
		return ErrCodeDatabaseConnection
	case strings.Contains(lowerMsg, "timeout"):
		return ErrCodeTimeoutError
	case strings.Contains(lowerMsg, "cancel"):
		return ErrCodeCancellationError
	case strings.Contains(lowerMsg, "unauthorized"):
		return ErrCodeUnauthorized
	case strings.Contains(lowerMsg, "forbidden"):
		return ErrCodeForbidden
	default:
		return ErrCodeInternalError
	}
}

// Common error constructors

func FileNotFound(filePath string) *TerratagError {
	return New(ErrCodeFileNotFound, "file not found").
		WithContext("file_path", filePath).
		WithFile(filePath, 0)
}

func FileAccessDenied(filePath string, cause error) *TerratagError {
	return Wrap(cause, ErrCodeFileAccessDenied, "file access denied").
		WithContext("file_path", filePath).
		WithFile(filePath, 0)
}

func ConfigurationInvalid(message string) *TerratagError {
	return New(ErrCodeConfigInvalid, message).
		WithComponent("configuration")
}

func TagStandardInvalid(filePath string, cause error) *TerratagError {
	return Wrap(cause, ErrCodeTagStandardInvalid, "tag standard file is invalid").
		WithContext("file_path", filePath).
		WithComponent("tag_standard")
}

func HCLParseError(filePath string, lineNumber int, cause error) *TerratagError {
	return Wrap(cause, ErrCodeHCLParseError, "HCL parse error").
		WithFile(filePath, lineNumber).
		WithComponent("hcl_parser")
}

func ValidationFailed(message string) *TerratagError {
	return New(ErrCodeValidationFailed, message).
		WithComponent("validator")
}

func DatabaseError(operation string, cause error) *TerratagError {
	return Wrap(cause, ErrCodeDatabaseConnection, "database operation failed").
		WithContext("operation", operation).
		WithComponent("database")
}

func Unauthorized(message string) *TerratagError {
	return New(ErrCodeUnauthorized, message).
		WithComponent("auth")
}

func InternalError(message string, cause error) *TerratagError {
	return Wrap(cause, ErrCodeInternalError, message).
		WithStackTrace()
}