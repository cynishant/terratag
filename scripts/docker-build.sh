#!/bin/bash
# Docker build script for Terratag
# Builds optimized Docker image with proper tagging

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
IMAGE_NAME="terratag"
TAG="latest"
PUSH=false
PLATFORM="linux/amd64"
BUILD_ARGS=""
CACHE_FROM=""

# Function to display usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Build Terratag Docker image with optimizations

OPTIONS:
    -t, --tag TAG           Image tag (default: latest)
    -n, --name NAME         Image name (default: terratag)
    -p, --push              Push image to registry after build
    -m, --multi-platform    Build for multiple platforms (linux/amd64,linux/arm64)
    --cache-from IMAGE      Use cache from existing image
    --no-cache              Build without cache
    --build-arg ARG=VALUE   Pass build argument
    -h, --help              Display this help message

EXAMPLES:
    # Basic build
    $0

    # Build with custom tag
    $0 -t v1.2.3

    # Build and push to registry
    $0 -t v1.2.3 -p

    # Multi-platform build
    $0 -m -t latest

    # Build with custom Terraform version
    $0 --build-arg TERRAFORM_VERSION=1.7.0

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

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--tag)
            TAG="$2"
            shift 2
            ;;
        -n|--name)
            IMAGE_NAME="$2"
            shift 2
            ;;
        -p|--push)
            PUSH=true
            shift
            ;;
        -m|--multi-platform)
            PLATFORM="linux/amd64,linux/arm64"
            shift
            ;;
        --cache-from)
            CACHE_FROM="--cache-from $2"
            shift 2
            ;;
        --no-cache)
            BUILD_ARGS="$BUILD_ARGS --no-cache"
            shift
            ;;
        --build-arg)
            BUILD_ARGS="$BUILD_ARGS --build-arg $2"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Change to project root
cd "$PROJECT_ROOT"

# Verify Docker is available
if ! command -v docker &> /dev/null; then
    log_error "Docker is not installed or not in PATH"
    exit 1
fi

# Create full image name
FULL_IMAGE_NAME="${IMAGE_NAME}:${TAG}"

log "Building Terratag Docker image..."
log "Image: $FULL_IMAGE_NAME"
log "Platform: $PLATFORM"

# Build the image
log "Starting Docker build..."

if [[ "$PLATFORM" == *","* ]]; then
    # Multi-platform build requires buildx
    if ! docker buildx version &> /dev/null; then
        log_error "Multi-platform build requires Docker Buildx"
        exit 1
    fi
    
    # Create builder if it doesn't exist
    if ! docker buildx inspect terratag-builder &> /dev/null; then
        log "Creating buildx builder..."
        docker buildx create --name terratag-builder --use
    fi
    
    BUILD_CMD="docker buildx build --platform $PLATFORM"
    if [[ "$PUSH" == true ]]; then
        BUILD_CMD="$BUILD_CMD --push"
    else
        BUILD_CMD="$BUILD_CMD --load"
    fi
else
    BUILD_CMD="docker build --platform $PLATFORM"
fi

# Execute build
BUILD_CMD="$BUILD_CMD $BUILD_ARGS $CACHE_FROM -t $FULL_IMAGE_NAME -f Dockerfile ."

log "Executing: $BUILD_CMD"

if eval "$BUILD_CMD"; then
    log_success "Docker image built successfully: $FULL_IMAGE_NAME"
else
    log_error "Docker build failed"
    exit 1
fi

# Verify the image
log "Verifying image..."
if docker run --rm "$FULL_IMAGE_NAME" -version; then
    log_success "Image verification successful"
else
    log_error "Image verification failed"
    exit 1
fi

# Push if requested (and not multi-platform)
if [[ "$PUSH" == true ]] && [[ "$PLATFORM" != *","* ]]; then
    log "Pushing image to registry..."
    if docker push "$FULL_IMAGE_NAME"; then
        log_success "Image pushed successfully: $FULL_IMAGE_NAME"
    else
        log_error "Failed to push image"
        exit 1
    fi
fi

# Display image information
log "Image build complete!"
echo
echo "Image Details:"
echo "  Name: $FULL_IMAGE_NAME"
echo "  Platform: $PLATFORM"
echo "  Size: $(docker images --format "table {{.Size}}" "$FULL_IMAGE_NAME" | tail -n1)"
echo
echo "Usage Examples:"
echo "  # Run Terratag validation"
echo "  docker run --rm -v \$(pwd):/workspace $FULL_IMAGE_NAME -validate-only -standard /standards/tag-standard.yaml"
echo
echo "  # Run with Docker Compose"
echo "  docker-compose --profile validate up"
echo
echo "  # Interactive shell"
echo "  docker run --rm -it -v \$(pwd):/workspace $FULL_IMAGE_NAME /bin/bash"

log_success "Build completed successfully!"