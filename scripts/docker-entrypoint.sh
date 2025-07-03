#!/bin/bash
# Docker entrypoint script for Terratag with auto-initialization

# Function to check and initialize Terraform directories
auto_init_terraform() {
    local dir="$1"
    
    # Check if directory exists and contains .tf files
    if [ -d "$dir" ] && find "$dir" -name "*.tf" -type f | grep -q .; then
        # Check if .terraform directory exists
        if [ ! -d "$dir/.terraform" ]; then
            echo "[AUTO-INIT] Initializing Terraform in $dir..."
            cd "$dir" && terraform init
            if [ $? -eq 0 ]; then
                echo "[AUTO-INIT] Successfully initialized $dir"
            else
                echo "[AUTO-INIT] Failed to initialize $dir"
            fi
            cd - > /dev/null
        else
            echo "[AUTO-INIT] $dir is already initialized"
        fi
    fi
}

# Auto-initialize demo deployment if it exists
if [ -d "/demo-deployment" ]; then
    auto_init_terraform "/demo-deployment"
fi

# Auto-initialize workspace if TERRATAG_AUTO_INIT is set
if [ "$TERRATAG_AUTO_INIT" = "true" ] && [ -n "$1" ]; then
    # Extract directory from arguments
    for arg in "$@"; do
        case "$arg" in
            -dir=*|--dir=*)
                dir="${arg#*=}"
                auto_init_terraform "$dir"
                ;;
        esac
    done
fi

# Execute the original command
exec "$@"