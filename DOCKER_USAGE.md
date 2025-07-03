# Docker Usage Guide

Terratag provides two Docker profiles for different use cases:

## UI Mode
Run the web interface and API server:
```bash
docker-compose --profile ui up
```
Access the UI at http://localhost:8080

To run in background:
```bash
docker-compose --profile ui up -d
```

## CLI Mode
Run an interactive shell for command-line usage:
```bash
docker-compose --profile cli run --rm terratag-cli
```

Inside the container, you can run terratag commands:
```bash
# Apply tags
terratag -tags='{"Environment":"prod","Owner":"devops"}' -dir=.

# Validate against standard
terratag -validate-only -standard=/standards/tag-standard.yaml -dir=.

# Generate report
terratag -validate-only -standard=/standards/tag-standard.yaml -report-format=json -report-output=/reports/compliance.json -dir=.
```

## Environment Variables
You can customize behavior using environment variables in a `.env` file:

```env
# UI Configuration
GIN_MODE=release          # Set to 'debug' for development
LOG_LEVEL=info           # Options: debug, info, warn, error

# Workspace Configuration
TERRATAG_SOURCE_DIR=./my-terraform-code
TERRATAG_STANDARDS_DIR=./my-standards
TERRATAG_REPORTS_DIR=./my-reports

# Cloud Credentials
AWS_PROFILE=my-profile
AWS_REGION=us-west-2
GOOGLE_PROJECT=my-project
```

## Stopping Services
```bash
# Stop UI
docker-compose --profile ui down

# Stop and remove volumes (clean start)
docker-compose --profile ui down -v
```