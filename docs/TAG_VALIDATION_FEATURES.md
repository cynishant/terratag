# Tag Validation Features

This document provides an overview of Terratag's comprehensive tag validation and standardization capabilities.

## Overview

Terratag now supports advanced tag validation and compliance checking in addition to its core tag application functionality. These features enable organizations to enforce tagging standards across their Infrastructure as Code.

## Key Features

### ✅ Multi-Tag Validation
- **Supports multiple tags per resource** (tested with up to 12 tags per resource)
- **Complex validation rules** including patterns, data types, and case sensitivity
- **Resource-specific requirements** with override capabilities
- **Comprehensive violation detection** and suggested fixes

### ✅ Tag Standardization
- **YAML-based standards** defining required and optional tags
- **Flexible validation rules** supporting various data types and constraints
- **Resource-type specific rules** for targeted compliance requirements
- **Global and local exclusions** for non-taggable resources

### ✅ AWS Resource Analysis
- **Comprehensive AWS coverage** - 1,506 resource types analyzed
- **Tagging support detection** - automatically identifies which resources support tags
- **Service-level breakdown** showing tagging capabilities across AWS services
- **Intelligent filtering** excludes non-taggable resources from compliance calculations

### ✅ Multiple Report Formats
- **Table format** - Human-readable console output
- **JSON format** - Machine-readable for automation and CI/CD
- **YAML format** - Structured data for further processing
- **Markdown format** - Documentation-ready compliance reports

## Quick Start

### 1. Basic Validation
```bash
# Validate existing tags against a standard
terratag -validate-only -standard tag-standard.yaml
```

### 2. Generate Compliance Report
```bash
# Create detailed compliance documentation
terratag -validate-only \
  -standard tag-standard.yaml \
  -report-format markdown \
  -report-output compliance-report.md
```

### 3. CI/CD Integration
```bash
# Strict mode for automated compliance checking
terratag -validate-only \
  -standard tag-standard.yaml \
  -strict-mode \
  -report-format json
```

## Documentation

### 📚 Complete Guides
- **[Getting Started Guide](GETTING_STARTED.md)** - Comprehensive tutorial with examples
- **[AWS Resource Tagging Support](AWS_RESOURCE_TAGGING.md)** - Complete AWS resource reference

### 🧪 Testing & Examples
- **Multi-tag test scenarios** - `test/validation-tests/multi-tag-scenarios/`
- **Enhanced tag standard** - `test/validation-tests/multi-tag-scenarios/enhanced-tag-standard.yaml`
- **Demo script** - `examples/demo-multi-tag-validation.sh`

### 📋 Sample Standards
- **Basic AWS standard** - `examples/aws-tag-standard.yaml`
- **Enhanced multi-tag standard** - Comprehensive validation rules with all features

## Test Results

The comprehensive multi-tag validation test demonstrates:

### 📊 Validation Capabilities
- **14 resources analyzed** across multiple AWS services
- **Complex violation detection** including:
  - Missing required tags
  - Format pattern violations (regex)
  - Data type mismatches (string vs numeric vs boolean vs date vs cron)
  - Case sensitivity violations
  - Length constraint violations
  - Invalid allowed values

### 🎯 Compliance Analysis
- **35.7% compliance rate** in test scenario (intentionally low for demonstration)
- **85.7% of resources support tagging** (AWS analysis)
- **Multiple violation types** categorized and prioritized
- **Actionable suggestions** for remediation

### 🔍 Violation Types Detected
- **Invalid value**: 6 occurrences
- **Invalid data type**: 5 occurrences  
- **Invalid format**: 3 occurrences
- **Length violations**: 1 occurrence

## AWS Resource Tagging Analysis

### 📈 Coverage Statistics
- **Total AWS Resources**: 1,506
- **Taggable Resources**: 736 (48.9%)
- **Non-Taggable Resources**: 770 (51.1%)

### 🏷️ Service Analysis
Services with excellent tagging support (>80%):
- **DataSync**: 100%
- **FSx**: 100%
- **SageMaker**: 83.3%

Services with mixed support (30-80%):
- **EC2**: 50.0%
- **VPC**: 43.6%
- **S3**: 33.3%

## Integration Examples

### GitHub Actions
```yaml
- name: Tag Compliance Check
  run: |
    terratag -validate-only \
      -standard .github/standards/aws-tags.yaml \
      -strict-mode \
      -report-format json \
      -report-output compliance.json
```

### GitLab CI
```yaml
tag-compliance:
  script:
    - terraform init
    - terratag -validate-only -standard tag-standard.yaml -strict-mode
  artifacts:
    reports:
      junit: compliance-report.xml
```

## Advanced Features

### Resource-Specific Rules
```yaml
resource_rules:
  # Database resources need backup tags
  - resource_types: ["aws_db_instance"]
    required_tags: ["Backup"]
    
  # Storage resources need data classification  
  - resource_types: ["aws_s3_bucket", "aws_ebs_volume"]
    required_tags: ["DataClassification"]
    
  # IAM resources exclude operational tags
  - resource_types: ["aws_iam_*"]
    excluded_tags: ["Backup", "MaintenanceWindow"]
```

### Complex Validation Rules
```yaml
tags:
  - key: "CostCenter"
    format: "^CC\\d{4}$"  # Must match CC1234 pattern
    
  - key: "Owner"
    data_type: "email"
    format: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
    
  - key: "Environment" 
    allowed_values: ["Production", "Staging", "Development"]
    case_sensitive: true
    
  - key: "IsActive"
    data_type: "boolean"
    allowed_values: ["true", "false"]
    
  - key: "MaintenanceWindow"
    data_type: "cron"
    description: "Maintenance schedule using cron expression"
    examples: ["0 2 * * 0", "30 3 * * 1-5", "*/15 * * * *"]
```

## Performance & Scalability

- **Concurrent processing** for large terraform codebases
- **Pre-compiled regex patterns** for validation performance
- **Efficient AWS resource mapping** with cached lookups
- **Streaming report generation** for large datasets

## Future Enhancements

- **Auto-fix capabilities** - Automatically correct violations when possible
- **Custom violation types** - Extensible validation framework
- **Multi-cloud standards** - GCP and Azure resource analysis
- **Policy as code integration** - OPA/Sentinel policy generation

## Support & Contribution

- **CLI Help**: `terratag -h`
- **GitHub Issues**: https://github.com/cloudyali/terratag/issues
- **Examples**: `examples/` directory
- **Tests**: `test/validation-tests/` directory

---

*This tag validation system represents a comprehensive solution for Infrastructure as Code compliance, enabling organizations to maintain consistent tagging standards across their cloud resources.*