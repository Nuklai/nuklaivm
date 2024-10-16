#!/usr/bin/env bash
# Copyright (C) 2024, Nuklai. All rights reserved.
# See the file LICENSE for licensing terms.

set -e

# Set the CGO flags to use the portable version of BLST
#
# We use "export" here instead of just setting a bash variable because we need
# to pass this flag to all child processes spawned by the shell.
export CGO_CFLAGS="-O -D__BLST_PORTABLE__" CGO_ENABLED=1

# to run E2E tests (terminates cluster afterwards)
# MODE=test ./scripts/run.sh
MODE=${MODE:-run}
if ! [[ "$0" =~ scripts/run.sh ]]; then
  echo "must be run from nuklaivm root"
  exit 255
fi

source ./scripts/hypersdk/common/utils.sh
source ./scripts/hypersdk/constants.sh

VERSION=v1.11.12-rc.2

############################
# build avalanchego
# https://github.com/ava-labs/avalanchego/releases
HYPERSDK_DIR=$HOME/.hypersdk

echo "working directory: $HYPERSDK_DIR"

AVALANCHEGO_PATH=${HYPERSDK_DIR}/avalanchego-${VERSION}/avalanchego
AVALANCHEGO_PLUGIN_DIR=${HYPERSDK_DIR}/avalanchego-${VERSION}/plugins

if [ ! -f "$AVALANCHEGO_PATH" ]; then
  echo "building avalanchego"
  CWD=$(pwd)

  # Clear old folders
  rm -rf "${HYPERSDK_DIR}"/avalanchego-"${VERSION}"
  mkdir -p "${HYPERSDK_DIR}"/avalanchego-"${VERSION}"
  rm -rf "${HYPERSDK_DIR}"/avalanchego-src
  mkdir -p "${HYPERSDK_DIR}"/avalanchego-src

  # Download src
  cd "${HYPERSDK_DIR}"/avalanchego-src
  git clone https://github.com/ava-labs/avalanchego.git
  cd avalanchego
  git checkout "${VERSION}"

  # Build avalanchego
  ./scripts/build.sh
  mv build/avalanchego "${HYPERSDK_DIR}"/avalanchego-"${VERSION}"

  cd "${CWD}"

  # Clear src
  rm -rf "${HYPERSDK_DIR}"/avalanchego-src
else
  echo "using previously built avalanchego"
fi

############################

echo "building nuklaivm"

# delete previous (if exists)
rm -f "${HYPERSDK_DIR}"/avalanchego-"${VERSION}"/plugins/qeX5BUxbiwUhSePncmz1C7RdH6njYYv6dNZhJrdeXRKMnTpKt

# rebuild with latest code
go build \
-o "${HYPERSDK_DIR}"/avalanchego-"${VERSION}"/plugins/qeX5BUxbiwUhSePncmz1C7RdH6njYYv6dNZhJrdeXRKMnTpKt \
./cmd/nuklaivm

############################
echo "building e2e.test"

prepare_ginkgo

ACK_GINKGO_RC=true ginkgo build ./tests/e2e
./tests/e2e/e2e.test --help

additional_args=("$@")

if [[ ${MODE} == "run" ]]; then
  echo "applying ginkgo.focus=Ping and --reuse-network to setup local network"
  additional_args+=("--ginkgo.focus=Ping")
  additional_args+=("--reuse-network")
fi

echo "running e2e tests"
./tests/e2e/e2e.test \
--ginkgo.v \
--avalanchego-path="${AVALANCHEGO_PATH}" \
--plugin-dir="${AVALANCHEGO_PLUGIN_DIR}" \
"${additional_args[@]}"
