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
EIP_FILE="./scripts/aws/elastic_ip_allocation.txt"

# Default values for addresses
INITIAL_OWNER_ADDRESS="00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9"
EMISSION_ADDRESS="00f3b89e583e3944dee8d45ca40ce30829eff47481bc45669d401c2f9cc2bc110d"
EXTERNAL_SUBSCRIBER_SERVER_ADDRESS=""

# Parse command-line arguments
while [[ "$#" -gt 0 ]]; do
  case $1 in
    --initial-owner-address) INITIAL_OWNER_ADDRESS="$2"; shift ;;
    --emission-address) EMISSION_ADDRESS="$2"; shift ;;
    --external-subscriber-server-address) EXTERNAL_SUBSCRIBER_SERVER_ADDRESS="$2"; shift ;;
  esac
  shift
done

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

# Check if an Elastic IP has already been allocated
if [ -f "$EIP_FILE" ]; then
  ALLOCATION_ID=$(cat $EIP_FILE)
  ELASTIC_IP=$(aws ec2 describe-addresses --allocation-ids $ALLOCATION_ID --region $REGION --query "Addresses[0].PublicIp" --output text)
  echo "Reusing existing Elastic IP: $ELASTIC_IP"
else
  echo "Allocating a new Elastic IP..."
  ALLOCATION_ID=$(aws ec2 allocate-address --region $REGION --query "AllocationId" --output text)
  ELASTIC_IP=$(aws ec2 describe-addresses --allocation-ids $ALLOCATION_ID --region $REGION --query "Addresses[0].PublicIp" --output text)
  echo $ALLOCATION_ID > $EIP_FILE
  echo "Allocated new Elastic IP: $ELASTIC_IP"
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

# Associate the allocated Elastic IP with the new instance
echo "Associating Elastic IP $ELASTIC_IP with instance $INSTANCE_ID."
aws ec2 associate-address --instance-id $INSTANCE_ID --allocation-id $ALLOCATION_ID --region $REGION

echo "Creating tarball of the project, excluding web_wallet and ignored files..."
EXCLUDES=$(cat .gitignore .dockerignore 2>/dev/null | grep -v '^#' | sed '/^$/d' | sed 's/^/--exclude=/' | tr '\n' ' ')
EXCLUDES+=" --exclude='./web_wallet' --exclude=$TARBALL"
eval "tar -czf $TARBALL $EXCLUDES -C . ."

echo "Transferring tarball and private key to the EC2 instance..."
scp -o "StrictHostKeyChecking=no" -i $KEY_NAME $TARBALL ec2-user@$ELASTIC_IP:/home/ec2-user/

echo "Connecting to the EC2 instance and deploying devnet..."
ssh -o "StrictHostKeyChecking=no" -i $KEY_NAME ec2-user@$ELASTIC_IP << EOF
  docker stop nuklai-devnet || true
  docker rm nuklai-devnet || true

  cd /home/ec2-user
  tar -xzf nuklaivm.tar.gz --strip-components=1

  # Build the Docker image
  docker build -t nuklai-devnet -f Dockerfile.devnet .

  # Run the Docker container with the provided arguments
  docker run -d --name nuklai-devnet -p 9650:9650 \
    -e INITIAL_OWNER_ADDRESS="$INITIAL_OWNER_ADDRESS" \
    -e EMISSION_ADDRESS="$EMISSION_ADDRESS" \
    -e EXTERNAL_SUBSCRIBER_SERVER_ADDRESS="$EXTERNAL_SUBSCRIBER_SERVER_ADDRESS" \
    nuklai-devnet

  echo "Devnet is running on the instance."
EOF

echo "Deployment completed. Access the devnet at: http://$ELASTIC_IP:9650"
