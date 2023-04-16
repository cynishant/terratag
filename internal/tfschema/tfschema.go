package tfschema

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

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
var providerSchemasMapLock sync.Mutex

var customSupportedProviderNames = [...]string{"google-beta"}

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

func IsTaggable(dir string, iacType common.IACType, resource hclwrite.Block) (bool, error) {
	var isTaggable bool
	resourceType := terraform.GetResourceType(resource)

	if providers.IsSupportedResource(resourceType) {
		resourceSchema, err := getResourceSchema(resourceType, resource, dir, iacType)
		if err != nil {
			if err == ErrResourceTypeNotFound {
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

func getTerragruntPluginPath(dir string) string {
	dir += "/.terragrunt-cache"
	ret := dir
	found := false

	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if found || err != nil {
			return filepath.SkipDir
		}

		// E.g. ./.terragrunt-cache/yHtqnMrVQOISIYxobafVvZbAAyU/ThyYwttwki6d6AS3aD5OwoyqIWA/.terraform
		if strings.HasSuffix(path, "/.terraform") {
			ret = strings.TrimSuffix(path, "/.terraform")
			found = true
		}

		return nil
	})

	return ret
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

func getResourceSchema(resourceType string, resource hclwrite.Block, dir string, iacType common.IACType) (*ResourceSchema, error) {
	if iacType == common.Terragrunt {
		// which mode of terragrunt it is (with or without cache folder).
		if _, err := os.Stat(dir + "/.terragrunt-cache"); err == nil {
			dir = getTerragruntPluginPath(dir)
		}
	}

	providerSchemasMapLock.Lock()
	defer providerSchemasMapLock.Unlock()

	providerSchemas, ok := providerSchemasMap[dir]
	if !ok {
		providerSchemas = &ProviderSchemas{}

		cmd := exec.Command("terraform", "providers", "schema", "-json")
		cmd.Dir = dir

		out, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to execute 'terraform providers schema -json' command: %w", err)
		}

		// Output can vary between operating systems. Get the correct output line.
		for _, line := range bytes.Split(out, []byte("\n")) {
			if len(line) > 0 && line[0] == '{' {
				out = line
				break
			}
		}

		if err := json.Unmarshal(out, providerSchemas); err != nil {
			if e, ok := err.(*json.SyntaxError); ok {
				log.Printf("syntax error at byte offset %d", e.Offset)
			}
			return nil, fmt.Errorf("failed to unmarshal returned provider schemas: %w", err)
		}

		providerSchemasMap[dir] = providerSchemas
	}

	detectedProviderName, _ := detectProviderName(resource)
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
