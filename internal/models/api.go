package models

import (
	"database/sql"
	"time"

	"github.com/cloudyali/terratag/internal/db"
)

// API request/response models

type CreateTagStandardRequest struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	CloudProvider string `json:"cloud_provider" binding:"required"`
	Version       int64  `json:"version"`
	Content       string `json:"content" binding:"required"`
}

type UpdateTagStandardRequest struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	CloudProvider string `json:"cloud_provider" binding:"required"`
	Version       int64  `json:"version"`
	Content       string `json:"content" binding:"required"`
}

type TagStandardResponse struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	CloudProvider string    `json:"cloud_provider"`
	Version       int64     `json:"version"`
	Content       string    `json:"content"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateOperationRequest struct {
	Type           string `json:"type" binding:"required"` // "validation" or "tagging"
	StandardID     int64  `json:"standard_id"`
	DirectoryPath  string `json:"directory_path" binding:"required"`
	FilterPattern  string `json:"filter_pattern"`
	SkipPattern    string `json:"skip_pattern"`
	Settings       string `json:"settings"` // JSON settings
}

type OperationResponse struct {
	ID            int64      `json:"id"`
	Type          string     `json:"type"`
	Status        string     `json:"status"`
	StandardID    *int64     `json:"standard_id"`
	DirectoryPath string     `json:"directory_path"`
	FilterPattern *string    `json:"filter_pattern"`
	SkipPattern   *string    `json:"skip_pattern"`
	Settings      *string    `json:"settings"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	StartedAt     *time.Time `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at"`
}

type OperationResultResponse struct {
	ID             int64     `json:"id"`
	OperationID    int64     `json:"operation_id"`
	FilePath       string    `json:"file_path"`
	ResourceType   *string   `json:"resource_type"`
	ResourceName   *string   `json:"resource_name"`
	LineNumber     *int64    `json:"line_number"`
	Snippet        *string   `json:"snippet"`
	Action         string    `json:"action"`
	ViolationType  *string   `json:"violation_type"`
	Details        *string   `json:"details"`
	CreatedAt      time.Time `json:"created_at"`
}

type OperationLogResponse struct {
	ID          int64     `json:"id"`
	OperationID int64     `json:"operation_id"`
	Level       string    `json:"level"`
	Message     string    `json:"message"`
	Details     *string   `json:"details"`
	CreatedAt   time.Time `json:"created_at"`
}

type OperationSummaryResponse struct {
	Operation    OperationResponse         `json:"operation"`
	Results      []OperationResultResponse `json:"results"`
	Logs         []OperationLogResponse    `json:"logs"`
	Summary      OperationStats            `json:"summary"`
}

type OperationStats struct {
	TotalFiles      int64 `json:"total_files"`
	ProcessedFiles  int64 `json:"processed_files"`
	TaggedResources int64 `json:"tagged_resources"`
	Violations      int64 `json:"violations"`
	Errors          int64 `json:"errors"`
}

type GenerateStandardRequest struct {
	DirectoryPath string `json:"directory_path" binding:"required"`
	CloudProvider string `json:"cloud_provider" binding:"required"`
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	AnalyzeTags   bool   `json:"analyze_tags"`
	IncludeCommon bool   `json:"include_common"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PaginationRequest struct {
	Page  int64 `form:"page" json:"page"`
	Limit int64 `form:"limit" json:"limit"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int64       `json:"page"`
	Limit      int64       `json:"limit"`
	Total      int64       `json:"total"`
	TotalPages int64       `json:"total_pages"`
}

// Conversion functions
func TagStandardFromDB(dbStandard db.TagStandard) TagStandardResponse {
	return TagStandardResponse{
		ID:            dbStandard.ID,
		Name:          dbStandard.Name,
		Description:   dbStandard.Description.String,
		CloudProvider: dbStandard.CloudProvider,
		Version:       dbStandard.Version,
		Content:       dbStandard.Content,
		CreatedAt:     dbStandard.CreatedAt.Time,
		UpdatedAt:     dbStandard.UpdatedAt.Time,
	}
}

func OperationFromDB(dbOperation db.Operation) OperationResponse {
	return OperationResponse{
		ID:            dbOperation.ID,
		Type:          dbOperation.Type,
		Status:        dbOperation.Status,
		StandardID:    nullInt64ToPtr(dbOperation.StandardID),
		DirectoryPath: dbOperation.DirectoryPath,
		FilterPattern: nullStringToPtr(dbOperation.FilterPattern),
		SkipPattern:   nullStringToPtr(dbOperation.SkipPattern),
		Settings:      nullStringToPtr(dbOperation.Settings),
		CreatedAt:     dbOperation.CreatedAt.Time,
		UpdatedAt:     dbOperation.UpdatedAt.Time,
		StartedAt:     nullTimeToPtr(dbOperation.StartedAt),
		CompletedAt:   nullTimeToPtr(dbOperation.CompletedAt),
	}
}

func OperationResultFromDB(dbResult db.OperationResult) OperationResultResponse {
	return OperationResultResponse{
		ID:             dbResult.ID,
		OperationID:    dbResult.OperationID,
		FilePath:       dbResult.FilePath,
		ResourceType:   nullStringToPtr(dbResult.ResourceType),
		ResourceName:   nullStringToPtr(dbResult.ResourceName),
		LineNumber:     nullInt64ToPtr(dbResult.LineNumber),
		Snippet:        nullStringToPtr(dbResult.Snippet),
		Action:         dbResult.Action,
		ViolationType:  nullStringToPtr(dbResult.ViolationType),
		Details:        nullStringToPtr(dbResult.Details),
		CreatedAt:      dbResult.CreatedAt.Time,
	}
}

func OperationLogFromDB(dbLog db.OperationLog) OperationLogResponse {
	return OperationLogResponse{
		ID:          dbLog.ID,
		OperationID: dbLog.OperationID,
		Level:       dbLog.Level,
		Message:     dbLog.Message,
		Details:     nullStringToPtr(dbLog.Details),
		CreatedAt:   dbLog.CreatedAt.Time,
	}
}

// Helper functions
func nullStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func nullInt64ToPtr(ni sql.NullInt64) *int64 {
	if ni.Valid {
		return &ni.Int64
	}
	return nil
}

func nullTimeToPtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}