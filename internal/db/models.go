// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package db

import (
	"database/sql"
)

type Operation struct {
	ID            int64          `db:"id" json:"id"`
	Type          string         `db:"type" json:"type"`
	Status        string         `db:"status" json:"status"`
	StandardID    sql.NullInt64  `db:"standard_id" json:"standard_id"`
	DirectoryPath string         `db:"directory_path" json:"directory_path"`
	FilterPattern sql.NullString `db:"filter_pattern" json:"filter_pattern"`
	SkipPattern   sql.NullString `db:"skip_pattern" json:"skip_pattern"`
	Settings      sql.NullString `db:"settings" json:"settings"`
	CreatedAt     sql.NullTime   `db:"created_at" json:"created_at"`
	UpdatedAt     sql.NullTime   `db:"updated_at" json:"updated_at"`
	StartedAt     sql.NullTime   `db:"started_at" json:"started_at"`
	CompletedAt   sql.NullTime   `db:"completed_at" json:"completed_at"`
}

type OperationLog struct {
	ID          int64          `db:"id" json:"id"`
	OperationID int64          `db:"operation_id" json:"operation_id"`
	Level       string         `db:"level" json:"level"`
	Message     string         `db:"message" json:"message"`
	Details     sql.NullString `db:"details" json:"details"`
	CreatedAt   sql.NullTime   `db:"created_at" json:"created_at"`
}

type OperationResult struct {
	ID            int64          `db:"id" json:"id"`
	OperationID   int64          `db:"operation_id" json:"operation_id"`
	FilePath      string         `db:"file_path" json:"file_path"`
	ResourceType  sql.NullString `db:"resource_type" json:"resource_type"`
	ResourceName  sql.NullString `db:"resource_name" json:"resource_name"`
	Action        string         `db:"action" json:"action"`
	ViolationType sql.NullString `db:"violation_type" json:"violation_type"`
	Details       sql.NullString `db:"details" json:"details"`
	CreatedAt     sql.NullTime   `db:"created_at" json:"created_at"`
	LineNumber    sql.NullInt64  `db:"line_number" json:"line_number"`
	Snippet       sql.NullString `db:"snippet" json:"snippet"`
}

type TagStandard struct {
	ID            int64          `db:"id" json:"id"`
	Name          string         `db:"name" json:"name"`
	Description   sql.NullString `db:"description" json:"description"`
	CloudProvider string         `db:"cloud_provider" json:"cloud_provider"`
	Version       int64          `db:"version" json:"version"`
	Content       string         `db:"content" json:"content"`
	CreatedAt     sql.NullTime   `db:"created_at" json:"created_at"`
	UpdatedAt     sql.NullTime   `db:"updated_at" json:"updated_at"`
}
