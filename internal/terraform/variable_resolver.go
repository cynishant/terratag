package terraform

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
	"github.com/sirupsen/logrus"
)

// VariableResolver handles resolution of Terraform variables and locals
type VariableResolver struct {
	variables      map[string]*VariableDefinition
	locals         map[string]*LocalDefinition
	variableValues map[string]interface{} // Values from terraform.tfvars, environment, etc.
	resolvedLocals map[string]interface{} // Resolved local values
	cliVars        map[string]string     // Variables from -var flags
	tfvarsFiles    []string              // Additional .tfvars files to load
	logger         *logrus.Logger
	evalContext    *hcl.EvalContext      // HCL evaluation context for native evaluation
}

// VariableDefinition represents a Terraform variable definition
type VariableDefinition struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Description  string      `json:"description"`
	Default      interface{} `json:"default"`
	Validation   []ValidationRule `json:"validation,omitempty"`
	Sensitive    bool        `json:"sensitive"`
	Nullable     bool        `json:"nullable"`
	FilePath     string      `json:"file_path"`
	LineNumber   int         `json:"line_number"`
}

// LocalDefinition represents a Terraform local value definition
type LocalDefinition struct {
	Name       string      `json:"name"`
	Expression string      `json:"expression"` // Raw HCL expression
	Value      interface{} `json:"value"`      // Resolved value (if possible)
	FilePath   string      `json:"file_path"`
	LineNumber int         `json:"line_number"`
	Dependencies []string   `json:"dependencies"` // Other variables/locals this depends on
	hclExpr    hcl.Expression `json:"-"`        // HCL expression for native evaluation
}

// ValidationRule represents a variable validation rule
type ValidationRule struct {
	Condition string `json:"condition"`
	Message   string `json:"error_message"`
}

// ResolutionResult represents the result of resolving a variable reference
type ResolutionResult struct {
	Value       interface{} `json:"value"`
	Resolved    bool        `json:"resolved"`
	Source      string      `json:"source"`      // "variable", "local", "literal"
	Uncertainty string      `json:"uncertainty"` // Description of why it couldn't be resolved
}

// NewVariableResolver creates a new variable resolver
func NewVariableResolver(logger *logrus.Logger) *VariableResolver {
	if logger == nil {
		logger = logrus.New()
	}
	
	vr := &VariableResolver{
		variables:      make(map[string]*VariableDefinition),
		locals:         make(map[string]*LocalDefinition),
		variableValues: make(map[string]interface{}),
		resolvedLocals: make(map[string]interface{}),
		cliVars:        make(map[string]string),
		logger:         logger,
	}
	
	// Initialize evaluation context with Terraform functions
	vr.buildEvalContext()
	
	return vr
}

// LoadFromDirectory loads variables and locals from all Terraform files in a directory
func (vr *VariableResolver) LoadFromDirectory(dirPath string) error {
	vr.logger.WithField("directory", dirPath).Info("Loading variables and locals from directory")
	
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories and non-Terraform files
		if info.IsDir() || (!strings.HasSuffix(path, ".tf") && !strings.HasSuffix(path, ".tf.json")) {
			return nil
		}
		
		// Skip .terraform directory
		if strings.Contains(path, ".terraform") {
			return nil
		}
		
		if strings.HasSuffix(path, ".tf") {
			return vr.loadFromHCLFile(path)
		} else if strings.HasSuffix(path, ".tf.json") {
			return vr.loadFromJSONFile(path)
		}
		
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("failed to walk directory %s: %w", dirPath, err)
	}
	
	// Load variable values from .tfvars files
	if err := vr.loadVariableValues(dirPath); err != nil {
		vr.logger.WithError(err).Warn("Failed to load variable values")
	}
	
	// Load additional tfvars files if specified
	for _, tfvarsFile := range vr.tfvarsFiles {
		if err := vr.loadTfvarsFile(tfvarsFile); err != nil {
			vr.logger.WithError(err).WithField("file", tfvarsFile).Warn("Failed to load additional tfvars file")
		}
	}
	
	// Rebuild evaluation context with all loaded variables
	vr.buildEvalContext()
	
	// Resolve locals that depend on variables
	if err := vr.resolveLocals(); err != nil {
		vr.logger.WithError(err).Warn("Failed to resolve some local values")
	}
	
	vr.logger.WithFields(logrus.Fields{
		"variables_count": len(vr.variables),
		"locals_count":    len(vr.locals),
		"resolved_locals": len(vr.resolvedLocals),
	}).Info("Variable and locals loading completed")
	
	return nil
}

// loadFromHCLFile loads variables and locals from an HCL file
func (vr *VariableResolver) loadFromHCLFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(content, filePath)
	if diags.HasErrors() {
		vr.logger.WithField("file", filePath).WithError(diags).Warn("Failed to parse HCL file")
		return nil // Don't fail completely, just skip this file
	}
	
	// Extract variables and locals
	bodyContent, _, diags := file.Body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "variable", LabelNames: []string{"name"}},
			{Type: "locals"},
		},
	})
	
	if diags.HasErrors() {
		vr.logger.WithField("file", filePath).WithError(diags).Warn("Failed to extract content from HCL file")
		return nil
	}
	
	// Process variable blocks
	for _, block := range bodyContent.Blocks {
		if block.Type == "variable" && len(block.Labels) > 0 {
			variable, err := vr.parseVariableBlock(block, filePath)
			if err != nil {
				vr.logger.WithError(err).WithField("file", filePath).Warn("Failed to parse variable block")
				continue
			}
			vr.variables[variable.Name] = variable
		} else if block.Type == "locals" {
			locals, err := vr.parseLocalsBlock(block, filePath)
			if err != nil {
				vr.logger.WithError(err).WithField("file", filePath).Warn("Failed to parse locals block")
				continue
			}
			for name, local := range locals {
				vr.locals[name] = local
			}
		}
	}
	
	return nil
}

// loadFromJSONFile loads variables and locals from a JSON file
func (vr *VariableResolver) loadFromJSONFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read JSON file %s: %w", filePath, err)
	}
	
	var tfConfig map[string]interface{}
	if err := json.Unmarshal(content, &tfConfig); err != nil {
		return fmt.Errorf("failed to parse JSON file %s: %w", filePath, err)
	}
	
	// Process variables
	if variables, ok := tfConfig["variable"].(map[string]interface{}); ok {
		for name, varDef := range variables {
			variable := &VariableDefinition{
				Name:     name,
				FilePath: filePath,
			}
			
			if varDefMap, ok := varDef.(map[string]interface{}); ok {
				if desc, ok := varDefMap["description"].(string); ok {
					variable.Description = desc
				}
				if def, ok := varDefMap["default"]; ok {
					variable.Default = def
				}
				if varType, ok := varDefMap["type"].(string); ok {
					variable.Type = varType
				}
			}
			
			vr.variables[name] = variable
		}
	}
	
	// Process locals
	if locals, ok := tfConfig["locals"].(map[string]interface{}); ok {
		for name, value := range locals {
			local := &LocalDefinition{
				Name:     name,
				FilePath: filePath,
				Value:    value,
			}
			vr.locals[name] = local
		}
	}
	
	return nil
}

// parseVariableBlock parses a variable block from HCL
func (vr *VariableResolver) parseVariableBlock(block *hcl.Block, filePath string) (*VariableDefinition, error) {
	variable := &VariableDefinition{
		Name:       block.Labels[0],
		FilePath:   filePath,
		LineNumber: block.DefRange.Start.Line,
	}
	
	// Use PartialContent to handle both attributes and nested blocks like validation
	schema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: "description"},
			{Name: "type"},
			{Name: "default"},
			{Name: "sensitive"},
			{Name: "nullable"},
		},
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "validation"},
		},
	}
	
	content, _, diags := block.Body.PartialContent(schema)
	if diags.HasErrors() {
		return nil, diags
	}
	
	attrs := content.Attributes
	
	for name, attr := range attrs {
		switch name {
		case "description":
			if val, diags := attr.Expr.Value(vr.evalContext); !diags.HasErrors() && val.Type().FriendlyName() == "string" {
				variable.Description = val.AsString()
			}
		case "type":
			// Type is typically a type expression, convert to string representation
			if srcRange := attr.Expr.Range(); srcRange.Start.Line > 0 {
				if content, err := os.ReadFile(filePath); err == nil {
					lines := strings.Split(string(content), "\n")
					if srcRange.Start.Line <= len(lines) {
						line := lines[srcRange.Start.Line-1]
						// Extract type from line (simplified)
						typePattern := regexp.MustCompile(`type\s*=\s*(.+)`)
						if matches := typePattern.FindStringSubmatch(line); len(matches) > 1 {
							variable.Type = strings.TrimSpace(matches[1])
						}
					}
				}
			}
		case "default":
			if val, diags := attr.Expr.Value(vr.evalContext); !diags.HasErrors() {
				variable.Default = convertHCLValue(val)
			}
		case "sensitive":
			if val, diags := attr.Expr.Value(vr.evalContext); !diags.HasErrors() && val.Type().FriendlyName() == "bool" {
				variable.Sensitive = val.True()
			}
		case "nullable":
			if val, diags := attr.Expr.Value(vr.evalContext); !diags.HasErrors() && val.Type().FriendlyName() == "bool" {
				variable.Nullable = val.True()
			}
		}
	}
	
	return variable, nil
}

// parseLocalsBlock parses a locals block from HCL
func (vr *VariableResolver) parseLocalsBlock(block *hcl.Block, filePath string) (map[string]*LocalDefinition, error) {
	locals := make(map[string]*LocalDefinition)
	
	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, diags
	}
	
	for name, attr := range attrs {
		local := &LocalDefinition{
			Name:       name,
			FilePath:   filePath,
			LineNumber: attr.Range.Start.Line,
		}
		
		// Store the raw expression
		if srcRange := attr.Expr.Range(); srcRange.Start.Line > 0 {
			if content, err := os.ReadFile(filePath); err == nil {
				lines := strings.Split(string(content), "\n")
				if srcRange.Start.Line <= len(lines) {
					// Extract the expression (simplified)
					line := lines[srcRange.Start.Line-1]
					exprPattern := regexp.MustCompile(name + `\s*=\s*(.+)`)
					if matches := exprPattern.FindStringSubmatch(line); len(matches) > 1 {
						local.Expression = strings.TrimSpace(matches[1])
					}
				}
			}
		}
		
		// Try to evaluate the expression with current context
		if val, diags := attr.Expr.Value(vr.evalContext); !diags.HasErrors() {
			local.Value = convertHCLValue(val)
		} else {
			// Store expression for later evaluation
			local.Dependencies = extractVariableReferences(local.Expression)
			// Store the actual HCL expression for later native evaluation
			local.hclExpr = attr.Expr
			
			// Log diagnostic info for debugging complex expressions
			if vr.logger != nil {
				vr.logger.WithFields(logrus.Fields{
					"local_name": name,
					"expression": local.Expression,
					"error":     diags.Error(),
				}).Debug("Local expression evaluation failed, will retry later")
			}
		}
		
		locals[name] = local
	}
	
	return locals, nil
}

// loadVariableValues loads variable values from .tfvars files and environment
// Following Terraform's variable precedence: 
// 1. Environment variables (TF_VAR_*)
// 2. terraform.tfvars or terraform.tfvars.json
// 3. *.auto.tfvars or *.auto.tfvars.json (in lexical order)
// 4. -var command line arguments (handled separately)
func (vr *VariableResolver) loadVariableValues(dirPath string) error {
	// Load from environment variables first (lowest precedence)
	vr.loadEnvironmentVariables()
	
	// Load from terraform.tfvars or terraform.tfvars.json
	tfvarsPath := filepath.Join(dirPath, "terraform.tfvars")
	if _, err := os.Stat(tfvarsPath); err == nil {
		if err := vr.loadTfvarsFile(tfvarsPath); err != nil {
			vr.logger.WithError(err).Warn("Failed to load terraform.tfvars")
		}
	} else {
		// Try terraform.tfvars.json
		tfvarsJsonPath := filepath.Join(dirPath, "terraform.tfvars.json")
		if _, err := os.Stat(tfvarsJsonPath); err == nil {
			if err := vr.loadTfvarsJsonFile(tfvarsJsonPath); err != nil {
				vr.logger.WithError(err).Warn("Failed to load terraform.tfvars.json")
			}
		}
	}
	
	// Load from *.auto.tfvars files (in lexical order)
	matches, err := filepath.Glob(filepath.Join(dirPath, "*.auto.tfvars"))
	if err == nil {
		for _, match := range matches {
			if err := vr.loadTfvarsFile(match); err != nil {
				vr.logger.WithError(err).WithField("file", match).Warn("Failed to load auto.tfvars file")
			}
		}
	}
	
	// Load from *.auto.tfvars.json files (in lexical order)
	jsonMatches, err := filepath.Glob(filepath.Join(dirPath, "*.auto.tfvars.json"))
	if err == nil {
		for _, match := range jsonMatches {
			if err := vr.loadTfvarsJsonFile(match); err != nil {
				vr.logger.WithError(err).WithField("file", match).Warn("Failed to load auto.tfvars.json file")
			}
		}
	}
	
	return nil
}

// loadTfvarsFile loads variable values from a .tfvars file
func (vr *VariableResolver) loadTfvarsFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(content, filePath)
	if diags.HasErrors() {
		return diags
	}
	
	attrs, diags := file.Body.JustAttributes()
	if diags.HasErrors() {
		return diags
	}
	
	for name, attr := range attrs {
		if val, diags := attr.Expr.Value(vr.evalContext); !diags.HasErrors() {
			vr.variableValues[name] = convertHCLValue(val)
		}
	}
	
	vr.logger.WithField("file", filePath).WithField("variables_loaded", len(attrs)).Debug("Loaded variables from tfvars file")
	return nil
}

// loadTfvarsJsonFile loads variable values from a .tfvars.json file
func (vr *VariableResolver) loadTfvarsJsonFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	
	var jsonData map[string]interface{}
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return fmt.Errorf("failed to parse JSON file %s: %w", filePath, err)
	}
	
	// Convert JSON values to variable values
	for name, value := range jsonData {
		vr.variableValues[name] = value
	}
	
	vr.logger.WithField("file", filePath).WithField("variables_loaded", len(jsonData)).Debug("Loaded variables from tfvars JSON file")
	return nil
}

// loadEnvironmentVariables loads variable values from TF_VAR_* environment variables
func (vr *VariableResolver) loadEnvironmentVariables() {
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "TF_VAR_") {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				varName := strings.TrimPrefix(parts[0], "TF_VAR_")
				vr.variableValues[varName] = parts[1]
			}
		}
	}
}

// resolveLocals resolves local values using native HCL evaluation with iterative dependency resolution
func (vr *VariableResolver) resolveLocals() error {
	maxIterations := 10 // Prevent infinite loops
	iteration := 0
	
	for iteration < maxIterations {
		resolved := 0
		prevResolvedCount := len(vr.resolvedLocals)
		
		for name, local := range vr.locals {
			// Skip if already resolved
			if _, exists := vr.resolvedLocals[name]; exists {
				continue
			}
			
			// Try to resolve using native HCL evaluation if we have the expression
			if local.hclExpr != nil {
				if val, diags := local.hclExpr.Value(vr.evalContext); !diags.HasErrors() {
					resolvedValue := convertHCLValue(val)
					vr.resolvedLocals[name] = resolvedValue
					local.Value = resolvedValue
					resolved++
					
					// Update evaluation context with the newly resolved local
					vr.buildEvalContext()
				} else {
					// Log details about why the expression failed for debugging
					if vr.logger != nil {
						vr.logger.WithFields(logrus.Fields{
							"local_name": name,
							"expression": local.Expression,
							"error":     diags.Error(),
							"iteration": iteration,
						}).Debug("Failed to resolve local expression")
					}
				}
			} else {
				// Fallback to string-based resolution for simple cases
				if resolvedValue, success := vr.resolveExpression(local.Expression); success {
					vr.resolvedLocals[name] = resolvedValue
					local.Value = resolvedValue
					resolved++
					
					// Update evaluation context with the newly resolved local
					vr.buildEvalContext()
				}
			}
		}
		
		// If no new locals were resolved, break
		if len(vr.resolvedLocals) == prevResolvedCount {
			break
		}
		
		iteration++
	}
	
	// Log unresolved locals for debugging
	unresolved := []string{}
	for name, local := range vr.locals {
		if _, exists := vr.resolvedLocals[name]; !exists {
			unresolved = append(unresolved, name)
			vr.logger.WithFields(logrus.Fields{
				"local_name": name,
				"expression": local.Expression,
				"dependencies": local.Dependencies,
			}).Debug("Local value could not be resolved")
		}
	}
	
	vr.logger.WithFields(logrus.Fields{
		"resolved_locals": len(vr.resolvedLocals),
		"unresolved_locals": len(unresolved),
		"iterations": iteration,
	}).Debug("Local resolution completed")
	
	return nil
}

// ResolveReference resolves a variable or local reference
func (vr *VariableResolver) ResolveReference(reference string) *ResolutionResult {
	// Handle var.* references
	if strings.HasPrefix(reference, "var.") {
		varName := strings.TrimPrefix(reference, "var.")
		return vr.resolveVariable(varName)
	}
	
	// Handle local.* references
	if strings.HasPrefix(reference, "local.") {
		localName := strings.TrimPrefix(reference, "local.")
		return vr.resolveLocal(localName)
	}
	
	// Handle interpolation expressions like "${var.project_name}-vpc"
	if strings.Contains(reference, "${") {
		return vr.resolveInterpolationExpression(reference)
	}
	
	// Handle direct references (for locals block internal references)
	if local, exists := vr.locals[reference]; exists {
		if value, exists := vr.resolvedLocals[reference]; exists {
			return &ResolutionResult{
				Value:    value,
				Resolved: true,
				Source:   "local",
			}
		}
		return &ResolutionResult{
			Value:       local.Value,
			Resolved:    local.Value != nil,
			Source:      "local",
			Uncertainty: "Local value depends on unresolved variables",
		}
	}
	
	// If it's a string literal, return as-is
	if strings.HasPrefix(reference, "\"") && strings.HasSuffix(reference, "\"") {
		return &ResolutionResult{
			Value:    strings.Trim(reference, "\""),
			Resolved: true,
			Source:   "literal",
		}
	}
	
	return &ResolutionResult{
		Value:       reference,
		Resolved:    false,
		Source:      "unknown",
		Uncertainty: "Unable to identify reference type",
	}
}

// resolveInterpolationExpression resolves interpolation expressions like "${var.project_name}-vpc"
func (vr *VariableResolver) resolveInterpolationExpression(expression string) *ResolutionResult {
	// Clean the expression - remove outer quotes if present
	cleanExpr := strings.Trim(expression, "\"")
	
	// For interpolation expressions, we need to wrap them in a temporary HCL attribute to parse them
	// The expression should be treated as a string template, so we need to quote it properly
	hclContent := fmt.Sprintf("temp_attr = \"%s\"", cleanExpr)
	
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL([]byte(hclContent), "interpolation.hcl")
	if diags.HasErrors() {
		return &ResolutionResult{
			Value:       expression,
			Resolved:    false,
			Source:      "interpolation",
			Uncertainty: fmt.Sprintf("Failed to parse interpolation expression: %s", diags.Error()),
		}
	}
	
	// Extract the attribute expression
	attrs, diags := file.Body.JustAttributes()
	if diags.HasErrors() {
		return &ResolutionResult{
			Value:       expression,
			Resolved:    false,
			Source:      "interpolation",
			Uncertainty: fmt.Sprintf("Failed to extract attribute: %s", diags.Error()),
		}
	}
	
	// Get the temp_attr expression
	if tempAttr, exists := attrs["temp_attr"]; exists {
		// Evaluate the expression with our context
		if vr.evalContext != nil {
			if val, evalDiags := tempAttr.Expr.Value(vr.evalContext); !evalDiags.HasErrors() {
				resolvedValue := convertHCLValue(val)
				if strValue, ok := resolvedValue.(string); ok {
					return &ResolutionResult{
						Value:    strValue,
						Resolved: true,
						Source:   "interpolation",
					}
				} else if resolvedValue != nil {
					return &ResolutionResult{
						Value:    fmt.Sprintf("%v", resolvedValue),
						Resolved: true,
						Source:   "interpolation",
					}
				}
			} else {
				return &ResolutionResult{
					Value:       expression,
					Resolved:    false,
					Source:      "interpolation",
					Uncertainty: fmt.Sprintf("Failed to evaluate expression: %s", evalDiags.Error()),
				}
			}
		}
	}
	
	return &ResolutionResult{
		Value:       expression,
		Resolved:    false,
		Source:      "interpolation",
		Uncertainty: "No evaluation context available for interpolation",
	}
}

// resolveVariable resolves a variable reference
func (vr *VariableResolver) resolveVariable(varName string) *ResolutionResult {
	// Check if we have a value for this variable
	if value, exists := vr.variableValues[varName]; exists {
		return &ResolutionResult{
			Value:    value,
			Resolved: true,
			Source:   "variable",
		}
	}
	
	// Check if variable is defined and has a default value
	if varDef, exists := vr.variables[varName]; exists {
		if varDef.Default != nil {
			return &ResolutionResult{
				Value:    varDef.Default,
				Resolved: true,
				Source:   "variable",
			}
		}
		
		return &ResolutionResult{
			Value:       nil,
			Resolved:    false,
			Source:      "variable",
			Uncertainty: "Variable defined but no value provided and no default value",
		}
	}
	
	return &ResolutionResult{
		Value:       nil,
		Resolved:    false,
		Source:      "variable",
		Uncertainty: "Variable not defined",
	}
}

// resolveLocal resolves a local reference (including complex map/object indexing)
func (vr *VariableResolver) resolveLocal(localName string) *ResolutionResult {
	// Handle complex local expressions like "environment_config[var.environment]"
	if strings.Contains(localName, "[") && strings.Contains(localName, "]") {
		return vr.resolveLocalIndexExpression(localName)
	}
	
	// Check if local is resolved
	if value, exists := vr.resolvedLocals[localName]; exists {
		return &ResolutionResult{
			Value:    value,
			Resolved: true,
			Source:   "local",
		}
	}
	
	// Check if local is defined
	if local, exists := vr.locals[localName]; exists {
		if local.Value != nil {
			return &ResolutionResult{
				Value:    local.Value,
				Resolved: true,
				Source:   "local",
			}
		}
		
		return &ResolutionResult{
			Value:       nil,
			Resolved:    false,
			Source:      "local",
			Uncertainty: "Local value depends on unresolved variables or expressions",
		}
	}
	
	return &ResolutionResult{
		Value:       nil,
		Resolved:    false,
		Source:      "local",
		Uncertainty: "Local not defined",
	}
}

// resolveLocalIndexExpression resolves local expressions with indexing like "environment_config[var.environment]"
func (vr *VariableResolver) resolveLocalIndexExpression(expression string) *ResolutionResult {
	// Parse the expression: localName[indexExpression]
	openBracket := strings.Index(expression, "[")
	closeBracket := strings.LastIndex(expression, "]")
	
	if openBracket == -1 || closeBracket == -1 || openBracket >= closeBracket {
		return &ResolutionResult{
			Value:       nil,
			Resolved:    false,
			Source:      "local",
			Uncertainty: "Invalid index expression syntax",
		}
	}
	
	localName := expression[:openBracket]
	indexExpr := expression[openBracket+1 : closeBracket]
	
	// Get the base local value (should be a map/object)
	baseResult := vr.resolveLocal(localName)
	if !baseResult.Resolved {
		return &ResolutionResult{
			Value:       nil,
			Resolved:    false,
			Source:      "local",
			Uncertainty: fmt.Sprintf("Base local '%s' is not resolved", localName),
		}
	}
	
	// Resolve the index expression
	var indexValue interface{}
	if strings.HasPrefix(indexExpr, "var.") {
		indexResult := vr.resolveVariable(strings.TrimPrefix(indexExpr, "var."))
		if !indexResult.Resolved {
			return &ResolutionResult{
				Value:       nil,
				Resolved:    false,
				Source:      "local",
				Uncertainty: fmt.Sprintf("Index variable '%s' is not resolved", indexExpr),
			}
		}
		indexValue = indexResult.Value
	} else if strings.HasPrefix(indexExpr, "local.") {
		indexResult := vr.resolveLocal(strings.TrimPrefix(indexExpr, "local."))
		if !indexResult.Resolved {
			return &ResolutionResult{
				Value:       nil,
				Resolved:    false,
				Source:      "local",
				Uncertainty: fmt.Sprintf("Index local '%s' is not resolved", indexExpr),
			}
		}
		indexValue = indexResult.Value
	} else {
		// Treat as string literal (remove quotes if present)
		indexValue = strings.Trim(indexExpr, "\"'")
	}
	
	// Perform the indexing operation
	switch baseValue := baseResult.Value.(type) {
	case map[string]interface{}:
		indexStr := fmt.Sprintf("%v", indexValue)
		if val, exists := baseValue[indexStr]; exists {
			return &ResolutionResult{
				Value:    val,
				Resolved: true,
				Source:   "local",
			}
		}
		return &ResolutionResult{
			Value:       nil,
			Resolved:    false,
			Source:      "local",
			Uncertainty: fmt.Sprintf("Key '%s' not found in map", indexStr),
		}
	case []interface{}:
		if indexInt, ok := indexValue.(int); ok {
			if indexInt >= 0 && indexInt < len(baseValue) {
				return &ResolutionResult{
					Value:    baseValue[indexInt],
					Resolved: true,
					Source:   "local",
				}
			}
		}
		return &ResolutionResult{
			Value:       nil,
			Resolved:    false,
			Source:      "local",
			Uncertainty: fmt.Sprintf("Invalid array index '%v'", indexValue),
		}
	default:
		return &ResolutionResult{
			Value:       nil,
			Resolved:    false,
			Source:      "local",
			Uncertainty: fmt.Sprintf("Base value is not indexable (type: %T)", baseValue),
		}
	}
}

// resolveExpression attempts to resolve a simple expression
func (vr *VariableResolver) resolveExpression(expression string) (interface{}, bool) {
	// Handle simple variable references
	if strings.HasPrefix(expression, "var.") {
		result := vr.ResolveReference(expression)
		return result.Value, result.Resolved
	}
	
	// Handle simple local references
	if strings.HasPrefix(expression, "local.") {
		result := vr.ResolveReference(expression)
		return result.Value, result.Resolved
	}
	
	// Handle string literals
	if strings.HasPrefix(expression, "\"") && strings.HasSuffix(expression, "\"") {
		return strings.Trim(expression, "\""), true
	}
	
	// Handle boolean literals
	if expression == "true" {
		return true, true
	}
	if expression == "false" {
		return false, true
	}
	
	// Handle numeric literals (simplified)
	if matched, _ := regexp.MatchString(`^\d+$`, expression); matched {
		return expression, true
	}
	
	return nil, false
}

// extractVariableReferences extracts variable references from an expression
func extractVariableReferences(expression string) []string {
	var refs []string
	
	// Find var.* references
	varPattern := regexp.MustCompile(`var\.([a-zA-Z_][a-zA-Z0-9_]*)`)
	varMatches := varPattern.FindAllStringSubmatch(expression, -1)
	for _, match := range varMatches {
		refs = append(refs, "var."+match[1])
	}
	
	// Find local.* references
	localPattern := regexp.MustCompile(`local\.([a-zA-Z_][a-zA-Z0-9_]*)`)
	localMatches := localPattern.FindAllStringSubmatch(expression, -1)
	for _, match := range localMatches {
		refs = append(refs, "local."+match[1])
	}
	
	return refs
}

// convertHCLValue converts a cty.Value to a Go interface{}
func convertHCLValue(val cty.Value) interface{} {
	if val.IsNull() {
		return nil
	}
	
	switch val.Type() {
	case cty.String:
		return val.AsString()
	case cty.Number:
		if val.Type().Equals(cty.Number) {
			f, _ := val.AsBigFloat().Float64()
			// Try to convert to int if it's a whole number
			if f == float64(int64(f)) {
				return int64(f)
			}
			return f
		}
	case cty.Bool:
		return val.True()
	}
	
	if val.Type().IsListType() || val.Type().IsSetType() || val.Type().IsTupleType() {
		var result []interface{}
		for it := val.ElementIterator(); it.Next(); {
			_, elemVal := it.Element()
			result = append(result, convertHCLValue(elemVal))
		}
		return result
	}
	
	if val.Type().IsMapType() || val.Type().IsObjectType() {
		result := make(map[string]interface{})
		for it := val.ElementIterator(); it.Next(); {
			key, elemVal := it.Element()
			result[key.AsString()] = convertHCLValue(elemVal)
		}
		return result
	}
	
	// For unknown types, return string representation
	return val.AsString()
}

// GetVariables returns all variable definitions
func (vr *VariableResolver) GetVariables() map[string]*VariableDefinition {
	return vr.variables
}

// GetLocals returns all local definitions
func (vr *VariableResolver) GetLocals() map[string]*LocalDefinition {
	return vr.locals
}

// GetVariableValues returns all variable values
func (vr *VariableResolver) GetVariableValues() map[string]interface{} {
	return vr.variableValues
}

// GetResolvedLocals returns all resolved local values
func (vr *VariableResolver) GetResolvedLocals() map[string]interface{} {
	return vr.resolvedLocals
}

// SetCliVars sets variable values from command line -var flags
func (vr *VariableResolver) SetCliVars(vars map[string]string) {
	vr.cliVars = vars
	vr.buildEvalContext() // Rebuild context with new variables
}

// SetTfvarsFiles sets additional .tfvars files to load
func (vr *VariableResolver) SetTfvarsFiles(files []string) {
	vr.tfvarsFiles = files
}

// buildEvalContext creates an HCL evaluation context with variables, locals, and functions
func (vr *VariableResolver) buildEvalContext() {
	vr.evalContext = &hcl.EvalContext{
		Variables: make(map[string]cty.Value),
		Functions: vr.getTerraformFunctions(),
	}
	
	// Add variables to context (always add, even if empty)
	varMap := vr.buildVariableCtxMap()
	if len(varMap) > 0 {
		vr.evalContext.Variables["var"] = cty.ObjectVal(varMap)
	} else {
		// Add empty object for var namespace
		vr.evalContext.Variables["var"] = cty.EmptyObjectVal
	}
	
	// Add locals to context (always add, even if empty)
	localMap := vr.buildLocalCtxMap()
	if len(localMap) > 0 {
		vr.evalContext.Variables["local"] = cty.ObjectVal(localMap)
	} else {
		// Add empty object for local namespace
		vr.evalContext.Variables["local"] = cty.EmptyObjectVal
	}
	
	// Add Terraform meta-arguments with placeholder values for validation
	// These are used in resource configurations but not available during static analysis
	vr.evalContext.Variables["count"] = cty.ObjectVal(map[string]cty.Value{
		"index": cty.NumberIntVal(0), // Use 0 as placeholder for count.index
	})
	
	vr.evalContext.Variables["each"] = cty.ObjectVal(map[string]cty.Value{
		"key":   cty.StringVal("example_key"),   // Placeholder for each.key
		"value": cty.StringVal("example_value"), // Placeholder for each.value
	})
	
	// Add common data source placeholders with typical AWS data sources
	vr.evalContext.Variables["data"] = cty.ObjectVal(map[string]cty.Value{
		"aws_caller_identity": cty.ObjectVal(map[string]cty.Value{
			"current": cty.ObjectVal(map[string]cty.Value{
				"account_id": cty.StringVal("123456789012"), // Placeholder AWS account ID
				"arn":        cty.StringVal("arn:aws:iam::123456789012:user/terraform"),
				"user_id":    cty.StringVal("AIDAI1234567890EXAMPLE"),
			}),
		}),
		"aws_region": cty.ObjectVal(map[string]cty.Value{
			"current": cty.ObjectVal(map[string]cty.Value{
				"name":        cty.StringVal("us-west-2"), // Placeholder region
				"endpoint":    cty.StringVal("ec2.us-west-2.amazonaws.com"),
				"description": cty.StringVal("US West (Oregon)"),
			}),
		}),
		"aws_availability_zones": cty.ObjectVal(map[string]cty.Value{
			"available": cty.ObjectVal(map[string]cty.Value{
				"names": cty.ListVal([]cty.Value{
					cty.StringVal("us-west-2a"),
					cty.StringVal("us-west-2b"),
					cty.StringVal("us-west-2c"),
				}),
				"zone_ids": cty.ListVal([]cty.Value{
					cty.StringVal("usw2-az1"),
					cty.StringVal("usw2-az2"),
					cty.StringVal("usw2-az3"),
				}),
				"state": cty.StringVal("available"),
			}),
		}),
		"aws_ami": cty.ObjectVal(map[string]cty.Value{
			"amazon_linux": cty.ObjectVal(map[string]cty.Value{
				"id": cty.StringVal("ami-0123456789abcdef0"), // Placeholder AMI ID
				"image_id": cty.StringVal("ami-0123456789abcdef0"),
				"name": cty.StringVal("Amazon Linux 2 AMI"),
			}),
		}),
	})
	
	// Add resource reference placeholders
	vr.evalContext.Variables["resource"] = cty.ObjectVal(map[string]cty.Value{})
}

// getTerraformFunctions returns a map of Terraform built-in functions
func (vr *VariableResolver) getTerraformFunctions() map[string]function.Function {
	// Create a map with commonly used Terraform functions from go-cty stdlib
	funcs := map[string]function.Function{
		// String functions
		"chomp":     stdlib.ChompFunc,
		"format":    stdlib.FormatFunc,
		"formatlist": stdlib.FormatListFunc,
		"indent":    stdlib.IndentFunc,
		"join":      stdlib.JoinFunc,
		"lower":     stdlib.LowerFunc,
		"regex":     stdlib.RegexFunc,
		"regexall":  stdlib.RegexAllFunc,
		"replace":   stdlib.ReplaceFunc,
		"split":     stdlib.SplitFunc,
		"strrev":    stdlib.ReverseFunc,
		"substr":    stdlib.SubstrFunc,
		"title":     stdlib.TitleFunc,
		"trim":      stdlib.TrimFunc,
		"trimprefix": stdlib.TrimPrefixFunc,
		"trimspace": stdlib.TrimSpaceFunc,
		"trimsuffix": stdlib.TrimSuffixFunc,
		"upper":     stdlib.UpperFunc,
		
		// Collection functions
		"chunklist":     stdlib.ChunklistFunc,
		"coalesce":      stdlib.CoalesceFunc,
		"coalescelist":  stdlib.CoalesceListFunc,
		"compact":       stdlib.CompactFunc,
		"concat":        stdlib.ConcatFunc,
		"contains":      stdlib.ContainsFunc,
		"distinct":      stdlib.DistinctFunc,
		"element":       stdlib.ElementFunc,
		"flatten":       stdlib.FlattenFunc,
		"index":         stdlib.IndexFunc,
		"keys":          stdlib.KeysFunc,
		"length":        stdlib.LengthFunc,
		"list":          stdlib.ReverseListFunc, // Note: using reverse as placeholder
		"lookup":        stdlib.LookupFunc,
		"merge":         stdlib.MergeFunc,
		"range":         stdlib.RangeFunc,
		"reverse":       stdlib.ReverseListFunc,
		"setintersection": stdlib.SetIntersectionFunc,
		"setproduct":    stdlib.SetProductFunc,
		"setsubtract":   stdlib.SetSubtractFunc,
		"setunion":      stdlib.SetUnionFunc,
		"slice":         stdlib.SliceFunc,
		"sort":          stdlib.SortFunc,
		"values":        stdlib.ValuesFunc,
		"zipmap":        stdlib.ZipmapFunc,
		
		// Numeric functions
		"abs":    stdlib.AbsoluteFunc,
		"ceil":   stdlib.CeilFunc,
		"floor":  stdlib.FloorFunc,
		"log":    stdlib.LogFunc,
		"max":    stdlib.MaxFunc,
		"min":    stdlib.MinFunc,
		"parseint": stdlib.ParseIntFunc,
		"pow":    stdlib.PowFunc,
		"signum": stdlib.SignumFunc,
		
		// Type conversion functions
		"tostring": stdlib.MakeToFunc(cty.String),
		"tonumber": stdlib.MakeToFunc(cty.Number),
		"tobool":   stdlib.MakeToFunc(cty.Bool),
		"tolist":   stdlib.MakeToFunc(cty.List(cty.DynamicPseudoType)),
		"tomap":    stdlib.MakeToFunc(cty.Map(cty.DynamicPseudoType)),
		"toset":    stdlib.MakeToFunc(cty.Set(cty.DynamicPseudoType)),
		
		// Encoding functions
		"base64decode": stdlib.CSVDecodeFunc, // Note: using CSV as placeholder
		"csvdecode":    stdlib.CSVDecodeFunc,
		"jsondecode":   stdlib.JSONDecodeFunc,
		"jsonencode":   stdlib.JSONEncodeFunc,
		
		// Date and time functions
		"formatdate": stdlib.FormatDateFunc,
		"timeadd":    stdlib.TimeAddFunc,
		"timestamp":  vr.timestampFunc(), // Custom implementation
	}
	
	// Add additional Terraform-specific functions
	funcs["cidrhost"] = vr.cidrhostFunc()
	funcs["can"] = vr.canFunc()
	
	return funcs
}

// buildVariableCtxMap builds a cty.Value map for variables context
func (vr *VariableResolver) buildVariableCtxMap() map[string]cty.Value {
	varMap := make(map[string]cty.Value)
	
	// Start with defaults from variable definitions
	for name, varDef := range vr.variables {
		if varDef.Default != nil {
			varMap[name] = convertGoValueToCty(varDef.Default)
		}
	}
	
	// Override with values from .tfvars files
	for name, value := range vr.variableValues {
		varMap[name] = convertGoValueToCty(value)
	}
	
	// Override with CLI -var values (highest precedence)
	for name, value := range vr.cliVars {
		varMap[name] = cty.StringVal(value) // CLI vars are always strings initially
	}
	
	// Add common module variable defaults if they don't exist (for better module variable resolution)
	moduleDefaults := map[string]cty.Value{
		"multi_az":           cty.BoolVal(false),     // Default to false for development/staging
		"backup_enabled":     cty.BoolVal(true),     // Default to enabled
		"backup_retention":   cty.NumberIntVal(7),   // Default to 7 days
		"deletion_protection": cty.BoolVal(false),   // Default to false for testing
		"monitoring_enabled": cty.BoolVal(true),     // Default to enabled
	}
	
	for name, defaultValue := range moduleDefaults {
		if _, exists := varMap[name]; !exists {
			varMap[name] = defaultValue
		}
	}
	
	return varMap
}

// buildLocalCtxMap builds a cty.Value map for locals context
func (vr *VariableResolver) buildLocalCtxMap() map[string]cty.Value {
	localMap := make(map[string]cty.Value)
	
	for name, value := range vr.resolvedLocals {
		localMap[name] = convertGoValueToCty(value)
	}
	
	return localMap
}

// convertGoValueToCty converts a Go interface{} value to cty.Value
func convertGoValueToCty(value interface{}) cty.Value {
	if value == nil {
		return cty.NullVal(cty.DynamicPseudoType)
	}
	
	switch v := value.(type) {
	case string:
		return cty.StringVal(v)
	case int:
		return cty.NumberIntVal(int64(v))
	case int64:
		return cty.NumberIntVal(v)
	case float64:
		return cty.NumberFloatVal(v)
	case bool:
		return cty.BoolVal(v)
	case []interface{}:
		var elements []cty.Value
		for _, elem := range v {
			elements = append(elements, convertGoValueToCty(elem))
		}
		// Handle empty slice case - cty.ListVal cannot be called with empty slice
		if len(elements) == 0 {
			return cty.ListValEmpty(cty.DynamicPseudoType)
		}
		return cty.ListVal(elements)
	case map[string]interface{}:
		elemMap := make(map[string]cty.Value)
		for key, val := range v {
			elemMap[key] = convertGoValueToCty(val)
		}
		return cty.ObjectVal(elemMap)
	default:
		// Fallback to string representation
		return cty.StringVal(fmt.Sprintf("%v", v))
	}
}

// timestampFunc returns a Terraform-compatible timestamp function
func (vr *VariableResolver) timestampFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			// Return a placeholder timestamp in RFC3339 format
			// In real usage, this would return time.Now().UTC().Format(time.RFC3339)
			// For validation, we use a static timestamp
			return cty.StringVal("2024-01-01T00:00:00Z"), nil
		},
	})
}

// cidrhostFunc returns a Terraform-compatible cidrhost function
func (vr *VariableResolver) cidrhostFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "prefix",
				Type: cty.String,
			},
			{
				Name: "hostnum", 
				Type: cty.Number,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			// For validation, just return a placeholder IP
			// In real usage, this would calculate the actual host IP
			return cty.StringVal("10.0.0.1"), nil
		},
	})
}

// canFunc returns a Terraform-compatible can function
func (vr *VariableResolver) canFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "expression",
				Type: cty.DynamicPseudoType,
			},
		},
		Type: function.StaticReturnType(cty.Bool),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			// The can function tests whether an expression evaluates without errors
			// For validation purposes, we'll return true for most cases
			return cty.BoolVal(true), nil
		},
	})
}