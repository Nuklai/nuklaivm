#!/usr/bin/env bash
# Copyright (C) 2024, Nuklai. All rights reserved.
# See the file LICENSE for licensing terms.

set -e

# Variables
IMAGE_NAME="nuklai-devnet"
CONTAINER_NAME="nuklai-devnet"

# Default values
INITIAL_OWNER_ADDRESS="00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9"
EMISSION_ADDRESS="00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9"
EXTERNAL_SUBSCRIBER_SERVER_ADDRESS=""

# Function to stop and remove any existing container
cleanup_docker() {
  echo "Stopping and removing any existing Docker container..."
  docker stop "$CONTAINER_NAME" || true
  docker rm "$CONTAINER_NAME" || true
}

# Function to build the Docker image
build_docker_image() {
  echo "Building the Docker image..."
  docker build -t "$IMAGE_NAME" -f Dockerfile.devnet .
}

# Function to start the Docker container with environment variables
start_container() {
  echo "Starting the Docker container with provided arguments..."

  docker run -d --name "$CONTAINER_NAME" \
    -p 9650:9650 \
    -e INITIAL_OWNER_ADDRESS="$INITIAL_OWNER_ADDRESS" \
    -e EMISSION_ADDRESS="$EMISSION_ADDRESS" \
    -e EXTERNAL_SUBSCRIBER_SERVER_ADDRESS="$EXTERNAL_SUBSCRIBER_SERVER_ADDRESS" \
    "$IMAGE_NAME"

  echo "Docker container started. Use './scripts/run_docker.sh logs' to view logs."
}

# Function to stop the Docker container
stop_container() {
  echo "Stopping the Docker container..."
  docker stop "$CONTAINER_NAME" || echo "No running container found."
}

# Function to view logs of the container
view_logs() {
  echo "Displaying logs for the container..."
  docker logs -f "$CONTAINER_NAME"
}

# Main logic to parse commands
case "$1" in
  start)
    shift  # Remove "start" from arguments
    # Parse command-line arguments
    while [[ "$#" -gt 0 ]]; do
      case $1 in
        --initial-owner-address) INITIAL_OWNER_ADDRESS="$2"; shift ;;
        --emission-address) EMISSION_ADDRESS="$2"; shift ;;
        --external-subscriber-server-address) EXTERNAL_SUBSCRIBER_SERVER_ADDRESS="$2"; shift ;;
      esac
      shift
    done
    cleanup_docker
    build_docker_image
    start_container "$@"
    ;;
  stop)
    stop_container
    ;;
  logs)
    view_logs
    ;;
  *)
    echo "Usage: $0 {start [--initial-owner-address VALUE] [--emission-address VALUE] [--external-subscriber-server-address VALUE] | stop | logs}"
    exit 1
    ;;
esac
