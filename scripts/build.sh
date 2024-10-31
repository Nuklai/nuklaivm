#!/usr/bin/env bash
# Copyright (C) 2024, Nuklai. All rights reserved.
# See the file LICENSE for licensing terms.

set -o errexit
set -o nounset
set -o pipefail

# Get the directory of the script, even if sourced from another directory
SCRIPT_DIR=$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd)

source "$SCRIPT_DIR"/hypersdk/common/build.sh
source "$SCRIPT_DIR"/hypersdk/constants.sh

# Construct the correct path to nuklaivm directory
NUKLAIVM_PATH=$(
  cd "$(dirname "${BASH_SOURCE[0]}")"
  cd .. && pwd
)

# Check if vmpath argument is provided
VMPATH=${1:-""}

build_project "$NUKLAIVM_PATH" "nuklaivm" "qeX5BUxbiwUhSePncmz1C7RdH6njYYv6dNZhJrdeXRKMnTpKt"

# If vmpath is provided, copy the binary to the specified vmpath
if [[ -n "$VMPATH" ]]; then
  echo "Copying binary to $VMPATH"
  cp "$NUKLAIVM_PATH/build/qeX5BUxbiwUhSePncmz1C7RdH6njYYv6dNZhJrdeXRKMnTpKt" "$VMPATH"
fi
