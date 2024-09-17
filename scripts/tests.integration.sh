#!/usr/bin/env bash
# Copyright (C) 2024, Nuklai. All rights reserved.
# See the file LICENSE for licensing terms.

set -e

# Set the CGO flags to use the portable version of BLST
export CGO_CFLAGS="-O -D__BLST_PORTABLE__" CGO_ENABLED=1

if ! [[ "$0" =~ scripts/tests.integration.sh ]]; then
  echo "must be run from repository root"
  exit 255
fi

# remove previous coverage reports
rm -f integration.coverage.out
rm -f integration.coverage.html

# to install the ginkgo binary (required for test build and run)
go install -v github.com/onsi/ginkgo/v2/ginkgo@v2.16.0 || true

# Check if a specific test name or pattern was provided
if [ "$#" -gt 0 ]; then
  TEST_NAME="$1"
  FOCUS_FLAG="--focus=${TEST_NAME}"
else
  FOCUS_FLAG=""
fi

echo "Running with focus: ${FOCUS_FLAG}"

# Run ginkgo with the appropriate focus flag
ACK_GINKGO_RC=true ginkgo \
run \
-v \
--fail-fast \
-cover \
-covermode=atomic \
-coverpkg=github.com/nuklai/nuklaivm/... \
-coverprofile=integration.coverage.out \
./tests/integration \
--vms 3 \
--min-price 1 \
${FOCUS_FLAG}

# Generate coverage HTML report
go tool cover -html=integration.coverage.out -o=integration.coverage.html
