package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ErrorHandler provides centralized error handling
type ErrorHandler struct {
	logger *logrus.Logger
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *logrus.Logger) *ErrorHandler {
	if logger == nil {
		logger = logrus.New()
	}
	return &ErrorHandler{logger: logger}
}

// HandleError logs and processes errors appropriately
func (eh *ErrorHandler) HandleError(err error, context string) {
	if err == nil {
		return
	}

	terratagErr := ErrorToTerratagError(err)
	
	// Log with appropriate level
	logFields := logrus.Fields{
		"error_code": terratagErr.Code,
		"context":    context,
		"component":  terratagErr.Component,
	}
	
	// Add additional context
	for k, v := range terratagErr.Context {
		logFields[k] = v
	}
	
	if terratagErr.FilePath != "" {
		logFields["file_path"] = terratagErr.FilePath
		if terratagErr.LineNumber > 0 {
			logFields["line_number"] = terratagErr.LineNumber
		}
	}

	// Choose log level based on error severity
	switch terratagErr.Code {
	case ErrCodeInternalError, ErrCodeDatabaseConnection, ErrCodeDatabaseMigration:
		eh.logger.WithFields(logFields).Error(terratagErr.Message)
	case ErrCodeFileNotFound, ErrCodeConfigMissing, ErrCodeTerraformNotFound:
		eh.logger.WithFields(logFields).Warn(terratagErr.Message)
	case ErrCodeUnauthorized, ErrCodeForbidden, ErrCodeRateLimited:
		eh.logger.WithFields(logFields).Info(terratagErr.Message)
	default:
		eh.logger.WithFields(logFields).Error(terratagErr.Message)
	}
	
	// Log stack trace for internal errors
	if terratagErr.Code == ErrCodeInternalError && len(terratagErr.StackTrace) > 0 {
		eh.logger.WithField("stack_trace", terratagErr.StackTrace).Debug("Stack trace")
	}
}

// HandleHTTPError handles errors in HTTP context
func (eh *ErrorHandler) HandleHTTPError(c *gin.Context, err error, context string) {
	if err == nil {
		return
	}

	terratagErr := ErrorToTerratagError(err)
	
	// Log the error
	eh.HandleError(terratagErr, context)
	
	// Determine HTTP status code
	statusCode := eh.getHTTPStatusCode(terratagErr.Code)
	
	// Create error response
	response := gin.H{
		"error": gin.H{
			"code":    terratagErr.Code,
			"message": terratagErr.Message,
		},
	}
	
	// Add details for client errors (4xx)
	if statusCode >= 400 && statusCode < 500 {
		if terratagErr.Details != "" {
			response["error"].(gin.H)["details"] = terratagErr.Details
		}
		
		// Add context for validation errors
		if terratagErr.Code == ErrCodeValidationFailed && len(terratagErr.Context) > 0 {
			response["error"].(gin.H)["context"] = terratagErr.Context
		}
	}
	
	// Don't expose internal details for server errors
	if statusCode >= 500 {
		response["error"].(gin.H)["message"] = "Internal server error"
		response["error"].(gin.H)["details"] = "An unexpected error occurred. Please try again later."
	}
	
	c.JSON(statusCode, response)
}

// HandleCLIError handles errors in CLI context
func (eh *ErrorHandler) HandleCLIError(err error, context string) {
	if err == nil {
		return
	}

	terratagErr := ErrorToTerratagError(err)
	
	// Log the error (this will go to stderr)
	eh.HandleError(terratagErr, context)
	
	// Print user-friendly message to stderr
	fmt.Fprintf(os.Stderr, "Error: %s\n", terratagErr.Message)
	
	if terratagErr.Details != "" {
		fmt.Fprintf(os.Stderr, "Details: %s\n", terratagErr.Details)
	}
	
	// Print file location if available
	if terratagErr.FilePath != "" {
		if terratagErr.LineNumber > 0 {
			fmt.Fprintf(os.Stderr, "Location: %s:%d\n", terratagErr.FilePath, terratagErr.LineNumber)
		} else {
			fmt.Fprintf(os.Stderr, "File: %s\n", terratagErr.FilePath)
		}
	}
	
	// Print suggestions based on error type
	suggestion := eh.getErrorSuggestion(terratagErr.Code)
	if suggestion != "" {
		fmt.Fprintf(os.Stderr, "Suggestion: %s\n", suggestion)
	}
}

// RecoverFromPanic recovers from panics and converts them to errors
func (eh *ErrorHandler) RecoverFromPanic() {
	if r := recover(); r != nil {
		var err error
		switch x := r.(type) {
		case string:
			err = InternalError(x, nil).WithStackTrace()
		case error:
			err = InternalError("panic occurred", x).WithStackTrace()
		default:
			err = InternalError(fmt.Sprintf("unknown panic: %v", x), nil).WithStackTrace()
		}
		
		eh.HandleError(err, "panic_recovery")
	}
}

// GinErrorHandler is a Gin middleware for error handling
func (eh *ErrorHandler) GinErrorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		var err error
		switch x := recovered.(type) {
		case string:
			err = InternalError(x, nil)
		case error:
			err = x
		default:
			err = InternalError(fmt.Sprintf("unknown error: %v", x), nil)
		}
		
		eh.HandleHTTPError(c, err, "gin_recovery")
		c.Abort()
	})
}

// getHTTPStatusCode maps error codes to HTTP status codes
func (eh *ErrorHandler) getHTTPStatusCode(code ErrorCode) int {
	switch code {
	case ErrCodeFileNotFound, ErrCodeConfigMissing:
		return http.StatusNotFound
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeInvalidRequest, ErrCodeConfigInvalid, ErrCodeTagStandardInvalid, 
		 ErrCodeValidationFailed, ErrCodeHCLParseError:
		return http.StatusBadRequest
	case ErrCodeRateLimited:
		return http.StatusTooManyRequests
	case ErrCodeTimeoutError:
		return http.StatusRequestTimeout
	case ErrCodeInternalError, ErrCodeDatabaseConnection, ErrCodeDatabaseMigration,
		 ErrCodeDatabaseQuery, ErrCodeSchemaLoadError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// getErrorSuggestion provides user-friendly suggestions based on error type
func (eh *ErrorHandler) getErrorSuggestion(code ErrorCode) string {
	switch code {
	case ErrCodeFileNotFound:
		return "Check that the file path is correct and the file exists"
	case ErrCodeFileAccessDenied:
		return "Check file permissions or run with appropriate privileges"
	case ErrCodeConfigMissing:
		return "Ensure all required configuration parameters are provided"
	case ErrCodeTagStandardInvalid:
		return "Validate your tag standard YAML file syntax and structure"
	case ErrCodeTerraformNotFound:
		return "Install Terraform/OpenTofu or ensure it's in your PATH"
	case ErrCodeTerraformInitError:
		return "Run 'terraform init' in your project directory"
	case ErrCodeHCLParseError:
		return "Check your Terraform file syntax for errors"
	case ErrCodeValidationFailed:
		return "Review the validation report for specific issues to fix"
	case ErrCodeDatabaseConnection:
		return "Check database connectivity and permissions"
	case ErrCodeUnauthorized:
		return "Provide valid authentication credentials"
	case ErrCodeRateLimited:
		return "Reduce request frequency or wait before retrying"
	default:
		return ""
	}
}

// ErrorSummary provides a summary of multiple errors
type ErrorSummary struct {
	TotalErrors   int                    `json:"total_errors"`
	ErrorsByCode  map[ErrorCode]int      `json:"errors_by_code"`
	ErrorsByFile  map[string]int         `json:"errors_by_file,omitempty"`
	FirstError    *TerratagError         `json:"first_error,omitempty"`
	SampleErrors  []*TerratagError       `json:"sample_errors,omitempty"`
}

// SummarizeErrors creates a summary of multiple errors
func SummarizeErrors(errors []error) *ErrorSummary {
	summary := &ErrorSummary{
		TotalErrors:  len(errors),
		ErrorsByCode: make(map[ErrorCode]int),
		ErrorsByFile: make(map[string]int),
	}

	if len(errors) == 0 {
		return summary
	}

	// Convert all errors to TerratagError and analyze
	terratagErrors := make([]*TerratagError, len(errors))
	for i, err := range errors {
		terratagErrors[i] = ErrorToTerratagError(err)
		
		// Count by error code
		summary.ErrorsByCode[terratagErrors[i].Code]++
		
		// Count by file
		if terratagErrors[i].FilePath != "" {
			summary.ErrorsByFile[terratagErrors[i].FilePath]++
		}
	}

	// Set first error
	summary.FirstError = terratagErrors[0]
	
	// Sample up to 5 errors
	sampleSize := len(terratagErrors)
	if sampleSize > 5 {
		sampleSize = 5
	}
	summary.SampleErrors = terratagErrors[:sampleSize]

	return summary
}

// PrintErrorSummary prints a formatted error summary
func PrintErrorSummary(summary *ErrorSummary) {
	if summary.TotalErrors == 0 {
		fmt.Println("No errors found.")
		return
	}

	fmt.Printf("Error Summary: %d total errors\n", summary.TotalErrors)
	fmt.Println(strings.Repeat("-", 40))

	// Print errors by code
	fmt.Println("Errors by type:")
	for code, count := range summary.ErrorsByCode {
		fmt.Printf("  %s: %d\n", code, count)
	}

	// Print errors by file if any
	if len(summary.ErrorsByFile) > 0 {
		fmt.Println("\nErrors by file:")
		for file, count := range summary.ErrorsByFile {
			fmt.Printf("  %s: %d\n", file, count)
		}
	}

	// Print sample errors
	if len(summary.SampleErrors) > 0 {
		fmt.Println("\nSample errors:")
		for i, err := range summary.SampleErrors {
			fmt.Printf("  %d. [%s] %s", i+1, err.Code, err.Message)
			if err.FilePath != "" {
				fmt.Printf(" (%s", err.FilePath)
				if err.LineNumber > 0 {
					fmt.Printf(":%d", err.LineNumber)
				}
				fmt.Printf(")")
			}
			fmt.Println()
		}
	}
}

// JSONErrorSummary returns a JSON representation of the error summary
func JSONErrorSummary(summary *ErrorSummary) (string, error) {
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}