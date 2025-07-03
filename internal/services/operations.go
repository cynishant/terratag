package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/sirupsen/logrus"
	"github.com/cloudyali/terratag/cli"
	"github.com/cloudyali/terratag/internal/db"
	"github.com/cloudyali/terratag/internal/file"
	"github.com/cloudyali/terratag/internal/models"
	"github.com/cloudyali/terratag/internal/providers"
	terratag "github.com/cloudyali/terratag"
	"github.com/cloudyali/terratag/internal/validation"
	"github.com/cloudyali/terratag/internal/standards"
	"github.com/cloudyali/terratag/internal/terraform"
	hclutil "github.com/cloudyali/terratag/internal/hcl"
)

type OperationsService struct {
	db               *DatabaseService
	tagStandardsService *TagStandardsService
}

// blockPos holds position information for a resource block
type blockPos struct {
	LineNumber int
	Snippet    string
}

func NewOperationsService(db *DatabaseService, tagStandardsService *TagStandardsService) *OperationsService {
	return &OperationsService{
		db: db,
		tagStandardsService: tagStandardsService,
	}
}

func (s *OperationsService) Create(ctx context.Context, req models.CreateOperationRequest) (*models.OperationResponse, error) {
	operation, err := s.db.Queries.CreateOperation(ctx, db.CreateOperationParams{
		Type:          req.Type,
		Status:        "pending",
		StandardID:    sql.NullInt64{Int64: req.StandardID, Valid: req.StandardID != 0},
		DirectoryPath: req.DirectoryPath,
		FilterPattern: sql.NullString{String: req.FilterPattern, Valid: req.FilterPattern != ""},
		SkipPattern:   sql.NullString{String: req.SkipPattern, Valid: req.SkipPattern != ""},
		Settings:      sql.NullString{String: req.Settings, Valid: req.Settings != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create operation: %w", err)
	}

	response := models.OperationFromDB(operation)
	return &response, nil
}

func (s *OperationsService) GetByID(ctx context.Context, id int64) (*models.OperationResponse, error) {
	operation, err := s.db.Queries.GetOperation(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("operation not found")
		}
		return nil, fmt.Errorf("failed to get operation: %w", err)
	}

	response := models.OperationFromDB(operation)
	return &response, nil
}

func (s *OperationsService) List(ctx context.Context, pagination models.PaginationRequest) ([]models.OperationResponse, error) {
	if pagination.Limit == 0 {
		pagination.Limit = 50
	}
	if pagination.Page == 0 {
		pagination.Page = 1
	}

	offset := (pagination.Page - 1) * pagination.Limit

	operations, err := s.db.Queries.ListOperations(ctx, db.ListOperationsParams{
		Limit:  pagination.Limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list operations: %w", err)
	}

	var response []models.OperationResponse
	for _, operation := range operations {
		response = append(response, models.OperationFromDB(operation))
	}

	return response, nil
}

func (s *OperationsService) GetSummary(ctx context.Context, id int64) (*models.OperationSummaryResponse, error) {
	// Get operation
	operation, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get results
	dbResults, err := s.db.Queries.GetOperationResults(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get operation results: %w", err)
	}

	var results []models.OperationResultResponse
	for _, result := range dbResults {
		results = append(results, models.OperationResultFromDB(result))
	}

	// Get logs
	dbLogs, err := s.db.Queries.GetOperationLogs(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get operation logs: %w", err)
	}

	var logs []models.OperationLogResponse
	for _, logEntry := range dbLogs {
		logs = append(logs, models.OperationLogFromDB(logEntry))
	}

	// Calculate summary
	summary := s.calculateStats(results)

	return &models.OperationSummaryResponse{
		Operation: *operation,
		Results:   results,
		Logs:      logs,
		Summary:   summary,
	}, nil
}

func (s *OperationsService) Execute(ctx context.Context, id int64) error {
	logger := logrus.WithFields(logrus.Fields{
		"component":   "operations",
		"action":      "execute",
		"operationId": id,
	})
	
	logger.Info("Starting execution for operation")
	
	// Get operation
	operation, err := s.GetByID(ctx, id)
	if err != nil {
		logger.WithError(err).Error("Failed to get operation")
		return err
	}
	
	logger.WithFields(logrus.Fields{
		"type":   operation.Type,
		"status": operation.Status,
	}).Info("Operation retrieved successfully")

	if operation.Status != "pending" {
		logger.WithField("status", operation.Status).Warn("Operation not in pending state")
		return fmt.Errorf("operation is not in pending state")
	}

	// Update status to running
	logger.Info("Updating operation status to running")
	_, err = s.db.Queries.UpdateOperationStarted(ctx, db.UpdateOperationStartedParams{
		Status: "running",
		ID:     id,
	})
	if err != nil {
		logger.WithError(err).Error("Failed to update operation status")
		return fmt.Errorf("failed to update operation status: %w", err)
	}
	logger.Info("Operation status updated to running")

	// Execute in goroutine with background context to avoid cancellation
	logger.Info("Starting background execution goroutine")
	go s.executeOperation(context.Background(), id, *operation)

	logger.Info("Execute method completed")
	return nil
}

func (s *OperationsService) executeOperation(ctx context.Context, operationID int64, operation models.OperationResponse) {
	var finalStatus = "completed"
	
	logger := logrus.WithFields(logrus.Fields{
		"component":   "operations",
		"action":      "executeOperation", 
		"operationId": operationID,
		"type":        operation.Type,
	})
	
	defer func() {
		if r := recover(); r != nil {
			finalStatus = "failed"
			logger.WithField("panic", r).Error("Operation panicked")
			s.logOperation(ctx, operationID, "error", fmt.Sprintf("Operation panicked: %v", r), nil)
		}
		
		// Update final status
		s.db.Queries.UpdateOperationCompleted(ctx, db.UpdateOperationCompletedParams{
			Status: finalStatus,
			ID:     operationID,
		})
		
		logger.WithField("finalStatus", finalStatus).Info("Operation execution finished")
	}()

	logger.Info("Starting operation execution")
	s.logOperation(ctx, operationID, "info", "Starting operation execution", nil)

	// Get tag standard if provided
	var standardContent string
	if operation.StandardID != nil {
		logger.WithField("standardId", *operation.StandardID).Info("Fetching tag standard")
		standard, err := s.tagStandardsService.GetByID(ctx, *operation.StandardID)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"standardId": *operation.StandardID,
				"error":      err.Error(),
			}).Error("Failed to fetch tag standard")
			s.logOperation(ctx, operationID, "error", "Failed to get tag standard", map[string]interface{}{"error": err.Error()})
			finalStatus = "failed"
			return
		}
		standardContent = standard.Content
		logger.WithFields(logrus.Fields{
			"standardId":    *operation.StandardID,
			"name":          standard.Name,
			"contentLength": len(standardContent),
		}).Info("Tag standard retrieved successfully")
	} else {
		logger.Info("No tag standard specified")
	}

	// Create CLI args
	logger.WithFields(logrus.Fields{
		"directory": operation.DirectoryPath,
		"type":      operation.Type,
	}).Info("Creating CLI args")
	
	args := cli.Args{
		Dir:           operation.DirectoryPath,
		Type:          "terraform", // Default to terraform
		ValidateOnly:  operation.Type == "validation",
		ReportFormat:  "json", // Default report format
	}
	
	logger.WithFields(logrus.Fields{
		"validateOnly":  args.ValidateOnly,
		"reportFormat":  args.ReportFormat,
		"directory":     args.Dir,
		"iacType":       args.Type,
	}).Info("Base CLI args created")

	if operation.FilterPattern != nil {
		args.Filter = *operation.FilterPattern
	} else {
		args.Filter = ".*"
	}

	if operation.SkipPattern != nil {
		args.Skip = *operation.SkipPattern
	}

	// Parse settings if provided
	if operation.Settings != nil && *operation.Settings != "" {
		var settings map[string]interface{}
		if err := json.Unmarshal([]byte(*operation.Settings), &settings); err != nil {
			s.logOperation(ctx, operationID, "error", "Failed to parse settings", map[string]interface{}{"error": err.Error()})
			finalStatus = "failed"
			return
		}

		// Apply settings to args
		if iacType, ok := settings["type"].(string); ok {
			args.Type = iacType
		}
		if verbose, ok := settings["verbose"].(bool); ok {
			args.Verbose = verbose
		}
		if rename, ok := settings["rename"].(bool); ok {
			args.Rename = rename
		}
		if reportFormat, ok := settings["report_format"].(string); ok {
			args.ReportFormat = reportFormat
		}
	}

	// Execute based on operation type
	logger.WithField("operationType", operation.Type).Info("Executing operation")
	switch operation.Type {
	case "validation":
		logger.Info("Starting validation execution")
		if standardContent != "" {
			// Write standard to temporary file
			standardFile := fmt.Sprintf("/tmp/standard_%d.yaml", operationID)
			logger.WithField("standardFile", standardFile).Info("Writing standard to temp file")
			if err := s.writeToTempFile(standardFile, standardContent); err != nil {
				logger.WithFields(logrus.Fields{
					"standardFile": standardFile,
					"error":        err.Error(),
				}).Error("Failed to write temp file")
				s.logOperation(ctx, operationID, "error", "Failed to write standard file", map[string]interface{}{"error": err.Error()})
				finalStatus = "failed"
				return
			}
			args.StandardFile = standardFile
			logger.WithField("standardFile", standardFile).Info("Standard file written successfully")
			// Clean up temp file after operation
			defer os.Remove(standardFile)
		}
		logger.Info("Calling validation executor")
		err := s.executeValidation(ctx, operationID, args)
		if err != nil {
			logger.WithError(err).Error("Validation execution failed")
			s.logOperation(ctx, operationID, "error", "Validation failed", map[string]interface{}{"error": err.Error()})
			finalStatus = "failed"
		} else {
			logger.Info("Validation execution completed successfully")
		}
	case "tagging":
		logger.Info("Starting tagging execution")
		if standardContent != "" {
			// Write standard to temporary file
			tagsFile := fmt.Sprintf("/tmp/tags_%d.yaml", operationID)
			logger.WithField("tagsFile", tagsFile).Info("Writing tags to temp file")
			if err := s.writeToTempFile(tagsFile, standardContent); err != nil {
				logger.WithFields(logrus.Fields{
					"tagsFile": tagsFile,
					"error":    err.Error(),
				}).Error("Failed to write tags file")
				s.logOperation(ctx, operationID, "error", "Failed to write tags file", map[string]interface{}{"error": err.Error()})
				finalStatus = "failed"
				return
			}
			args.TagsFile = tagsFile
			logger.WithField("tagsFile", tagsFile).Info("Tags file written successfully")
			// Clean up temp file after operation
			defer os.Remove(tagsFile)
		}
		err := s.executeTagging(ctx, operationID, args)
		if err != nil {
			logger.WithError(err).Error("Tagging execution failed")
			s.logOperation(ctx, operationID, "error", "Tagging failed", map[string]interface{}{"error": err.Error()})
			finalStatus = "failed"
		} else {
			logger.Info("Tagging execution completed successfully")
		}
	default:
		logger.WithField("operationType", operation.Type).Error("Unknown operation type")
		s.logOperation(ctx, operationID, "error", "Unknown operation type", map[string]interface{}{"type": operation.Type})
		finalStatus = "failed"
	}

	logger.WithField("status", finalStatus).Info("Operation execution completed")
	s.logOperation(ctx, operationID, "info", "Operation execution completed", map[string]interface{}{"status": finalStatus})
}

func (s *OperationsService) executeValidation(ctx context.Context, operationID int64, args cli.Args) error {
	logger := logrus.WithFields(logrus.Fields{
		"component":   "operations",
		"action":      "executeValidation",
		"operationId": operationID,
	})
	
	logger.Info("Starting validation")
	s.logOperation(ctx, operationID, "info", "Starting validation", nil)
	
	// Check if terraform init is required (but don't auto-init)
	logger.Info("Checking terraform initialization status")
	if !s.isTerraformInitialized(args.Dir, args.Type) {
		logger.Warn("Terraform directory not initialized")
		s.logOperation(ctx, operationID, "warning", "Terraform directory not initialized", map[string]interface{}{
			"directory": args.Dir,
			"type": args.Type,
			"suggestion": "Run 'terraform init' in the directory before running validation",
		})
		// Continue anyway - let terraform/terratag handle the error
	} else {
		logger.Info("Terraform directory is properly initialized")
		s.logOperation(ctx, operationID, "info", "Terraform directory is properly initialized", nil)
	}
	
	// Capture logs from the validation process
	logger.WithFields(logrus.Fields{
		"standardFile": args.StandardFile,
		"directory":    args.Dir,
		"filter":       args.Filter,
		"skip":         args.Skip,
	}).Info("Calling core Terratag validation engine")
	
	s.logOperation(ctx, operationID, "info", "Calling core Terratag validation engine", map[string]interface{}{
		"standardFile": args.StandardFile,
		"directory": args.Dir,
		"filter": args.Filter,
		"skip": args.Skip,
	})
	
	// Store original log flags and prefix to restore later
	originalFlags := log.Flags()
	originalPrefix := log.Prefix()
	
	// Set log prefix to identify validation logs
	log.SetPrefix(fmt.Sprintf("[TERRATAG-VALIDATION-OP-%d] ", operationID))
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	
	// Restore original log settings after validation
	defer func() {
		log.SetFlags(originalFlags)
		log.SetPrefix(originalPrefix)
	}()
	
	// Log validation start details using both logrus and standard log for core Terratag
	logger.WithFields(logrus.Fields{
		"standardFile": args.StandardFile,
		"directory":    args.Dir,
		"filter":       args.Filter,
		"skip":         args.Skip,
	}).Info("Starting core Terratag validation process")
	
	log.Printf("Starting Terratag validation process with standard: %s", args.StandardFile)
	log.Printf("Validation directory: %s", args.Dir)
	log.Printf("Filter pattern: %s", args.Filter)
	log.Printf("Skip pattern: %s", args.Skip)
	
	err := validation.ValidateStandards(args)
	if err != nil {
		log.Printf("Terratag validation failed: %v", err)
		logger.WithError(err).Error("Core validation engine failed")
		s.logOperation(ctx, operationID, "error", "Core validation engine failed", map[string]interface{}{"error": err.Error()})
		return err
	}

	log.Printf("Terratag validation completed successfully")
	logger.Info("Core validation engine completed successfully")
	s.logOperation(ctx, operationID, "info", "Core validation engine completed successfully", nil)

	// Parse and store validation results
	logger.Info("Parsing and storing validation results")
	s.logOperation(ctx, operationID, "info", "Parsing and storing validation results", nil)
	if err := s.parseAndStoreValidationResults(ctx, operationID, args); err != nil {
		logger.WithError(err).Warn("Failed to parse validation results")
		s.logOperation(ctx, operationID, "warning", "Failed to parse validation results", map[string]interface{}{"error": err.Error()})
		// Don't fail the operation if result parsing fails
	}
	
	logger.Info("Validation completed successfully")
	s.logOperation(ctx, operationID, "info", "Validation completed successfully", nil)
	return nil
}

// extractAllVariablesInfo extracts information about all variables and locals in the codebase
func (s *OperationsService) extractAllVariablesInfo(resolver *terraform.VariableResolver) map[string]interface{} {
	allVars := make(map[string]interface{})
	
	// Extract variables
	variables := make(map[string]interface{})
	for name, varDef := range resolver.GetVariables() {
		varInfo := map[string]interface{}{
			"name":        name,
			"type":        varDef.Type,
			"description": varDef.Description,
			"default":     varDef.Default,
			"file_path":   varDef.FilePath,
			"line_number": varDef.LineNumber,
			"resolved":    false,
		}
		
		// Check if variable has a resolved value
		if values := resolver.GetVariableValues(); values != nil {
			if value, exists := values[name]; exists {
				varInfo["value"] = value
				varInfo["resolved"] = true
			} else if varDef.Default != nil {
				// Use default value if no explicit value
				varInfo["value"] = varDef.Default
				varInfo["resolved"] = true
			}
		}
		
		variables[name] = varInfo
	}
	
	// Extract locals
	locals := make(map[string]interface{})
	for name, localDef := range resolver.GetLocals() {
		localInfo := map[string]interface{}{
			"name":        name,
			"expression":  localDef.Expression,
			"file_path":   localDef.FilePath,
			"line_number": localDef.LineNumber,
			"resolved":    false,
		}
		
		// Check if local has a resolved value
		if resolvedLocals := resolver.GetResolvedLocals(); resolvedLocals != nil {
			if value, exists := resolvedLocals[name]; exists {
				localInfo["value"] = value
				localInfo["resolved"] = true
			}
		}
		
		// If not resolved but has a value, use it
		if !localInfo["resolved"].(bool) && localDef.Value != nil {
			localInfo["value"] = localDef.Value
			localInfo["resolved"] = true
		}
		
		locals[name] = localInfo
	}
	
	allVars["variables"] = variables
	allVars["locals"] = locals
	
	return allVars
}

func (s *OperationsService) executeTagging(ctx context.Context, operationID int64, args cli.Args) error {
	logger := logrus.WithFields(logrus.Fields{
		"component":   "operations",
		"action":      "executeTagging",
		"operationId": operationID,
	})
	
	logger.Info("Starting tagging")
	s.logOperation(ctx, operationID, "info", "Starting tagging", nil)
	
	// Capture logs from the tagging process
	logger.WithFields(logrus.Fields{
		"tagsFile":  args.TagsFile,
		"directory": args.Dir,
		"filter":    args.Filter,
		"skip":      args.Skip,
	}).Info("Calling core Terratag engine")
	
	s.logOperation(ctx, operationID, "info", "Calling core Terratag engine", map[string]interface{}{
		"tagsFile": args.TagsFile,
		"directory": args.Dir,
		"filter": args.Filter,
		"skip": args.Skip,
	})
	
	// Store original log flags and prefix to restore later
	originalFlags := log.Flags()
	originalPrefix := log.Prefix()
	
	// Set log prefix to identify tagging logs
	log.SetPrefix(fmt.Sprintf("[TERRATAG-TAGGING-OP-%d] ", operationID))
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	
	// Restore original log settings after tagging
	defer func() {
		log.SetFlags(originalFlags)
		log.SetPrefix(originalPrefix)
	}()
	
	// Log tagging start details using both logrus and standard log for core Terratag
	logger.WithFields(logrus.Fields{
		"tagsFile":  args.TagsFile,
		"directory": args.Dir,
		"filter":    args.Filter,
		"skip":      args.Skip,
	}).Info("Starting core Terratag tagging process")
	
	log.Printf("Starting Terratag tagging process with tags file: %s", args.TagsFile)
	log.Printf("Tagging directory: %s", args.Dir)
	log.Printf("Filter pattern: %s", args.Filter)
	log.Printf("Skip pattern: %s", args.Skip)
	
	err := terratag.Terratag(args)
	if err != nil {
		log.Printf("Terratag tagging failed: %v", err)
		logger.WithError(err).Error("Core tagging engine failed")
		s.logOperation(ctx, operationID, "error", "Core tagging engine failed", map[string]interface{}{"error": err.Error()})
		return err
	}

	log.Printf("Terratag tagging completed successfully")
	logger.Info("Core tagging engine completed successfully")
	s.logOperation(ctx, operationID, "info", "Core tagging engine completed successfully", nil)

	// Parse and store tagging results
	logger.Info("Parsing and storing tagging results")
	s.logOperation(ctx, operationID, "info", "Parsing and storing tagging results", nil)
	if err := s.parseAndStoreTaggingResults(ctx, operationID, args); err != nil {
		logger.WithError(err).Warn("Failed to parse tagging results")
		s.logOperation(ctx, operationID, "warning", "Failed to parse tagging results", map[string]interface{}{"error": err.Error()})
		// Don't fail the operation if result parsing fails
	}
	
	logger.Info("Tagging completed successfully")
	s.logOperation(ctx, operationID, "info", "Tagging completed successfully", nil)
	return nil
}

func (s *OperationsService) logOperation(ctx context.Context, operationID int64, level, message string, details map[string]interface{}) {
	// Use a background context for logging to prevent context cancellation issues
	logCtx := context.Background()
	
	// Create logger with operation context
	logger := logrus.WithFields(logrus.Fields{
		"component":   "operations",
		"action":      "logOperation",
		"operationId": operationID,
		"level":       level,
	})
	
	// Add details to logger fields if provided
	if details != nil {
		for k, v := range details {
			logger = logger.WithField(k, v)
		}
	}
	
	// Log using logrus based on level
	switch level {
	case "error":
		logger.Error(message)
	case "warning", "warn":
		logger.Warn(message)
	case "info":
		logger.Info(message)
	case "debug":
		logger.Debug(message)
	default:
		logger.Info(message)
	}
	
	var detailsJSON string
	if details != nil {
		if data, err := json.Marshal(details); err == nil {
			detailsJSON = string(data)
		}
	}

	// Store in database for operation history
	_, err := s.db.Queries.CreateOperationLog(logCtx, db.CreateOperationLogParams{
		OperationID: operationID,
		Level:       level,
		Message:     message,
		Details:     sql.NullString{String: detailsJSON, Valid: detailsJSON != ""},
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component":   "operations",
			"operationId": operationID,
			"error":       err.Error(),
		}).Error("Failed to store operation log in database")
	}
}

func (s *OperationsService) calculateStats(results []models.OperationResultResponse) models.OperationStats {
	stats := models.OperationStats{}
	
	fileSet := make(map[string]bool)
	for _, result := range results {
		fileSet[result.FilePath] = true
		
		switch result.Action {
		case "tagged":
			stats.TaggedResources++
		case "violation":
			stats.Violations++
		case "error":
			stats.Errors++
		}
	}
	
	stats.TotalFiles = int64(len(fileSet))
	stats.ProcessedFiles = stats.TotalFiles
	
	return stats
}

func (s *OperationsService) GetResults(ctx context.Context, id int64, pagination models.PaginationRequest) ([]models.OperationResultResponse, error) {
	// Check if operation exists
	_, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if pagination.Limit == 0 {
		pagination.Limit = 50
	}
	if pagination.Page == 0 {
		pagination.Page = 1
	}

	// Get all results for the operation
	dbResults, err := s.db.Queries.GetOperationResults(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get operation results: %w", err)
	}

	// Manual pagination
	offset := (pagination.Page - 1) * pagination.Limit
	end := offset + pagination.Limit
	if end > int64(len(dbResults)) {
		end = int64(len(dbResults))
	}

	var results []models.OperationResultResponse
	if offset < int64(len(dbResults)) {
		for _, result := range dbResults[offset:end] {
			results = append(results, models.OperationResultFromDB(result))
		}
	}

	return results, nil
}

func (s *OperationsService) GetLogs(ctx context.Context, id int64, pagination models.PaginationRequest) ([]models.OperationLogResponse, error) {
	// Check if operation exists
	_, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if pagination.Limit == 0 {
		pagination.Limit = 50
	}
	if pagination.Page == 0 {
		pagination.Page = 1
	}

	// Get all logs for the operation
	dbLogs, err := s.db.Queries.GetOperationLogs(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get operation logs: %w", err)
	}

	// Manual pagination
	offset := (pagination.Page - 1) * pagination.Limit
	end := offset + pagination.Limit
	if end > int64(len(dbLogs)) {
		end = int64(len(dbLogs))
	}

	var logs []models.OperationLogResponse
	if offset < int64(len(dbLogs)) {
		for _, logEntry := range dbLogs[offset:end] {
			logs = append(logs, models.OperationLogFromDB(logEntry))
		}
	}

	return logs, nil
}

func (s *OperationsService) Retry(ctx context.Context, id int64) error {
	// Get operation
	operation, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if operation.Status != "failed" {
		return fmt.Errorf("operation is not in failed state")
	}

	// Update status to pending
	_, err = s.db.Queries.UpdateOperationStatus(ctx, db.UpdateOperationStatusParams{
		Status: "pending",
		ID:     id,
	})
	if err != nil {
		return fmt.Errorf("failed to update operation status: %w", err)
	}

	// Execute the operation
	return s.Execute(ctx, id)
}

func (s *OperationsService) Delete(ctx context.Context, id int64) error {
	// Check if exists
	_, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	err = s.db.Queries.DeleteOperation(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete operation: %w", err)
	}

	return nil
}

// Helper function to write content to temporary file
func (s *OperationsService) writeToTempFile(filePath, content string) error {
	// Ensure temp directory exists
	tempDir := filepath.Dir(filePath)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	
	// Write content to file
	if err := ioutil.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	return nil
}

// Parse and store validation results in database
func (s *OperationsService) parseAndStoreValidationResults(ctx context.Context, operationID int64, args cli.Args) error {
	logger := logrus.WithFields(logrus.Fields{
		"component":   "operations",
		"action":      "parseAndStoreValidationResults",
		"operationId": operationID,
	})
	
	logger.Info("ZZZZZ VERY START OF parseAndStoreValidationResults ZZZZZ")
	
	// Load standard if provided
	var tagStandard *standards.TagStandard
	var validator *standards.TagValidator
	
	if args.StandardFile != "" {
		logger.WithField("standardFile", args.StandardFile).Info("Loading tag standard")
		var err error
		tagStandard, err = standards.LoadStandard(args.StandardFile)
		if err != nil {
			return fmt.Errorf("failed to load standard: %w", err)
		}
		
		validator, err = standards.NewTagValidator(tagStandard)
		if err != nil {
			return fmt.Errorf("failed to create validator: %w", err)
		}
		
		// Load variables and locals from the directory for proper variable resolution
		logger.WithField("directory", args.Dir).Info("Loading variables and locals into validator")
		if err := validator.LoadVariablesFromDirectory(args.Dir); err != nil {
			logger.WithError(err).Warn("Failed to load variables and locals into validator")
			// Continue anyway - validation will show unresolvable variables
		} else {
			logger.Info("Successfully loaded variables and locals into validator")
		}
		
		logger.Info("Tag standard loaded and validator created")
	}
	
	// Get terraform files in directory
	logger.WithFields(logrus.Fields{
		"directory": args.Dir,
		"filter":    args.Filter,
		"skip":      args.Skip,
	}).Info("Getting terraform files in directory")
	
	files, err := s.getTerraformFiles(args.Dir, args.Filter, args.Skip)
	if err != nil {
		return fmt.Errorf("failed to get terraform files: %w", err)
	}
	
	logger.WithField("fileCount", len(files)).Info("Found terraform files")
	
	logger.Info("XXXXX IMMEDIATE AFTER FOUND FILES XXXXX")
	
	// Always extract all variables information for the UI, regardless of whether there's a standard
	logger.Info("About to start variable extraction logic")
	var allVariablesValidator *standards.TagValidator
	if validator != nil {
		allVariablesValidator = validator
	} else {
		// Create a minimal validator just for variable extraction
		emptyStandard := &standards.TagStandard{
			Version:      1,
			CloudProvider: "aws", // Default to aws for variable extraction
			RequiredTags: []standards.TagSpec{},
			OptionalTags: []standards.TagSpec{},
		}
		var err error
		allVariablesValidator, err = standards.NewTagValidator(emptyStandard)
		if err != nil {
			logger.WithError(err).Warn("Failed to create validator for variable extraction")
		} else {
			// Load variables and locals from the directory
			if err := allVariablesValidator.LoadVariablesFromDirectory(args.Dir); err != nil {
				logger.WithError(err).Warn("Failed to load variables and locals for extraction")
			} else {
				logger.Info("Successfully loaded variables and locals for extraction")
			}
		}
	}
	
	// Store all variables information once at operation level
	logger.WithFields(logrus.Fields{
		"allVariablesValidator_nil": allVariablesValidator == nil,
		"getVariableResolver_nil": allVariablesValidator != nil && allVariablesValidator.GetVariableResolver() == nil,
	}).Info("Checking conditions for storing operation_summary")
	
	if allVariablesValidator != nil && allVariablesValidator.GetVariableResolver() != nil {
		allVarsInfo := s.extractAllVariablesInfo(allVariablesValidator.GetVariableResolver())
		allVarsDetails := map[string]interface{}{
			"all_variables": allVarsInfo,
		}
		
		logger.Info("About to store operation_summary result")
		
		// Store as a special "summary" entry
		logger.Info("Calling storeOperationResult for operation_summary")
		s.storeOperationResult(ctx, operationID, "", "operation_summary.variables_summary", "variables_summary", "All variables and locals from codebase", allVarsDetails, 0, "")
		logger.Info("Completed call to storeOperationResult for operation_summary")
		logger.Info("Stored all variables information at operation level")
	} else {
		logger.Info("Conditions not met for storing operation_summary - skipping variable extraction")
	}

	// Use the validation system to extract resources with line numbers and snippets
	logger.Info("Using validation system to extract resources with enhanced information")
	logger.WithFields(logrus.Fields{
		"tagStandard_nil": tagStandard == nil,
		"validator_nil": validator == nil,
	}).Info("Checking conditions for resource validation")
	
	// Call validation system directly to get enhanced resource information
	if tagStandard != nil && validator != nil {
		logger.WithField("cloudProvider", tagStandard.CloudProvider).Info("Collecting resources using validation system")
		
		// Create a modified args for resource extraction
		extractArgs := args
		extractArgs.IsSkipTerratagFiles = true // Skip .terratag.tf files
		
		// Import the validation package to use its enhanced resource extraction
		allResources, err := s.collectResourcesWithEnhancedInfo(files, extractArgs, tagStandard.CloudProvider)
		if err != nil {
			logger.WithError(err).Error("Failed to collect resources with enhanced info")
			return fmt.Errorf("failed to collect resources: %w", err)
		}
		
		logger.WithField("resourceCount", len(allResources)).Info("Successfully collected resources with enhanced information")
		
		// Validate each resource and store results
		for _, resource := range allResources {
			relativePath := strings.TrimPrefix(resource.FilePath, args.Dir+"/")
			
			resourceLogger := logger.WithFields(logrus.Fields{
				"file":         relativePath,
				"resourceType": resource.Type,
				"resourceName": resource.Name,
				"lineNumber":   resource.LineNumber,
			})
			
			resourceLogger.Info("Validating resource with enhanced information")
			result := validator.ValidateResourceTags(resource.Type, resource.Name, resource.FilePath, resource.Tags)
			
			resourceLogger.WithFields(logrus.Fields{
				"compliant":      result.IsCompliant,
				"missingCount":   len(result.MissingTags),
				"violationCount": len(result.Violations),
				"extraCount":     len(result.ExtraTags),
			}).Info("MODIFIED Validation result MODIFIED")
			
			if len(result.MissingTags) > 0 {
				resourceLogger.WithField("missingTags", result.MissingTags).Warn("Missing required tags")
			}
			if len(result.Violations) > 0 {
				resourceLogger.WithField("violations", result.Violations).Warn("Tag violations")
			}
			
			action := "compliant"
			if !result.IsCompliant {
				action = "violation"
			}
			
			// Extract variable information from tags
			tagsInterface := make(map[string]interface{})
			for k, v := range resource.Tags {
				tagsInterface[k] = v
			}
			variableInfo := s.extractVariableResolutionInfo(tagsInterface)
			
			details := map[string]interface{}{
				"violations":        result.Violations,
				"compliance_status": result.IsCompliant,
				"supports_tagging":  result.SupportsTagging,
				"missing_tags":      result.MissingTags,
				"extra_tags":        result.ExtraTags,
				"variable_resolution": variableInfo,
			}
			
			message := "Resource is compliant"
			if !result.IsCompliant {
				var issues []string
				if len(result.MissingTags) > 0 {
					issues = append(issues, fmt.Sprintf("%d missing required tags", len(result.MissingTags)))
				}
					if len(result.Violations) > 0 {
						issues = append(issues, fmt.Sprintf("%d tag violations", len(result.Violations)))
					}
					if len(result.ExtraTags) > 0 {
						issues = append(issues, fmt.Sprintf("%d extra tags", len(result.ExtraTags)))
					}
					if len(issues) > 0 {
						message = fmt.Sprintf("Resource has issues: %s", strings.Join(issues, ", "))
					} else {
						message = "Resource is non-compliant (reason unknown)"
					}
				}
				
				// Enhance snippet with resolved tag values if available
				enhancedSnippet := s.enhanceSnippetWithResolvedTags(resource.Snippet, resource.Type, resource.Tags, validator)
				s.storeOperationResult(ctx, operationID, relativePath, resource.Type+"."+resource.Name, action, message, details, resource.LineNumber, enhancedSnippet)
			}
	} else {
		// No standard provided - just record files as processed
		logger.Info("No tag standard provided, recording files as processed")
		for _, file := range files {
			relativePath := strings.TrimPrefix(file, args.Dir+"/")
			s.storeOperationResult(ctx, operationID, relativePath, "", "processed", "File processed without standard", nil, 0, "")
		}
	}
	
	return nil
}

// Parse and store tagging results in database
func (s *OperationsService) parseAndStoreTaggingResults(ctx context.Context, operationID int64, args cli.Args) error {
	// Look for .terratag.tf files generated by terratag
	terratagFiles, err := filepath.Glob(filepath.Join(args.Dir, "**/*.terratag.tf"))
	if err != nil {
		return fmt.Errorf("failed to find terratag files: %w", err)
	}
	
	// Also look for backup files to see what was modified
	backupFiles, err := filepath.Glob(filepath.Join(args.Dir, "**/*.tf.bak"))
	if err != nil {
		return fmt.Errorf("failed to find backup files: %w", err)
	}
	
	// Process terratag files
	for _, file := range terratagFiles {
		relativePath := strings.TrimPrefix(file, args.Dir+"/")
		originalFile := strings.TrimSuffix(file, ".terratag.tf") + ".tf"
		
		// Count resources in the file (simplified)
		resources, err := s.extractResourcesFromFile(file)
		if err != nil {
			s.storeOperationResult(ctx, operationID, relativePath, "", "error", err.Error(), nil, 0, "")
			continue
		}
		
		for _, resource := range resources {
			details := map[string]interface{}{
				"original_file": filepath.Base(originalFile),
				"resource_type": resource.Type,
			}
			
			s.storeOperationResult(ctx, operationID, relativePath, resource.Type+"."+resource.Name, "tagged", "Resource successfully tagged", details, resource.LineNumber, resource.Snippet)
		}
	}
	
	// Process backup files to see what was modified
	for _, backupFile := range backupFiles {
		relativePath := strings.TrimPrefix(backupFile, args.Dir+"/")
		originalFile := strings.TrimSuffix(backupFile, ".bak")
		
		details := map[string]interface{}{
			"original_file": filepath.Base(originalFile),
			"backup_created": true,
		}
		
		s.storeOperationResult(ctx, operationID, relativePath, "", "backed_up", "Original file backed up", details, 0, "")
	}
	
	return nil
}

// Helper to get terraform files matching patterns
// collectResourcesWithEnhancedInfo uses the validation system to extract resources with line numbers and snippets
func (s *OperationsService) collectResourcesWithEnhancedInfo(filePaths []string, args cli.Args, cloudProvider string) ([]standards.ResourceInfo, error) {
	var allResources []standards.ResourceInfo
	
	for _, filePath := range filePaths {
		// Skip .terratag.tf files if requested
		if args.IsSkipTerratagFiles && strings.HasSuffix(filePath, ".terratag.tf") {
			continue
		}
		
		// Call the validation system's enhanced resource extraction
		// Since the extractResourcesFromFile function is not exported, we need to replicate its logic
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		// Parse with hcl for position information
		parser := hclparse.NewParser()
		hclFile, diags := parser.ParseHCL(content, filePath)
		if diags.HasErrors() {
			// Log error but continue with other files
			continue
		}

		// Parse with hclwrite for tag extraction
		hclWriteFile, err := file.ReadHCLFile(filePath)
		if err != nil {
			continue
		}

		// Create maps to correlate blocks between parsers
		blockPositions := make(map[string]blockPos) // key: "resourceType.resourceName"

		// Extract position information from hcl parser
		if hclFile.Body != nil {
			bodyContent, _, _ := hclFile.Body.PartialContent(&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}},
				},
			})

			for _, block := range bodyContent.Blocks {
				if len(block.Labels) >= 2 {
					key := block.Labels[0] + "." + block.Labels[1]
					lineNumber := block.DefRange.Start.Line
					
					// Extract snippet (block definition) - use DefRange for the entire block
					snippet := extractBlockSnippet(content, block.DefRange, block.DefRange)
					
					blockPositions[key] = blockPos{
						LineNumber: lineNumber,
						Snippet:    snippet,
					}
				}
			}
		}

		for _, block := range hclWriteFile.Body().Blocks() {
			if block.Type() != "resource" {
				continue
			}

			if len(block.Labels()) < 2 {
				continue
			}

			resourceType := block.Labels()[0]
			resourceName := block.Labels()[1]

			// Apply filter if specified
			if args.Filter != "" {
				matched, err := regexp.MatchString(args.Filter, resourceType)
				if err != nil {
					return nil, fmt.Errorf("invalid filter regex: %w", err)
				}
				if !matched {
					continue
				}
			}

			// Apply skip filter if specified
			if args.Skip != "" {
				matched, err := regexp.MatchString(args.Skip, resourceType)
				if err != nil {
					return nil, fmt.Errorf("invalid skip regex: %w", err)
				}
				if matched {
					continue
				}
			}

			// Check if resource supports tagging
			if !standards.IsTaggableResource(resourceType, cloudProvider) {
				continue
			}

			// Extract existing tags
			tags, err := extractTagsFromResource(block, resourceType)
			if err != nil {
				// Only continue with empty tags for complex expressions, fail for actual errors
				if strings.Contains(err.Error(), "complex expression") {
					tags = make(map[string]string)
				} else {
					return nil, fmt.Errorf("failed to extract tags from %s.%s: %w", resourceType, resourceName, err)
				}
			}

			// Get position information for this resource
			resourceKey := resourceType + "." + resourceName
			pos := blockPositions[resourceKey]
			
			allResources = append(allResources, standards.ResourceInfo{
				Type:       resourceType,
				Name:       resourceName,
				FilePath:   filePath,
				Tags:       tags,
				LineNumber: pos.LineNumber,
				Snippet:    pos.Snippet,
			})
		}
	}
	
	return allResources, nil
}

// extractBlockSnippet extracts the complete resource block from HCL content
func extractBlockSnippet(content []byte, defRange hcl.Range, bodyRange hcl.Range) string {
	lines := strings.Split(string(content), "\n")
	
	// Extract lines from start to end (1-based to 0-based conversion)
	startIdx := defRange.Start.Line - 1
	if startIdx < 0 {
		startIdx = 0
	}
	
	// Find the actual end of the resource block by looking for the closing brace
	endIdx := findResourceBlockEnd(lines, startIdx)
	if endIdx == -1 {
		// Fallback to bodyRange.End.Line if we can't find the closing brace
		endIdx = bodyRange.End.Line
		if endIdx > len(lines) {
			endIdx = len(lines)
		}
	}
	
	// Ensure we don't go beyond the file
	if endIdx > len(lines) {
		endIdx = len(lines)
	}
	
	// Extract the complete resource block
	snippet := strings.Join(lines[startIdx:endIdx], "\n")
	
	// Clean up the snippet - remove excessive leading/trailing whitespace
	snippet = strings.TrimSpace(snippet)
	
	// Only apply reasonable size limits for extremely large blocks (10KB+)
	// This allows for complex resources while preventing memory issues
	maxSize := 10240 // 10KB
	if len(snippet) > maxSize {
		// Find a good truncation point (preferably at a line boundary)
		truncateAt := maxSize
		for i := maxSize - 100; i < maxSize && i < len(snippet); i++ {
			if snippet[i] == '\n' {
				truncateAt = i
				break
			}
		}
		snippet = snippet[:truncateAt] + "\n  # ... (truncated for display)"
	}
	
	return snippet
}

// findResourceBlockEnd finds the actual end line of a resource block by counting braces
func findResourceBlockEnd(lines []string, startIdx int) int {
	if startIdx >= len(lines) {
		return -1
	}
	
	braceCount := 0
	inResource := false
	
	for i := startIdx; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		
		// Start counting braces once we find the opening brace of the resource
		if !inResource && strings.Contains(line, "{") {
			inResource = true
		}
		
		if inResource {
			// Count opening and closing braces
			for _, char := range line {
				switch char {
				case '{':
					braceCount++
				case '}':
					braceCount--
					if braceCount == 0 {
						// Found the closing brace of the resource block
						return i + 1 // Include the line with the closing brace
					}
				}
			}
		}
	}
	
	// If we couldn't find the closing brace, return -1 to use fallback
	return -1
}

// extractTagsFromResource extracts tags from a terraform resource block
func extractTagsFromResource(block *hclwrite.Block, resourceType string) (map[string]string, error) {
	tags := make(map[string]string)

	// Determine the correct tag attribute name based on provider
	tagAttrName := providers.GetTagIdByResource(resourceType)

	// Find the tags attribute in the resource block
	for attrName, attr := range block.Body().Attributes() {
		if attrName == tagAttrName {
			// Parse the attribute value
			tagValue := attr.Expr()
			if tagValue == nil {
				continue
			}

			// Get the tokens and parse them
			tokens := tagValue.BuildTokens(nil)
			
			// Use shared HCL parsing utility
			parsedTags, err := hclutil.ParseHclMapToStringMap(tokens)
			if err != nil {
				if strings.Contains(err.Error(), "complex expression") {
					// For complex expressions (variables, functions), we skip validation
					continue
				} else {
					return nil, fmt.Errorf("failed to parse tags from %s: %w", attrName, err)
				}
			}
			
			// Merge parsed tags
			for k, v := range parsedTags {
				tags[k] = v
			}
		}
	}

	return tags, nil
}

func (s *OperationsService) getTerraformFiles(dir, filter, skip string) ([]string, error) {
	var files []string
	
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && strings.HasSuffix(path, ".tf") {
			// Apply filter pattern
			if filter != "" && filter != ".*" {
				matched, _ := filepath.Match(filter, filepath.Base(path))
				if !matched {
					return nil
				}
			}
			
			// Apply skip pattern
			if skip != "" {
				matched, _ := filepath.Match(skip, filepath.Base(path))
				if matched {
					return nil
				}
			}
			
			files = append(files, path)
		}
		
		return nil
	})
	
	return files, err
}

// Simplified resource extraction (in production would use proper HCL parsing)
type TerraformResource struct {
	Type       string
	Name       string
	Tags       map[string]interface{}
	LineNumber int
	Snippet    string
}

func (s *OperationsService) extractResourcesFromFile(filePath string) ([]TerraformResource, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	
	// Simple regex-based extraction (in production would use HCL parser)
	var resources []TerraformResource
	lines := strings.Split(string(content), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "resource \"") {
			// Extract resource type and name from: resource "aws_instance" "example" {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				resourceType := strings.Trim(parts[1], "\"")
				resourceName := strings.Trim(parts[2], "\"")
				
				resources = append(resources, TerraformResource{
					Type: resourceType,
					Name: resourceName,
					Tags: make(map[string]interface{}),
				})
			}
		}
	}
	
	return resources, nil
}

// Extract variable resolution information from resource tags
func (s *OperationsService) extractVariableResolutionInfo(tags map[string]interface{}) map[string]interface{} {
	variableInfo := map[string]interface{}{
		"resolved_variables":    make(map[string]interface{}),
		"unresolved_references": make([]interface{}, 0),
		"literal_values":        make(map[string]interface{}),
	}
	
	resolvedVars := make(map[string]interface{})
	unresolvedRefs := make([]interface{}, 0)
	literalVals := make(map[string]interface{})
	
	for tagKey, tagValue := range tags {
		if tagValue == nil {
			continue
		}
		
		tagValueStr := fmt.Sprintf("%v", tagValue)
		
		// Check if it's a variable reference
		if strings.Contains(tagValueStr, "var.") {
			// Extract variable name
			if strings.HasPrefix(tagValueStr, "var.") {
				varName := strings.TrimPrefix(tagValueStr, "var.")
				resolvedVars[varName] = map[string]interface{}{
					"reference":    tagValueStr,
					"tag_location": tagKey,
					"resolved":     true,
					"source":       "variable",
					"value":        "resolved at runtime",
				}
			} else if strings.Contains(tagValueStr, "${var.") {
				// Handle interpolation like "${var.environment}"
				resolvedVars[tagValueStr] = map[string]interface{}{
					"reference":    tagValueStr,
					"tag_location": tagKey,
					"resolved":     true,
					"source":       "interpolation",
					"value":        "resolved at runtime",
				}
			}
		} else if strings.Contains(tagValueStr, "local.") {
			// Local reference
			resolvedVars[tagValueStr] = map[string]interface{}{
				"reference":    tagValueStr,
				"tag_location": tagKey,
				"resolved":     true,
				"source":       "local",
				"value":        "resolved at runtime",
			}
		} else {
			// Literal value
			literalVals[tagKey] = tagValueStr
		}
	}
	
	variableInfo["resolved_variables"] = resolvedVars
	variableInfo["unresolved_references"] = unresolvedRefs
	variableInfo["literal_values"] = literalVals
	
	return variableInfo
}

// Store operation result in database
func (s *OperationsService) storeOperationResult(ctx context.Context, operationID int64, filePath, resourceName, action, message string, details map[string]interface{}, lineNumber int, snippet string) {
	// Debug logging for operation_summary entries using logrus
	if strings.Contains(resourceName, "operation_summary") {
		logrus.WithFields(logrus.Fields{
			"component":    "operations", 
			"action":       "storeOperationResult",
			"operationId":  operationID,
			"filePath":     filePath,
			"resourceName": resourceName,
			"actionParam":  action,
			"message":      message,
		}).Info("DEBUG: Storing operation_summary result")
	}
	
	var detailsJSON string
	if details != nil {
		// Include the message in the details since there's no separate Message field
		details["message"] = message
		if data, err := json.Marshal(details); err == nil {
			detailsJSON = string(data)
		}
	} else {
		// Create details with just the message
		detailsMap := map[string]interface{}{"message": message}
		if data, err := json.Marshal(detailsMap); err == nil {
			detailsJSON = string(data)
		}
	}
	
	// Parse resource type and name from combined resourceName (format: "resource_type.resource_name")
	var resourceType, resourceNameOnly string
	if resourceName != "" && strings.Contains(resourceName, ".") {
		parts := strings.SplitN(resourceName, ".", 2)
		if len(parts) == 2 {
			resourceType = parts[0]
			resourceNameOnly = parts[1]
		} else {
			// Fallback if splitting fails
			resourceNameOnly = resourceName
		}
	} else {
		resourceNameOnly = resourceName
	}
	
	// More debug logging for operation_summary entries
	if strings.Contains(resourceName, "operation_summary") {
		logrus.WithFields(logrus.Fields{
			"component":       "operations", 
			"action":          "storeOperationResult",
			"operationId":     operationID,
			"resourceType":    resourceType,
			"resourceNameOnly": resourceNameOnly,
			"detailsJSONLen":  len(detailsJSON),
		}).Info("DEBUG: Parsed operation_summary parameters")
	}
	
	_, err := s.db.Queries.CreateOperationResult(ctx, db.CreateOperationResultParams{
		OperationID:   operationID,
		FilePath:      filePath,
		ResourceType:  sql.NullString{String: resourceType, Valid: resourceType != ""},
		ResourceName:  sql.NullString{String: resourceNameOnly, Valid: resourceNameOnly != ""},
		LineNumber:    sql.NullInt64{Int64: int64(lineNumber), Valid: lineNumber > 0},
		Snippet:       sql.NullString{String: snippet, Valid: snippet != ""},
		Action:        action,
		ViolationType: sql.NullString{}, // Will be set by existing logic if needed
		Details:       sql.NullString{String: detailsJSON, Valid: detailsJSON != ""},
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component":   "operations", 
			"action":      "storeOperationResult",
			"operationId": operationID,
			"error":       err.Error(),
		}).Error("Failed to store operation result")
	} else if strings.Contains(resourceName, "operation_summary") {
		logrus.WithFields(logrus.Fields{
			"component":   "operations", 
			"action":      "storeOperationResult",
			"operationId": operationID,
		}).Info("DEBUG: Successfully stored operation_summary result")
	}
}

// Helper to convert map[string]interface{} tags to map[string]string
func (s *OperationsService) convertTagsToString(tags map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range tags {
		if str, ok := v.(string); ok {
			result[k] = str
		} else {
			result[k] = fmt.Sprintf("%v", v)
		}
	}
	return result
}

// isTerraformInitialized checks if terraform/terragrunt is already initialized in the directory
func (s *OperationsService) isTerraformInitialized(dir string, iacType string) bool {
	// Check for terraform initialization files
	initPaths := []string{
		filepath.Join(dir, ".terraform"),
		filepath.Join(dir, ".terraform.lock.hcl"),
	}
	
	// For terragrunt, also check terragrunt cache
	if iacType == "terragrunt" {
		initPaths = append(initPaths, filepath.Join(dir, ".terragrunt-cache"))
	}
	
	// Check if any of the initialization indicators exist
	for _, path := range initPaths {
		if _, err := os.Stat(path); err == nil {
			logrus.WithFields(logrus.Fields{
				"component": "operations",
				"directory": dir,
				"iacType":   iacType,
				"foundPath": path,
			}).Debug("Found terraform initialization indicator")
			return true
		}
	}
	
	logrus.WithFields(logrus.Fields{
		"component": "operations",
		"directory": dir,
		"iacType":   iacType,
		"checkedPaths": initPaths,
	}).Debug("No terraform initialization indicators found")
	
	return false
}

// enhanceSnippetWithResolvedTags enhances a resource code snippet with Git diff-style resolved value blocks
func (s *OperationsService) enhanceSnippetWithResolvedTags(snippet, resourceType string, extractedTags map[string]string, validator *standards.TagValidator) string {
	if snippet == "" {
		return snippet
	}
	
	// Try to resolve tag expressions using the validator's variable resolver
	resolvedTags := make(map[string]interface{})
	if validator != nil && validator.GetVariableResolver() != nil {
		resolver := validator.GetVariableResolver()
		
		// Look for tag expressions in the snippet and try to resolve them
		resolvedTags = s.resolveTagExpressions(snippet, resourceType, resolver)
	}
	
	// Collect all resolved expressions in the snippet
	resolutions := s.collectResolvedExpressions(snippet, resolvedTags, validator)
	
	// If no resolutions found, return original snippet
	if len(resolutions) == 0 {
		return snippet
	}
	
	// Create enhanced snippet with original code + side-by-side cards
	return s.createEnhancedSnippet(snippet, resolutions)
}

// addInlineResolvedValues adds resolved values inline to a code line
func (s *OperationsService) addInlineResolvedValues(line string, resolvedTags map[string]interface{}, validator *standards.TagValidator) string {
	originalLine := line
	trimmedLine := strings.TrimSpace(line)
	
	// Skip empty lines, comments, and lines that don't contain assignments
	if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") || strings.HasPrefix(trimmedLine, "//") {
		return originalLine
	}
	
	// Check for merge expressions first
	if strings.Contains(trimmedLine, "merge(") {
		resolvedValue := s.findResolvedValueForLine(trimmedLine, resolvedTags)
		if resolvedValue != "" {
			return originalLine + "  // → " + resolvedValue
		}
	}
	
	// Check for individual variable/expression patterns in any line that has variables
	if strings.Contains(trimmedLine, "var.") || strings.Contains(trimmedLine, "local.") || strings.Contains(trimmedLine, "${") {
		if validator != nil && validator.GetVariableResolver() != nil {
			if resolvedExpression := s.resolveLineExpressions(trimmedLine, validator.GetVariableResolver()); resolvedExpression != "" {
				return originalLine + "  // → " + resolvedExpression
			}
		}
	}
	
	// Check for tag assignments with resolved values
	if strings.Contains(trimmedLine, "=") && !strings.Contains(trimmedLine, "merge(") {
		resolvedValue := s.findResolvedValueForLine(trimmedLine, resolvedTags)
		if resolvedValue != "" {
			return originalLine + "  // → " + resolvedValue
		}
	}
	
	return originalLine
}

// findResolvedValueForLine finds the resolved value for a specific line
func (s *OperationsService) findResolvedValueForLine(line string, resolvedTags map[string]interface{}) string {
	// Look for tag assignments like: Name = "value"
	if strings.Contains(line, "=") && !strings.Contains(line, "merge(") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			// Remove quotes from key if present
			key = strings.Trim(key, "\"")
			
			if value, exists := resolvedTags[key]; exists {
				return fmt.Sprintf("\"%v\"", value)
			}
		}
	}
	
	// Look for merge expressions
	if strings.Contains(line, "merge(") {
		// For merge expressions, show the combined result
		var tagPairs []string
		for key, value := range resolvedTags {
			tagPairs = append(tagPairs, fmt.Sprintf("%s = \"%v\"", key, value))
		}
		if len(tagPairs) > 0 {
			// Limit the display length
			result := "{ " + strings.Join(tagPairs, ", ") + " }"
			if len(result) > 80 {
				result = result[:77] + "..."
			}
			return result
		}
	}
	
	return ""
}

// resolveLineExpressions resolves individual expressions in a line
func (s *OperationsService) resolveLineExpressions(line string, resolver *terraform.VariableResolver) string {
	if resolver == nil {
		return ""
	}
	
	// Look for complex interpolations like "${var.project_name}-${var.environment}-vpc"
	if strings.Contains(line, "${") && strings.Contains(line, "}") {
		// Try to resolve the entire interpolated string
		result := s.resolveInterpolatedString(line, resolver)
		if result != "" {
			return result
		}
	}
	
	// Look for simple variable references like var.something or local.something
	varPattern := regexp.MustCompile(`var\.[\w.]+|local\.[\w.]+`)
	matches := varPattern.FindAllString(line, -1)
	
	if len(matches) == 0 {
		return ""
	}
	
	// Try to resolve the first variable found
	for _, match := range matches {
		result := resolver.ResolveReference(match)
		if result.Resolved {
			if strValue, ok := result.Value.(string); ok {
				return fmt.Sprintf("\"%s\"", strValue)
			} else if result.Value != nil {
				return fmt.Sprintf("%v", result.Value)
			}
		}
	}
	
	return ""
}

// resolveInterpolatedString resolves complex interpolated strings
func (s *OperationsService) resolveInterpolatedString(line string, resolver *terraform.VariableResolver) string {
	// Extract the value part from assignment (right side of =)
	if strings.Contains(line, "=") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			value := strings.TrimSpace(parts[1])
			// Remove trailing comma if present
			value = strings.TrimSuffix(value, ",")
			// Remove quotes if present
			value = strings.Trim(value, "\"")
			
			// Try to resolve this value
			result := resolver.ResolveReference(value)
			if result.Resolved {
				if strValue, ok := result.Value.(string); ok {
					return fmt.Sprintf("\"%s\"", strValue)
				} else if result.Value != nil {
					return fmt.Sprintf("\"%v\"", result.Value)
				}
			}
		}
	}
	
	return ""
}

// Resolution represents a resolved expression with its context
type Resolution struct {
	Original    string // Original expression
	Resolved    string // Resolved value
	LineNumber  int    // Line number (1-based)
	StartPos    int    // Start position in line
	EndPos      int    // End position in line
	Type        string // Type of resolution (variable, local, interpolation, merge)
}

// collectResolvedExpressions finds all resolvable expressions in the snippet
func (s *OperationsService) collectResolvedExpressions(snippet string, resolvedTags map[string]interface{}, validator *standards.TagValidator) []Resolution {
	var resolutions []Resolution
	lines := strings.Split(snippet, "\n")
	
	for lineNum, line := range lines {
		// Find all resolvable expressions in this line
		lineResolutions := s.findExpressionsInLine(line, lineNum+1, resolvedTags, validator)
		resolutions = append(resolutions, lineResolutions...)
	}
	
	return resolutions
}

// findExpressionsInLine finds all resolvable expressions in a single line
func (s *OperationsService) findExpressionsInLine(line string, lineNum int, resolvedTags map[string]interface{}, validator *standards.TagValidator) []Resolution {
	var resolutions []Resolution
	
	// Skip empty lines and comments
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
		return resolutions
	}
	
	// Find variable references: ${var.something}
	interpolationPattern := regexp.MustCompile(`\$\{([^}]+)\}`)
	matches := interpolationPattern.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			fullMatch := match[0]
			expression := match[1]
			
			if validator != nil && validator.GetVariableResolver() != nil {
				result := validator.GetVariableResolver().ResolveReference(expression)
				if result.Resolved {
					startPos := strings.Index(line, fullMatch)
					resolutions = append(resolutions, Resolution{
						Original:   fullMatch,
						Resolved:   fmt.Sprintf("%v", result.Value),
						LineNumber: lineNum,
						StartPos:   startPos,
						EndPos:     startPos + len(fullMatch),
						Type:       "interpolation",
					})
				}
			}
		}
	}
	
	// Find simple variable references: var.something
	varPattern := regexp.MustCompile(`\b(var\.[a-zA-Z_][a-zA-Z0-9_]*)\b`)
	matches = varPattern.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			varRef := match[1]
			
			if validator != nil && validator.GetVariableResolver() != nil {
				result := validator.GetVariableResolver().ResolveReference(varRef)
				if result.Resolved {
					startPos := strings.Index(line, varRef)
					resolutions = append(resolutions, Resolution{
						Original:   varRef,
						Resolved:   fmt.Sprintf("%v", result.Value),
						LineNumber: lineNum,
						StartPos:   startPos,
						EndPos:     startPos + len(varRef),
						Type:       "variable",
					})
				}
			}
		}
	}
	
	// Find local references: local.something
	localPattern := regexp.MustCompile(`\b(local\.[a-zA-Z_][a-zA-Z0-9_]*(?:\[[^\]]+\])?)\b`)
	matches = localPattern.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			localRef := match[1]
			
			if validator != nil && validator.GetVariableResolver() != nil {
				result := validator.GetVariableResolver().ResolveReference(localRef)
				if result.Resolved {
					startPos := strings.Index(line, localRef)
					resolutions = append(resolutions, Resolution{
						Original:   localRef,
						Resolved:   fmt.Sprintf("%v", result.Value),
						LineNumber: lineNum,
						StartPos:   startPos,
						EndPos:     startPos + len(localRef),
						Type:       "local",
					})
				}
			}
		}
	}
	
	// Find merge expressions
	if strings.Contains(line, "merge(") && len(resolvedTags) > 0 {
		startPos := strings.Index(line, "merge(")
		if startPos != -1 {
			// Find the complete merge expression (simplified)
			mergeExpr := "merge(...)"
			resolvedValue := s.formatResolvedTags(resolvedTags)
			
			resolutions = append(resolutions, Resolution{
				Original:   mergeExpr,
				Resolved:   resolvedValue,
				LineNumber: lineNum,
				StartPos:   startPos,
				EndPos:     startPos + 5, // Just highlight "merge"
				Type:       "merge",
			})
		}
	}
	
	return resolutions
}

// formatResolvedTags formats resolved tags for display
func (s *OperationsService) formatResolvedTags(resolvedTags map[string]interface{}) string {
	if len(resolvedTags) == 0 {
		return "{}"
	}
	
	var pairs []string
	for key, value := range resolvedTags {
		pairs = append(pairs, fmt.Sprintf(`%s = "%v"`, key, value))
	}
	
	result := "{ " + strings.Join(pairs, ", ") + " }"
	if len(result) > 100 {
		result = result[:97] + "..."
	}
	
	return result
}

// createEnhancedSnippet creates a snippet with original code followed by modern visual cards for resolved parts
func (s *OperationsService) createEnhancedSnippet(originalSnippet string, resolutions []Resolution) string {
	if len(resolutions) == 0 {
		return originalSnippet
	}
	
	var result strings.Builder
	
	// First, add the complete original code snippet
	result.WriteString(originalSnippet)
	result.WriteString("\n\n")
	
	// Add clean separator
	result.WriteString("\n── Variable Resolutions ──\n\n")
	
	// Now create modern visual cards for resolved parts
	lines := strings.Split(originalSnippet, "\n")
	
	// Group resolutions by line number
	resolutionsByLine := make(map[int][]Resolution)
	for _, res := range resolutions {
		resolutionsByLine[res.LineNumber] = append(resolutionsByLine[res.LineNumber], res)
	}
	
	// Create modern cards for each resolved section
	cardCount := 0
	for lineNum := 1; lineNum <= len(lines); lineNum++ {
		if resolutions, exists := resolutionsByLine[lineNum]; exists {
			if cardCount > 0 {
				result.WriteString("\n")
			}
			
			line := lines[lineNum-1]
			
			// Create modern visual representation
			result.WriteString(s.createModernResolutionDisplay(line, resolutions))
			cardCount++
		}
	}
	
	return result.String()
}

// createModernResolutionDisplay creates a modern visual display for resolved variables
func (s *OperationsService) createModernResolutionDisplay(line string, resolutions []Resolution) string {
	var result strings.Builder
	
	// For each resolution, create a modern visual representation
	for i, res := range resolutions {
		if i > 0 {
			result.WriteString("\n")
		}
		
		// Create clean visual boxes similar to the screenshots
		switch res.Type {
		case "merge":
			// For merge expressions, show in a visual box format
			result.WriteString(s.createVisualBox(res.Original, res.Resolved, "merge"))
			
		case "variable", "local":
			// For simple variables/locals - compact arrow format
			result.WriteString(s.createVisualBox(res.Original, res.Resolved, res.Type))
			
		case "interpolation":
			// For interpolated strings
			result.WriteString(s.createVisualBox(res.Original, res.Resolved, "interpolation"))
		}
	}
	
	return result.String()
}

// createVisualBox creates a clean visual display for resolved variables
func (s *OperationsService) createVisualBox(original, resolved, resType string) string {
	// Create clean, simple display inspired by the screenshots
	return fmt.Sprintf("  %s → %s\n", original, resolved)
}

// createCard creates a bordered card with title and content
func (s *OperationsService) createCard(title, content string) []string {
	lines := strings.Split(content, "\n")
	
	// Calculate the width needed (minimum 30, maximum 60)
	maxWidth := len(title) + 4 // title + padding
	for _, line := range lines {
		if len(line)+4 > maxWidth {
			maxWidth = len(line) + 4
		}
	}
	if maxWidth < 30 {
		maxWidth = 30
	}
	if maxWidth > 60 {
		maxWidth = 60
	}
	
	var card []string
	
	// Top border
	card = append(card, "┌─"+strings.Repeat("─", maxWidth-2)+"┐")
	
	// Title line
	titlePadding := maxWidth - len(title) - 2
	leftPad := titlePadding / 2
	rightPad := titlePadding - leftPad
	card = append(card, "│"+strings.Repeat(" ", leftPad)+title+strings.Repeat(" ", rightPad)+"│")
	
	// Separator
	card = append(card, "├─"+strings.Repeat("─", maxWidth-2)+"┤")
	
	// Content lines
	for _, line := range lines {
		if len(line) > maxWidth-4 {
			// Truncate long lines
			line = line[:maxWidth-7] + "..."
		}
		padding := maxWidth - len(line) - 2
		card = append(card, "│ "+line+strings.Repeat(" ", padding-1)+"│")
	}
	
	// Add empty line if content is too short
	if len(lines) < 2 {
		card = append(card, "│"+strings.Repeat(" ", maxWidth-2)+"│")
	}
	
	// Bottom border
	card = append(card, "└─"+strings.Repeat("─", maxWidth-2)+"┘")
	
	return card
}

// createResolvedContent creates the resolved content text for a line
func (s *OperationsService) createResolvedContent(line string, resolutions []Resolution) string {
	var parts []string
	
	for _, res := range resolutions {
		switch res.Type {
		case "merge":
			// For merge expressions, show the combined result
			parts = append(parts, res.Resolved)
		case "interpolation", "variable", "local":
			// For individual expressions, show the resolved value
			parts = append(parts, fmt.Sprintf("%s = %s", res.Original, res.Resolved))
		default:
			parts = append(parts, res.Resolved)
		}
	}
	
	if len(parts) == 0 {
		return "No resolutions found"
	}
	
	return strings.Join(parts, "\n")
}

// combineSideBySide combines two card arrays side by side
func (s *OperationsService) combineSideBySide(leftCard, rightCard []string) string {
	var result strings.Builder
	
	// Make both cards the same height
	maxHeight := len(leftCard)
	if len(rightCard) > maxHeight {
		maxHeight = len(rightCard)
	}
	
	// Pad shorter card with empty lines
	for len(leftCard) < maxHeight {
		if len(leftCard) > 0 {
			width := len(leftCard[0])
			leftCard = append(leftCard, "│"+strings.Repeat(" ", width-2)+"│")
		}
	}
	for len(rightCard) < maxHeight {
		if len(rightCard) > 0 {
			width := len(rightCard[0])
			rightCard = append(rightCard, "│"+strings.Repeat(" ", width-2)+"│")
		}
	}
	
	// Combine line by line
	for i := 0; i < maxHeight; i++ {
		leftLine := ""
		rightLine := ""
		
		if i < len(leftCard) {
			leftLine = leftCard[i]
		}
		if i < len(rightCard) {
			rightLine = rightCard[i]
		}
		
		result.WriteString(leftLine + "    " + rightLine)
		if i < maxHeight-1 {
			result.WriteString("\n")
		}
	}
	
	return result.String()
}

// highlightResolvedLine highlights resolved expressions in a line
func (s *OperationsService) highlightResolvedLine(line string, resolutions []Resolution) string {
	// Sort resolutions by start position
	sort.Slice(resolutions, func(i, j int) bool {
		return resolutions[i].StartPos < resolutions[j].StartPos
	})
	
	var result strings.Builder
	lastPos := 0
	
	for _, res := range resolutions {
		// Add text before the resolved expression
		if res.StartPos > lastPos {
			result.WriteString(line[lastPos:res.StartPos])
		}
		
		// Add highlighted resolved expression
		result.WriteString("⟨")
		result.WriteString(res.Original)
		result.WriteString("⟩")
		
		lastPos = res.EndPos
	}
	
	// Add remaining text
	if lastPos < len(line) {
		result.WriteString(line[lastPos:])
	}
	
	return result.String()
}

// resolveTagExpressions attempts to resolve tag expressions in a code snippet
func (s *OperationsService) resolveTagExpressions(snippet, resourceType string, resolver *terraform.VariableResolver) map[string]interface{} {
	resolvedTags := make(map[string]interface{})
	
	// Look for common tag expression patterns
	lines := strings.Split(snippet, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Look for merge() function calls
		if strings.Contains(line, "merge(") {
			// Try to resolve the merge expression
			if resolved := s.tryResolveMergeExpression(line, resolver); resolved != nil {
				for k, v := range resolved {
					resolvedTags[k] = v
				}
			}
		}
		
		// Look for direct tag assignments
		if strings.Contains(line, "=") && !strings.HasPrefix(line, "#") {
			if key, value := s.tryResolveDirectTagAssignment(line, resolver); key != "" {
				resolvedTags[key] = value
			}
		}
	}
	
	return resolvedTags
}

// tryResolveMergeExpression attempts to resolve a merge() expression
func (s *OperationsService) tryResolveMergeExpression(line string, resolver *terraform.VariableResolver) map[string]interface{} {
	// Extract the merge expression
	mergeStart := strings.Index(line, "merge(")
	if mergeStart == -1 {
		return nil
	}
	
	// Find the matching closing parenthesis
	parenCount := 0
	var mergeEnd int
	for i := mergeStart + 6; i < len(line); i++ {
		if line[i] == '(' {
			parenCount++
		} else if line[i] == ')' {
			if parenCount == 0 {
				mergeEnd = i
				break
			}
			parenCount--
		}
	}
	
	if mergeEnd == 0 {
		return nil
	}
	
	mergeExpr := line[mergeStart:mergeEnd+1]
	
	// Try to resolve common patterns
	result := make(map[string]interface{})
	
	// Look for local.common_tags references
	if strings.Contains(mergeExpr, "local.common_tags") {
		if commonTags := s.resolveLocalReference("common_tags", resolver); commonTags != nil {
			if tagsMap, ok := commonTags.(map[string]interface{}); ok {
				for k, v := range tagsMap {
					result[k] = v
				}
			}
		}
	}
	
	// Look for inline object literals like { Name = "value" }
	if objStart := strings.Index(mergeExpr, "{"); objStart != -1 {
		if objEnd := strings.LastIndex(mergeExpr, "}"); objEnd != -1 && objEnd > objStart {
			objContent := mergeExpr[objStart+1 : objEnd]
			if inlineTags := s.parseInlineTagObject(objContent, resolver); inlineTags != nil {
				for k, v := range inlineTags {
					result[k] = v
				}
			}
		}
	}
	
	return result
}

// tryResolveDirectTagAssignment attempts to resolve a direct tag assignment
func (s *OperationsService) tryResolveDirectTagAssignment(line string, resolver *terraform.VariableResolver) (string, interface{}) {
	// Look for pattern like: key = value
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return "", nil
	}
	
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	
	// Remove quotes from key if present
	key = strings.Trim(key, "\"")
	
	// Try to resolve the value
	if resolved := s.resolveValue(value, resolver); resolved != nil {
		return key, resolved
	}
	
	return "", nil
}

// resolveLocalReference resolves a local.* reference
func (s *OperationsService) resolveLocalReference(localName string, resolver *terraform.VariableResolver) interface{} {
	if resolver == nil {
		return nil
	}
	
	result := resolver.ResolveReference("local." + localName)
	if result != nil && result.Resolved {
		return result.Value
	}
	
	return nil
}

// parseInlineTagObject parses an inline tag object like { Name = "value", Type = "resource" }
func (s *OperationsService) parseInlineTagObject(content string, resolver *terraform.VariableResolver) map[string]interface{} {
	result := make(map[string]interface{})
	
	// Split by commas but be careful about nested structures
	assignments := strings.Split(content, ",")
	
	for _, assignment := range assignments {
		assignment = strings.TrimSpace(assignment)
		if assignment == "" {
			continue
		}
		
		parts := strings.SplitN(assignment, "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Remove quotes from key
		key = strings.Trim(key, "\"")
		
		// Try to resolve the value
		if resolved := s.resolveValue(value, resolver); resolved != nil {
			result[key] = resolved
		}
	}
	
	return result
}

// resolveValue attempts to resolve a value expression
func (s *OperationsService) resolveValue(value string, resolver *terraform.VariableResolver) interface{} {
	if resolver == nil {
		return value
	}
	
	value = strings.TrimSpace(value)
	
	// Remove trailing comma if present
	if strings.HasSuffix(value, ",") {
		value = strings.TrimSuffix(value, ",")
		value = strings.TrimSpace(value)
	}
	
	// Handle string literals
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		return strings.Trim(value, "\"")
	}
	
	// Handle variable references
	if strings.HasPrefix(value, "var.") || strings.HasPrefix(value, "local.") {
		result := resolver.ResolveReference(value)
		if result != nil && result.Resolved {
			return result.Value
		}
	}
	
	// Handle interpolation expressions
	if strings.Contains(value, "${") {
		result := resolver.ResolveReference(value)
		if result != nil && result.Resolved {
			return result.Value
		}
	}
	
	// Return as string if we can't resolve it
	return value
}
