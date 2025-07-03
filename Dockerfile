# Multi-stage Dockerfile for Terratag
# Optimized for size and security with minimal runtime dependencies

# UI Build stage
FROM node:18-alpine AS ui-builder

WORKDIR /app/web/ui

# Copy package files
COPY web/ui/package*.json ./

# Install dependencies (including devDependencies for build)
RUN npm ci

# Copy UI source
COPY web/ui/ ./

# Build UI for production
RUN npm run build

# Go Build stage
FROM golang:1.23-alpine AS go-builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the CLI binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o terratag \
    cmd/terratag/main.go

# Build the API server binary (requires CGO for SQLite)
RUN CGO_ENABLED=1 go build \
    -ldflags='-w -s' \
    -o terratag-api \
    cmd/api/main.go

# Verify the binary
RUN ./terratag -version

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    git \
    openssh-client \
    curl \
    bash \
    jq \
    sqlite \
    && rm -rf /var/cache/apk/*

# Install Terraform
ARG TERRAFORM_VERSION=1.7.5
RUN wget -O terraform.zip "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip" \
    && unzip terraform.zip \
    && mv terraform /usr/local/bin/ \
    && rm terraform.zip \
    && terraform version

# Install OpenTofu (alternative to Terraform)
ARG OPENTOFU_VERSION=1.8.5
RUN wget -O tofu.tar.gz "https://github.com/opentofu/opentofu/releases/download/v${OPENTOFU_VERSION}/tofu_${OPENTOFU_VERSION}_linux_amd64.tar.gz" \
    && tar -xzf tofu.tar.gz \
    && mv tofu /usr/local/bin/ \
    && rm tofu.tar.gz \
    && tofu version

# Create non-root user for security
RUN addgroup -g 1000 terratag \
    && adduser -D -s /bin/bash -u 1000 -G terratag terratag

# Copy the binaries from go builder stage
COPY --from=go-builder /app/terratag /usr/local/bin/terratag
COPY --from=go-builder /app/terratag-api /usr/local/bin/terratag-api

# Copy the built UI from ui builder stage
COPY --from=ui-builder /app/web/ui/build /usr/share/terratag/web/ui/build

# Copy database migrations
COPY --from=go-builder /app/db /db

# Make binaries executable
RUN chmod +x /usr/local/bin/terratag /usr/local/bin/terratag-api

# Create working directories
RUN mkdir -p /workspace /standards /reports /demo-deployment /data \
    && chown -R terratag:terratag /workspace /standards /reports /demo-deployment /data

# Set working directory
WORKDIR /workspace

# Copy demo deployment files as root before switching user
COPY --chown=terratag:terratag demo-deployment /demo-deployment

# Copy entrypoint script
COPY --chown=terratag:terratag scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

# Switch to non-root user
USER terratag

# Verify installation
RUN terratag -version && terraform version && tofu version

# Initialize the demo deployment during build
RUN cd /demo-deployment && terraform init

# Set default command
ENTRYPOINT ["terratag"]
CMD ["--help"]

# Metadata
LABEL org.opencontainers.image.title="Terratag" \
      org.opencontainers.image.description="CLI tool for applying and validating tags across Terraform/OpenTofu files" \
      org.opencontainers.image.vendor="cloudyali" \
      org.opencontainers.image.url="https://github.com/cloudyali/terratag" \
      org.opencontainers.image.documentation="https://github.com/cloudyali/terratag/blob/main/README.md" \
      org.opencontainers.image.source="https://github.com/cloudyali/terratag" \
      org.opencontainers.image.licenses="Apache-2.0"