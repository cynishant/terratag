#!/bin/bash

# Setup Demo Environment for Terratag
# This script prepares the demo environment and ensures all required files are in place

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
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

# Function to check if running from correct directory
check_directory() {
    if [[ ! -f "docker-compose.yml" ]] || [[ ! -d "demo-deployment" ]]; then
        print_error "Please run this script from the terratag root directory"
        print_info "Expected structure:"
        print_info "  terratag/"
        print_info "  ├── docker-compose.yml"
        print_info "  ├── demo-deployment/"
        print_info "  └── scripts/setup-demo.sh"
        exit 1
    fi
}

# Function to create required directories
create_directories() {
    print_info "Creating required directories..."
    
    mkdir -p standards
    mkdir -p reports
    
    print_success "Directories created: standards/, reports/"
}

# Function to copy demo files
setup_demo_files() {
    print_info "Setting up demo files..."
    
    # Copy tag standard to standards directory
    if [[ -f "demo-deployment/tag-standard.yaml" ]]; then
        cp demo-deployment/tag-standard.yaml standards/
        print_success "Copied tag-standard.yaml to standards directory"
    else
        print_error "Demo tag standard file not found: demo-deployment/tag-standard.yaml"
        exit 1
    fi
    
    # Create environment file if it doesn't exist
    if [[ ! -f ".env" ]]; then
        cp .env.example .env
        print_success "Created .env file from .env.example"
        print_warning "Please review and customize .env file as needed"
    else
        print_info ".env file already exists"
    fi
}

# Function to verify demo deployment
verify_demo_deployment() {
    print_info "Verifying demo deployment files..."
    
    local required_files=(
        "demo-deployment/main.tf"
        "demo-deployment/variables.tf"
        "demo-deployment/outputs.tf"
        "demo-deployment/tag-standard.yaml"
    )
    
    local missing_files=()
    
    for file in "${required_files[@]}"; do
        if [[ ! -f "$file" ]]; then
            missing_files+=("$file")
        fi
    done
    
    if [[ ${#missing_files[@]} -gt 0 ]]; then
        print_error "Missing required demo files:"
        for file in "${missing_files[@]}"; do
            print_error "  - $file"
        done
        exit 1
    fi
    
    print_success "All demo deployment files verified"
}

# Function to check Docker
check_docker() {
    print_info "Checking Docker installation..."
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed or not in PATH"
        print_info "Please install Docker from: https://docs.docker.com/get-docker/"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        print_error "Docker daemon is not running"
        print_info "Please start Docker and try again"
        exit 1
    fi
    
    print_success "Docker is installed and running"
}

# Function to build Docker image
build_docker_image() {
    print_info "Building Terratag Docker image..."
    
    if docker build -t terratag:latest . > /dev/null 2>&1; then
        print_success "Docker image built successfully"
    else
        print_error "Failed to build Docker image"
        print_info "Try running manually: docker build -t terratag:latest ."
        exit 1
    fi
}

# Function to show demo instructions
show_demo_instructions() {
    print_success "Demo environment setup complete!"
    echo
    print_info "You can now run demos using the following commands:"
    echo
    print_info "CLI Demos:"
    echo "  ./scripts/docker-demo.sh demo-basic          # Basic tag application"
    echo "  ./scripts/docker-demo.sh demo-validation     # Tag validation"
    echo "  ./scripts/docker-demo.sh demo-interactive    # Interactive shell"
    echo
    print_info "Web UI Demo:"
    echo "  docker-compose --profile ui up               # Start web interface"
    echo "  # Then open: http://localhost:8080"
    echo
    print_info "Other Profiles:"
    echo "  docker-compose --profile validate up         # Validation demo"
    echo "  docker-compose --profile dev up              # Development mode"
    echo "  docker-compose --profile cicd up             # CI/CD strict mode"
    echo
    print_info "Documentation:"
    echo "  DEMO_GUIDE.md                                # Complete demo guide"
    echo "  docs/DOCKER_DEMO.md                          # Docker-specific demos"
    echo
    print_warning "Note: The demo uses a complete AWS infrastructure example."
    print_warning "No real AWS resources will be created - it's for tagging demonstration only."
}

# Function to test demo setup
test_demo_setup() {
    print_info "Testing demo setup..."
    
    # Test basic tag application
    print_info "Running quick demo test..."
    
    if docker run --rm \
        -v "$(pwd)/demo-deployment:/demo-deployment" \
        -v "$(pwd)/reports:/reports" \
        -v "$(pwd)/standards:/standards" \
        terratag:latest \
        -dir=/demo-deployment \
        -tags='{"DemoTest":"SetupVerification"}' \
        --dry-run > /dev/null 2>&1; then
        print_success "Demo test passed"
    else
        print_warning "Demo test had issues, but setup is complete"
        print_info "You can still proceed with manual testing"
    fi
}

# Main execution
main() {
    echo "================================================"
    echo "       Terratag Demo Environment Setup"
    echo "================================================"
    echo
    
    check_directory
    check_docker
    create_directories
    setup_demo_files
    verify_demo_deployment
    
    # Ask user if they want to build Docker image
    read -p "Build Docker image now? (y/N): " -r
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        build_docker_image
        test_demo_setup
    else
        print_info "Skipping Docker image build"
        print_info "Run 'docker build -t terratag:latest .' when ready"
    fi
    
    echo
    show_demo_instructions
}

# Run main function
main "$@"