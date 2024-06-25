#!/usr/bin/env bash

# Check if Docker is installed
if ! command -v docker &>/dev/null; then
	echo "Docker is not installed. Please install Docker and try again."
	exit 1
fi

# Check if Docker Compose is installed
if ! docker compose version &>/dev/null; then
	echo "Docker Compose is not installed or not configured correctly. Please install Docker Compose and try again."
	exit 1
fi

# Provide docker-compose systemctl unit file
cat <<EOF | sudo tee /etc/systemd/system/avalanche-cli-docker.service
[Unit]
Description=Avalanche CLI Docker Compose Service
Requires=docker.service
After=docker.service

[Service]
User=ubuntu
Group=ubuntu
Restart=on-failure
ExecStart=/usr/bin/docker compose -f /home/ubuntu/.avalanche-cli/services/docker-compose.yml up 
ExecStop=/usr/bin/docker compose -f /home/ubuntu/.avalanche-cli/services/docker-compose.yml down

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd manager configuration
sudo systemctl daemon-reload

# Enable the new service
sudo systemctl enable avalanche-cli-docker.service

echo "Service created and enabled successfully."
