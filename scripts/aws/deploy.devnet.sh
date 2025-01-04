#!/bin/bash
# Copyright (C) 2024, Nuklai. All rights reserved.
# See the file LICENSE for licensing terms.

# deploy.devnet.sh

if [[ $(basename "$PWD") != "nuklaivm" ]]; then
  echo "Error: This script must be executed from the repository root (nuklaivm/)."
  exit 1
fi

KEY_NAME="../ssh-keys/nuklaivm-nodes-devnet.pem"
REGION="eu-west-1"
INSTANCE_NAME="nuklaivm-nodes-devnet"
INSTANCE_PROFILE_NAME="nuklaivm-nodes-devnet"
TARBALL="nuklaivm.tar.gz"
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
RETRIES=3
for ((i=1; i<=RETRIES; i++)); do
  INSTANCE_ID=$(aws ec2 describe-instances --region $REGION \
    --filters "Name=tag:Name,Values=$INSTANCE_NAME" "Name=instance-state-name,Values=running" \
    --query "Reservations[0].Instances[0].InstanceId" --output text)
  if [[ "$INSTANCE_ID" != "None" ]]; then
    echo "Existing instance found. Terminating it..."
    aws ec2 terminate-instances --instance-ids $INSTANCE_ID --region $REGION
    aws ec2 wait instance-terminated --instance-ids $INSTANCE_ID --region $REGION
    echo "Instance terminated."
    break
  elif [[ "$INSTANCE_ID" = "None" ]]; then
    echo "No instance running"
    break
  elif [[ $i -eq $RETRIES ]]; then
    echo "Failed to describe instances after $RETRIES attempts."
    exit 1
  fi
  sleep $((2**i))
done

# Check if an Elastic IP has already been allocated
if [ -f "$EIP_FILE" ]; then
  ALLOCATION_ID=$(cat $EIP_FILE)
  if [ -z "$ALLOCATION_ID" ]; then
    echo "Error: Allocation ID file exists but is empty. Allocating a new Elastic IP..."
    ALLOCATION_ID=$(aws ec2 allocate-address --region $REGION --query "AllocationId" --output text)
    echo $ALLOCATION_ID > $EIP_FILE
  fi
  ELASTIC_IP=$(aws ec2 describe-addresses --allocation-ids $ALLOCATION_ID --region $REGION --query "Addresses[0].PublicIp" --output text)
  echo "Reusing existing Elastic IP: $ELASTIC_IP"
else
  echo "Allocating a new Elastic IP..."
  ALLOCATION_ID=$(aws ec2 allocate-address --region $REGION --query "AllocationId" --output text)
  if [ -n "$ALLOCATION_ID" ]; then
    echo $ALLOCATION_ID > $EIP_FILE
    ELASTIC_IP=$(aws ec2 describe-addresses --allocation-ids $ALLOCATION_ID --region $REGION --query "Addresses[0].PublicIp" --output text)
    echo "Allocated new Elastic IP: $ELASTIC_IP"
  else
    echo "Error: Failed to allocate a new Elastic IP."
    exit 1
  fi
fi

# Launch a new EC2 instance
echo "Launching a new EC2 instance..."
INSTANCE_ID=$(aws ec2 run-instances \
  --region $REGION \
  --launch-template "LaunchTemplateName=$INSTANCE_PROFILE_NAME,Version=$Latest" \
  --query "Instances[0].InstanceId" --output text)


echo "Waiting for the instance to start..."
aws ec2 wait instance-status-ok --instance-ids $INSTANCE_ID --region $REGION
echo "Instance started. ID: $INSTANCE_ID"

# Associate the allocated Elastic IP with the new instance
echo "Associating Elastic IP $ELASTIC_IP with instance $INSTANCE_ID."
aws ec2 associate-address --instance-id $INSTANCE_ID --allocation-id $ALLOCATION_ID --region $REGION

echo "Creating tarball of the project, excluding web_wallet and ignored files..."
EXCLUDES=$(cat .gitignore .dockerignore 2>/dev/null | grep -v '^#' | sed '/^$/d' | sed 's/^/--exclude=/' | tr '\n' ' ')
EXCLUDES+=" --exclude='./web_wallet' --exclude=$TARBALL"
eval "tar -czf $TARBALL $EXCLUDES -C . ."

# Use $HOME instead of ~ to avoid any potential path resolution issues
if [ -f "$HOME/.ssh/known_hosts" ]; then
  ssh-keygen -f "$HOME/.ssh/known_hosts" -R $ELASTIC_IP
fi

echo "Transferring tarball and private key to the EC2 instance..."
scp -o "StrictHostKeyChecking=no" -o "UserKnownHostsFile=/dev/null" -i $KEY_NAME $TARBALL ec2-user@$ELASTIC_IP:/home/ec2-user/

echo "Connecting to the EC2 instance and deploying devnet..."
ssh -o "StrictHostKeyChecking=no" -o "UserKnownHostsFile=/dev/null" -i $KEY_NAME ec2-user@$ELASTIC_IP << EOF
  echo "Waiting for Docker installation to complete..."
  TIMEOUT=180  # Set a timeout in seconds to wait for Docker installation completion
  SECONDS=0
  while [[ ! -f /home/ec2-user/docker_installed && SECONDS -lt TIMEOUT ]]; do
    echo "Docker installation not complete, waiting..."
    sleep 5
  done

  if [[ ! -f /home/ec2-user/docker_installed ]]; then
    echo "Docker installation did not complete within the timeout period. Exiting."
    exit 1
  fi

  echo "Docker installation complete. Proceeding with deployment..."

  docker stop nuklai-devnet || true
  docker rm nuklai-devnet || true

  cd /home/ec2-user
  tar -xzf nuklaivm.tar.gz --strip-components=1

  # Build the Docker image
  docker build -t nuklai-devnet -f Dockerfile.devnet .

  # Run the Docker container with the provided arguments
  docker run -d --name nuklai-devnet -p 9650:9650 \
    --restart unless-stopped \
    -e INITIAL_OWNER_ADDRESS="$INITIAL_OWNER_ADDRESS" \
    -e EMISSION_ADDRESS="$EMISSION_ADDRESS" \
    -e EXTERNAL_SUBSCRIBER_SERVER_ADDRESS="$EXTERNAL_SUBSCRIBER_SERVER_ADDRESS" \
    nuklai-devnet

  echo "Checking if the blockchain is fully started..."
  TIMEOUT=300  # Timeout for blockchain readiness check
  SECONDS=0
  SUCCESS=false
  until [[ "\$SUCCESS" == true || SECONDS -ge TIMEOUT ]]; do
    RESPONSE=\$(curl -s -X POST --data '{
        "jsonrpc":"2.0",
        "id"     :1,
        "method" :"hypersdk.network",
        "params" : {}
      }' -H 'content-type:application/json;' 127.0.0.1:9650/ext/bc/nuklaivm/coreapi)

    # Check if the response contains a successful result
    if echo "\$RESPONSE" | grep -q '"result"'; then
      SUCCESS=true
      echo "Blockchain is fully started."
    else
      echo "Blockchain not yet started, retrying..."
      sleep 5
    fi
  done

  if [[ "\$SUCCESS" == true ]]; then
    echo "Devnet is ready and running on the instance."
  else
    echo "Blockchain did not start within the timeout period. Exiting."
    exit 1
  fi
EOF

echo "Deployment completed. Access the devnet at: http://$ELASTIC_IP:9650"