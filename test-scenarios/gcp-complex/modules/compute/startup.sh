#!/bin/bash

# Startup script for ${project_name} ${environment} instances
# This script installs and configures basic web server

# Update system
apt-get update -y

# Install nginx and monitoring agent
apt-get install -y nginx curl wget

# Install Cloud Ops Agent for monitoring
curl -sSO https://dl.google.com/cloudagents/add-google-cloud-ops-agent-repo.sh
bash add-google-cloud-ops-agent-repo.sh --also-install

# Configure nginx
cat > /etc/nginx/sites-available/default << 'EOF'
server {
    listen 80 default_server;
    listen [::]:80 default_server;

    root /var/www/html;
    index index.html index.htm index.nginx-debian.html;

    server_name _;

    location / {
        try_files $uri $uri/ =404;
    }

    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }
}
EOF

# Create basic index page
cat > /var/www/html/index.html << EOF
<!DOCTYPE html>
<html>
<head>
    <title>${project_name} ${environment}</title>
</head>
<body>
    <h1>${project_name} - ${environment}</h1>
    <p>Instance: $(hostname)</p>
    <p>Environment: ${environment}</p>
    <p>Timestamp: $(date)</p>
</body>
</html>
EOF

# Start and enable nginx
systemctl start nginx
systemctl enable nginx

# Create application user
useradd -m -s /bin/bash app

# Set up application directory
mkdir -p /opt/app
chown app:app /opt/app

# Log startup completion
echo "Startup script completed for ${project_name} ${environment}" >> /var/log/startup.log