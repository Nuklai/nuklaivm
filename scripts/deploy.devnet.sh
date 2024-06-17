#!/usr/bin/env bash
# Copyright (C) 2024, Nuklai. All rights reserved.
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
AVALANCHEGO_VERSION=v1.11.6

# Create temporary directory for the deployment
TMPDIR=/tmp/nuklaivm-deploy
rm -rf $TMPDIR && mkdir -p $TMPDIR
echo -e "${YELLOW}set working directory:${NC} $TMPDIR"

# Install avalanche-cli
LOCAL_CLI_COMMIT=v1.6.0
REMOTE_CLI_COMMIT=v1.6.0
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

# Generate genesis file and configs
MIN_BLOCK_GAP=250
MIN_UNIT_PRICE="100,100,100,100,100"
WINDOW_TARGET_UNITS="40000000,450000,450000,450000,450000"
MAX_UINT64=18446744073709551615
MAX_BLOCK_UNITS="1800000,${MAX_UINT64},${MAX_UINT64},${MAX_UINT64},${MAX_UINT64}"

INITIAL_OWNER_ADDRESS=${INITIAL_OWNER_ADDRESS:-nuklai1qpg4ecapjymddcde8sfq06dshzpxltqnl47tvfz0hnkesjz7t0p35d5fnr3}
EMISSION_ADDRESS=${EMISSION_ADDRESS:-nuklai1qr4hhj8vfrnmzghgfnqjss0ns9tv7pjhhhggfm2zeagltnlmu4a6sgh6dqn}
# Sum of allocations must be less than uint64 max
cat <<EOF > "${TMPDIR}"/allocations.json
[
  {"address":"${INITIAL_OWNER_ADDRESS}", "balance":853000000000000000}
]
EOF
# maxSupply: 10 billion NAI
cat <<EOF > "${TMPDIR}"/emission-balancer.json
{
  "maxSupply":  10000000000000000000,
  "emissionAddress":"${EMISSION_ADDRESS}"
}
EOF

"${TMPDIR}"/nuklai-cli genesis generate "${TMPDIR}"/allocations.json "${TMPDIR}"/emission-balancer.json \
--min-unit-price "${MIN_UNIT_PRICE}" \
--window-target-units ${WINDOW_TARGET_UNITS} \
--max-block-units ${MAX_BLOCK_UNITS} \
--min-block-gap "${MIN_BLOCK_GAP}" \
--genesis-file "${TMPDIR}"/nuklaivm.genesis

# TODO: find a smarter way to split auth cores between exec and RPC
# TODO: we limit root generation cores because it can cause network handling to stop (exhausts all CPU for a few seconds)
cat <<EOF > "${TMPDIR}"/nuklaivm.config
{
  "chunkBuildFrequency": 250,
  "targetChunkBuildDuration": 250,
  "blockBuildFrequency": 250,
  "mempoolSize": 2147483648,
  "mempoolSponsorSize": 10000000,
  "authExecutionCores": 2,
  "precheckCores": 2,
  "actionExecutionCores": 2,
  "missingChunkFetchers": 48,
  "verifyAuth": true,
  "authRPCCores": 2,
  "authRPCBacklog": 10000000,
  "authGossipCores": 2,
  "authGossipBacklog": 10000000,
  "chunkStorageCores": 2,
  "chunkStorageBacklog": 10000000,
  "streamingBacklogSize": 10000000,
  "continuousProfilerDir":"/home/ubuntu/nuklaivm-profiles",
  "logLevel": "INFO",
  "mempoolExemptSponsors": [
    "nuklai1qpg4ecapjymddcde8sfq06dshzpxltqnl47tvfz0hnkesjz7t0p35d5fnr3",
    "nuklai1qr4hhj8vfrnmzghgfnqjss0ns9tv7pjhhhggfm2zeagltnlmu4a6sgh6dqn"
  ],
  "authVerificationCores": 2,
  "rootGenerationCores": 2,
  "transactionExecutionCores": 2,
  "storeTransactions": false,
  "stateSyncServerDelay": 0
}
EOF

cat <<EOF > "${TMPDIR}"/nuklaivm.subnet
{
  "proposerMinBlockDelay": 250,
  "proposerNumHistoricalBlocks": 1000000
}
EOF

cat <<EOF > "${TMPDIR}"/node.config
{
  "log-level":"INFO",
  "log-display-level":"INFO",
  "proposervm-use-current-height":true,
  "throttler-inbound-validator-alloc-size":"10737418240",
  "throttler-inbound-at-large-alloc-size":"10737418240",
  "throttler-inbound-node-max-processing-msgs":"1000000",
	"throttler-inbound-node-max-at-large-bytes":"10737418240",
  "throttler-inbound-bandwidth-refill-rate":"1073741824",
  "throttler-inbound-bandwidth-max-burst-size":"1073741824",
  "throttler-inbound-cpu-validator-alloc":"100000",
  "throttler-inbound-cpu-max-non-validator-usage":"100000",
  "throttler-inbound-cpu-max-non-validator-node-usage":"100000",
  "throttler-inbound-disk-validator-alloc":"10737418240000",
  "throttler-outbound-validator-alloc-size":"10737418240",
  "throttler-outbound-at-large-alloc-size":"10737418240",
  "throttler-outbound-node-max-at-large-bytes":"10737418240",
  "consensus-on-accept-gossip-validator-size":"10",
  "consensus-on-accept-gossip-peer-size":"10",
  "network-compression-type":"zstd",
  "consensus-app-concurrency":"128",
  "profile-continuous-enabled":true,
  "profile-continuous-freq":"1m",
  "http-host":"",
  "http-allowed-origins": "*",
  "http-allowed-hosts": "*"
}
EOF

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
$TMPDIR/avalanche node devnet wiz ${CLUSTER} ${VMID} --force-subnet-create=true --authorize-access=true --aws --node-type t4g.medium --num-apis 0 --num-validators 10 --region eu-west-1 --use-static-ip=false --enable-monitoring=false --default-validator-params=true --custom-avalanchego-version $AVALANCHEGO_VERSION --custom-vm-repo-url="https://www.github.com/nuklai/nuklaivm" --custom-vm-branch $VM_COMMIT --custom-vm-build-script="scripts/build.sh" --custom-subnet=true --subnet-genesis="${TMPDIR}/nuklaivm.genesis" --subnet-config="${TMPDIR}/nuklaivm.subnet" --chain-config="${TMPDIR}/nuklaivm.config" --node-config="${TMPDIR}/node.config" --config="${TMPDIR}/node.config" --remote-cli-version $REMOTE_CLI_COMMIT --add-grafana-dashboard="${TMPDIR}/nuklaivm/grafana.json" --log-level DEBUG

# Import the cluster into nuklai-cli for local interaction
$TMPDIR/nuklai-cli chain import-cli $HOME/.avalanche-cli/nodes/inventories/$CLUSTER/clusterInfo.yaml

# Extract Subnet ID, Chain ID, Validator IPs, and API IPs
SUBNET_ID=$(yq e '.SUBNET_ID' $HOME/.avalanche-cli/nodes/inventories/$CLUSTER/clusterInfo.yaml)
CHAIN_ID=$(yq e '.CHAIN_ID' $HOME/.avalanche-cli/nodes/inventories/$CLUSTER/clusterInfo.yaml)
VALIDATOR_IPS=($(yq e '.VALIDATOR[].IP' $HOME/.avalanche-cli/nodes/inventories/$CLUSTER/clusterInfo.yaml))
API_IPS=($(yq e '.API[].IP' $HOME/.avalanche-cli/nodes/inventories/$CLUSTER/clusterInfo.yaml))

# Print some info
echo -e "\n${CYAN}Cluster:${NC} $CLUSTER"
echo -e "\n${CYAN}VM ID:${NC} $VMID"
echo -e "\n${CYAN}VM Commit:${NC} $VM_COMMIT"
echo -e "\n${CYAN}Subnet ID:${NC} $SUBNET_ID"
echo -e "${CYAN}Chain ID:${NC} $CHAIN_ID"

# Print Validator and API IPs in required format
echo -e "${CYAN}RPC URLs:${NC}"
echo "RPC_URLS=("
for ip in "${VALIDATOR_IPS[@]}"; do
  echo "  $ip"
done
for ip in "${API_IPS[@]}"; do
  echo "  $ip"
done
echo ")"


# Start load test on dedicated machine
# sleep 30
# Zipf parameters expected to lead to ~1M active accounts per 60s
#echo -e "\n${YELLOW}starting load test...${NC}"
#$TMPDIR/avalanche node loadtest start "default" ${CLUSTER} ${VMID} --region eu-west-1 --aws --node-type t4g.medium --load-test-repo="https://github.com/nuklai/nuklaivm" --load-test-branch=$VM_COMMIT --load-test-build-cmd="cd /home/ubuntu/nuklai/nuklaivm; CGO_CFLAGS=\"-O -D__BLST_PORTABLE__\" go build -o ~/simulator ./cmd/nuklai-cli" --load-test-cmd="/home/ubuntu/simulator spam run ed25519 --accounts=10000000 --txs-per-second=100000 --min-capacity=15000 --step-size=1000 --s-zipf=1.0001 --v-zipf=2.7 --conns-per-host=10 --cluster-info=/home/ubuntu/clusterInfo.yaml --private-key=eef3a4b4f3ab5e277d2ea90952bd59300f849ad339bca5e499d1474ac7aa1e836de59563045c6b2792f12b2d0301b73db650c0291f739205afeea8e44000cf75"

# Log dashboard information
echo -e "\n\n${CYAN}dashboards:${NC} (username: admin, password: admin)"
echo "* nuklai devnet (metrics): http://$(yq e '.MONITOR.IP' $HOME/.avalanche-cli/nodes/inventories/$CLUSTER/clusterInfo.yaml):3000/d/nuklai-devnet/nuklai-devnet"
echo "* nuklaivm (logs): http://$(yq e '.MONITOR.IP' ~/.avalanche-cli/nodes/inventories/$CLUSTER/clusterInfo.yaml):3000/d/avalanche-loki-logs/avalanche-logs?var-app=subnet"
# echo "* load test (logs): http://$(yq e '.MONITOR.IP' ~/.avalanche-cli/nodes/inventories/$CLUSTER/clusterInfo.yaml):3000/d/avalanche-loki-logs/avalanche-logs?var-app=loadtest"


