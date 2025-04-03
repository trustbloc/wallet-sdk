#!/bin/bash
#
# Copyright Avast Software. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -eo pipefail

# Function to wait for a service to be available
wait_for_service() {
  local host=$1
  local port=$2
  local timeout=$3
  local start_time=$(date +%s)

  echo "Waiting for $host:$port..."
  while ! nc -z $host $port; do
    sleep 1
    local current_time=$(date +%s)
    local elapsed_time=$((current_time - start_time))

    if [ $elapsed_time -gt $timeout ]; then
      echo "Service $host:$port not available after $timeout seconds"
      return 1
    fi
  done
  echo "$host:$port is available"
}

# Function to check simulator status
wait_for_simulator() {
  local device_name="iPhone 15"
  local max_attempts=10
  local attempt=0

  echo "Waiting for simulator to boot..."
  while [ "$(xcrun simctl list devices | grep -c "$device_name.*Booted")" -eq 0 ]; do
    if [ $attempt -ge $max_attempts ]; then
      echo "Simulator failed to boot after $max_attempts attempts"
      return 1
    fi

    sleep 10
    attempt=$((attempt + 1))
    echo "Attempt $attempt/$max_attempts - Simulator not yet booted..."
  done
  echo "Simulator is booted"
}

# Function to capture diagnostics
capture_diagnostics() {
  echo "### Capturing diagnostics ###"
  xcrun simctl spawn "iPhone 15" log show --predicate 'process == "Runner"' --last 5m > simulator_logs.txt
  docker ps -a > docker_containers.txt
  netstat -an | grep "LISTEN" > ports_listening.txt

  # Upload artifacts if running in CI
  if [ -n "$GITHUB_ACTIONS" ]; then
    mkdir -p diagnostics
    mv *.txt diagnostics/
    tar -czvf diagnostics.tar.gz diagnostics/
  fi
}

# Main execution
try_main() {
  # Wait for critical services
  wait_for_service localhost 8075 30 || return 1  # mock-login-consent
  wait_for_service localhost 8072 30 || return 1  # mock-trust-registry
  wait_for_service localhost 8097 30 || return 1  # mock-attestation

  cd test/integration
  INITIATE_ISSUANCE_URL="$(../../build/bin/integration-cli issuance bank_issuer)"
  INITIATE_VERIFICATION_URL="$(../../build/bin/integration-cli verification v_myprofile_jwt_verified_employee#withScope=registration+testscope)"
  INITIATE_ISSUANCE_URLS_AUTH_CODE_FLOW="$(../../build/bin/integration-cli auth-code-flow bank_issuer did_ion_issuer)"

  echo "Generated Test URLs:"
  echo "INITIATE_ISSUANCE_URL:${INITIATE_ISSUANCE_URL}"
  echo "INITIATE_VERIFICATION_URL:${INITIATE_VERIFICATION_URL}"
  echo "INITIATE_ISSUANCE_URLS_AUTH_CODE_FLOW:${INITIATE_ISSUANCE_URLS_AUTH_CODE_FLOW}"

  cd ../../demo/app

  # Build with verbose output
  echo "Building Flutter iOS app..."
  flutter build ios --simulator --verbose

  # Boot simulator
  echo "Booting simulator..."
  xcrun simctl boot "iPhone 15" || true
  wait_for_simulator || return 1

  # Configure simulator
  echo "Configuring simulator..."
  xcrun simctl status_bar "iPhone 15" override \
    --time "9:41" \
    --dataNetwork wifi \
    --wifiMode active \
    --wifiBars 3 \
    --cellularMode active \
    --cellularBars 4 \
    --batteryState charged \
    --batteryLevel 100

  cd ios

  # Run tests with retries
  MAX_RETRIES=3
  ATTEMPT=0
  TEST_RESULT=1

  while [ $ATTEMPT -lt $MAX_RETRIES ] && [ $TEST_RESULT -ne 0 ]; do
    ATTEMPT=$((ATTEMPT + 1))
    echo "Running test attempt $ATTEMPT/$MAX_RETRIES..."

    set +e
    xcodebuild test \
      -workspace Runner.xcworkspace \
      -scheme Runner \
      -destination 'platform=iOS Simulator,name=iPhone 15,OS=17.0.1' \
      -test-timeouts-enabled YES \
      -maximum-test-execution-time-allowance 600 \
      -only-testing:RunnerTests/IntegrationTest/testFullFlow \
      -only-testing:RunnerTests/IntegrationTest/testAuthFlow \
      | tee xcodebuild.log
    TEST_RESULT=$?
    set -e

    if [ $TEST_RESULT -ne 0 ]; then
      echo "Test attempt $ATTEMPT failed"
      capture_diagnostics

      if [ $ATTEMPT -lt $MAX_RETRIES ]; then
        echo "Preparing for retry..."
        xcrun simctl shutdown "iPhone 15" || true
        sleep 10
        xcrun simctl boot "iPhone 15" || true
        wait_for_simulator || return 1
        sleep 5
      fi
    fi
  done

  return $TEST_RESULT
}

# Cleanup function
cleanup() {
  echo "Cleaning up..."
  xcrun simctl shutdown "iPhone 15" || true
  # Add any additional cleanup here
}

# Main execution with error handling
trap cleanup EXIT
if ! try_main; then
  echo "Test execution failed"
  capture_diagnostics
  exit 1
fi