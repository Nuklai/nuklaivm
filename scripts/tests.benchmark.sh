#!/usr/bin/env bash
# Copyright (C) 2024, Nuklai. All rights reserved.
# See the file LICENSE for licensing terms.

set -e

if ! [[ "$0" =~ scripts/tests.benchmark.sh ]]; then
  echo "must be run from nuklaivm root"
  exit 255
fi

source ./scripts/hypersdk/common/utils.sh
source ./scripts/hypersdk/constants.sh

# Provision of the list of tests requires word splitting, so disable the shellcheck
# shellcheck disable=SC2046
go test -bench=. -benchmem $(find . -name "*.go" | grep -v "./cmd" | grep -v "./tests" | xargs -n1 dirname | sort -u | xargs)

