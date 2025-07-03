# Docker Examples for Terratag

This directory contains practical Docker and Docker Compose examples for using Terratag in various scenarios.

## Quick Start Examples

### 1. Basic Tag Validation

```bash
# Validate AWS infrastructure
docker run --rm \
  -v $(pwd):/workspace \
  -v $(pwd)/standards:/standards:ro \
  terratag:latest \
  -validate-only -standard /standards/aws-standard.yaml

# With Docker Compose
docker-compose --profile validate up
```

### 2. Apply Tags to Resources

```bash
# Apply production tags
docker run --rm \
  -v $(pwd):/workspace \
  terratag:latest \
  -tags='{"Environment":"production","Owner":"devops@company.com","CostCenter":"CC1001"}'

# With Docker Compose
export ENVIRONMENT=production
export OWNER=devops@company.com
export COST_CENTER=CC1001
docker-compose --profile apply up
```

### 3. Generate Compliance Reports

```bash
# Generate JSON report
docker run --rm \
  -v $(pwd):/workspace \
  -v $(pwd)/standards:/standards:ro \
  -v $(pwd)/reports:/reports \
  terratag:latest \
  -validate-only \
  -standard /standards/compliance-standard.yaml \
  -report-format json \
  -report-output /reports/compliance-report.json
```

## Scenario-Based Examples

### AWS Production Environment

```yaml
# docker-compose.aws.yml
version: '3.8'
services:
  aws-prod-validation:
    image: terratag:latest
    command: >
      -validate-only
      -standard /standards/aws-production.yaml
      -report-format markdown
      -report-output /reports/aws-prod-compliance.md
      -strict-mode
    volumes:
      - ./aws-infrastructure:/workspace:ro
      - ./standards:/standards:ro
      - ./reports:/reports
      - ~/.aws:/home/terratag/.aws:ro
    environment:
      - AWS_PROFILE=production
      - AWS_REGION=us-east-1
```

Usage:
```bash
docker-compose -f docker-compose.aws.yml up
```

### GCP Multi-Project Setup

```yaml
# docker-compose.gcp.yml
version: '3.8'
services:
  gcp-validation:
    image: terratag:latest
    command: >
      -validate-only
      -standard /standards/gcp-multi-project.yaml
      -report-format json
      -report-output /reports/gcp-compliance.json
    volumes:
      - ./gcp-projects:/workspace:ro
      - ./standards:/standards:ro
      - ./reports:/reports
      - ~/.config/gcloud:/home/terratag/.config/gcloud:ro
    environment:
      - GOOGLE_APPLICATION_CREDENTIALS=/home/terratag/.config/gcloud/credentials.json
      - GOOGLE_PROJECT=my-primary-project
```

### CI/CD Pipeline Integration

```yaml
# docker-compose.cicd.yml
version: '3.8'
services:
  cicd-validation:
    image: terratag:latest
    command: >
      -validate-only
      -standard /standards/cicd-standard.yaml
      -report-format json
      -report-output /reports/pipeline-compliance.json
      -strict-mode
    volumes:
      - ${CI_WORKSPACE:-./}:/workspace:ro
      - ./standards:/standards:ro
      - ./reports:/reports
    environment:
      - CI=true
      - PIPELINE_ID=${CI_PIPELINE_ID:-local}
```

## Advanced Examples

### 1. Multi-Environment Batch Processing

```bash
#!/bin/bash
# batch-validate.sh

environments=("dev" "staging" "prod")

for env in "${environments[@]}"; do
  echo "Validating $env environment..."
  
  docker run --rm \
    -v $(pwd)/${env}-infrastructure:/workspace:ro \
    -v $(pwd)/standards:/standards:ro \
    -v $(pwd)/reports:/reports \
    terratag:latest \
    -validate-only \
    -standard /standards/${env}-standard.yaml \
    -report-format json \
    -report-output /reports/${env}-compliance.json \
    -strict-mode
    
  echo "$env validation completed."
done

# Generate summary report
docker run --rm \
  -v $(pwd)/reports:/reports \
  -v $(pwd)/scripts:/scripts:ro \
  terratag:latest \
  /scripts/generate-summary.sh
```

### 2. Custom Tag Application with Filtering

```bash
# Apply tags only to compute and storage resources
docker run --rm \
  -v $(pwd):/workspace \
  terratag:latest \
  -tags='{"Backup":"required","Monitoring":"enabled"}' \
  -filter="aws_(instance|ebs_volume|s3_bucket)" \
  -verbose

# Apply different tags based on resource type
docker run --rm \
  -v $(pwd):/workspace \
  terratag:latest \
  -tags='{"SecurityLevel":"high"}' \
  -filter="aws_(security_group|iam_)" \
  -verbose
```

### 3. Compliance Audit Workflow

```yaml
# docker-compose.audit.yml
version: '3.8'
services:
  compliance-audit:
    image: terratag:latest
    command: >
      /bin/bash -c "
      echo 'Starting compliance audit...';
      
      # Run comprehensive validation
      terratag -validate-only 
        -standard /standards/audit-standard.yaml 
        -report-format markdown 
        -report-output /reports/audit-$(date +%Y%m%d).md 
        -strict-mode 
        -verbose 
        /workspace;
        
      # Generate summary statistics
      terratag -validate-only 
        -standard /standards/audit-standard.yaml 
        -report-format json 
        -report-output /reports/audit-data-$(date +%Y%m%d).json 
        /workspace;
        
      echo 'Audit completed. Reports generated.';
      "
    volumes:
      - ./infrastructure:/workspace:ro
      - ./standards:/standards:ro
      - ./reports:/reports
    environment:
      - AUDIT_DATE=$(date +%Y%m%d)
      - AUDITOR=automated-system
```

### 4. Development Environment Setup

```yaml
# docker-compose.dev.yml
version: '3.8'
services:
  dev-environment:
    image: terratag:latest
    command: /bin/bash
    stdin_open: true
    tty: true
    volumes:
      - ./:/workspace
      - ./standards:/standards
      - ./reports:/reports
      - ~/.aws:/home/terratag/.aws:ro
      - ~/.gitconfig:/home/terratag/.gitconfig:ro
      - ~/.ssh:/home/terratag/.ssh:ro
    environment:
      - AWS_PROFILE=dev
      - TERRATAG_VERBOSE=true
    working_dir: /workspace
```

## Integration Examples

### GitHub Actions

```yaml
# .github/workflows/terratag-validation.yml
name: Terratag Validation

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  validate-tags:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
        
      - name: Build Terratag image
        run: docker build -t terratag:ci .
        
      - name: Validate infrastructure tags
        run: |
          docker run --rm \
            -v ${{ github.workspace }}:/workspace:ro \
            -v ${{ github.workspace }}/standards:/standards:ro \
            -v ${{ github.workspace }}/reports:/reports \
            terratag:ci \
            -validate-only \
            -standard /standards/production-standard.yaml \
            -report-format json \
            -report-output /reports/validation-${{ github.sha }}.json \
            -strict-mode
            
      - name: Upload validation report
        uses: actions/upload-artifact@v3
        if: always()
        with:
          name: validation-report
          path: reports/validation-*.json
          
      - name: Comment PR with results
        if: github.event_name == 'pull_request'
        uses: actions/github-script@v6
        with:
          script: |
            const fs = require('fs');
            const report = JSON.parse(fs.readFileSync('reports/validation-${{ github.sha }}.json'));
            
            const comment = `
            ## 🏷️ Terratag Validation Results
            
            **Compliance Rate:** ${(report.compliance_rate * 100).toFixed(1)}%
            **Total Resources:** ${report.total_resources}
            **Compliant:** ${report.compliant_resources}
            **Non-compliant:** ${report.non_compliant_resources}
            
            ${report.non_compliant_resources > 0 ? '❌ Some resources are not compliant with tagging standards.' : '✅ All resources are compliant!'}
            `;
            
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: comment
            });
```

### GitLab CI

```yaml
# .gitlab-ci.yml
stages:
  - build
  - validate
  - report

variables:
  DOCKER_DRIVER: overlay2
  DOCKER_TLS_CERTDIR: "/certs"

build-terratag:
  stage: build
  services:
    - docker:20.10.16-dind
  script:
    - docker build -t $CI_REGISTRY_IMAGE/terratag:$CI_COMMIT_SHA .
    - docker push $CI_REGISTRY_IMAGE/terratag:$CI_COMMIT_SHA

validate-infrastructure:
  stage: validate
  services:
    - docker:20.10.16-dind
  script:
    - docker pull $CI_REGISTRY_IMAGE/terratag:$CI_COMMIT_SHA
    - |
      docker run --rm \
        -v $CI_PROJECT_DIR:/workspace:ro \
        -v $CI_PROJECT_DIR/standards:/standards:ro \
        -v $CI_PROJECT_DIR/reports:/reports \
        $CI_REGISTRY_IMAGE/terratag:$CI_COMMIT_SHA \
        -validate-only \
        -standard /standards/gitlab-standard.yaml \
        -report-format json \
        -report-output /reports/gitlab-compliance.json \
        -strict-mode
  artifacts:
    reports:
      junit: reports/gitlab-compliance.xml
    paths:
      - reports/
    expire_in: 30 days

generate-report:
  stage: report
  dependencies:
    - validate-infrastructure
  script:
    - |
      docker run --rm \
        -v $CI_PROJECT_DIR/reports:/reports \
        $CI_REGISTRY_IMAGE/terratag:$CI_COMMIT_SHA \
        -validate-only \
        -standard /standards/gitlab-standard.yaml \
        -report-format markdown \
        -report-output /reports/compliance-report.md \
        /workspace
  artifacts:
    paths:
      - reports/compliance-report.md
    expire_in: 30 days
```

### Jenkins Pipeline

```groovy
// Jenkinsfile
pipeline {
    agent any
    
    environment {
        TERRATAG_IMAGE = "terratag:${env.BUILD_ID}"
    }
    
    stages {
        stage('Build Terratag Image') {
            steps {
                script {
                    docker.build(env.TERRATAG_IMAGE)
                }
            }
        }
        
        stage('Validate Tags') {
            parallel {
                stage('AWS Infrastructure') {
                    steps {
                        script {
                            docker.image(env.TERRATAG_IMAGE).inside(
                                "-v ${workspace}/aws:/workspace:ro " +
                                "-v ${workspace}/standards:/standards:ro " +
                                "-v ${workspace}/reports:/reports"
                            ) {
                                sh '''
                                    terratag -validate-only \
                                      -standard /standards/aws-standard.yaml \
                                      -report-format json \
                                      -report-output /reports/aws-compliance.json \
                                      -strict-mode \
                                      /workspace
                                '''
                            }
                        }
                    }
                }
                
                stage('GCP Infrastructure') {
                    steps {
                        script {
                            docker.image(env.TERRATAG_IMAGE).inside(
                                "-v ${workspace}/gcp:/workspace:ro " +
                                "-v ${workspace}/standards:/standards:ro " +
                                "-v ${workspace}/reports:/reports"
                            ) {
                                sh '''
                                    terratag -validate-only \
                                      -standard /standards/gcp-standard.yaml \
                                      -report-format json \
                                      -report-output /reports/gcp-compliance.json \
                                      -strict-mode \
                                      /workspace
                                '''
                            }
                        }
                    }
                }
            }
        }
        
        stage('Generate Summary') {
            steps {
                script {
                    docker.image(env.TERRATAG_IMAGE).inside(
                        "-v ${workspace}/reports:/reports"
                    ) {
                        sh '''
                            echo "Generating compliance summary..."
                            # Custom script to merge reports
                            /scripts/merge-compliance-reports.sh
                        '''
                    }
                }
            }
        }
    }
    
    post {
        always {
            archiveArtifacts artifacts: 'reports/**/*.json', fingerprint: true
            publishHTML([
                allowMissing: false,
                alwaysLinkToLastBuild: false,
                keepAll: true,
                reportDir: 'reports',
                reportFiles: '*.html',
                reportName: 'Terratag Compliance Report'
            ])
        }
        
        failure {
            emailext (
                subject: "Terratag Validation Failed: ${env.JOB_NAME} - ${env.BUILD_NUMBER}",
                body: "The Terratag validation pipeline has failed. Please check the build logs.",
                to: "${env.CHANGE_AUTHOR_EMAIL}"
            )
        }
    }
}
```

## Utility Scripts

### Report Analysis Script

```bash
#!/bin/bash
# scripts/analyze-reports.sh

REPORTS_DIR="./reports"

echo "Terratag Compliance Analysis"
echo "============================"

# Find all JSON reports
reports=$(find "$REPORTS_DIR" -name "*.json" -type f)

total_resources=0
total_compliant=0
total_non_compliant=0

for report in $reports; do
    if [[ -f "$report" ]]; then
        echo "Processing: $(basename "$report")"
        
        resources=$(jq -r '.total_resources // 0' "$report")
        compliant=$(jq -r '.compliant_resources // 0' "$report")
        non_compliant=$(jq -r '.non_compliant_resources // 0' "$report")
        
        echo "  Resources: $resources"
        echo "  Compliant: $compliant"
        echo "  Non-compliant: $non_compliant"
        echo ""
        
        total_resources=$((total_resources + resources))
        total_compliant=$((total_compliant + compliant))
        total_non_compliant=$((total_non_compliant + non_compliant))
    fi
done

echo "Summary"
echo "-------"
echo "Total Resources: $total_resources"
echo "Total Compliant: $total_compliant"
echo "Total Non-compliant: $total_non_compliant"

if [[ $total_resources -gt 0 ]]; then
    compliance_rate=$(echo "scale=2; $total_compliant * 100 / $total_resources" | bc)
    echo "Compliance Rate: ${compliance_rate}%"
fi
```

### Multi-Environment Validation

```bash
#!/bin/bash
# scripts/validate-all-environments.sh

environments=("development" "staging" "production")
standards_dir="./standards"
reports_dir="./reports"

mkdir -p "$reports_dir"

for env in "${environments[@]}"; do
    echo "Validating $env environment..."
    
    if [[ -f "$standards_dir/${env}-standard.yaml" ]]; then
        docker run --rm \
            -v "$(pwd)/${env}-infrastructure":/workspace:ro \
            -v "$(pwd)/$standards_dir":/standards:ro \
            -v "$(pwd)/$reports_dir":/reports \
            terratag:latest \
            -validate-only \
            -standard "/standards/${env}-standard.yaml" \
            -report-format json \
            -report-output "/reports/${env}-compliance-$(date +%Y%m%d).json" \
            -strict-mode \
            -verbose
    else
        echo "Warning: Standard file for $env not found"
    fi
done

echo "All environments validated. Reports available in $reports_dir/"
```

## Best Practices

### 1. Image Management
- Use specific tags instead of `latest` in production
- Implement image scanning for security vulnerabilities
- Use multi-stage builds to minimize image size
- Cache images in CI/CD pipelines for performance

### 2. Volume Mounts
- Use read-only mounts for input directories
- Separate standards and reports directories
- Mount credentials securely and read-only
- Use named volumes for cache persistence

### 3. Security
- Run containers as non-root user
- Limit container capabilities
- Use secrets management for sensitive data
- Scan images for vulnerabilities

### 4. Performance
- Use `.dockerignore` to exclude unnecessary files
- Implement build caching strategies
- Optimize Docker layer ordering
- Use specific CPU and memory limits

## Troubleshooting

### Common Issues

1. **Permission Errors**
   ```bash
   # Fix file permissions
   docker run --rm -v $(pwd):/workspace --user root terratag:latest chown -R 1000:1000 /workspace
   ```

2. **Volume Mount Issues**
   ```bash
   # Use absolute paths
   docker run --rm -v "$(realpath .):/workspace" terratag:latest -version
   ```

3. **Credential Problems**
   ```bash
   # Verify AWS credentials
   docker run --rm -v ~/.aws:/home/terratag/.aws:ro -e AWS_PROFILE=default terratag:latest aws sts get-caller-identity
   ```

4. **Memory Issues**
   ```bash
   # Increase container memory
   docker run --rm --memory="1g" -v $(pwd):/workspace terratag:latest -validate-only -standard /standards/large.yaml
   ```

For more examples and documentation, see the [main Docker usage guide](../../docs/DOCKER_USAGE.md).