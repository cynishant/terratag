# Terratag Demo Quick Start

Get the Terratag demo running in 3 simple steps!

## Prerequisites

- **Docker** installed and running
- **Git** repository cloned locally

## Quick Setup

### Step 1: Run Setup Script
```bash
# Navigate to terratag directory
cd terratag

# Run automated setup
./scripts/setup-demo.sh
```

This script will:
- ✅ Create required directories (`standards/`, `reports/`)
- ✅ Copy demo tag standard file
- ✅ Create `.env` configuration
- ✅ Verify all demo files
- ✅ Build Docker image
- ✅ Test the setup

### Step 2: Choose Your Demo Mode

#### **Option A: Web UI Demo** (Recommended for presentations)
```bash
docker-compose --profile ui up
```
Then open: **http://localhost:8080**

#### **Option B: CLI Demo** (Quick command-line demo)
```bash
./scripts/docker-demo.sh demo-basic
```

#### **Option C: Interactive Demo** (Hands-on exploration)
```bash
./scripts/docker-demo.sh demo-interactive
```

### Step 3: Explore and Demonstrate

## 🚀 Enhanced Standard Management Workflow

### **Creating Your First Standard (Web UI)**
1. **Start UI**: `docker-compose --profile ui up`
2. **Open**: http://localhost:8080
3. **Two ways to create standards**:
   - **Operations Page**: Click "➕ Create New" when selecting standards
   - **Standards Page**: Click "Create New Standard" button
4. **Choose from templates**: AWS Basic, GCP Basic, or start from scratch
5. **Customize**: Edit tags, validation rules, and resource-specific requirements
6. **Use immediately**: Select your new standard in operations

### **Loading Demo Standards (Optional)**
```bash
# Auto-load example standards into database
./scripts/load-demo-standards.sh
```
This loads:
- **AWS Demo Standard** (comprehensive example)
- **AWS Basic Template** (simple starter)
- **GCP Basic Template** (GCP equivalent)

## Web UI Features (http://localhost:8080)

### 📊 **Dashboard**
- Real-time tagging operations
- Validation status overview
- Resource compliance metrics

### 🏷️ **Tag Standards Management**
- **➕ Create new standards** directly from operations page
- **✏️ Edit existing standards** inline
- **📋 Select standards** for validation and tagging
- **🔄 Auto-refresh** standards list
- **📝 Template library** with AWS/GCP examples

### ✅ **Operations & Validation**
- **Smart standard selection** with provider filtering
- **Required/optional indicators** for validation vs tagging
- **Integrated standard creation** when none exist
- **Rich standard details** showing version, provider, description
- Interactive compliance reports and drill-down violation details

### 📁 **Resource Explorer**
- Browse Terraform resources
- Filter by type and compliance
- Bulk tag operations

## CLI Demonstration Commands

### Basic Tag Application
```bash
# Apply demo tags to all resources
./scripts/docker-demo.sh demo-basic

# View generated files
find demo-deployment -name "*.terratag.tf" | head -5
```

### Tag Validation
```bash
# Validate against tag standards
./scripts/docker-demo.sh demo-validation

# Generate JSON compliance report
docker run --rm \
  -v $(pwd)/demo-deployment:/demo-deployment \
  -v $(pwd)/reports:/reports \
  terratag:latest \
  -validate-only \
  -standard=/demo-deployment/tag-standard.yaml \
  -dir=/demo-deployment \
  -report-format=json \
  -report-output=/reports/compliance.json
```

### Interactive Exploration
```bash
# Start interactive shell
./scripts/docker-demo.sh demo-interactive

# Inside the container:
terratag --help
terratag -validate-only -standard=/demo-deployment/tag-standard.yaml -dir=/demo-deployment
exit
```

## Demo Infrastructure

The demo includes a complete AWS multi-tier application:

- **🌐 VPC** with public/private subnets
- **⚖️ Application Load Balancer** 
- **🖥️ Auto Scaling Group** with EC2 instances
- **🗄️ RDS MySQL database**
- **📦 S3 buckets** for data, logs, and backups
- **🔒 Security groups** and IAM roles
- **📈 CloudWatch monitoring**

## Quick Demo Script (2 minutes)

```bash
# 1. Apply tags and show changes
./scripts/docker-demo.sh demo-basic
echo "✅ Tags applied! Check demo-deployment/*.terratag.tf"

# 2. Validate compliance
./scripts/docker-demo.sh demo-validation
echo "✅ Validation complete! Check reports/"

# 3. Start web UI
docker-compose --profile ui up -d
echo "✅ Web UI started at http://localhost:8080"
```

## Troubleshooting

### Common Issues

**Port 8080 in use:**
```bash
docker-compose --profile ui up
# Change port in docker-compose.yml or:
docker run -p 8081:8080 ... # Use port 8081
```

**Permission denied:**
```bash
sudo chown -R $(id -u):$(id -g) demo-deployment reports standards
```

**Demo files missing:**
```bash
# Re-run setup
./scripts/setup-demo.sh
```

**Docker build fails:**
```bash
# Clean build
docker build --no-cache -t terratag:latest .
```

## What's Included

### 📁 **File Structure**
```
terratag/
├── demo-deployment/          # Complete Terraform example
│   ├── main.tf              # VPC and networking
│   ├── compute.tf            # EC2 and auto scaling  
│   ├── database.tf           # RDS configuration
│   ├── storage.tf            # S3 buckets
│   └── tag-standard.yaml     # Validation rules
├── standards/                # Mounted tag standards
├── reports/                  # Generated compliance reports
└── scripts/
    ├── setup-demo.sh         # Automated setup
    └── docker-demo.sh        # Demo commands
```

### 🏷️ **Sample Tags Applied**
```json
{
  "Environment": "Demo",
  "Owner": "demo@company.com",
  "Project": "TerratagDemo",
  "ManagedBy": "Terraform",
  "DemoRun": "2024-01-01-123456"
}
```

### ✅ **Tag Validation Rules**
- **Required**: Environment, Owner, Project, ManagedBy
- **Optional**: CostCenter, BackupSchedule, DataClassification
- **Resource-specific**: Additional rules for EC2, RDS, S3
- **Format validation**: Email addresses, cost center codes

## Next Steps

1. **📖 Read Full Guide**: See `DEMO_GUIDE.md` for detailed scenarios
2. **🐳 Docker Details**: Check `docs/DOCKER_DEMO.md` for Docker specifics  
3. **⚙️ Customize**: Modify `demo-deployment/tag-standard.yaml` for your rules
4. **🚀 Production**: Use with your own Terraform files

## Cleanup

```bash
# Stop all services
docker-compose down

# Remove demo containers
docker container prune -f

# Remove demo image (optional)
docker rmi terratag:latest
```

---

**🎯 Ready to demo Terratag in under 5 minutes!**

For detailed documentation, see:
- `DEMO_GUIDE.md` - Complete demonstration scenarios
- `docs/DOCKER_DEMO.md` - Docker-specific instructions
- `demo-deployment/README.md` - Infrastructure details