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

build_project "$NUKLAIVM_PATH" "nuklai" "qeX5BUxbiwUhSePncmz1C7RdH6njYYv6dNZhJrdeXRKMnTpKt"
