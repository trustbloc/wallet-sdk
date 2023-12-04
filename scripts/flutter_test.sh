#!/bin/bash
#
# Copyright Avast Software. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

cd test/integration
INITIATE_ISSUANCE_URLS=$(../../build/bin/integration-cli issuance bank_issuer did_ion_issuer did_ion_issuer drivers_license_issuer)
INITIATE_VERIFICATION_URLS=$(../../build/bin/integration-cli verification v_myprofile_jwt_verified_employee#withScope=registration v_myprofile_jwt_verified_employee#withScope=registration v_myprofile_jwt_verified_employee#withScope=registration v_myprofile_jwt_drivers_license#withScope=registration)

echo "INITIATE_ISSUANCE_URLS:${INITIATE_ISSUANCE_URLS}"
echo "INITIATE_VERIFICATION_URLS:${INITIATE_VERIFICATION_URLS}"

INITIATE_ISSUANCE_URLS_MULTIPLE_CREDS=$(../../build/bin/integration-cli issuance bank_issuer did_ion_issuer did_ion_issuer drivers_license_issuer)
INITIATE_VERIFICATION_URLS_MULTIPLE_CREDS=$(../../build/bin/integration-cli verification v_myprofile_jwt_verified_employee#withScope=registration+testscope)

echo "INITIATE_ISSUANCE_URLS_MULTIPLE_CREDS:${INITIATE_ISSUANCE_URLS_MULTIPLE_CREDS}"
echo "INITIATE_VERIFICATION_URLS_MULTIPLE_CREDS:${INITIATE_VERIFICATION_URLS_MULTIPLE_CREDS}"

INITIATE_ISSUANCE_URLS_AUTH_CODE_FLOW=$(../../build/bin/integration-cli auth-code-flow bank_issuer did_ion_issuer did_ion_issuer drivers_license_issuer)

echo "INITIATE_ISSUANCE_URLS_AUTH_CODE_FLOW:${INITIATE_ISSUANCE_URLS_AUTH_CODE_FLOW}"

cd ../../demo/app
flutter test integration_test "$@" --dart-define=INITIATE_ISSUANCE_URLS="${INITIATE_ISSUANCE_URLS}" --dart-define=WALLET_DID_METHODS="ion key jwk ion" --dart-define=DID_RESOLVER_URI="http://localhost:8072/1.0/identifiers" --dart-define=INITIATE_VERIFICATION_URLS="${INITIATE_VERIFICATION_URLS}" --dart-define=INITIATE_ISSUANCE_URLS_MULTIPLE_CREDS="${INITIATE_ISSUANCE_URLS_MULTIPLE_CREDS}" --dart-define=INITIATE_VERIFICATION_URLS_MULTIPLE_CREDS="${INITIATE_VERIFICATION_URLS_MULTIPLE_CREDS}" --dart-define=INITIATE_ISSUANCE_URLS_AUTH_CODE_FLOW="${INITIATE_ISSUANCE_URLS_AUTH_CODE_FLOW}"