#!/bin/bash

# Copyright Avast Software. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -e

cd test/integration

# Generate URLs - only generate what's needed based on test selection
if [[ -z "$TEST_NAME" || "$TEST_NAME" == *"single credential"* ]]; then
  INITIATE_ISSUANCE_URLS=$(../../build/bin/integration-cli issuance bank_issuer did_ion_issuer did_ion_issuer drivers_license_issuer)
  INITIATE_VERIFICATION_URLS=$(../../build/bin/integration-cli verification v_myprofile_jwt_verified_employee#withScope=registration v_myprofile_jwt_verified_employee#withScope=registration v_myprofile_jwt_verified_employee#withScope=registration v_myprofile_jwt_drivers_license#withScope=registration)
  echo "Generated single credential URLs"
fi

if [[ -z "$TEST_NAME" || "$TEST_NAME" == *"multiple credentials"* ]]; then
  INITIATE_ISSUANCE_URLS_MULTIPLE_CREDS=$(../../build/bin/integration-cli issuance bank_issuer did_ion_issuer did_ion_issuer drivers_license_issuer)
  INITIATE_VERIFICATION_URLS_MULTIPLE_CREDS=$(../../build/bin/integration-cli verification v_myprofile_jwt_verified_employee#withScope=registration+testscope)
  echo "Generated multiple credentials URLs"
fi

if [[ -z "$TEST_NAME" || "$TEST_NAME" == *"auth code flow"* ]]; then
  INITIATE_ISSUANCE_URLS_AUTH_CODE_FLOW=$(../../build/bin/integration-cli auth-code-flow bank_issuer did_ion_issuer did_ion_issuer drivers_license_issuer)
  echo "Generated auth code flow URLs"
fi

cd ../../demo/app

# Run specific test or all tests
if [ -z "$TEST_NAME" ]; then
  echo "Running all integration tests"
  flutter test integration_test \
    --dart-define=INITIATE_ISSUANCE_URLS="${INITIATE_ISSUANCE_URLS}" \
    --dart-define=WALLET_DID_METHODS="ion key jwk ion" \
    --dart-define=DID_RESOLVER_URI="http://localhost:8072/1.0/identifiers" \
    --dart-define=INITIATE_VERIFICATION_URLS="${INITIATE_VERIFICATION_URLS}" \
    --dart-define=INITIATE_ISSUANCE_URLS_MULTIPLE_CREDS="${INITIATE_ISSUANCE_URLS_MULTIPLE_CREDS}" \
    --dart-define=INITIATE_VERIFICATION_URLS_MULTIPLE_CREDS="${INITIATE_VERIFICATION_URLS_MULTIPLE_CREDS}" \
    --dart-define=INITIATE_ISSUANCE_URLS_AUTH_CODE_FLOW="${INITIATE_ISSUANCE_URLS_AUTH_CODE_FLOW}" \
    "$@"
else
  echo "Running specific test: $TEST_NAME"
  flutter test integration_test \
    --plain-name="$TEST_NAME" \
    --dart-define=INITIATE_ISSUANCE_URLS="${INITIATE_ISSUANCE_URLS}" \
    --dart-define=WALLET_DID_METHODS="ion key jwk ion" \
    --dart-define=DID_RESOLVER_URI="http://localhost:8072/1.0/identifiers" \
    --dart-define=INITIATE_VERIFICATION_URLS="${INITIATE_VERIFICATION_URLS}" \
    --dart-define=INITIATE_ISSUANCE_URLS_MULTIPLE_CREDS="${INITIATE_ISSUANCE_URLS_MULTIPLE_CREDS}" \
    --dart-define=INITIATE_VERIFICATION_URLS_MULTIPLE_CREDS="${INITIATE_VERIFICATION_URLS_MULTIPLE_CREDS}" \
    --dart-define=INITIATE_ISSUANCE_URLS_AUTH_CODE_FLOW="${INITIATE_ISSUANCE_URLS_AUTH_CODE_FLOW}" \
    "$@"
fi