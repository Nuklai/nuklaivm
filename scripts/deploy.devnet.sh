#!/usr/bin/env bash
# Copyright (C) 2024, AllianceBlock. All rights reserved.
# See the file LICENSE for licensing terms.

set -e

# Set the CGO flags to use the portable version of BLST
#
# We use "export" here instead of just setting a bash variable because we need
# to pass this flag to all child processes spawned by the shell.
export CGO_CFLAGS="-O -D__BLST_PORTABLE__" CGO_ENABLED=1

# Set console colors
RED='\033[1;31m'
YELLOW='\033[1;33m'
CYAN='\033[1;36m'
NC='\033[0m'

# Ensure we return back to the original directory
pw=$(pwd)
function cleanup() {
  cd "$pw"
}
trap cleanup EXIT

# Ensure that the script is being run from the repository root
if ! [[ "$0" =~ scripts/deploy.devnet.sh ]]; then
  echo -e "${RED}must be run from repository root${NC}"
  exit 1
fi

# Ensure required software is installed and aws credentials are set
if ! command -v go >/dev/null 2>&1 ; then
    echo -e "${RED}golang is not installed. exiting...${NC}"
    exit 1
fi
if ! aws sts get-caller-identity >/dev/null 2>&1 ; then
    echo -e "${RED}aws credentials not set. exiting...${NC}"
    exit 1
fi

# Set AvalancheGo Build (should have canPop disabled)
AVALANCHEGO_VERSION=v1.10.18

# Create temporary directory for the deployment
TMPDIR=/tmp/nuklaivm-deploy
rm -rf $TMPDIR && mkdir -p $TMPDIR
echo -e "${YELLOW}set working directory:${NC} $TMPDIR"

# Install avalanche-cli
LOCAL_CLI_COMMIT=v1.5.2
REMOTE_CLI_COMMIT=v1.5.2
cd $TMPDIR
git clone https://github.com/ava-labs/avalanche-cli
cd avalanche-cli
git checkout $LOCAL_CLI_COMMIT
./scripts/build.sh
mv ./bin/avalanche "${TMPDIR}/avalanche"
cd $pw

# Install nuklai-cli
NUKLAI_VM_COMMIT=main
echo -e "${YELLOW}building nuklai-cli${NC}"
echo "set working directory: $TMPDIR"
cd $TMPDIR
echo "cloning nuklaivm commit: $NUKLAI_VM_COMMIT"
git clone https://github.com/nuklai/nuklaivm
cd nuklaivm
echo "checking out nuklaivm commit: $NUKLAI_VM_COMMIT"
git checkout $NUKLAI_VM_COMMIT
echo "building nuklaivm"
VMID=$(git rev-parse --short HEAD) # ensure we use a fresh vm
VM_COMMIT=$(git rev-parse HEAD)
./scripts/build.sh
echo "moving nuklai-cli to $TMPDIR"
mv ./build/nuklai-cli "${TMPDIR}/nuklai-cli"
cd $pw

# Setup devnet
CLUSTER="nuklai-$(date +%s)"
function cleanup {
  echo -e "\n\n${RED}run this command to destroy the devnet:${NC} ${TMPDIR}/avalanche node destroy ${CLUSTER}\n"
}
trap cleanup EXIT

# List of supported instances in each AWS region: https://docs.aws.amazon.com/ec2/latest/instancetypes/ec2-instance-regions.html
#
# It is not recommended to use an instance with burstable network performance.
echo -e "${YELLOW}creating devnet${NC}"
$TMPDIR/avalanche node devnet wiz ${CLUSTER} ${VMID} --force-subnet-create=true --authorize-access=true --aws --node-type t4g.medium --num-apis 1 --num-validators 5 --region eu-west-1 --use-static-ip=true --enable-monitoring=true --default-validator-params --custom-avalanchego-version $AVALANCHEGO_VERSION --custom-vm-repo-url="https://www.github.com/nuklai/nuklaivm" --custom-vm-branch $VM_COMMIT --custom-vm-build-script="scripts/build.sh" --custom-subnet=true --subnet-genesis="${TMPDIR}/nuklaivm/docs/deployment/nuklaivm_genesis.json" --subnet-config="${TMPDIR}/nuklaivm/docs/deployment/nuklaivm_genesis.json" --chain-config="${TMPDIR}/nuklaivm/docs/deployment/nuklaivm_chain.json" --node-config="${TMPDIR}/nuklaivm/docs/deployment/nuklaivm_avago.json" --config="${TMPDIR}/nuklaivm/docs/deployment/nuklaivm_avago.json" --remote-cli-version $REMOTE_CLI_COMMIT --add-grafana-dashboard="${TMPDIR}/nuklaivm/grafana.json" --log-level DEBUG
EPOCH_WAIT_START=$(date +%s)

# Import the cluster into nuklai-cli for local interaction
$TMPDIR/nuklai-cli chain import-ops ~/.avalanche-cli/nodes/inventories/$CLUSTER/clusterInfo.yaml

# We use a shorter EPOCH_DURATION and VALIDITY_WINDOW to speed up devnet
# startup. In a production environment, these should be set to longer values.
#
EPOCH_DURATION=60000
VALIDITY_WINDOW=59000

# Wait for epoch initialization
SLEEP_DUR=$(($EPOCH_DURATION / 1000 * 3))
EPOCH_SEC=$(($EPOCH_DURATION / 1000))
VALIDITY_WINDOW_SEC=$(($VALIDITY_WINDOW / 1000))
echo -e "\n${YELLOW}waiting for epoch initialization:${NC} $SLEEP_DUR seconds"
echo "We use a shorter EPOCH_DURATION ($EPOCH_SEC seconds) and VALIDITY_WINDOW ($VALIDITY_WINDOW_SEC seconds) to speed up devnet startup. In a production environment, these should be set to larger values."
sleep $SLEEP_DUR

# Log dashboard information
echo -e "\n\n${CYAN}dashboards:${NC} (username: admin, password: admin)"
echo "* nuklaivm (metrics): http://$(yq e '.MONITOR.IP' ~/.avalanche-cli/nodes/inventories/$CLUSTER/clusterInfo.yaml):3000/d/vryx-poc"
echo "* nuklaivm (logs): http://$(yq e '.MONITOR.IP' ~/.avalanche-cli/nodes/inventories/$CLUSTER/clusterInfo.yaml):3000/d/avalanche-loki-logs/avalanche-logs?var-app=subnet"
