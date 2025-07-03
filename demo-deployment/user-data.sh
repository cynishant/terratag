#!/bin/bash
yum update -y
yum install -y httpd php php-mysql

# Start and enable Apache
systemctl start httpd
systemctl enable httpd

# Create a simple PHP application
cat > /var/www/html/index.php << 'EOF'
<?php
$db_host = "${db_endpoint}";
$s3_bucket = "${s3_bucket}";
?>
<!DOCTYPE html>
<html>
<head>
    <title>Sample Web Application</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 800px; margin: 0 auto; }
        .info { background: #f0f0f0; padding: 20px; margin: 20px 0; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Sample Web Application</h1>
        <div class="info">
            <h2>Server Information</h2>
            <p><strong>Server:</strong> <?php echo gethostname(); ?></p>
            <p><strong>PHP Version:</strong> <?php echo phpversion(); ?></p>
            <p><strong>Current Time:</strong> <?php echo date('Y-m-d H:i:s'); ?></p>
        </div>
        
        <div class="info">
            <h2>Configuration</h2>
            <p><strong>Database Endpoint:</strong> <?php echo $db_host; ?></p>
            <p><strong>S3 Bucket:</strong> <?php echo $s3_bucket; ?></p>
        </div>
        
        <div class="info">
            <h2>Health Check</h2>
            <p style="color: green;"><strong>Status:</strong> OK</p>
        </div>
    </div>
</body>
</html>
EOF

# Set proper permissions
chown apache:apache /var/www/html/index.php
chmod 644 /var/www/html/index.php

# Install CloudWatch agent
wget https://s3.amazonaws.com/amazoncloudwatch-agent/amazon_linux/amd64/latest/amazon-cloudwatch-agent.rpm
rpm -U ./amazon-cloudwatch-agent.rpm

# Configure CloudWatch agent
cat > /opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json << 'EOF'
{
    "metrics": {
        "namespace": "CWAgent",
        "metrics_collected": {
            "cpu": {
                "measurement": [
                    "cpu_usage_idle",
                    "cpu_usage_iowait",
                    "cpu_usage_user",
                    "cpu_usage_system"
                ],
                "metrics_collection_interval": 60
            },
            "disk": {
                "measurement": [
                    "used_percent"
                ],
                "metrics_collection_interval": 60,
                "resources": [
                    "*"
                ]
            },
            "mem": {
                "measurement": [
                    "mem_used_percent"
                ],
                "metrics_collection_interval": 60
            }
        }
    }
}
EOF

# Start CloudWatch agent
/opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl -a fetch-config -m ec2 -c file:/opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json -s