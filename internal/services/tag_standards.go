package services

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudyali/terratag/internal/db"
	"github.com/cloudyali/terratag/internal/models"
	"github.com/cloudyali/terratag/internal/standards"
	"gopkg.in/yaml.v2"
)

type TagStandardsService struct {
	db *DatabaseService
}

func NewTagStandardsService(db *DatabaseService) *TagStandardsService {
	return &TagStandardsService{db: db}
}

func (s *TagStandardsService) Create(ctx context.Context, req models.CreateTagStandardRequest) (*models.TagStandardResponse, error) {
	log.Printf("[TAG_STANDARDS] Creating tag standard: name=%s, provider=%s, version=%d", req.Name, req.CloudProvider, req.Version)
	
	// Set default version if not provided
	if req.Version == 0 {
		req.Version = 1
		log.Printf("[TAG_STANDARDS] Set default version to 1")
	}

	// Validate YAML content before storing
	log.Printf("[TAG_STANDARDS] Validating YAML content: length=%d", len(req.Content))
	if err := s.validateYamlContent(req.Content, req.CloudProvider); err != nil {
		log.Printf("[TAG_STANDARDS] YAML validation failed: %v", err)
		return nil, fmt.Errorf("invalid YAML content: %w", err)
	}
	log.Printf("[TAG_STANDARDS] YAML validation passed")

	log.Printf("[TAG_STANDARDS] Inserting standard into database")
	standard, err := s.db.Queries.CreateTagStandard(ctx, db.CreateTagStandardParams{
		Name:          req.Name,
		Description:   sql.NullString{String: req.Description, Valid: req.Description != ""},
		CloudProvider: req.CloudProvider,
		Version:       req.Version,
		Content:       req.Content,
	})
	if err != nil {
		log.Printf("[TAG_STANDARDS] Database insert failed: %v", err)
		return nil, fmt.Errorf("failed to create tag standard: %w", err)
	}

	log.Printf("[TAG_STANDARDS] Tag standard created successfully: id=%d", standard.ID)
	response := models.TagStandardFromDB(standard)
	return &response, nil
}

func (s *TagStandardsService) GetByID(ctx context.Context, id int64) (*models.TagStandardResponse, error) {
	log.Printf("[TAG_STANDARDS] Fetching tag standard by ID: id=%d", id)
	
	standard, err := s.db.Queries.GetTagStandard(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[TAG_STANDARDS] Tag standard not found: id=%d", id)
			return nil, fmt.Errorf("tag standard not found")
		}
		log.Printf("[TAG_STANDARDS] Database query failed: id=%d, error=%v", id, err)
		return nil, fmt.Errorf("failed to get tag standard: %w", err)
	}

	log.Printf("[TAG_STANDARDS] Tag standard retrieved successfully: id=%d, name=%s", standard.ID, standard.Name)
	response := models.TagStandardFromDB(standard)
	return &response, nil
}

func (s *TagStandardsService) GetByName(ctx context.Context, name string) (*models.TagStandardResponse, error) {
	standard, err := s.db.Queries.GetTagStandardByName(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tag standard not found")
		}
		return nil, fmt.Errorf("failed to get tag standard: %w", err)
	}

	response := models.TagStandardFromDB(standard)
	return &response, nil
}

func (s *TagStandardsService) List(ctx context.Context) ([]models.TagStandardResponse, error) {
	standards, err := s.db.Queries.ListTagStandards(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tag standards: %w", err)
	}

	var response []models.TagStandardResponse
	for _, standard := range standards {
		response = append(response, models.TagStandardFromDB(standard))
	}

	return response, nil
}

func (s *TagStandardsService) ListByProvider(ctx context.Context, provider string) ([]models.TagStandardResponse, error) {
	standards, err := s.db.Queries.ListTagStandardsByProvider(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to list tag standards by provider: %w", err)
	}

	var response []models.TagStandardResponse
	for _, standard := range standards {
		response = append(response, models.TagStandardFromDB(standard))
	}

	return response, nil
}

func (s *TagStandardsService) Update(ctx context.Context, id int64, req models.UpdateTagStandardRequest) (*models.TagStandardResponse, error) {
	// Check if exists
	_, err := s.db.Queries.GetTagStandard(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tag standard not found")
		}
		return nil, fmt.Errorf("failed to check tag standard existence: %w", err)
	}

	// Validate YAML content before updating
	if err := s.validateYamlContent(req.Content, req.CloudProvider); err != nil {
		return nil, fmt.Errorf("invalid YAML content: %w", err)
	}

	standard, err := s.db.Queries.UpdateTagStandard(ctx, db.UpdateTagStandardParams{
		ID:            id,
		Name:          req.Name,
		Description:   sql.NullString{String: req.Description, Valid: req.Description != ""},
		CloudProvider: req.CloudProvider,
		Version:       req.Version,
		Content:       req.Content,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update tag standard: %w", err)
	}

	response := models.TagStandardFromDB(standard)
	return &response, nil
}

func (s *TagStandardsService) Delete(ctx context.Context, id int64) error {
	// Check if exists
	_, err := s.db.Queries.GetTagStandard(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("tag standard not found")
		}
		return fmt.Errorf("failed to check tag standard existence: %w", err)
	}

	err = s.db.Queries.DeleteTagStandard(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete tag standard: %w", err)
	}

	return nil
}

func (s *TagStandardsService) GenerateFromDirectory(ctx context.Context, req models.GenerateStandardRequest) (*models.TagStandardResponse, error) {
	// Get terraform files in directory
	files, err := s.getTerraformFiles(req.DirectoryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get terraform files: %w", err)
	}

	// Analyze existing tags from terraform files
	existingTags := make(map[string]map[string]bool) // resource_type -> tag_key -> exists
	resourceTypes := make(map[string]bool)

	for _, file := range files {
		resources, tags, err := s.analyzeFile(file)
		if err != nil {
			continue // Skip files with errors
		}

		for _, resource := range resources {
			resourceTypes[resource] = true
			if existingTags[resource] == nil {
				existingTags[resource] = make(map[string]bool)
			}
		}

		for resourceType, resourceTags := range tags {
			if existingTags[resourceType] == nil {
				existingTags[resourceType] = make(map[string]bool)
			}
			for tagKey := range resourceTags {
				existingTags[resourceType][tagKey] = true
			}
		}
	}

	// Generate tag standard based on analysis
	standard := s.generateStandard(req, existingTags, resourceTypes)

	// Convert to YAML
	content, err := yaml.Marshal(standard)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal standard to YAML: %w", err)
	}

	// Create the standard in database
	createReq := models.CreateTagStandardRequest{
		Name:          req.Name,
		Description:   req.Description,
		CloudProvider: req.CloudProvider,
		Version:       1,
		Content:       string(content),
	}

	return s.Create(ctx, createReq)
}

// Helper to get terraform files in directory
func (s *TagStandardsService) getTerraformFiles(dir string) ([]string, error) {
	var files []string
	
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && strings.HasSuffix(path, ".tf") {
			files = append(files, path)
		}
		
		return nil
	})
	
	return files, err
}

// Analyze a terraform file for resources and existing tags
func (s *TagStandardsService) analyzeFile(filePath string) ([]string, map[string]map[string]string, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, nil, err
	}

	lines := strings.Split(string(content), "\n")
	var resources []string
	tags := make(map[string]map[string]string)
	
	var currentResource string
	inTagsBlock := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Look for resource declarations
		if strings.HasPrefix(line, "resource \"") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				resourceType := strings.Trim(parts[1], "\"")
				resources = append(resources, resourceType)
				currentResource = resourceType
			}
		}
		
		// Look for tags blocks
		if currentResource != "" && strings.Contains(line, "tags") && strings.Contains(line, "{") {
			inTagsBlock = true
			if tags[currentResource] == nil {
				tags[currentResource] = make(map[string]string)
			}
		}
		
		// Parse individual tag lines
		if inTagsBlock && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(strings.Trim(parts[0], "\""))
				value := strings.TrimSpace(strings.Trim(parts[1], "\""))
				if tags[currentResource] == nil {
					tags[currentResource] = make(map[string]string)
				}
				tags[currentResource][key] = value
			}
		}
		
		// End of tags block
		if inTagsBlock && strings.Contains(line, "}") {
			inTagsBlock = false
		}
		
		// End of resource block
		if currentResource != "" && strings.HasPrefix(line, "}") && !inTagsBlock {
			currentResource = ""
		}
	}
	
	return resources, tags, nil
}

// Generate a tag standard based on analysis
func (s *TagStandardsService) generateStandard(req models.GenerateStandardRequest, existingTags map[string]map[string]bool, resourceTypes map[string]bool) *standards.TagStandard {
	standard := &standards.TagStandard{
		Version: 1,
		Metadata: standards.Metadata{
			Description: req.Description,
			Author:      "Generated by Terratag",
		},
		CloudProvider: req.CloudProvider,
	}

	// Common required tags based on cloud provider
	commonTags := s.getCommonTags(req.CloudProvider)
	
	// Analyze existing tags to find common patterns
	if req.AnalyzeTags {
		commonTags = append(commonTags, s.analyzeCommonTags(existingTags)...)
	}
	
	// Add common tags if requested
	if req.IncludeCommon {
		standard.RequiredTags = commonTags
		standard.OptionalTags = s.getOptionalTags(req.CloudProvider)
	} else {
		standard.RequiredTags = commonTags[:min(len(commonTags), 3)] // Limit to top 3
	}

	// Add resource-specific rules for taggable resources
	for resourceType := range resourceTypes {
		if standards.IsTaggableResource(resourceType, req.CloudProvider) {
			rule := standards.ResourceRule{
				ResourceTypes: []string{resourceType},
			}
			standard.ResourceRules = append(standard.ResourceRules, rule)
		}
	}

	return standard
}

// Get common required tags for a cloud provider
func (s *TagStandardsService) getCommonTags(provider string) []standards.TagSpec {
	switch provider {
	case "aws":
		return []standards.TagSpec{
			{Key: "Environment", AllowedValues: []string{"Production", "Staging", "Development"}, CaseSensitive: false},
			{Key: "Owner", DataType: standards.DataTypeEmail, Format: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"},
			{Key: "CostCenter", DataType: standards.DataTypeString, MaxLength: 50},
		}
	case "gcp":
		return []standards.TagSpec{
			{Key: "environment", AllowedValues: []string{"prod", "staging", "dev"}, CaseSensitive: false},
			{Key: "team", DataType: standards.DataTypeString, MaxLength: 50},
			{Key: "project", DataType: standards.DataTypeString, MaxLength: 100},
		}
	case "azure":
		return []standards.TagSpec{
			{Key: "Environment", AllowedValues: []string{"Production", "Staging", "Development"}, CaseSensitive: false},
			{Key: "Owner", DataType: standards.DataTypeString, MaxLength: 100},
			{Key: "CostCenter", DataType: standards.DataTypeString, MaxLength: 50},
		}
	default:
		return []standards.TagSpec{
			{Key: "Environment", AllowedValues: []string{"Production", "Staging", "Development"}, CaseSensitive: false},
			{Key: "Owner", DataType: standards.DataTypeString, MaxLength: 100},
		}
	}
}

// Get optional tags for a cloud provider
func (s *TagStandardsService) getOptionalTags(provider string) []standards.TagSpec {
	return []standards.TagSpec{
		{Key: "Project", DataType: standards.DataTypeString, MaxLength: 100},
		{Key: "Description", DataType: standards.DataTypeString, MaxLength: 255},
		{Key: "CreatedBy", DataType: standards.DataTypeString, MaxLength: 100},
		{Key: "CreatedDate", DataType: standards.DataTypeString, Format: "^\\d{4}-\\d{2}-\\d{2}$"},
	}
}

// Analyze existing tags to find common patterns
func (s *TagStandardsService) analyzeCommonTags(existingTags map[string]map[string]bool) []standards.TagSpec {
	tagCounts := make(map[string]int)
	totalResources := len(existingTags)
	
	// Count tag occurrences across resources
	for _, resourceTags := range existingTags {
		for tagKey := range resourceTags {
			tagCounts[tagKey]++
		}
	}
	
	var commonTags []standards.TagSpec
	
	// Consider tags that appear in at least 50% of resources as common
	threshold := totalResources / 2
	for tagKey, count := range tagCounts {
		if count >= threshold {
			commonTags = append(commonTags, standards.TagSpec{
				Key:       tagKey,
				DataType:  standards.DataTypeString,
				MaxLength: 100,
			})
		}
	}
	
	return commonTags
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// validateYamlContent validates that the provided YAML content is valid and conforms to tag standard schema
func (s *TagStandardsService) validateYamlContent(content string, cloudProvider string) error {
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("content cannot be empty")
	}

	// Parse YAML to check basic syntax
	var standard standards.TagStandard
	if err := yaml.Unmarshal([]byte(content), &standard); err != nil {
		return fmt.Errorf("invalid YAML syntax: %w", err)
	}

	// Validate cloud provider matches
	if standard.CloudProvider != cloudProvider {
		return fmt.Errorf("cloud provider in content (%s) does not match specified provider (%s)", 
			standard.CloudProvider, cloudProvider)
	}

	// Use the existing validation logic from standards package
	if err := standards.ValidateStandard(&standard); err != nil {
		return fmt.Errorf("invalid tag standard: %w", err)
	}

	return nil
}

// ValidateContent validates YAML content without storing it (public interface for API)
func (s *TagStandardsService) ValidateContent(content string, cloudProvider string) error {
	log.Printf("[TAG_STANDARDS] Validating content: provider=%s, length=%d", cloudProvider, len(content))
	err := s.validateYamlContent(content, cloudProvider)
	if err != nil {
		log.Printf("[TAG_STANDARDS] Content validation failed: provider=%s, error=%v", cloudProvider, err)
	} else {
		log.Printf("[TAG_STANDARDS] Content validation passed: provider=%s", cloudProvider)
	}
	return err
}