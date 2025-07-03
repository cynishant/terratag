#!/bin/bash

# Load Demo Standards Script
# This script loads the demo tag standards into the Terratag database

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if API is running
check_api() {
    local max_attempts=30
    local attempt=1
    
    print_info "Checking if Terratag API is running..."
    
    while [[ $attempt -le $max_attempts ]]; do
        if curl -s http://localhost:8080/api/v1/standards > /dev/null 2>&1; then
            print_success "API is running"
            return 0
        fi
        
        if [[ $attempt -eq 1 ]]; then
            print_info "API not ready, waiting..."
        fi
        
        sleep 2
        ((attempt++))
    done
    
    print_error "API is not running or not responding"
    print_info "Please start the API first:"
    print_info "  docker-compose --profile ui up"
    return 1
}

# Function to load demo standard
load_demo_standard() {
    local standard_file="$1"
    local standard_name="$2"
    local description="$3"
    local cloud_provider="$4"
    
    print_info "Loading standard: $standard_name"
    
    # Read the YAML content
    if [[ ! -f "$standard_file" ]]; then
        print_error "Standard file not found: $standard_file"
        return 1
    fi
    
    local yaml_content=$(cat "$standard_file")
    
    # Create the JSON payload
    local json_payload=$(cat << EOF
{
  "name": "$standard_name",
  "description": "$description",
  "cloud_provider": "$cloud_provider",
  "content": $(echo "$yaml_content" | jq -Rs .)
}
EOF
)
    
    # Post to API
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$json_payload" \
        http://localhost:8080/api/v1/standards)
    
    if echo "$response" | jq -e '.id' > /dev/null 2>&1; then
        local standard_id=$(echo "$response" | jq -r '.id')
        print_success "Standard loaded with ID: $standard_id"
        return 0
    else
        print_error "Failed to load standard: $response"
        return 1
    fi
}

# Function to check if standard already exists
standard_exists() {
    local standard_name="$1"
    
    local response=$(curl -s http://localhost:8080/api/v1/standards)
    echo "$response" | jq -e ".[] | select(.name == \"$standard_name\")" > /dev/null 2>&1
}

# Function to load all demo standards
load_all_standards() {
    print_info "Loading demo tag standards into database..."
    
    # Load main demo standard
    if standard_exists "AWS Demo Standard"; then
        print_warning "AWS Demo Standard already exists, skipping"
    else
        load_demo_standard \
            "demo-deployment/tag-standard.yaml" \
            "AWS Demo Standard" \
            "Comprehensive AWS tagging standard for demonstration purposes" \
            "aws"
    fi
    
    # Load AWS basic template
    if [[ -f "standards/aws-basic-template.yaml" ]]; then
        if standard_exists "AWS Basic Template"; then
            print_warning "AWS Basic Template already exists, skipping"
        else
            load_demo_standard \
                "standards/aws-basic-template.yaml" \
                "AWS Basic Template" \
                "Basic AWS tagging template for quick start" \
                "aws"
        fi
    fi
    
    # Load GCP basic template
    if [[ -f "standards/gcp-basic-template.yaml" ]]; then
        if standard_exists "GCP Basic Template"; then
            print_warning "GCP Basic Template already exists, skipping"
        else
            load_demo_standard \
                "standards/gcp-basic-template.yaml" \
                "GCP Basic Template" \
                "Basic GCP labeling template for quick start" \
                "gcp"
        fi
    fi
    
    # Load standards directory standard if different from demo
    if [[ -f "standards/tag-standard.yaml" ]]; then
        if ! cmp -s "demo-deployment/tag-standard.yaml" "standards/tag-standard.yaml"; then
            if standard_exists "Custom Demo Standard"; then
                print_warning "Custom Demo Standard already exists, skipping"
            else
                load_demo_standard \
                    "standards/tag-standard.yaml" \
                    "Custom Demo Standard" \
                    "Custom tag standard from standards directory" \
                    "aws"
            fi
        fi
    fi
    
    print_success "Demo standards loading complete"
}

# Function to show loaded standards
show_standards() {
    print_info "Currently loaded standards:"
    
    local response=$(curl -s http://localhost:8080/api/v1/standards)
    
    if echo "$response" | jq -e '. | length > 0' > /dev/null 2>&1; then
        echo "$response" | jq -r '.[] | "  • \(.name) (\(.cloud_provider)) - ID: \(.id)"'
    else
        print_warning "No standards currently loaded"
    fi
}

# Function to start API if not running
start_api_if_needed() {
    if ! curl -s http://localhost:8080/api/v1/standards > /dev/null 2>&1; then
        print_info "API not running, starting in background..."
        
        # Start API in background
        docker-compose --profile ui up -d
        
        # Wait for API to be ready
        check_api
    fi
}

# Main function
main() {
    echo "================================================"
    echo "       Load Demo Standards into Database"
    echo "================================================"
    echo
    
    # Check if we're in the right directory
    if [[ ! -f "docker-compose.yml" ]] || [[ ! -f "demo-deployment/tag-standard.yaml" ]]; then
        print_error "Please run this script from the terratag root directory"
        print_info "Expected files:"
        print_info "  - docker-compose.yml"
        print_info "  - demo-deployment/tag-standard.yaml"
        exit 1
    fi
    
    # Check if jq is available
    if ! command -v jq &> /dev/null; then
        print_error "jq is required but not installed"
        print_info "Install jq: https://stedolan.github.io/jq/download/"
        exit 1
    fi
    
    # Start API if needed or check if running
    start_api_if_needed
    
    # Load standards
    load_all_standards
    
    echo
    show_standards
    
    echo
    print_success "Demo standards are now available in the UI!"
    print_info "Open http://localhost:8080 to use them"
    print_info "Standards will appear in the dropdown when creating operations"
}

# Run main function
main "$@"