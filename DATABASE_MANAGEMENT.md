# Database Management Guide

This guide explains how to manage the SQLite database in the Terratag Docker environment.

## SQLite Volume Configuration

### **Where SQLite is Stored**
```yaml
# In docker-compose.yml
environment:
  - DB_PATH=/data/terratag.db    # SQLite file inside container

volumes:
  - terratag-data:/data          # Named volume for database
```

The SQLite database is stored in a **Docker named volume** called `terratag-data`, not in your host directory.

## Database Management Commands

### **Reset Database (Remove All Data)**
```bash
# Stop services and remove all volumes (including database)
docker-compose down -v
```

This command:
- ✅ Stops all running containers
- ✅ Removes all named volumes (including `terratag-data`)
- ✅ Completely resets the database
- ✅ Preserves your source files and demo deployment

### **Restart with Fresh Database**
```bash
# Reset and restart
docker-compose down -v
docker-compose --profile ui up
```

### **Check Database Status**
```bash
# View volume information
docker volume ls | grep terratag

# Inspect volume details
docker volume inspect terratag_terratag-data
```

### **Stop Without Removing Database**
```bash
# Stop services but keep database
docker-compose down
```

## What Gets Reset vs Preserved

### **Reset with `docker-compose down -v`:**
- ❌ **Database**: All stored tag standards, operations, and results
- ❌ **UI Settings**: Any custom configurations in the web interface
- ❌ **Operation History**: All previous tagging and validation operations

### **Always Preserved:**
- ✅ **Demo Files**: `demo-deployment/` directory and all Terraform files
- ✅ **Generated Files**: `*.terratag.tf` and `*.tf.bak` files in demo-deployment
- ✅ **Reports**: Files in `reports/` directory
- ✅ **Standards Files**: YAML files in `standards/` directory
- ✅ **Docker Image**: Built terratag image

## Common Workflows

### **1. Start Fresh Demo**
```bash
# Complete reset and restart
docker-compose down -v
docker-compose --profile ui up
```

### **2. Reload Demo Standards**
```bash
# Reset database and load demo standards
docker-compose down -v
docker-compose --profile ui up -d
./scripts/load-demo-standards.sh
```

### **3. Quick Restart (Keep Database)**
```bash
# Restart without losing data
docker-compose restart
```

### **4. Clean Generated Files Only**
```bash
# Remove generated Terraform files but keep database
find demo-deployment -name "*.terratag.tf" -delete
find demo-deployment -name "*.tf.bak" -delete
rm -rf reports/*
```

## Benefits of Named Volumes

### **Advantages:**
- 🚀 **Easy Reset**: Single command `docker-compose down -v`
- 🔒 **Isolated**: Database isolated from host filesystem
- 📦 **Portable**: Volume can be backed up and restored
- 🧹 **Clean**: No database files cluttering your project directory

### **Volume Management:**
```bash
# List all volumes
docker volume ls

# Remove specific volume
docker volume rm terratag_terratag-data

# Remove all unused volumes
docker volume prune
```

## Backup and Restore (Optional)

### **Backup Database**
```bash
# Create backup
docker run --rm \
  -v terratag_terratag-data:/data:ro \
  -v $(pwd):/backup \
  alpine tar czf /backup/terratag-db-backup.tar.gz -C /data .
```

### **Restore Database**
```bash
# Restore from backup
docker run --rm \
  -v terratag_terratag-data:/data \
  -v $(pwd):/backup \
  alpine tar xzf /backup/terratag-db-backup.tar.gz -C /data
```

## Troubleshooting

### **Database Permission Issues**
```bash
# Reset volumes and restart
docker-compose down -v
docker-compose --profile ui up
```

### **Volume Not Found**
```bash
# Recreate volumes
docker-compose down -v
docker-compose up --force-recreate
```

### **Database Corruption**
```bash
# Complete reset
docker-compose down -v
docker volume prune -f
docker-compose --profile ui up
```

## Summary

| Command | Effect |
|---------|--------|
| `docker-compose down -v` | **Complete reset** - Removes database and all volumes |
| `docker-compose down` | **Soft stop** - Preserves database and volumes |
| `docker-compose restart` | **Quick restart** - Keeps everything intact |
| `docker volume prune` | **Clean unused volumes** - Removes orphaned volumes |

**🎯 For demo purposes, use `docker-compose down -v` for a clean slate every time!**