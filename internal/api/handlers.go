package api

import (
	"log"
	"net/http"
	"strconv"
	"time"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/cloudyali/terratag/internal/models"
	"github.com/cloudyali/terratag/internal/services"
)

type Handlers struct {
	tagStandardsService *services.TagStandardsService
	operationsService   *services.OperationsService
	databaseService     *services.DatabaseService
}

func NewHandlers(
	tagStandardsService *services.TagStandardsService,
	operationsService *services.OperationsService,
	databaseService *services.DatabaseService,
) *Handlers {
	return &Handlers{
		tagStandardsService: tagStandardsService,
		operationsService:   operationsService,
		databaseService:     databaseService,
	}
}

// Helper function for structured logging
func (h *Handlers) logRequest(c *gin.Context, operation string, details map[string]interface{}) {
	requestID := c.GetString("requestId")
	if requestID == "" {
		requestID = fmt.Sprintf("%d", time.Now().UnixNano())
		c.Set("requestId", requestID)
	}
	
	logData := map[string]interface{}{
		"requestId": requestID,
		"method":    c.Request.Method,
		"path":      c.Request.URL.Path,
		"operation": operation,
		"clientIP":  c.ClientIP(),
		"userAgent": c.Request.UserAgent(),
	}
	
	// Merge additional details
	for k, v := range details {
		logData[k] = v
	}
	
	log.Printf("[API] %s: %+v", operation, logData)
}

func (h *Handlers) logResponse(c *gin.Context, operation string, statusCode int, duration time.Duration, details map[string]interface{}) {
	requestID := c.GetString("requestId")
	
	logData := map[string]interface{}{
		"requestId":  requestID,
		"operation":  operation,
		"statusCode": statusCode,
		"duration":   duration.String(),
	}
	
	// Merge additional details
	for k, v := range details {
		logData[k] = v
	}
	
	log.Printf("[API] %s completed: %+v", operation, logData)
}

// Health check endpoint
func (h *Handlers) HealthCheck(c *gin.Context) {
	start := time.Now()
	h.logRequest(c, "HealthCheck", nil)
	
	if err := h.databaseService.HealthCheck(); err != nil {
		h.logResponse(c, "HealthCheck", http.StatusServiceUnavailable, time.Since(start), map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error:   "Database connection failed",
			Message: err.Error(),
			Code:    http.StatusServiceUnavailable,
		})
		return
	}

	h.logResponse(c, "HealthCheck", http.StatusOK, time.Since(start), map[string]interface{}{
		"dbStatus": "healthy",
	})
	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Service is healthy",
		Data: map[string]string{
			"status": "ok",
			"version": "1.0.0",
		},
	})
}

// Tag Standards endpoints

func (h *Handlers) CreateTagStandard(c *gin.Context) {
	start := time.Now()
	h.logRequest(c, "CreateTagStandard", nil)
	
	var req models.CreateTagStandardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logResponse(c, "CreateTagStandard", http.StatusBadRequest, time.Since(start), map[string]interface{}{
			"error": "Invalid JSON request",
			"details": err.Error(),
		})
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	h.logRequest(c, "CreateTagStandard", map[string]interface{}{
		"name": req.Name,
		"cloudProvider": req.CloudProvider,
		"version": req.Version,
		"contentLength": len(req.Content),
	})

	standard, err := h.tagStandardsService.Create(c.Request.Context(), req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "tag standard not found" {
			status = http.StatusNotFound
		}
		h.logResponse(c, "CreateTagStandard", status, time.Since(start), map[string]interface{}{
			"error": err.Error(),
			"name": req.Name,
		})
		c.JSON(status, models.ErrorResponse{
			Error:   "Failed to create tag standard",
			Message: err.Error(),
			Code:    status,
		})
		return
	}

	h.logResponse(c, "CreateTagStandard", http.StatusCreated, time.Since(start), map[string]interface{}{
		"standardId": standard.ID,
		"name": standard.Name,
	})
	c.JSON(http.StatusCreated, models.SuccessResponse{
		Message: "Tag standard created successfully",
		Data:    standard,
	})
}

func (h *Handlers) GetTagStandard(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid ID",
			Message: "ID must be a valid integer",
			Code:    http.StatusBadRequest,
		})
		return
	}

	standard, err := h.tagStandardsService.GetByID(c.Request.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "tag standard not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, models.ErrorResponse{
			Error:   "Failed to get tag standard",
			Message: err.Error(),
			Code:    status,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Tag standard retrieved successfully",
		Data:    standard,
	})
}

func (h *Handlers) ListTagStandards(c *gin.Context) {
	provider := c.Query("provider")
	
	var standards []models.TagStandardResponse
	var err error
	
	if provider != "" {
		standards, err = h.tagStandardsService.ListByProvider(c.Request.Context(), provider)
	} else {
		standards, err = h.tagStandardsService.List(c.Request.Context())
	}
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to list tag standards",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Tag standards retrieved successfully",
		Data:    standards,
	})
}

func (h *Handlers) UpdateTagStandard(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid ID",
			Message: "ID must be a valid integer",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req models.UpdateTagStandardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	standard, err := h.tagStandardsService.Update(c.Request.Context(), id, req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "tag standard not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, models.ErrorResponse{
			Error:   "Failed to update tag standard",
			Message: err.Error(),
			Code:    status,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Tag standard updated successfully",
		Data:    standard,
	})
}

func (h *Handlers) DeleteTagStandard(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid ID",
			Message: "ID must be a valid integer",
			Code:    http.StatusBadRequest,
		})
		return
	}

	err = h.tagStandardsService.Delete(c.Request.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "tag standard not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, models.ErrorResponse{
			Error:   "Failed to delete tag standard",
			Message: err.Error(),
			Code:    status,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Tag standard deleted successfully",
	})
}

func (h *Handlers) ValidateTagStandardContent(c *gin.Context) {
	start := time.Now()
	h.logRequest(c, "ValidateTagStandardContent", nil)
	
	var req struct {
		Content       string `json:"content" binding:"required"`
		CloudProvider string `json:"cloud_provider" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logResponse(c, "ValidateTagStandardContent", http.StatusBadRequest, time.Since(start), map[string]interface{}{
			"error": "Invalid JSON request",
			"details": err.Error(),
		})
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	h.logRequest(c, "ValidateTagStandardContent", map[string]interface{}{
		"cloudProvider": req.CloudProvider,
		"contentLength": len(req.Content),
		"contentPreview": func() string {
			if len(req.Content) > 100 {
				return req.Content[:100] + "..."
			}
			return req.Content
		}(),
	})

	// Use the tag standards service to validate the content
	err := h.tagStandardsService.ValidateContent(req.Content, req.CloudProvider)
	if err != nil {
		h.logResponse(c, "ValidateTagStandardContent", http.StatusBadRequest, time.Since(start), map[string]interface{}{
			"error": err.Error(),
			"cloudProvider": req.CloudProvider,
			"validationFailed": true,
		})
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid content",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	h.logResponse(c, "ValidateTagStandardContent", http.StatusOK, time.Since(start), map[string]interface{}{
		"cloudProvider": req.CloudProvider,
		"validationPassed": true,
	})
	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Content is valid",
		Data: map[string]interface{}{
			"valid": true,
		},
	})
}

// Operations endpoints

func (h *Handlers) CreateOperation(c *gin.Context) {
	var req models.CreateOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	operation, err := h.operationsService.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to create operation",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse{
		Message: "Operation created successfully",
		Data:    operation,
	})
}

func (h *Handlers) GetOperation(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid ID",
			Message: "ID must be a valid integer",
			Code:    http.StatusBadRequest,
		})
		return
	}

	operation, err := h.operationsService.GetByID(c.Request.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "operation not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, models.ErrorResponse{
			Error:   "Failed to get operation",
			Message: err.Error(),
			Code:    status,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Operation retrieved successfully",
		Data:    operation,
	})
}

func (h *Handlers) ListOperations(c *gin.Context) {
	var pagination models.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid pagination parameters",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	operations, err := h.operationsService.List(c.Request.Context(), pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to list operations",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Operations retrieved successfully",
		Data:    operations,
	})
}

func (h *Handlers) GetOperationSummary(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid ID",
			Message: "ID must be a valid integer",
			Code:    http.StatusBadRequest,
		})
		return
	}

	summary, err := h.operationsService.GetSummary(c.Request.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "operation not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, models.ErrorResponse{
			Error:   "Failed to get operation summary",
			Message: err.Error(),
			Code:    status,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Operation summary retrieved successfully",
		Data:    summary,
	})
}

func (h *Handlers) ExecuteOperation(c *gin.Context) {
	start := time.Now()
	h.logRequest(c, "ExecuteOperation", nil)
	
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.logResponse(c, "ExecuteOperation", http.StatusBadRequest, time.Since(start), map[string]interface{}{
			"error": "Invalid operation ID",
			"providedId": c.Param("id"),
		})
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid ID",
			Message: "ID must be a valid integer",
			Code:    http.StatusBadRequest,
		})
		return
	}

	h.logRequest(c, "ExecuteOperation", map[string]interface{}{
		"operationId": id,
	})

	err = h.operationsService.Execute(c.Request.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "operation not found" {
			status = http.StatusNotFound
		} else if err.Error() == "operation is not in pending state" {
			status = http.StatusConflict
		}
		h.logResponse(c, "ExecuteOperation", status, time.Since(start), map[string]interface{}{
			"error": err.Error(),
			"operationId": id,
			"executionFailed": true,
		})
		c.JSON(status, models.ErrorResponse{
			Error:   "Failed to execute operation",
			Message: err.Error(),
			Code:    status,
		})
		return
	}

	h.logResponse(c, "ExecuteOperation", http.StatusOK, time.Since(start), map[string]interface{}{
		"operationId": id,
		"executionStarted": true,
	})
	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Operation execution started",
	})
}

func (h *Handlers) DeleteOperation(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid ID",
			Message: "ID must be a valid integer",
			Code:    http.StatusBadRequest,
		})
		return
	}

	err = h.operationsService.Delete(c.Request.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "operation not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, models.ErrorResponse{
			Error:   "Failed to delete operation",
			Message: err.Error(),
			Code:    status,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Operation deleted successfully",
	})
}

func (h *Handlers) GetOperationResults(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid ID",
			Message: "ID must be a valid integer",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var pagination models.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid pagination parameters",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	results, err := h.operationsService.GetResults(c.Request.Context(), id, pagination)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "operation not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, models.ErrorResponse{
			Error:   "Failed to get operation results",
			Message: err.Error(),
			Code:    status,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Operation results retrieved successfully",
		Data:    results,
	})
}

func (h *Handlers) GetOperationLogs(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid ID",
			Message: "ID must be a valid integer",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var pagination models.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid pagination parameters",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	logs, err := h.operationsService.GetLogs(c.Request.Context(), id, pagination)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "operation not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, models.ErrorResponse{
			Error:   "Failed to get operation logs",
			Message: err.Error(),
			Code:    status,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Operation logs retrieved successfully",
		Data:    logs,
	})
}

func (h *Handlers) RetryOperation(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid ID",
			Message: "ID must be a valid integer",
			Code:    http.StatusBadRequest,
		})
		return
	}

	err = h.operationsService.Retry(c.Request.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "operation not found" {
			status = http.StatusNotFound
		} else if err.Error() == "operation is not in failed state" {
			status = http.StatusConflict
		}
		c.JSON(status, models.ErrorResponse{
			Error:   "Failed to retry operation",
			Message: err.Error(),
			Code:    status,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Operation retry started",
	})
}

func (h *Handlers) GenerateTagStandard(c *gin.Context) {
	var req models.GenerateStandardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	standard, err := h.tagStandardsService.GenerateFromDirectory(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to generate tag standard",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Tag standard generated successfully",
		Data:    standard,
	})
}