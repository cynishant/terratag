# Terratag Comprehensive Test Report

## Test Overview

This report summarizes the results of comprehensive testing of Terratag across multiple scenarios including AWS, GCP, and mixed-provider configurations.

## Test Execution Details

- **Test Run ID**: 20250703_171022
- **Test Date**: Thu Jul  3 17:10:58 IST 2025
- **Terratag Binary**: /Users/nishant/Documents/GitHub/terratag/test-scenarios/../terratag
- **Terraform Version**: Terraform v1.12.2

## Scenarios Tested

### Simple AWS Test

```
Terratag Test Analysis - simple-aws
Generated: Thu Jul  3 17:10:32 IST 2025

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

```
Terratag Test Analysis - aws-complex
Generated: Thu Jul  3 17:10:41 IST 2025

SCENARIO OVERVIEW:
- Scenario: aws-complex
- Directory: /Users/nishant/Documents/GitHub/terratag/test-scenarios/aws-complex
- Total Resources:       67
- Tagged Resources: 0

VALIDATION RESULTS:
{
  "compliance_rate": 0,
  "most_common_violations": null,
  "resource_type_breakdown": {
    "aws_cloudfront_distribution": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    },
    "aws_cloudtrail": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    },
    "aws_cloudwatch_log_group": {
      "total": 5,
      "compliant": 0,
      "rate": 0
    },
    "aws_cloudwatch_metric_alarm": {
      "total": 5,
      "compliant": 0,
      "rate": 0
    },
    "aws_db_instance": {
      "total": 2,
      "compliant": 0,
      "rate": 0
    },
    "aws_db_option_group": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    },
    "aws_db_parameter_group": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    },
    "aws_db_subnet_group": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    },
    "aws_iam_instance_profile": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    },
    "aws_iam_role": {
      "total": 2,
      "compliant": 0,
      "rate": 0
    },
    "aws_internet_gateway": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    },
    "aws_kms_key": {
      "total": 2,
      "compliant": 0,
      "rate": 0
    },
    "aws_launch_template": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    },
    "aws_lb": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    },
    "aws_lb_listener": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    },
    "aws_lb_target_group": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    },
    "aws_route_table": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    },
    "aws_s3_bucket": {
      "total": 4,
      "compliant": 0,
      "rate": 0
    },
    "aws_secretsmanager_secret": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    },
    "aws_security_group": {
      "total": 2,
      "compliant": 0,
      "rate": 0
    },
    "aws_sns_topic": {
      "total": 2,
      "compliant": 0,
      "rate": 0
    },
    "aws_subnet": {
      "total": 2,
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
modules/compute
modules/database
modules/monitoring
modules/storage

VARIABLES ANALYSIS:
Total variables defined: 687

TAG/LABEL PATTERNS:
      72 AWS tag blocks found
       0 GCP label blocks found
      62 merge() function calls found

COMPLEX FEATURES:
      32 Resources using count
       1 Resources using for_each
       3 Dynamic blocks found
       3 Locals blocks found

```

### GCP Complex Multi-Module Test

```
Terratag Test Analysis - gcp-complex
Generated: Thu Jul  3 17:10:46 IST 2025

SCENARIO OVERVIEW:
- Scenario: gcp-complex
- Directory: /Users/nishant/Documents/GitHub/terratag/test-scenarios/gcp-complex
- Total Resources:       71
- Tagged Resources: 0

VALIDATION RESULTS:


FILES CREATED:


TERRAFORM MODULES:
modules/compute
modules/database
modules/monitoring
modules/storage

VARIABLES ANALYSIS:
Total variables defined: 411

TAG/LABEL PATTERNS:
       5 AWS tag blocks found
      61 GCP label blocks found
      52 merge() function calls found

COMPLEX FEATURES:
      30 Resources using count
       7 Resources using for_each
       8 Dynamic blocks found
       1 Locals blocks found

```

### Mixed AWS/GCP Provider Test

```
Terratag Test Analysis - mixed-providers
Generated: Thu Jul  3 17:10:58 IST 2025

SCENARIO OVERVIEW:
- Scenario: mixed-providers
- Directory: /Users/nishant/Documents/GitHub/terratag/test-scenarios/mixed-providers
- Total Resources:       21
- Tagged Resources: 0

VALIDATION RESULTS:
{
  "compliance_rate": 0,
  "most_common_violations": null,
  "resource_type_breakdown": {
    "aws_cloudwatch_log_group": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    },
    "aws_db_instance": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    },
    "aws_iam_role": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    },
    "aws_internet_gateway": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    },
    "aws_lambda_function": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    },
    "aws_s3_bucket": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    },
    "aws_subnet": {
      "total": 2,
      "compliant": 0,
      "rate": 0
    },
    "aws_vpc": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    },
    "aws_vpn_gateway": {
      "total": 1,
      "compliant": 0,
      "rate": 0
    }
  }
}

FILES CREATED:


TERRAFORM MODULES:


VARIABLES ANALYSIS:
Total variables defined: 128

TAG/LABEL PATTERNS:
      12 AWS tag blocks found
      11 GCP label blocks found
      21 merge() function calls found

COMPLEX FEATURES:
       2 Resources using count
       0 Resources using for_each
       0 Dynamic blocks found
       1 Locals blocks found

```

## Test Files Generated

The following test result files were generated:

- `aws-complex_validation.json`
- `aws-complex_analysis.txt`
- `simple-aws_validation.json`
- `simple-aws_analysis.txt`
- `gcp-complex_analysis.txt`
- `mixed-providers_validation.json`
- `mixed-providers_analysis.txt`
- `comprehensive_test_report_20250703_171022.md`
