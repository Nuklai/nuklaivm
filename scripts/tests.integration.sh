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

# Ensure cleanup on script exit
cleanup() {
  echo "Performing cleanup..."
  rm -f total_results.txt  # Remove the temporary results file
  rm -f tests/integration/NodeID-*  # Remove NodeID logs
  echo "Cleanup completed."
}
# Register the cleanup function to be called on script exit
trap cleanup EXIT

# Function to run ginkgo tests and capture results
run_tests() {
  local focus=$1
  local output=$(mktemp)

  echo "Running tests with focus: ${focus}"
  ACK_GINKGO_RC=true ginkgo run -v --fail-fast --cover --covermode=atomic \
    --coverpkg=github.com/nuklai/nuklaivm/... \
    --coverprofile=integration.coverage.out \
    --focus="${focus}" ./tests/integration --vms 3 --min-price 1 | tee "${output}"

  # Extract the total number of tests run from the output
  local result=$(grep "Ran [0-9]\+ of [0-9]\+ Specs" "${output}" | tail -n 1)
  rm -f "${output}" # Clean up the temporary file
  echo "${result}"
  echo "${result}" >> total_results.txt
}

# Check if a specific test name or pattern was provided
if [ "$#" -gt 0 ]; then
  FOCUS_FLAG="$1"
fi

# If FOCUS_FLAG is set, only run the tests that match the pattern
# Otherwise, run all tests
if [ -n "${FOCUS_FLAG}" ]; then
  echo "Running integration tests with focus on: ${FOCUS_FLAG}"
  run_tests "${FOCUS_FLAG}"
else
  echo "Running all integration tests"
  # Clear previous results
  > total_results.txt

  # Run tests in the specified order
  run_tests "ping"
  run_tests "network"
  run_tests "tx_processing"
  run_tests "assets"
  run_tests "datasets"
  run_tests "marketplace"

  # Summarize total results
  total_run=0
  total_specs=0

  while IFS= read -r line; do
    specs_run=$(echo "$line" | grep -oP 'Ran \K[0-9]+')
    specs_total=$(echo "$line" | grep -oP 'of \K[0-9]+')
    total_run=$((total_run + specs_run))
  done < total_results.txt

  echo "Total: Ran ${total_run} of ${specs_total} Specs"
fi

# Generate coverage HTML report
go tool cover -html=integration.coverage.out -o=integration.coverage.html
