package providers

import (
	"testing"
)

func TestGetTagIdByResource(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		expected     string
	}{
		{
			name:         "AWS instance",
			resourceType: "aws_instance",
			expected:     "tags",
		},
		{
			name:         "AWS S3 bucket",
			resourceType: "aws_s3_bucket",
			expected:     "tags",
		},
		{
			name:         "AWS ELB",
			resourceType: "aws_elb",
			expected:     "tags",
		},
		{
			name:         "AWS Auto Scaling Group",
			resourceType: "aws_autoscaling_group",
			expected:     "tag",
		},
		{
			name:         "Google compute instance",
			resourceType: "google_compute_instance",
			expected:     "labels",
		},
		{
			name:         "Google storage bucket",
			resourceType: "google_storage_bucket",
			expected:     "labels",
		},
		{
			name:         "Azure resource group",
			resourceType: "azurerm_resource_group",
			expected:     "tags",
		},
		{
			name:         "Azure virtual machine",
			resourceType: "azurerm_virtual_machine",
			expected:     "tags",
		},
		{
			name:         "Unknown resource",
			resourceType: "unknown_resource",
			expected:     "tags",
		},
		{
			name:         "Local file",
			resourceType: "local_file",
			expected:     "tags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTagIdByResource(tt.resourceType)
			if result != tt.expected {
				t.Errorf("GetTagIdByResource(%s) = %v, want %v", tt.resourceType, result, tt.expected)
			}
		})
	}
}

func TestIsResourceTaggable(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		expected     bool
	}{
		{
			name:         "AWS instance is taggable",
			resourceType: "aws_instance",
			expected:     true,
		},
		{
			name:         "AWS S3 bucket is taggable",
			resourceType: "aws_s3_bucket",
			expected:     true,
		},
		{
			name:         "Google compute instance supports labels",
			resourceType: "google_compute_instance",
			expected:     true,
		},
		{
			name:         "Azure resource group supports tags",
			resourceType: "azurerm_resource_group",
			expected:     true,
		},
		{
			name:         "Local file is not taggable",
			resourceType: "local_file",
			expected:     false,
		},
		{
			name:         "Random string is not taggable",
			resourceType: "random_string",
			expected:     false,
		},
		{
			name:         "Data source is not taggable",
			resourceType: "data.aws_ami",
			expected:     false,
		},
		{
			name:         "Terraform resource is not taggable",
			resourceType: "terraform_remote_state",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsResourceTaggable(tt.resourceType)
			if result != tt.expected {
				t.Errorf("IsResourceTaggable(%s) = %v, want %v", tt.resourceType, result, tt.expected)
			}
		})
	}
}

func TestGetProviderByResource(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		expected     string
	}{
		{
			name:         "AWS resource",
			resourceType: "aws_instance",
			expected:     "aws",
		},
		{
			name:         "Google resource",
			resourceType: "google_compute_instance",
			expected:     "google",
		},
		{
			name:         "Azure resource",
			resourceType: "azurerm_resource_group",
			expected:     "azure",
		},
		{
			name:         "AzureStack resource",
			resourceType: "azurestack_virtual_machine",
			expected:     "azure",
		},
		{
			name:         "Unknown resource",
			resourceType: "unknown_resource",
			expected:     "unknown",
		},
		{
			name:         "Local resource",
			resourceType: "local_file",
			expected:     "local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetProviderByResource(tt.resourceType)
			if result != tt.expected {
				t.Errorf("GetProviderByResource(%s) = %v, want %v", tt.resourceType, result, tt.expected)
			}
		})
	}
}

func TestGetResourceCategory(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		expected     string
	}{
		{
			name:         "AWS compute resource",
			resourceType: "aws_instance",
			expected:     "compute",
		},
		{
			name:         "AWS storage resource",
			resourceType: "aws_s3_bucket",
			expected:     "storage",
		},
		{
			name:         "AWS network resource",
			resourceType: "aws_vpc",
			expected:     "network",
		},
		{
			name:         "AWS security resource",
			resourceType: "aws_security_group",
			expected:     "security",
		},
		{
			name:         "AWS database resource",
			resourceType: "aws_db_instance",
			expected:     "database",
		},
		{
			name:         "Unknown resource",
			resourceType: "unknown_resource",
			expected:     "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetResourceCategory(tt.resourceType)
			if result != tt.expected {
				t.Errorf("GetResourceCategory(%s) = %v, want %v", tt.resourceType, result, tt.expected)
			}
		})
	}
}

func TestSpecialTagHandling(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		description  string
	}{
		{
			name:         "AWS Auto Scaling Group uses singular tag",
			resourceType: "aws_autoscaling_group",
			description:  "Uses 'tag' blocks instead of 'tags' attribute",
		},
		{
			name:         "AWS ECS Service uses tags",
			resourceType: "aws_ecs_service",
			description:  "Uses standard 'tags' attribute",
		},
		{
			name:         "Google resources use labels",
			resourceType: "google_compute_instance",
			description:  "Uses 'labels' instead of 'tags'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tagId := GetTagIdByResource(tt.resourceType)
			
			switch tt.resourceType {
			case "aws_autoscaling_group":
				if tagId != "tag" {
					t.Errorf("Expected 'tag' for %s, got %s", tt.resourceType, tagId)
				}
			case "google_compute_instance":
				if tagId != "labels" {
					t.Errorf("Expected 'labels' for %s, got %s", tt.resourceType, tagId)
				}
			default:
				if tagId != "tags" {
					t.Errorf("Expected 'tags' for %s, got %s", tt.resourceType, tagId)
				}
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		description  string
	}{
		{
			name:         "empty string",
			resourceType: "",
			description:  "Should handle empty resource type gracefully",
		},
		{
			name:         "mixed case",
			resourceType: "AWS_Instance",
			description:  "Should handle mixed case resource types",
		},
		{
			name:         "with spaces",
			resourceType: " aws_instance ",
			description:  "Should handle resource types with spaces",
		},
		{
			name:         "data source",
			resourceType: "data.aws_ami.latest",
			description:  "Should handle data sources",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// These should not panic and should return reasonable defaults
			tagId := GetTagIdByResource(tt.resourceType)
			if tagId == "" {
				t.Errorf("GetTagIdByResource(%s) returned empty string", tt.resourceType)
			}
			
			isTaggable := IsResourceTaggable(tt.resourceType)
			if tt.resourceType == "" || tt.resourceType == " aws_instance " {
				// Empty or malformed types should not be taggable
				if isTaggable {
					t.Errorf("IsResourceTaggable(%s) should return false", tt.resourceType)
				}
			}
		})
	}
}