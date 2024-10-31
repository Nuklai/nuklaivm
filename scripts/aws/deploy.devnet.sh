#!/bin/bash
# Copyright (C) 2024, Nuklai. All rights reserved.
# See the file LICENSE for licensing terms.

if [[ $(basename "$PWD") != "nuklaivm" ]]; then
  echo "Error: This script must be executed from the repository root (nuklaivm/)."
  exit 1
fi

KEY_NAME="./scripts/aws/nuklaivm-aws-key.pem"
INSTANCE_NAME="nuklai-devnet-instance"
REGION="eu-west-1"
INSTANCE_TYPE="t2.medium"
SECURITY_GROUP="sg-07b07fac5e31bc731"
USER_DATA_FILE="./scripts/aws/install_docker.sh"
TARBALL="nuklaivm.tar.gz"
AMI_ID="ami-008d05461f83df5b1"

# Accept arguments or use default values
INITIAL_OWNER_ADDRESS=${1:-"002b5d019495996310f81c6a26a4dd9eeb9a3f3be1bac0a9294436713aecc84496"}
EMISSION_ADDRESS=${2:-"00f3b89e583e3944dee8d45ca40ce30829eff47481bc45669d401c2f9cc2bc110d"}

echo "Using AMI ID: $AMI_ID"

# Check if an instance is already running
INSTANCE_ID=$(aws ec2 describe-instances --region $REGION \
  --filters "Name=tag:Name,Values=$INSTANCE_NAME" "Name=instance-state-name,Values=running" \
  --query "Reservations[0].Instances[0].InstanceId" --output text)

if [ "$INSTANCE_ID" != "None" ]; then
  echo "Existing instance found. Terminating it..."
  aws ec2 terminate-instances --instance-ids $INSTANCE_ID --region $REGION
  aws ec2 wait instance-terminated --instance-ids $INSTANCE_ID --region $REGION
  echo "Instance terminated."
fi

# Launch a new EC2 instance
echo "Launching a new EC2 instance..."
INSTANCE_ID=$(aws ec2 run-instances --region $REGION \
  --image-id $AMI_ID --count 1 --instance-type $INSTANCE_TYPE \
  --key-name nuklaivm-aws-key --security-group-ids $SECURITY_GROUP \
  --associate-public-ip-address \
  --block-device-mappings 'DeviceName=/dev/xvda,Ebs={VolumeSize=1000,VolumeType=gp3,DeleteOnTermination=true}' \
  --tag-specifications "ResourceType=instance,Tags=[{Key=Name,Value=$INSTANCE_NAME}]" \
  --user-data file://$USER_DATA_FILE \
  --query "Instances[0].InstanceId" --output text)

echo "Waiting for the instance to start..."
aws ec2 wait instance-running --instance-ids $INSTANCE_ID --region $REGION
echo "Instance started. ID: $INSTANCE_ID"

PUBLIC_IP=$(aws ec2 describe-instances --instance-ids $INSTANCE_ID --region $REGION \
  --query "Reservations[0].Instances[0].PublicIpAddress" --output text)

echo "EC2 instance public IP: $PUBLIC_IP"
sleep 60

# Create a tarball of the project root, excluding web_wallet and ignored files
echo "Creating tarball of the project, excluding web_wallet and ignored files..."
EXCLUDES=$(cat .gitignore .dockerignore 2>/dev/null | grep -v '^#' | sed '/^$/d' | sed 's/^/--exclude=/' | tr '\n' ' ')
EXCLUDES+=" --exclude='./web_wallet' --exclude=$TARBALL"
eval "tar -czf $TARBALL $EXCLUDES -C . ."

echo "Transferring tarball and private key to the EC2 instance..."
scp -o "StrictHostKeyChecking=no" -i $KEY_NAME $TARBALL ec2-user@$PUBLIC_IP:/home/ec2-user/

echo "Connecting to the EC2 instance and deploying devnet..."
ssh -o "StrictHostKeyChecking=no" -i $KEY_NAME ec2-user@$PUBLIC_IP << EOF
  docker stop nuklai-devnet || true
  docker rm nuklai-devnet || true

  cd /home/ec2-user
  tar -xzf nuklaivm.tar.gz --strip-components=1

  # Build the Docker image
  docker build -t nuklai-devnet -f Dockerfile.devnet .

  # Run the Docker container with the passed arguments as environment variables
  docker run -d --name nuklai-devnet -p 9650:9650 \
    -e INITIAL_OWNER_ADDRESS="$INITIAL_OWNER_ADDRESS" \
    -e EMISSION_ADDRESS="$EMISSION_ADDRESS" \
    nuklai-devnet

  echo "Devnet is running on the instance."
EOF

echo "Deployment completed. Access the devnet at: http://$PUBLIC_IP:9650"
