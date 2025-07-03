#!/bin/bash

# Docker Demo Script for Terratag
# This script provides easy ways to demonstrate Terratag functionality with Docker

set -e

# Default values
DEMO_DEPLOYMENT_DIR="./demo-deployment"
REPORTS_DIR="./reports"
STANDARDS_DIR="./standards"
CONTAINER_NAME="terratag-demo"
IMAGE_NAME="terratag:latest"

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

# Function to show usage
show_usage() {
    cat << EOF
Docker Demo Script for Terratag

Usage: $0 [command] [options]

Commands:
    build               Build the Terratag Docker image
    demo-basic          Run basic tagging demonstration
    demo-validation     Run tag validation demonstration
    demo-interactive    Start interactive shell in container
    demo-ui             Start UI/API service for demonstration
    clean               Clean up containers and volumes
    help                Show this help message

Options:
    -d, --dir DIR       Demo deployment directory (default: ./demo-deployment)
    -r, --reports DIR   Reports output directory (default: ./reports)
    -s, --standards DIR Standards directory (default: ./standards)
    -v, --verbose       Enable verbose output
    --no-build          Skip building the image

Examples:
    # Build image and run basic demo
    $0 build
    $0 demo-basic

    # Run validation demo with custom directories
    $0 demo-validation -d ./my-terraform -r ./my-reports

    # Start interactive shell for manual demonstration
    $0 demo-interactive

    # Clean up everything
    $0 clean

EOF
}

# Function to build Docker image
build_image() {
    print_info "Building Terratag Docker image..."
    docker build -t "$IMAGE_NAME" .
    print_success "Docker image built successfully"
}

# Function to ensure directories exist
ensure_directories() {
    mkdir -p "$REPORTS_DIR"
    mkdir -p "$STANDARDS_DIR"
    
    # Copy tag standard if it doesn't exist in standards directory
    if [[ ! -f "$STANDARDS_DIR/tag-standard.yaml" && -f "$DEMO_DEPLOYMENT_DIR/tag-standard.yaml" ]]; then
        cp "$DEMO_DEPLOYMENT_DIR/tag-standard.yaml" "$STANDARDS_DIR/"
        print_info "Copied tag-standard.yaml to standards directory"
    fi
}

# Function to run basic tagging demonstration
demo_basic() {
    print_info "Running basic tagging demonstration..."
    
    ensure_directories
    
    # Run terratag with demo tags
    docker run --rm \
        -v "$(pwd)/$DEMO_DEPLOYMENT_DIR:/demo-deployment" \
        -v "$(pwd)/$REPORTS_DIR:/reports" \
        -v "$(pwd)/$STANDARDS_DIR:/standards" \
        --name "$CONTAINER_NAME" \
        "$IMAGE_NAME" \
        -dir=/demo-deployment \
        -tags='{"Environment":"Demo","Owner":"demo@example.com","DemoRun":"'$(date +%Y%m%d-%H%M%S)'"}'
    
    print_success "Basic tagging demonstration completed"
    
    # Show generated files
    print_info "Generated files:"
    find "$DEMO_DEPLOYMENT_DIR" -name "*.terratag.tf" -o -name "*.tf.bak" | head -10
}

# Function to run validation demonstration
demo_validation() {
    print_info "Running tag validation demonstration..."
    
    ensure_directories
    
    # Run validation
    docker run --rm \
        -v "$(pwd)/$DEMO_DEPLOYMENT_DIR:/demo-deployment" \
        -v "$(pwd)/$REPORTS_DIR:/reports" \
        -v "$(pwd)/$STANDARDS_DIR:/standards" \
        --name "$CONTAINER_NAME" \
        "$IMAGE_NAME" \
        -validate-only \
        -standard=/demo-deployment/tag-standard.yaml \
        -dir=/demo-deployment \
        -report-format=json \
        -report-output=/reports/validation-report.json \
        -verbose
    
    print_success "Validation demonstration completed"
    
    # Show report if it exists
    if [[ -f "$REPORTS_DIR/validation-report.json" ]]; then
        print_info "Validation report generated:"
        echo "Location: $REPORTS_DIR/validation-report.json"
        if command -v jq &> /dev/null; then
            echo "Summary:"
            jq '.summary' "$REPORTS_DIR/validation-report.json" 2>/dev/null || echo "Report generated but couldn't parse with jq"
        fi
    fi
}

# Function to start interactive shell
demo_interactive() {
    print_info "Starting interactive shell in Terratag container..."
    
    ensure_directories
    
    docker run -it --rm \
        -v "$(pwd)/$DEMO_DEPLOYMENT_DIR:/demo-deployment" \
        -v "$(pwd)/$REPORTS_DIR:/reports" \
        -v "$(pwd)/$STANDARDS_DIR:/standards" \
        -v "$(pwd):/workspace" \
        --name "$CONTAINER_NAME" \
        --entrypoint /bin/bash \
        "$IMAGE_NAME"
}

# Function to start UI service
demo_ui() {
    print_info "Starting Terratag UI/API service..."
    
    ensure_directories
    
    print_info "UI will be available at: http://localhost:8080"
    print_warning "Press Ctrl+C to stop the service"
    
    docker run --rm \
        -v "$(pwd)/$DEMO_DEPLOYMENT_DIR:/demo-deployment" \
        -v "$(pwd)/$REPORTS_DIR:/reports" \
        -v "$(pwd)/$STANDARDS_DIR:/standards" \
        -v "$(pwd):/workspace" \
        -p 8080:8080 \
        --name "${CONTAINER_NAME}-ui" \
        --entrypoint /usr/local/bin/terratag-api \
        "$IMAGE_NAME"
}

# Function to clean up
clean_up() {
    print_info "Cleaning up Docker containers and volumes..."
    
    # Stop and remove containers
    docker ps -a --filter "name=terratag" --format "{{.Names}}" | xargs -r docker rm -f
    
    # Remove dangling volumes
    docker volume prune -f
    
    # Optionally remove the image
    read -p "Remove Terratag Docker image? (y/N): " -r
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker rmi "$IMAGE_NAME" 2>/dev/null || true
    fi
    
    print_success "Cleanup completed"
}

# Function to check prerequisites
check_prerequisites() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    if [[ ! -d "$DEMO_DEPLOYMENT_DIR" ]]; then
        print_error "Demo deployment directory not found: $DEMO_DEPLOYMENT_DIR"
        print_info "Please run this script from the terratag root directory"
        exit 1
    fi
}

# Parse command line arguments
VERBOSE=false
NO_BUILD=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--dir)
            DEMO_DEPLOYMENT_DIR="$2"
            shift 2
            ;;
        -r|--reports)
            REPORTS_DIR="$2"
            shift 2
            ;;
        -s|--standards)
            STANDARDS_DIR="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        --no-build)
            NO_BUILD=true
            shift
            ;;
        -h|--help|help)
            show_usage
            exit 0
            ;;
        build|demo-basic|demo-validation|demo-interactive|demo-ui|clean)
            COMMAND="$1"
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Set default command if none provided
if [[ -z ${COMMAND:-} ]]; then
    COMMAND="help"
fi

# Run prerequisite checks
check_prerequisites

# Execute the command
case $COMMAND in
    build)
        build_image
        ;;
    demo-basic)
        if [[ "$NO_BUILD" == "false" ]]; then
            build_image
        fi
        demo_basic
        ;;
    demo-validation)
        if [[ "$NO_BUILD" == "false" ]]; then
            build_image
        fi
        demo_validation
        ;;
    demo-interactive)
        if [[ "$NO_BUILD" == "false" ]]; then
            build_image
        fi
        demo_interactive
        ;;
    demo-ui)
        if [[ "$NO_BUILD" == "false" ]]; then
            build_image
        fi
        demo_ui
        ;;
    clean)
        clean_up
        ;;
    help)
        show_usage
        ;;
    *)
        print_error "Unknown command: $COMMAND"
        show_usage
        exit 1
        ;;
esac