#!/bin/bash
# Docker run script for Terratag
# Provides convenient wrappers for common Terratag operations

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
IMAGE_NAME="terratag:latest"
WORKSPACE_DIR="${TERRATAG_SOURCE_DIR:-$(pwd)}"
WORKSPACE_SUBDIR="${TERRATAG_WORKSPACE_SUBDIR:-}"
STANDARDS_DIR="${TERRATAG_STANDARDS_DIR:-./standards}"
REPORTS_DIR="${TERRATAG_REPORTS_DIR:-./reports}"
OPERATION=""
STANDARD_FILE=""
TAGS=""
EXTRA_ARGS=""

# Function to display usage
usage() {
    cat << EOF
Usage: $0 OPERATION [OPTIONS]

Run Terratag in Docker container with convenient operations

OPERATIONS:
    validate        Validate tags against a standard
    apply           Apply tags to Terraform files
    shell          Start interactive shell in container
    version        Show Terratag version
    help           Show Terratag help

OPTIONS:
    -s, --standard FILE     Tag standard file (for validate operation)
    -t, --tags TAGS         Tags to apply (for apply operation)
    -w, --workspace DIR     Workspace directory (default: current directory or TERRATAG_SOURCE_DIR)
    --subdir SUBDIR         Subdirectory within workspace to process (or TERRATAG_WORKSPACE_SUBDIR)
    -r, --reports DIR       Reports output directory (default: ./reports or TERRATAG_REPORTS_DIR)
    --standards DIR         Standards directory (default: ./standards or TERRATAG_STANDARDS_DIR)
    --image IMAGE           Docker image to use (default: terratag:latest)
    --format FORMAT         Report format (table, json, yaml, markdown)
    --output FILE           Report output file
    --strict                Enable strict mode
    --verbose               Enable verbose output
    -h, --help              Display this help message

EXAMPLES:
    # Validate tags with AWS standard
    $0 validate -s standards/aws-standard.yaml

    # Apply tags to all resources
    $0 apply -t '{"Environment":"production","Owner":"devops@company.com"}'

    # Validate specific subdirectory with JSON output
    $0 validate -s standards/gcp-standard.yaml --subdir infrastructure/aws --format json --output reports/compliance.json

    # Validate different workspace directory
    $0 validate -s standards/tag-standard.yaml -w /path/to/my/terraform/project

    # Start interactive shell
    $0 shell

    # Validate with strict mode
    $0 validate -s standards/tag-standard.yaml --strict --verbose

ENVIRONMENT VARIABLES:
    TERRATAG_SOURCE_DIR      Source directory containing Terraform files
    TERRATAG_WORKSPACE_SUBDIR  Subdirectory within source to process
    TERRATAG_STANDARDS_DIR   Directory containing tag standard files
    TERRATAG_REPORTS_DIR     Directory for output reports

EOF
}

# Function to log messages
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] ✓${NC} $1"
}

log_error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ✗${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] ⚠${NC} $1"
}

# Check if operation is provided
if [[ $# -eq 0 ]]; then
    log_error "No operation specified"
    usage
    exit 1
fi

OPERATION="$1"
shift

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -s|--standard)
            STANDARD_FILE="$2"
            shift 2
            ;;
        -t|--tags)
            TAGS="$2"
            shift 2
            ;;
        -w|--workspace)
            WORKSPACE_DIR="$2"
            shift 2
            ;;
        --subdir)
            WORKSPACE_SUBDIR="$2"
            shift 2
            ;;
        -r|--reports)
            REPORTS_DIR="$2"
            shift 2
            ;;
        --standards)
            STANDARDS_DIR="$2"
            shift 2
            ;;
        --image)
            IMAGE_NAME="$2"
            shift 2
            ;;
        --format)
            EXTRA_ARGS="$EXTRA_ARGS -report-format $2"
            shift 2
            ;;
        --output)
            EXTRA_ARGS="$EXTRA_ARGS -report-output $2"
            shift 2
            ;;
        --strict)
            EXTRA_ARGS="$EXTRA_ARGS -strict-mode"
            shift
            ;;
        --verbose)
            EXTRA_ARGS="$EXTRA_ARGS -verbose"
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            EXTRA_ARGS="$EXTRA_ARGS $1"
            shift
            ;;
    esac
done

# Verify Docker is available
if ! command -v docker &> /dev/null; then
    log_error "Docker is not installed or not in PATH"
    exit 1
fi

# Create directories if they don't exist
mkdir -p "$REPORTS_DIR"
if [[ -n "$STANDARDS_DIR" ]] && [[ ! -d "$STANDARDS_DIR" ]]; then
    mkdir -p "$STANDARDS_DIR"
fi

# Build base Docker command
DOCKER_CMD="docker run --rm -it"
DOCKER_CMD="$DOCKER_CMD -v $(realpath "$WORKSPACE_DIR"):/workspace"
DOCKER_CMD="$DOCKER_CMD -v $(realpath "$REPORTS_DIR"):/reports"

# Mount standards directory if it exists
if [[ -d "$STANDARDS_DIR" ]]; then
    DOCKER_CMD="$DOCKER_CMD -v $(realpath "$STANDARDS_DIR"):/standards:ro"
fi

# Set working directory inside container based on subdirectory
if [[ -n "$WORKSPACE_SUBDIR" ]]; then
    DOCKER_CMD="$DOCKER_CMD -w /workspace/$WORKSPACE_SUBDIR"
else
    DOCKER_CMD="$DOCKER_CMD -w /workspace"
fi

# Mount cloud provider credentials if they exist
if [[ -d "$HOME/.aws" ]]; then
    DOCKER_CMD="$DOCKER_CMD -v $HOME/.aws:/home/terratag/.aws:ro"
fi

if [[ -d "$HOME/.config/gcloud" ]]; then
    DOCKER_CMD="$DOCKER_CMD -v $HOME/.config/gcloud:/home/terratag/.config/gcloud:ro"
fi

# Mount SSH keys if they exist
if [[ -d "$HOME/.ssh" ]]; then
    DOCKER_CMD="$DOCKER_CMD -v $HOME/.ssh:/home/terratag/.ssh:ro"
fi

# Set environment variables
DOCKER_CMD="$DOCKER_CMD -e AWS_PROFILE=${AWS_PROFILE:-}"
DOCKER_CMD="$DOCKER_CMD -e AWS_REGION=${AWS_REGION:-us-east-1}"
DOCKER_CMD="$DOCKER_CMD -e GOOGLE_APPLICATION_CREDENTIALS=${GOOGLE_APPLICATION_CREDENTIALS:-}"
DOCKER_CMD="$DOCKER_CMD -e GOOGLE_PROJECT=${GOOGLE_PROJECT:-}"

# Add image name
DOCKER_CMD="$DOCKER_CMD $IMAGE_NAME"

# Execute operation
case $OPERATION in
    validate)
        if [[ -z "$STANDARD_FILE" ]]; then
            log_error "Standard file is required for validate operation"
            echo "Use: $0 validate -s <standard-file>"
            exit 1
        fi
        
        log "Running tag validation..."
        TERRATAG_CMD="-validate-only -standard /standards/$(basename "$STANDARD_FILE") $EXTRA_ARGS ."
        FINAL_CMD="$DOCKER_CMD $TERRATAG_CMD"
        ;;
        
    apply)
        if [[ -z "$TAGS" ]]; then
            log_error "Tags are required for apply operation"
            echo "Use: $0 apply -t '{\"key\":\"value\"}'"
            exit 1
        fi
        
        log "Applying tags to Terraform files..."
        TERRATAG_CMD="-tags='$TAGS' $EXTRA_ARGS"
        FINAL_CMD="$DOCKER_CMD $TERRATAG_CMD"
        ;;
        
    shell)
        log "Starting interactive shell..."
        FINAL_CMD="$DOCKER_CMD /bin/bash"
        ;;
        
    version)
        log "Getting Terratag version..."
        FINAL_CMD="$DOCKER_CMD -version"
        ;;
        
    help)
        log "Getting Terratag help..."
        FINAL_CMD="$DOCKER_CMD --help"
        ;;
        
    *)
        log_error "Unknown operation: $OPERATION"
        usage
        exit 1
        ;;
esac

# Display command being executed
log "Executing: $FINAL_CMD"
echo

# Execute the command
if eval "$FINAL_CMD"; then
    echo
    log_success "Operation completed successfully!"
    
    # Show report location if validation was run
    if [[ "$OPERATION" == "validate" ]] && [[ "$EXTRA_ARGS" == *"-report-output"* ]]; then
        echo
        echo "Report saved to: $REPORTS_DIR/"
        ls -la "$REPORTS_DIR/"
    fi
else
    echo
    log_error "Operation failed!"
    exit 1
fi