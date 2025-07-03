#!/bin/bash

# Terratag Tag Standardization Demo Script
echo "🏷️  Terratag Tag Standardization Demo"
echo "====================================="
echo

# Build terratag
echo "📦 Building Terratag..."
go build ./cmd/terratag
echo "✅ Build complete!"
echo

# Demo 1: Compliant resources (should show 100% compliance)
echo "🎯 Demo 1: Validating COMPLIANT resources"
echo "----------------------------------------"
./terratag -validate-only -standard=examples/aws-tag-standard.yaml -dir=test/validation-tests/compliant
echo

# Demo 2: Non-compliant resources (should show violations)
echo "❌ Demo 2: Validating NON-COMPLIANT resources"
echo "---------------------------------------------"
./terratag -validate-only -standard=examples/aws-tag-standard.yaml -dir=test/validation-tests/non-compliant
echo

# Demo 3: JSON Report generation
echo "📊 Demo 3: Generating JSON report for mixed scenario"
echo "---------------------------------------------------"
./terratag -validate-only -standard=examples/aws-tag-standard.yaml -dir=test/validation-tests/mixed -report-format=json -report-output=compliance.json
echo "✅ JSON report saved to compliance.json"
echo

# Demo 4: Filtering specific resource types
echo "🔍 Demo 4: Filtering only AWS instances"
echo "--------------------------------------"
./terratag -validate-only -standard=examples/aws-tag-standard.yaml -dir=test/validation-tests/non-compliant -filter="aws_instance"
echo

# Demo 5: Strict mode (should exit with error)
echo "⚠️  Demo 5: Strict mode validation (will exit with error)"
echo "--------------------------------------------------------"
./terratag -validate-only -standard=examples/aws-tag-standard.yaml -dir=test/validation-tests/non-compliant -strict-mode || echo "❌ Strict mode detected violations and exited with error code $?"
echo

echo "🎉 Tag Standardization Demo Complete!"
echo
echo "🚀 New Terratag Features Demonstrated:"
echo "  ✅ Tag standardization with YAML configuration"
echo "  ✅ Comprehensive validation with detailed violations"
echo "  ✅ Multiple report formats (table, json, yaml, markdown)"
echo "  ✅ Resource filtering and strict mode"
echo "  ✅ Suggested fixes for non-compliant resources"
echo "  ✅ Resource-specific rules and global exclusions"
echo "  ✅ Pattern matching, data type validation, and allowed values"
echo
echo "📚 Usage Examples:"
echo "  terratag -validate-only -standard=tag-standard.yaml -dir=."
echo "  terratag -validate-only -standard=tag-standard.yaml -report-format=json -report-output=report.json"
echo "  terratag -validate-only -standard=tag-standard.yaml -strict-mode"
echo "  terratag -validate-only -standard=tag-standard.yaml -filter=\"aws_instance|aws_s3_bucket\""