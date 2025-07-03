package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// ProviderSchema represents the structure of terraform provider schema
type ProviderSchema struct {
	ProviderSchemas map[string]Provider `json:"provider_schemas"`
}

type Provider struct {
	ResourceSchemas map[string]ResourceSchema `json:"resource_schemas"`
}

type ResourceSchema struct {
	Block Block `json:"block"`
}

type Block struct {
	Attributes map[string]interface{} `json:"attributes"`
}

// AWSResourceTaggingInfo contains information about AWS resource tagging support
type AWSResourceTaggingInfo struct {
	ResourceType         string `json:"resource_type"`
	SupportsTagAttribute bool   `json:"supports_tag_attribute"`
	Category             string `json:"category"`
	Service              string `json:"service"`
}

func main() {
	// Get Terraform provider schema from file if it exists, otherwise run command
	var output []byte
	var err error
	
	if _, statErr := os.Stat("/tmp/aws-schema.json"); statErr == nil {
		output, err = os.ReadFile("/tmp/aws-schema.json")
		if err != nil {
			log.Fatalf("Failed to read schema file: %v", err)
		}
	} else {
		cmd := exec.Command("terraform", "providers", "schema", "-json")
		cmd.Dir = "test/validation-tests/compliant" // Set working directory
		output, err = cmd.Output()
		if err != nil {
			log.Fatalf("Failed to get provider schema: %v", err)
		}
	}

	var schema ProviderSchema
	if err := json.Unmarshal(output, &schema); err != nil {
		log.Fatalf("Failed to parse schema: %v", err)
	}

	// Extract AWS provider schema
	awsProvider, exists := schema.ProviderSchemas["registry.terraform.io/hashicorp/aws"]
	if !exists {
		log.Fatal("AWS provider not found in schema")
	}

	var taggableResources []string
	var nonTaggableResources []string
	var allResources []AWSResourceTaggingInfo

	for resourceType, resourceSchema := range awsProvider.ResourceSchemas {
		supportsTagging := resourceSchema.Block.Attributes["tags"] != nil
		
		info := AWSResourceTaggingInfo{
			ResourceType:         resourceType,
			SupportsTagAttribute: supportsTagging,
			Service:              extractServiceName(resourceType),
			Category:             categorizeResource(resourceType, supportsTagging),
		}
		
		allResources = append(allResources, info)
		
		if supportsTagging {
			taggableResources = append(taggableResources, resourceType)
		} else {
			nonTaggableResources = append(nonTaggableResources, resourceType)
		}
	}

	// Sort resources
	sort.Strings(taggableResources)
	sort.Strings(nonTaggableResources)

	// Generate comprehensive data file
	generateResourceTaggingData(allResources)
	
	// Print summary
	fmt.Printf("AWS Provider Resource Tagging Analysis\n")
	fmt.Printf("=====================================\n\n")
	fmt.Printf("Total AWS Resources: %d\n", len(allResources))
	fmt.Printf("Resources Supporting Tags: %d (%.1f%%)\n", len(taggableResources), float64(len(taggableResources))/float64(len(allResources))*100)
	fmt.Printf("Resources NOT Supporting Tags: %d (%.1f%%)\n\n", len(nonTaggableResources), float64(len(nonTaggableResources))/float64(len(allResources))*100)

	// Service breakdown
	serviceBreakdown := make(map[string]struct {
		total      int
		taggable   int
		percentage float64
	})
	
	for _, resource := range allResources {
		stats := serviceBreakdown[resource.Service]
		stats.total++
		if resource.SupportsTagAttribute {
			stats.taggable++
		}
		stats.percentage = float64(stats.taggable) / float64(stats.total) * 100
		serviceBreakdown[resource.Service] = stats
	}

	fmt.Printf("Service Breakdown (Top 15 by resource count):\n")
	fmt.Printf("%-20s %-8s %-8s %-8s\n", "Service", "Total", "Taggable", "Rate")
	fmt.Printf("%-20s %-8s %-8s %-8s\n", "-------", "-----", "--------", "----")
	
	// Sort services by total count
	type serviceStats struct {
		name string
		stats struct {
			total      int
			taggable   int
			percentage float64
		}
	}
	var services []serviceStats
	for service, stats := range serviceBreakdown {
		services = append(services, serviceStats{name: service, stats: stats})
	}
	sort.Slice(services, func(i, j int) bool {
		return services[i].stats.total > services[j].stats.total
	})
	
	for i, service := range services {
		if i >= 15 {
			break
		}
		fmt.Printf("%-20s %-8d %-8d %.1f%%\n", 
			service.name, service.stats.total, service.stats.taggable, service.stats.percentage)
	}
}

func extractServiceName(resourceType string) string {
	// Remove aws_ prefix and extract service name
	if strings.HasPrefix(resourceType, "aws_") {
		parts := strings.Split(resourceType[4:], "_")
		if len(parts) > 0 {
			return parts[0]
		}
	}
	return "unknown"
}

func categorizeResource(resourceType string, supportsTagging bool) string {
	if supportsTagging {
		return "taggable"
	}
	
	// Categorize non-taggable resources
	if strings.Contains(resourceType, "_attachment") ||
		strings.Contains(resourceType, "_association") ||
		strings.Contains(resourceType, "_permission") {
		return "association"
	}
	
	if strings.Contains(resourceType, "_validation") ||
		strings.Contains(resourceType, "_certificate") {
		return "validation"
	}
	
	if strings.Contains(resourceType, "_deployment") ||
		strings.Contains(resourceType, "_method") ||
		strings.Contains(resourceType, "_integration") {
		return "configuration"
	}
	
	return "non-taggable"
}

func generateResourceTaggingData(resources []AWSResourceTaggingInfo) {
	// Sort by resource type
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].ResourceType < resources[j].ResourceType
	})

	// Generate Go code for AWS resource tagging support
	var goCode strings.Builder
	goCode.WriteString(`package aws

// Generated AWS resource tagging support matrix
// This file is auto-generated from AWS provider schema

var AWSResourceTaggingSupport = map[string]bool{
`)

	for _, resource := range resources {
		goCode.WriteString(fmt.Sprintf("\t\"%s\": %t,\n", resource.ResourceType, resource.SupportsTagAttribute))
	}

	goCode.WriteString("}\n\n")
	
	// Add service mapping
	goCode.WriteString("var AWSResourceToService = map[string]string{\n")
	for _, resource := range resources {
		goCode.WriteString(fmt.Sprintf("\t\"%s\": \"%s\",\n", resource.ResourceType, resource.Service))
	}
	goCode.WriteString("}\n")

	// Write to file
	err := os.WriteFile("internal/aws/resource_tagging.go", []byte(goCode.String()), 0644)
	if err != nil {
		log.Printf("Failed to write resource tagging data: %v", err)
	} else {
		fmt.Printf("Generated resource tagging data file: internal/aws/resource_tagging.go\n\n")
	}

	// Generate JSON data for external tools
	jsonData, _ := json.MarshalIndent(resources, "", "  ")
	err = os.WriteFile("examples/aws-resource-tagging-support.json", jsonData, 0644)
	if err != nil {
		log.Printf("Failed to write JSON data: %v", err)
	} else {
		fmt.Printf("Generated JSON data file: examples/aws-resource-tagging-support.json\n\n")
	}
}