#!/bin/bash
# Copyright (C) 2024, Nuklai. All rights reserved.
# See the file LICENSE for licensing terms.

# install_docker.sh

set -e
LOGFILE="/var/log/install_docker.log"
exec > >(tee -a "$LOGFILE") 2>&1
date
echo "Starting Docker installation..."

echo "Updating packages..."
yum update -y

echo "Installing Docker and Git..."
yum install -y docker git

echo "Enabling Docker service..."
systemctl enable docker

echo "Starting Docker service..."
systemctl start docker

echo "Adding ec2-user to Docker group..."
usermod -aG docker ec2-user

echo "Setting Docker socket permissions..."
chmod 666 /var/run/docker.sock

echo "Waiting for Docker socket..."
until [ -S /var/run/docker.sock ]; do
  echo "Docker socket not yet available, retrying..."
  sleep 5
done

# Create a completion flag
touch /home/ec2-user/docker_installed

echo "Docker installation completed at $(date)"
