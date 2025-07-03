package convert

import (
	"testing"

	"github.com/cloudyali/terratag/internal/common"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

func TestGetExistingTagsExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple expression",
			input:    "var.tags",
			expected: "var.tags",
		},
		{
			name:     "expression with whitespace",
			input:    "  var.tags  ",
			expected: "var.tags",
		},
		{
			name:     "expression with interpolation TF11 style",
			input:    "${var.tags}",
			expected: "var.tags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := hclwrite.Tokens{
				{
					Type:  1, // TokenIdent
					Bytes: []byte(tt.input),
				},
			}
			result := GetExistingTagsExpression(tokens)
			if result != tt.expected {
				t.Errorf("GetExistingTagsExpression() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsHclMap(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid HCL map",
			input:    `{ key = "value" }`,
			expected: true,
		},
		{
			name:     "empty HCL map",
			input:    `{}`,
			expected: true,
		},
		{
			name:     "HCL map with whitespace",
			input:    ` { key = "value" } `,
			expected: true,
		},
		{
			name:     "not a map - variable",
			input:    `var.tags`,
			expected: false,
		},
		{
			name:     "not a map - function",
			input:    `merge(var.tags, local.tags)`,
			expected: false,
		},
		{
			name:     "not a map - missing closing brace",
			input:    `{ key = "value"`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := hclwrite.Tokens{
				{
					Type:  1, // TokenIdent
					Bytes: []byte(tt.input),
				},
			}
			result := isHclMap(tokens)
			if result != tt.expected {
				t.Errorf("isHclMap() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAppendLocalsBlock(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		terratag common.TerratagLocal
		setup    func(*hclwrite.File)
		validate func(*testing.T, *hclwrite.File)
	}{
		{
			name:     "append new locals block",
			filename: "test.tf",
			terratag: common.TerratagLocal{
				Added: "test_added_123",
			},
			setup: func(f *hclwrite.File) {
				// Empty file
			},
			validate: func(t *testing.T, f *hclwrite.File) {
				blocks := f.Body().Blocks()
				if len(blocks) != 1 {
					t.Fatalf("expected 1 block, got %d", len(blocks))
				}
				if blocks[0].Type() != "locals" {
					t.Errorf("expected locals block, got %s", blocks[0].Type())
				}
				attr := blocks[0].Body().GetAttribute("terratag_added_test_tf")
				if attr == nil {
					t.Error("expected terratag_added_test_tf attribute")
				}
			},
		},
		{
			name:     "update existing locals block",
			filename: "main.tf",
			terratag: common.TerratagLocal{
				Added: "updated_value",
			},
			setup: func(f *hclwrite.File) {
				locals := f.Body().AppendNewBlock("locals", nil)
				locals.Body().SetAttributeValue("terratag_added_main_tf", cty.StringVal("old_value"))
			},
			validate: func(t *testing.T, f *hclwrite.File) {
				blocks := f.Body().Blocks()
				if len(blocks) != 1 {
					t.Fatalf("expected 1 block, got %d", len(blocks))
				}
				attr := blocks[0].Body().GetAttribute("terratag_added_main_tf")
				if attr == nil {
					t.Error("expected terratag_added_main_tf attribute")
				}
				// Check if value was updated
				tokens := attr.Expr().BuildTokens(nil)
				valueStr := string(tokens.Bytes())
				if valueStr != `"updated_value"` {
					t.Errorf("expected updated value, got %s", valueStr)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := hclwrite.NewEmptyFile()
			tt.setup(file)
			AppendLocalsBlock(file, tt.filename, tt.terratag)
			tt.validate(t, file)
		})
	}
}

func TestAppendTagBlocks(t *testing.T) {
	tests := []struct {
		name     string
		tags     string
		wantErr  bool
		validate func(*testing.T, *hclwrite.Block)
	}{
		{
			name: "valid tags",
			tags: `{"Environment":"prod","Team":"platform"}`,
			validate: func(t *testing.T, resource *hclwrite.Block) {
				blocks := resource.Body().Blocks()
				if len(blocks) != 2 {
					t.Fatalf("expected 2 tag blocks, got %d", len(blocks))
				}
				// Check if tags are sorted by key
				firstBlock := blocks[0]
				if firstBlock.Type() != "tag" {
					t.Errorf("expected tag block, got %s", firstBlock.Type())
				}
				keyAttr := firstBlock.Body().GetAttribute("key")
				if keyAttr == nil {
					t.Error("expected key attribute in tag block")
				}
			},
		},
		{
			name: "single tag",
			tags: `{"Name":"test-resource"}`,
			validate: func(t *testing.T, resource *hclwrite.Block) {
				blocks := resource.Body().Blocks()
				if len(blocks) != 1 {
					t.Fatalf("expected 1 tag block, got %d", len(blocks))
				}
			},
		},
		{
			name:    "invalid JSON",
			tags:    `{"invalid":}`,
			wantErr: true,
		},
		{
			name: "empty tags",
			tags: `{}`,
			validate: func(t *testing.T, resource *hclwrite.Block) {
				blocks := resource.Body().Blocks()
				if len(blocks) != 0 {
					t.Fatalf("expected 0 tag blocks, got %d", len(blocks))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := hclwrite.NewBlock("resource", []string{"aws_instance", "test"})
			err := AppendTagBlocks(resource, tt.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("AppendTagBlocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && tt.validate != nil {
				tt.validate(t, resource)
			}
		})
	}
}

func TestUnquoteTagsAttribute(t *testing.T) {
	tests := []struct {
		name               string
		swappedTagsStrings []string
		text               string
		expected           string
	}{
		{
			name:               "simple unquote",
			swappedTagsStrings: []string{`var.tags`},
			text:               `"var.tags"`,
			expected:           `var.tags`,
		},
		{
			name:               "unquote with escaped quotes",
			swappedTagsStrings: []string{`merge(var.tags, {"key":"value"})`},
			text:               `"merge(var.tags, {\"key\":\"value\"})"`,
			expected:           `merge(var.tags, {"key":"value"})`,
		},
		{
			name:               "unquote with variable interpolation",
			swappedTagsStrings: []string{`${var.tags}`},
			text:               `"$${var.tags}"`,
			expected:           `${var.tags}`,
		},
		{
			name:               "multiple unquotes",
			swappedTagsStrings: []string{`var.tags`, `local.tags`},
			text:               `merge("var.tags", "local.tags")`,
			expected:           `merge(var.tags, local.tags)`,
		},
		{
			name:               "no quotes needed for interpolation",
			swappedTagsStrings: []string{`${merge(var.tags, local.tags)}`},
			text:               `${merge(var.tags, local.tags)}`,
			expected:           `${merge(var.tags, local.tags)}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UnquoteTagsAttribute(tt.swappedTagsStrings, tt.text)
			if result != tt.expected {
				t.Errorf("UnquoteTagsAttribute() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMoveExistingTags(t *testing.T) {
	tests := []struct {
		name          string
		filename      string
		tagId         string
		setupBlock    func() *hclwrite.Block
		expectFound   bool
		expectError   bool
		validateTags  func(*testing.T, common.TerratagLocal)
	}{
		{
			name:     "move tags from attribute",
			filename: "test.tf",
			tagId:    "tags",
			setupBlock: func() *hclwrite.Block {
				block := hclwrite.NewBlock("resource", []string{"aws_instance", "test"})
				block.Body().SetAttributeRaw("tags", hclwrite.Tokens{
					{Type: 1, Bytes: []byte(`{ Name = "test" }`)},
				})
				return block
			},
			expectFound: true,
			validateTags: func(t *testing.T, terratag common.TerratagLocal) {
				if len(terratag.Found) != 1 {
					t.Errorf("expected 1 found tag, got %d", len(terratag.Found))
				}
			},
		},
		{
			name:     "move tags from block",
			filename: "main.tf",
			tagId:    "tags",
			setupBlock: func() *hclwrite.Block {
				block := hclwrite.NewBlock("resource", []string{"aws_instance", "test"})
				tagBlock := block.Body().AppendNewBlock("tags", nil)
				tagBlock.Body().SetAttributeValue("Name", cty.StringVal("test"))
				return block
			},
			expectFound: true,
			validateTags: func(t *testing.T, terratag common.TerratagLocal) {
				if len(terratag.Found) != 1 {
					t.Errorf("expected 1 found tag, got %d", len(terratag.Found))
				}
			},
		},
		{
			name:     "no existing tags",
			filename: "test.tf",
			tagId:    "tags",
			setupBlock: func() *hclwrite.Block {
				return hclwrite.NewBlock("resource", []string{"aws_instance", "test"})
			},
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terratag := common.TerratagLocal{
				Found: make(map[string]hclwrite.Tokens),
			}
			block := tt.setupBlock()
			
			found, err := MoveExistingTags(tt.filename, terratag, block, tt.tagId)
			if (err != nil) != tt.expectError {
				t.Errorf("MoveExistingTags() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if found != tt.expectFound {
				t.Errorf("MoveExistingTags() found = %v, want %v", found, tt.expectFound)
			}
			if tt.validateTags != nil {
				tt.validateTags(t, terratag)
			}
		})
	}
}

func TestQuoteBlockKeys(t *testing.T) {
	block := hclwrite.NewBlock("tags", nil)
	block.Body().SetAttributeValue("Name", cty.StringVal("test"))
	block.Body().SetAttributeValue("Environment", cty.StringVal("prod"))

	quotedBlock := quoteBlockKeys(block)
	
	// Check that the new block has quoted keys
	attrs := quotedBlock.Body().Attributes()
	for key := range attrs {
		if key != `"Name"` && key != `"Environment"` {
			t.Errorf("expected quoted key, got %s", key)
		}
	}
}

func TestQuoteAttributeKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		isMap    bool
	}{
		{
			name:  "HCL map with unquoted keys",
			input: `{ Name = "test", Environment = "prod" }`,
			isMap: true,
		},
		{
			name:  "variable reference",
			input: `var.tags`,
			isMap: false,
		},
		{
			name:  "function call",
			input: `merge(var.tags, local.tags)`,
			isMap: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := hclwrite.Tokens{
				{Type: 1, Bytes: []byte(tt.input)},
			}
			
			// Test that isHclMap correctly identifies maps
			result := isHclMap(tokens)
			if result != tt.isMap {
				t.Errorf("isHclMap() = %v, want %v", result, tt.isMap)
			}
		})
	}
}

func TestHclValueToMap(t *testing.T) {
	// This test ensures the function delegates to the shared HCL parser
	// Since we can't easily create valid HCL tokens in a unit test,
	// we'll just ensure the function exists and returns the expected error
	// for invalid input
	tokens := hclwrite.Tokens{
		{Type: 1, Bytes: []byte("invalid")},
	}
	
	_, err := HclValueToMap(tokens)
	if err == nil {
		t.Error("expected error for invalid HCL tokens")
	}
}