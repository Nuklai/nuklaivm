#!/usr/bin/env bash
# Copyright (C) 2024, Nuklai. All rights reserved.
# See the file LICENSE for licensing terms.

set -e

# Set the CGO flags to use the portable version of BLST
#
# We use "export" here instead of just setting a bash variable because we need
# to pass this flag to all child processes spawned by the shell.
export CGO_CFLAGS="-O -D__BLST_PORTABLE__" CGO_ENABLED=1

if ! [[ "$0" =~ scripts/tests.unit.sh ]]; then
  echo "must be run from nuklaivm root"
  exit 255
fi

source ./scripts/hypersdk/common/utils.sh
source ./scripts/hypersdk/constants.sh

# Provision of the list of tests requires word splitting, so disable the shellcheck
# shellcheck disable=SC2046
go test -race -timeout="10m" -coverprofile="coverage.out" -covermode="atomic" $(find . -name "*.go" | grep -v "./cmd" | grep -v "./tests" | xargs -n1 dirname | sort -u | xargs)
