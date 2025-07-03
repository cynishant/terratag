#!/bin/bash

# Multi-Tag Validation Demo Script
# Demonstrates comprehensive tag validation capabilities with multiple tags per resource

set -e

echo "🏷️  Terratag Multi-Tag Validation Demo"
echo "====================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if terratag is available
if ! command -v terratag &> /dev/null; then
    echo -e "${RED}❌ Terratag not found. Please build and install terratag first:${NC}"
    echo "   go build -o terratag cmd/terratag/main.go"
    echo "   sudo mv terratag /usr/local/bin/"
    exit 1
fi

# Check if we're in the right directory
if [ ! -f "test/validation-tests/multi-tag-scenarios/main.tf" ]; then
    echo -e "${RED}❌ Please run this script from the terratag project root directory${NC}"
    exit 1
fi

echo -e "${BLUE}📋 Running comprehensive multi-tag validation scenarios...${NC}"
echo ""

# Navigate to test directory
cd test/validation-tests/multi-tag-scenarios

# Initialize terraform if needed
if [ ! -d ".terraform" ]; then
    echo -e "${YELLOW}🔧 Initializing Terraform...${NC}"
    terraform init >/dev/null 2>&1
    echo -e "${GREEN}✅ Terraform initialized${NC}"
    echo ""
fi

echo -e "${BLUE}📊 Test Scenario 1: Comprehensive Multi-Tag Analysis${NC}"
echo "================================================"
echo ""

# Run validation with table format (human readable)
echo -e "${YELLOW}Running validation with enhanced tag standard...${NC}"
echo ""

terratag -validate-only \
    -standard enhanced-tag-standard.yaml \
    -report-format table \
    -verbose \
    . 2>/dev/null | head -80

echo ""
echo -e "${BLUE}📄 Test Scenario 2: JSON Report for Automation${NC}"
echo "=============================================="
echo ""

# Generate JSON report
terratag -validate-only \
    -standard enhanced-tag-standard.yaml \
    -report-format json \
    -report-output multi-tag-report.json \
    . 2>/dev/null

if [ -f "multi-tag-report.json" ]; then
    echo -e "${GREEN}✅ JSON report generated: multi-tag-report.json${NC}"
    echo ""
    echo -e "${YELLOW}Sample JSON output:${NC}"
    head -20 multi-tag-report.json | jq '.' 2>/dev/null || head -20 multi-tag-report.json
    echo ""
else
    echo -e "${RED}❌ Failed to generate JSON report${NC}"
fi

echo -e "${BLUE}📝 Test Scenario 3: Markdown Report for Documentation${NC}"
echo "=================================================="
echo ""

# Generate Markdown report
terratag -validate-only \
    -standard enhanced-tag-standard.yaml \
    -report-format markdown \
    -report-output multi-tag-compliance.md \
    . 2>/dev/null

if [ -f "multi-tag-compliance.md" ]; then
    echo -e "${GREEN}✅ Markdown report generated: multi-tag-compliance.md${NC}"
    echo ""
    echo -e "${YELLOW}Sample Markdown output:${NC}"
    head -30 multi-tag-compliance.md
    echo ""
else
    echo -e "${RED}❌ Failed to generate Markdown report${NC}"
fi

echo -e "${BLUE}🔍 Test Scenario 4: Violation Analysis${NC}"
echo "====================================="
echo ""

# Show specific violation details
echo -e "${YELLOW}Analyzing tag violations detected:${NC}"
echo ""

# Extract violation summary from JSON report
if [ -f "multi-tag-report.json" ]; then
    echo "📈 Compliance Statistics:"
    jq -r '"Total Resources: " + (.total_resources | tostring)' multi-tag-report.json 2>/dev/null || grep -o '"total_resources":[0-9]*' multi-tag-report.json
    jq -r '"Compliant: " + (.compliant_resources | tostring)' multi-tag-report.json 2>/dev/null || grep -o '"compliant_resources":[0-9]*' multi-tag-report.json
    jq -r '"Non-Compliant: " + (.non_compliant_resources | tostring)' multi-tag-report.json 2>/dev/null || grep -o '"non_compliant_resources":[0-9]*' multi-tag-report.json
    echo ""
    
    echo "🏷️  AWS Tagging Support Analysis:"
    jq -r '"Resources Supporting Tags: " + (.tagging_support.resources_supporting_tags | tostring)' multi-tag-report.json 2>/dev/null || echo "Check JSON report for tagging support details"
    jq -r '"Tagging Support Rate: " + ((.tagging_support.tagging_support_rate * 100) | tostring) + "%"' multi-tag-report.json 2>/dev/null || echo "Check JSON report for tagging rate"
    echo ""
fi

echo -e "${BLUE}🧪 Test Scenario 5: Different Validation Types Detected${NC}"
echo "====================================================="
echo ""

echo "The multi-tag test configuration demonstrates:"
echo ""
echo -e "${GREEN}✅ Valid configurations:${NC}"
echo "  • aws_instance.fully_compliant - 12 tags, all requirements met"
echo "  • aws_s3_bucket.data_lake - 12 tags with comprehensive data classification"
echo ""
echo -e "${YELLOW}⚠️  Format violations:${NC}"
echo "  • Environment: 'prod' vs 'Production' (case sensitivity)"
echo "  • CostCenter: 'INVALID123' vs 'CC\d{4}' pattern"
echo "  • Owner: 'platform' vs email format required"
echo ""
echo -e "${RED}❌ Data type violations:${NC}"
echo "  • PortCount: 'twenty' vs numeric required"
echo "  • CreatedDate: 'not-a-date' vs YYYY-MM-DD format"
echo "  • IsActive: 'maybe' vs boolean true/false"
echo ""
echo -e "${BLUE}📝 Length violations:${NC}"
echo "  • Description: 'A' vs minimum 5 characters"
echo "  • LongDescription: 200+ chars vs maximum 100"
echo ""
echo -e "${RED}🚫 Missing required tags:${NC}"
echo "  • Project tag missing on database instance"
echo "  • Backup tag missing on storage resources"
echo ""

echo -e "${BLUE}🎯 Summary & Next Steps${NC}"
echo "======================"
echo ""
echo "This demo showcased Terratag's comprehensive multi-tag validation capabilities:"
echo ""
echo "1. ✅ Multi-tag resource validation (up to 12 tags per resource)"
echo "2. ✅ Complex validation rules (patterns, data types, case sensitivity)"
echo "3. ✅ AWS resource tagging support analysis"
echo "4. ✅ Multiple report formats (Table, JSON, Markdown)"
echo "5. ✅ Resource-specific tag requirements"
echo "6. ✅ Violation categorization and suggestions"
echo ""
echo -e "${GREEN}📁 Generated files:${NC}"
echo "  • multi-tag-report.json - Machine-readable compliance data"
echo "  • multi-tag-compliance.md - Human-readable documentation"
echo ""
echo -e "${YELLOW}🔧 To integrate into CI/CD:${NC}"
echo "  terratag -validate-only -standard enhanced-tag-standard.yaml -strict-mode -report-format json"
echo ""
echo -e "${BLUE}📚 For more information:${NC}"
echo "  • docs/GETTING_STARTED.md - Complete usage guide"
echo "  • docs/AWS_RESOURCE_TAGGING.md - AWS tagging reference"
echo "  • examples/aws-tag-standard.yaml - Standard template"
echo ""

cd ../../.. # Return to project root

echo -e "${GREEN}🎉 Multi-tag validation demo completed successfully!${NC}"