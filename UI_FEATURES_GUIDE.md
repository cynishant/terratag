# Terratag UI Features Guide

Complete guide to creating, selecting, and managing tag standards directly from the Terratag web interface.

## 🎯 Overview

The Terratag UI now provides **integrated tag standard management** directly within the operations workflow, making it easy to create, select, and use standards for validation and tagging operations.

## 🏷️ Tag Standard Management Features

### **1. Integrated Standard Creation**

#### **From Operations Page**
- **➕ Create New Button**: Available in the standard selection section
- **Smart Context**: Automatically appears when creating validation or tagging operations
- **Instant Usage**: Newly created standards immediately available for selection

#### **From Standards Page**
- **Dedicated Management**: Full-featured standard creation and editing
- **Bulk Management**: View, edit, delete multiple standards
- **Template Library**: Start from pre-built templates

### **2. Enhanced Standard Selection**

#### **Visual Design**
- 🎨 **Blue highlighted section** for prominence
- 📋 **Clear labeling** with emojis and context
- ✅ **Rich feedback** showing selected standard details
- ⚠️ **Smart warnings** for missing requirements

#### **Smart Behaviors**
- **Required for Validation**: Clear indication and validation
- **Optional for Tagging**: Helpful but not required
- **Auto-filtering**: Standards filtered by relevant cloud provider
- **Rich descriptions**: Full standard details in dropdown

#### **Action Buttons**
- **➕ Create New**: Launch standard editor modal
- **✏️ Edit**: Edit currently selected standard
- **🔄 Refresh**: Reload standards from database

### **3. Contextual Help & Guidance**

#### **Empty State Handling**
```
📝 No tag standards found. Create your first standard to get started.
```

#### **Validation Requirements**
```
⚠️ A tag standard is required for validation operations. Create one now
```

#### **Standard Selection Feedback**
```
✅ Selected Standard: AWS Demo Standard
Provider: AWS | Version: 1 | Description: Comprehensive AWS tagging standard...
```

## 📋 Creating Standards: Step-by-Step

### **Method 1: From Operations Page (Recommended)**

1. **Navigate to Operations**
   - Open http://localhost:8080
   - Go to "Operations" page

2. **Start Creating Operation**
   - Click "Create New Operation"
   - Select "Validation" or "Tagging"

3. **Create Standard**
   - In the Tag Standard section, click "➕ Create New"
   - Standard editor modal opens
   - Fill in details and save
   - Standard immediately available for selection

### **Method 2: From Standards Page**

1. **Navigate to Standards**
   - Go to "Standards" page
   - Click "Create New Standard"

2. **Use Template or Start Fresh**
   - Choose from AWS/GCP templates
   - Or start with blank standard

3. **Customize Standard**
   - Define required/optional tags
   - Set validation rules
   - Add resource-specific requirements

### **Method 3: Load Demo Standards**

```bash
# Load pre-built standards into database
./scripts/load-demo-standards.sh
```

This automatically loads:
- **AWS Demo Standard**: Comprehensive example from demo-deployment
- **AWS Basic Template**: Simple starter template
- **GCP Basic Template**: GCP labeling equivalent

## 🛠️ Standard Templates

### **AWS Basic Template** (`standards/aws-basic-template.yaml`)
```yaml
required_tags:
  - Environment (Production/Staging/Development/Testing)
  - Owner (email format)
  - Project (string, 2-50 chars)
  - ManagedBy (Terraform/CloudFormation/Manual)

optional_tags:
  - CostCenter (CC-#### format)
  - BackupSchedule (Daily/Weekly/Monthly/None)

resource_rules:
  - EC2 instances require BackupSchedule
  - RDS instances require BackupSchedule
```

### **GCP Basic Template** (`standards/gcp-basic-template.yaml`)
```yaml
required_tags:
  - environment (lowercase, production/staging/development/testing)
  - owner (email format)
  - project (lowercase, hyphenated format)
  - managed-by (terraform/deployment-manager/manual)

optional_tags:
  - cost-center (cc-#### format)
  - backup-schedule (daily/weekly/monthly/none)

resource_rules:
  - Compute instances require backup-schedule
  - SQL instances require backup-schedule
```

## 🔄 Workflow Examples

### **Example 1: First-Time User**

1. **Start Fresh**
   ```bash
   docker-compose down -v  # Reset database
   docker-compose --profile ui up
   ```

2. **Create First Standard**
   - Open http://localhost:8080
   - Go to Operations → Create New Operation
   - Select "Validation"
   - See message: "No tag standards found"
   - Click "Create your first standard"

3. **Use Template**
   - Choose "AWS Basic Template"
   - Customize for your organization
   - Save and immediately use for validation

### **Example 2: Team Collaboration**

1. **Load Organization Standards**
   ```bash
   ./scripts/load-demo-standards.sh
   ```

2. **Select and Customize**
   - Operations page shows available standards
   - Select "AWS Demo Standard"
   - Click "✏️ Edit" to customize
   - Save changes for team use

3. **Run Validation**
   - Standard pre-selected
   - Configure directory path
   - Execute validation with team standard

### **Example 3: Multi-Cloud Environment**

1. **Create AWS Standard**
   - Use AWS Basic Template
   - Customize for AWS resources

2. **Create GCP Standard**
   - Use GCP Basic Template
   - Adapt for GCP naming conventions

3. **Select by Context**
   - UI automatically filters by cloud provider
   - Choose appropriate standard for each operation

## 🚀 Advanced Features

### **Standard Editing Workflow**
1. **Select Standard**: Choose from dropdown
2. **Click Edit**: "✏️ Edit" button appears
3. **Modify**: Change validation rules, tags, descriptions
4. **Save**: Changes immediately available
5. **Use**: Updated standard ready for operations

### **Database Management**
```bash
# Reset everything (including standards)
docker-compose down -v

# Keep standards, restart services
docker-compose restart

# Reload demo standards after reset
docker-compose --profile ui up -d
./scripts/load-demo-standards.sh
```

### **Standard Validation Features**
- **Real-time validation** of YAML syntax
- **Provider-specific rules** (AWS vs GCP vs Azure)
- **Resource type validation** for applicable resources
- **Tag format validation** (email, regex patterns, etc.)

## 📊 UI Benefits Summary

| Feature | Benefit |
|---------|---------|
| **Integrated Creation** | No context switching - create standards when needed |
| **Smart Selection** | Context-aware filtering and validation |
| **Rich Feedback** | Clear indication of requirements and status |
| **Template Library** | Quick start with proven patterns |
| **Inline Editing** | Modify standards without leaving workflow |
| **Auto-refresh** | Always see latest standards |
| **Empty State Guidance** | Clear next steps when no standards exist |
| **Validation Integration** | Immediate feedback on standard requirements |

## 🎯 Best Practices

### **For Organizations**
1. **Start with Templates**: Use provided AWS/GCP templates as base
2. **Customize Gradually**: Add organization-specific requirements over time
3. **Test First**: Validate with small directory before organization-wide rollout
4. **Document Standards**: Use description fields to explain requirements

### **For Developers**
1. **Create Development Standards**: Relaxed rules for development environments
2. **Use Validation Mode**: Check compliance before applying tags
3. **Iterate Standards**: Edit and refine based on validation results
4. **Load Demo Standards**: Learn from comprehensive examples

### **For Demos**
1. **Reset Between Demos**: `docker-compose down -v` for clean slate
2. **Load Examples**: `./scripts/load-demo-standards.sh` for variety
3. **Show Creation Flow**: Demonstrate standard creation during presentation
4. **Highlight Integration**: Show seamless standard selection workflow

The enhanced UI makes tag standard management a seamless part of the Terratag workflow, enabling teams to quickly create, customize, and apply consistent tagging standards across their infrastructure.