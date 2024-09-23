#!/usr/bin/env bash
# Copyright (C) 2024, Nuklai. All rights reserved.
# See the file LICENSE for licensing terms.

set -e

VMWITHCONTRACTS_PATH=$(
  cd "$(dirname "${BASH_SOURCE[0]}")"
  cd .. && pwd
)

ginkgo -v "$VMWITHCONTRACTS_PATH"/tests/e2e/e2e.test -- --stop-network
