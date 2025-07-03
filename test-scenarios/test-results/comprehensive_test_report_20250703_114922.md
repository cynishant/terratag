# Terratag Comprehensive Test Report

## Test Overview

This report summarizes the results of comprehensive testing of Terratag across multiple scenarios including AWS, GCP, and mixed-provider configurations.

## Test Execution Details

- **Test Run ID**: 20250703_114922
- **Test Date**: Thu Jul  3 11:49:31 IST 2025
- **Terratag Binary**: /Users/nishant/Documents/GitHub/terratag/test-scenarios/../terratag
- **Terraform Version**: Terraform v1.12.2

## Scenarios Tested

### Simple AWS Test

```
Terratag Test Analysis - simple-aws
Generated: Thu Jul  3 11:49:31 IST 2025

SCENARIO OVERVIEW:
- Scenario: simple-aws
- Directory: /Users/nishant/Documents/GitHub/terratag/test-scenarios/simple-aws
- Total Resources:        7
- Tagged Resources: 0

VALIDATION RESULTS:
{
  "compliance_rate": 0,
  "most_common_violations": null,
  "resource_type_breakdown": {
    "aws_instance": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    },
    "aws_s3_bucket": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    },
    "aws_security_group": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    },
    "aws_subnet": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    },
    "aws_vpc": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    }
  }
}

FILES CREATED:


TERRAFORM MODULES:


VARIABLES ANALYSIS:
Total variables defined: 

TAG/LABEL PATTERNS:
       7 AWS tag blocks found
       0 GCP label blocks found
       5 merge() function calls found

COMPLEX FEATURES:
       0 Resources using count
       0 Resources using for_each
       0 Dynamic blocks found
       1 Locals blocks found

```

### AWS Complex Multi-Module Test

### GCP Complex Multi-Module Test

### Mixed AWS/GCP Provider Test

## Test Files Generated

The following test result files were generated:

- `simple-aws_validation.json`
- `simple-aws_analysis.txt`
- `comprehensive_test_report_20250703_114922.md`
