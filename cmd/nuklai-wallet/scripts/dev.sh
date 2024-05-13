#!/usr/bin/env bash
# Copyright (C) 2024, AllianceBlock. All rights reserved.
# See the file LICENSE for licensing terms.

set -o errexit
set -o nounset
set -o pipefail

# Set the CGO flags to use the portable version of BLST
export CGO_CFLAGS="-O -D__BLST_PORTABLE__"  CGO_ENABLED=1

# Create a temporary directory to store the updated config.json files
nuklai_wallet_dir=/tmp/nuklaivm/nuklai-wallet
mkdir -p "$nuklai_wallet_dir"

# Install wails
go install -v github.com/wailsapp/wails/v2/cmd/wails@v2.8.0

# Go up two directories
pushd ../..

# Build nuklai-feed and nuklai-faucet
./scripts/build.sh

# Run nuklai-cli and capture the output
output=$(./build/nuklai-cli chain import-anr)

# Extract the first URI using grep and awk
nuklai_rpc_uri=$(echo "$output" | grep -o 'http://127.0.0.1:[0-9]*/ext/bc/[a-zA-Z0-9]*' | head -n 1)

# Check if nuklai_rpc_uri is empty and exit if it is
if [ -z "$nuklai_rpc_uri" ]; then
    echo "Error: No URI found in the output of nuklai-cli chain import-anr."
    exit 1
fi

# Use sed to update nuklaiRPC field in config.json files and write changes to new files
sed "s|\"nuklaiRPC\": \".*\"|\"nuklaiRPC\": \"$nuklai_rpc_uri\"|" cmd/nuklai-feed/config.json > ${nuklai_wallet_dir}/nuklai-feed-config.json
sed "s|\"nuklaiRPC\": \".*\"|\"nuklaiRPC\": \"$nuklai_rpc_uri\"|" cmd/nuklai-faucet/config.json > ${nuklai_wallet_dir}/nuklai-faucet-config.json

# Create a directory for nuklai-wallet if it does not exist
mkdir -p cmd/nuklai-wallet/.nuklai-wallet
sed "s|\"nuklaiRPC\": \".*\"|\"nuklaiRPC\": \"$nuklai_rpc_uri\"|" cmd/nuklai-wallet/config.json > cmd/nuklai-wallet/.nuklai-wallet/config.json

# Run nuklai-feed in the background and capture its PID
nuklai_feed_log=cmd/nuklai-wallet/.nuklai-wallet/nuklai-feed.log
rm -f ${nuklai_feed_log}
nohup ./build/nuklai-feed ${nuklai_wallet_dir}/nuklai-feed-config.json > ${nuklai_feed_log} 2>&1 &
nuklai_feed_pid=$!
echo "Nuklai Feed started with PID: $nuklai_feed_pid"

# Run nuklai-faucet in the background and capture its PID
nuklai_faucet_log=cmd/nuklai-wallet/.nuklai-wallet/nuklai-faucet.log
rm -f ${nuklai_faucet_log}
nohup ./build/nuklai-faucet ${nuklai_wallet_dir}/nuklai-faucet-config.json > ${nuklai_faucet_log} 2>&1 &
nuklai_faucet_pid=$!
echo "Nuklai Faucet started with PID: $nuklai_faucet_pid"

# Go back to the original directory
popd

# Function to kill background processes
cleanup() {
    echo "Killing nuklai-feed and nuklai-faucet..."
    kill $nuklai_feed_pid
    kill $nuklai_faucet_pid
    rm -rf ${nuklai_wallet_dir}
}

# Trap EXIT signal to ensure cleanup runs when the script exits
trap cleanup EXIT

# Start development environment
wails dev
