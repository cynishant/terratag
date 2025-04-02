package tfschema

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"

	"github.com/env0/terratag/internal/common"
	"github.com/env0/terratag/internal/providers"
	"github.com/env0/terratag/internal/tagging"
	"github.com/env0/terratag/internal/terraform"
	"github.com/thoas/go-funk"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

var ErrResourceTypeNotFound = errors.New("resource type not found")

var providerSchemasMap map[string]*ProviderSchemas = map[string]*ProviderSchemas{}

var customSupportedProviderNames = [...]string{"google-beta"}

// jsonRegex matches JSON objects in mixed output
var jsonRegex = regexp.MustCompile(`(?s)\{.*?\}`)

type Attribute struct {
	Type      cty.Type `json:"type"`
	Required  bool     `json:"required"`
	Optional  bool     `json:"optional"`
	Computed  bool     `json:"computed"`
	Sensitive bool     `json:"sensitive"`
}

type Block struct {
	Attributes map[string]*Attribute `json:"attributes"`
}

type ResourceSchema struct {
	Block Block `json:"block"`
}

type ProviderSchema struct {
	ResourceSchemas map[string]*ResourceSchema `json:"resource_schemas"`
}

type ProviderSchemas struct {
	ProviderSchemas map[string]*ProviderSchema `json:"provider_schemas"`
}

// InitProviderSchemas fetches and stores the provider schemas for a directory
// This can be called ahead of time to pre-populate the schemas cache
func InitProviderSchemas(dir string, iacType common.IACType, defaultToTerraform bool, isRunAll bool) error {
	// Use tofu by default (if it exists).
	name := "terraform"
	// For terragrunt - use terragrunt.
	if iacType == common.Terragrunt {
		name = "terragrunt"
	} else if _, err := exec.LookPath("tofu"); !defaultToTerraform && err == nil {
		name = "tofu"
	}

	log.Print("[INFO] Fetching provider schemas for directory: ", dir)

	var cmd *exec.Cmd
	if iacType == common.Terragrunt && isRunAll {
		log.Print("[INFO] Using terragrunt run-all mode")
		cmd = exec.Command(name, "run-all", "providers", "schema", "-json")
	} else {
		cmd = exec.Command(name, "providers", "schema", "-json")
	}
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && ee.Stderr != nil {
			log.Println("===============================================")
			log.Printf("Error output: %s\n", string(ee.Stderr))
			log.Println("===============================================")
		}

		log.Println("===============================================")
		log.Printf("Standard output: %s\n", string(out))
		log.Println("===============================================")

		return fmt.Errorf("failed to execute '%s providers schema -json' command in directory '%s': %w", name, dir, err)
	}

	// Create a new provider schemas object
	mergedProviderSchemas := &ProviderSchemas{
		ProviderSchemas: make(map[string]*ProviderSchema),
	}

	if iacType == common.Terragrunt && isRunAll {
		// In run-all mode, we need to parse multiple JSON objects from the output
		jsonMatches := jsonRegex.FindAll(out, -1)
		log.Printf("[INFO] Found %d JSON schema objects in terragrunt run-all output", len(jsonMatches))

		for i, jsonData := range jsonMatches {
			log.Printf("[INFO] Processing schema object %d", i+1)

			providerSchemas := &ProviderSchemas{}
			if err := json.Unmarshal(jsonData, providerSchemas); err != nil {
				log.Printf("[WARN] Failed to unmarshal schema object %d: %v", i+1, err)
				continue
			}

			// Merge this schema into our accumulated schemas
			mergeProviderSchemas(mergedProviderSchemas, providerSchemas)
		}
	} else {
		// Standard mode - just parse the single JSON object
		// Output can vary between operating systems. Get the correct output line.
		for _, line := range bytes.Split(out, []byte("\n")) {
			if len(line) > 0 && line[0] == '{' {
				out = line
				break
			}
		}

		if err := json.Unmarshal(out, mergedProviderSchemas); err != nil {
			if e, ok := err.(*json.SyntaxError); ok {
				log.Printf("syntax error at byte offset %d", e.Offset)
			}
			return fmt.Errorf("failed to unmarshal returned provider schemas: %w", err)
		}
	}

	providerSchemasMap[dir] = mergedProviderSchemas
	log.Printf("[INFO] Successfully initialized provider schemas for directory: %s with %d providers",
		dir, len(mergedProviderSchemas.ProviderSchemas))

	return nil
}

// mergeProviderSchemas merges the source provider schemas into the target
func mergeProviderSchemas(target, source *ProviderSchemas) {
	for providerName, providerSchema := range source.ProviderSchemas {
		// If this provider already exists in the target, merge their resource schemas
		if existingProvider, exists := target.ProviderSchemas[providerName]; exists {
			if existingProvider.ResourceSchemas == nil {
				existingProvider.ResourceSchemas = make(map[string]*ResourceSchema)
			}

			// Copy all resource schemas from source to target
			for resourceType, resourceSchema := range providerSchema.ResourceSchemas {
				existingProvider.ResourceSchemas[resourceType] = resourceSchema
			}
		} else {
			// Otherwise, just add this provider to the target
			target.ProviderSchemas[providerName] = providerSchema
		}
	}
}

// IsTaggable checks if a resource can be tagged
func IsTaggable(dir string, resource hclwrite.Block) (bool, error) {
	var isTaggable bool

	resourceType := terraform.GetResourceType(resource)

	if providers.IsSupportedResource(resourceType, resource) {
		resourceSchema, err := getResourceSchema(resourceType, resource, dir)
		if err != nil {
			if errors.Is(err, ErrResourceTypeNotFound) {
				log.Print("[WARN] Skipped ", resourceType, " as it is not YET supported")

				return false, nil
			}

			return false, err
		}

		for attribute := range resourceSchema.Block.Attributes {
			if providers.IsTaggableByAttribute(resourceType, attribute) {
				isTaggable = true
			}
		}

		if tagging.HasResourceTagFn(resourceType) {
			isTaggable = true
		}
	}

	return isTaggable, nil
}

type TfSchemaAttribute struct {
	Name string
	Type string
}

func extractProviderNameFromResourceType(resourceType string) (string, error) {
	s := strings.SplitN(resourceType, "_", 2)
	if len(s) < 2 {
		return "", fmt.Errorf("failed to detect a provider name: %s", resourceType)
	}

	return s[0], nil
}

func detectProviderName(resource hclwrite.Block) (string, error) {
	providerAttribute := resource.Body().GetAttribute("provider")

	if providerAttribute != nil {
		providerTokens := providerAttribute.Expr().BuildTokens(hclwrite.Tokens{})
		providerName := strings.Trim(string(providerTokens.Bytes()), "\" ")

		if funk.Contains(customSupportedProviderNames, providerName) {
			return providerName, nil
		}
	}

	return extractProviderNameFromResourceType(terraform.GetResourceType(resource))
}

func getResourceSchema(resourceType string, resource hclwrite.Block, dir string) (*ResourceSchema, error) {
	detectedProviderName, err := detectProviderName(resource)
	if err != nil {
		return nil, fmt.Errorf("failed to detect provider name for resource %s: %w", resourceType, err)
	}

	providerSchemas := providerSchemasMap[dir]

	// Search through all providers.
	for providerName, providerSchema := range providerSchemas.ProviderSchemas {
		if len(detectedProviderName) > 0 && providerName != detectedProviderName && !strings.HasSuffix(providerName, "/"+detectedProviderName) {
			// Not the correct provider (based on name). Skip.
			continue
		}

		resourceSchema, ok := providerSchema.ResourceSchemas[resourceType]
		if ok {
			return resourceSchema, nil
		}
	}

	return nil, ErrResourceTypeNotFound
}
