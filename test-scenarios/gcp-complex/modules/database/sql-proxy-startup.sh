#!/bin/bash

# SQL Proxy startup script for ${project_id}

# Update system
apt-get update -y

# Install Cloud SQL Proxy
curl -o cloud_sql_proxy https://dl.google.com/cloudsql/cloud_sql_proxy.linux.amd64
chmod +x cloud_sql_proxy
mv cloud_sql_proxy /usr/local/bin/

# Create service user
useradd -m -s /bin/bash sqlproxy

# Create systemd service
cat > /etc/systemd/system/cloud-sql-proxy.service << EOF
[Unit]
Description=Google Cloud SQL Proxy
After=network.target

[Service]
Type=simple
User=sqlproxy
ExecStart=/usr/local/bin/cloud_sql_proxy -instances=${connection_name}=tcp:0.0.0.0:3306
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Enable and start the service
systemctl daemon-reload
systemctl enable cloud-sql-proxy
systemctl start cloud-sql-proxy

# Log startup completion
echo "SQL Proxy startup completed for ${project_id}:${instance_name}" >> /var/log/startup.log