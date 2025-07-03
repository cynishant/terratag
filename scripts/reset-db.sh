#!/bin/bash

# Reset Terratag Database Script
# This script cleans up the SQLite database and resets the demo environment

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

# Function to stop running containers
stop_containers() {
    print_info "Stopping running Terratag containers..."
    
    # Stop docker-compose services
    if docker-compose ps | grep -q "Up"; then
        docker-compose down
        print_success "Docker Compose services stopped"
    fi
    
    # Stop any standalone containers
    local containers=$(docker ps --filter "name=terratag" --format "{{.Names}}" 2>/dev/null || true)
    if [[ -n "$containers" ]]; then
        docker stop $containers
        print_success "Standalone containers stopped"
    fi
}

# Function to remove SQLite database
remove_database() {
    print_info "Removing SQLite database..."
    
    local db_files=(
        "terratag.db"
        "terratag.db-shm"
        "terratag.db-wal"
    )
    
    for db_file in "${db_files[@]}"; do
        if [[ -f "$db_file" ]]; then
            rm -f "$db_file"
            print_success "Removed $db_file"
        fi
    done
}

# Function to clean generated files
clean_generated_files() {
    print_info "Cleaning generated files..."
    
    # Remove terratag generated files
    find demo-deployment -name "*.terratag.tf" -delete 2>/dev/null || true
    find demo-deployment -name "*.tf.bak" -delete 2>/dev/null || true
    
    # Clean reports directory
    if [[ -d "reports" ]]; then
        rm -rf reports/*
        print_success "Cleaned reports directory"
    fi
    
    print_success "Cleaned generated files"
}

# Function to reset demo standards
reset_demo_standards() {
    print_info "Resetting demo standards..."
    
    # Recreate standards directory
    rm -rf standards
    mkdir -p standards
    
    # Copy fresh demo standard
    if [[ -f "demo-deployment/tag-standard.yaml" ]]; then
        cp demo-deployment/tag-standard.yaml standards/
        print_success "Reset demo tag standard"
    else
        print_warning "Demo tag standard not found"
    fi
}

# Function to show reset options
show_reset_options() {
    cat << EOF
Reset Options:

1. Database only     - Remove SQLite database, keep standards and reports
2. Full reset        - Remove database, generated files, and reset standards  
3. Soft reset        - Keep database, clean generated files only
4. Exit              - Cancel operation

EOF
}

# Function to perform database-only reset
reset_database_only() {
    print_warning "This will remove all stored standards, operations, and results from the database."
    read -p "Continue? (y/N): " -r
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        stop_containers
        remove_database
        print_success "Database reset complete"
        print_info "Standards and reports preserved"
    else
        print_info "Database reset cancelled"
    fi
}

# Function to perform full reset
reset_full() {
    print_warning "This will remove ALL data: database, standards, generated files, and reports."
    read -p "Continue? (y/N): " -r
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        stop_containers
        remove_database
        clean_generated_files
        reset_demo_standards
        print_success "Full reset complete"
        print_info "Demo environment restored to initial state"
    else
        print_info "Full reset cancelled"
    fi
}

# Function to perform soft reset
reset_soft() {
    print_info "Cleaning generated files while preserving database and standards..."
    clean_generated_files
    print_success "Soft reset complete"
    print_info "Database and standards preserved"
}

# Main menu
main() {
    echo "================================================"
    echo "       Terratag Database Reset Tool"
    echo "================================================"
    echo
    
    show_reset_options
    
    read -p "Select option (1-4): " -r choice
    
    case $choice in
        1)
            reset_database_only
            ;;
        2)
            reset_full
            ;;
        3)
            reset_soft
            ;;
        4)
            print_info "Reset cancelled"
            exit 0
            ;;
        *)
            print_error "Invalid option: $choice"
            exit 1
            ;;
    esac
    
    echo
    print_info "You can now restart the demo environment:"
    print_info "  docker-compose --profile ui up"
    print_info "  ./scripts/docker-demo.sh demo-basic"
}

# Check if running from correct directory
if [[ ! -f "docker-compose.yml" ]]; then
    print_error "Please run this script from the terratag root directory"
    exit 1
fi

# Run main function
main "$@"