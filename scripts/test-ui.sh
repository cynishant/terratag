#!/bin/bash

# Test UI Script - Quick verification of the UI setup

set -e

echo "🧪 Testing Terratag UI Setup"
echo "============================="

# Check if required directories exist
echo "📁 Checking directories..."
if [[ ! -d "demo-deployment" ]]; then
    echo "❌ demo-deployment directory not found"
    exit 1
fi

if [[ ! -d "standards" ]]; then
    echo "❌ standards directory not found"
    exit 1
fi

if [[ ! -d "reports" ]]; then
    echo "❌ reports directory not found"
    exit 1
fi

echo "✅ All directories found"

# Check if tag standard exists
echo "📋 Checking tag standard..."
if [[ ! -f "standards/tag-standard.yaml" ]]; then
    echo "❌ Tag standard not found in standards/"
    echo "📋 Copying from demo-deployment..."
    cp demo-deployment/tag-standard.yaml standards/
fi

echo "✅ Tag standard verified"

# Build the image if it doesn't exist
echo "🐳 Checking Docker image..."
if ! docker image inspect terratag:latest >/dev/null 2>&1; then
    echo "🔨 Building Docker image..."
    docker build -t terratag:latest .
fi

echo "✅ Docker image ready"

# Start the UI service directly
echo "🚀 Starting UI service..."
echo "📍 UI will be available at: http://localhost:8080"
echo "🛑 Press Ctrl+C to stop"

docker run --rm \
  -v "$(pwd)/demo-deployment:/demo-deployment" \
  -v "$(pwd)/standards:/standards:ro" \
  -v "$(pwd)/reports:/reports" \
  -v "$(pwd):/workspace" \
  -p 8080:8080 \
  -e PORT=8080 \
  -e DB_PATH=/workspace/terratag.db \
  -e GIN_MODE=debug \
  -e TERRATAG_DEMO_MODE=true \
  -e TERRATAG_DEMO_DIR=/demo-deployment \
  -e TERRATAG_STANDARDS_DIR=/standards \
  --name terratag-ui-test \
  --entrypoint /usr/local/bin/terratag-api \
  terratag:latest