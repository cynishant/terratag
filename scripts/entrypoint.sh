#!/bin/bash
set -e

echo "Starting Terratag API Server..."

# Ensure terraform is initialized in demo deployment
if [ -d "/demo-deployment" ] && [ ! -d "/demo-deployment/.terraform" ]; then
    echo "Initializing Terraform in demo deployment..."
    cd /demo-deployment && terraform init
fi

# Start the API server
exec "$@"