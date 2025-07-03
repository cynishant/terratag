#!/bin/bash
# User Data Script for ${project_name} ${environment} instances

# Update system
yum update -y

# Install required packages
yum install -y \
    aws-cli \
    htop \
    nginx \
    docker

# Install CloudWatch agent if monitoring is enabled
%{ if enable_monitoring ~}
wget https://s3.amazonaws.com/amazoncloudwatch-agent/amazon_linux/amd64/latest/amazon-cloudwatch-agent.rpm
rpm -U ./amazon-cloudwatch-agent.rpm

# Configure CloudWatch agent
cat > /opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json << 'EOF'
{
    "agent": {
        "metrics_collection_interval": 60,
        "run_as_user": "cwagent"
    },
    "logs": {
        "logs_collected": {
            "files": {
                "collect_list": [
                    {
                        "file_path": "/var/log/messages",
                        "log_group_name": "/aws/ec2/${project_name}-${environment}",
                        "log_stream_name": "{instance_id}/messages"
                    },
                    {
                        "file_path": "/var/log/nginx/access.log",
                        "log_group_name": "/aws/ec2/${project_name}-${environment}",
                        "log_stream_name": "{instance_id}/nginx-access"
                    },
                    {
                        "file_path": "/var/log/nginx/error.log",
                        "log_group_name": "/aws/ec2/${project_name}-${environment}",
                        "log_stream_name": "{instance_id}/nginx-error"
                    }
                ]
            }
        }
    },
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
/opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl \
    -a fetch-config \
    -m ec2 \
    -c file:/opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json \
    -s
%{ endif ~}

# Configure nginx
systemctl start nginx
systemctl enable nginx

# Create a simple index page
cat > /var/www/html/index.html << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>${project_name} - ${environment}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { color: #333; border-bottom: 1px solid #ccc; padding-bottom: 10px; }
        .info { background: #f5f5f5; padding: 20px; margin: 20px 0; border-radius: 5px; }
        .status { color: green; font-weight: bold; }
    </style>
</head>
<body>
    <div class="header">
        <h1>${project_name} Application</h1>
        <h2>Environment: ${environment}</h2>
    </div>
    
    <div class="info">
        <h3>Server Information</h3>
        <p><strong>Instance ID:</strong> <span id="instance-id">Loading...</span></p>
        <p><strong>Availability Zone:</strong> <span id="az">Loading...</span></p>
        <p><strong>Local IP:</strong> <span id="local-ip">Loading...</span></p>
        <p><strong>Status:</strong> <span class="status">Running</span></p>
    </div>
    
    <div class="info">
        <h3>Application Status</h3>
        <p>This is a test application for Terratag validation.</p>
        <p>Monitoring: ${enable_monitoring ? "Enabled" : "Disabled"}</p>
    </div>

    <script>
        // Fetch instance metadata
        fetch('/latest/meta-data/instance-id')
            .then(response => response.text())
            .then(data => document.getElementById('instance-id').textContent = data)
            .catch(err => document.getElementById('instance-id').textContent = 'Unknown');
            
        fetch('/latest/meta-data/placement/availability-zone')
            .then(response => response.text())
            .then(data => document.getElementById('az').textContent = data)
            .catch(err => document.getElementById('az').textContent = 'Unknown');
            
        fetch('/latest/meta-data/local-ipv4')
            .then(response => response.text())
            .then(data => document.getElementById('local-ip').textContent = data)
            .catch(err => document.getElementById('local-ip').textContent = 'Unknown');
    </script>
</body>
</html>
EOF

# Start Docker service
systemctl start docker
systemctl enable docker

# Add ec2-user to docker group
usermod -a -G docker ec2-user

# Create application directory
mkdir -p /opt/app
chown ec2-user:ec2-user /opt/app

# Install application dependencies (example)
# This would be replaced with actual application deployment

# Signal completion
/opt/aws/bin/cfn-signal -e $? --stack ${AWS::StackName} --resource AutoScalingGroup --region ${AWS::Region} || true

echo "User data script completed successfully" > /var/log/user-data.log