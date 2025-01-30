#!/bin/bash
#
# Copyright Avast Software. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

cd test/integration
INITIATE_ISSUANCE_URL="$(../../build/bin/integration-cli issuance bank_issuer)"
INITIATE_VERIFICATION_URL="$(../../build/bin/integration-cli verification v_myprofile_jwt_verified_employee#withScope=registration+testscope)"
INITIATE_ISSUANCE_URLS_AUTH_CODE_FLOW="$(../../build/bin/integration-cli auth-code-flow bank_issuer did_ion_issuer)"

echo "INITIATE_ISSUANCE_URL:${INITIATE_ISSUANCE_URL}"
echo "INITIATE_VERIFICATION_URL:${INITIATE_VERIFICATION_URL}"
echo "INITIATE_ISSUANCE_URLS_AUTH_CODE_FLOW:${INITIATE_ISSUANCE_URLS_AUTH_CODE_FLOW}"

cd ../../demo/app

flutter build ios --simulator
cd ios
INITIATE_ISSUANCE_URL="${INITIATE_ISSUANCE_URL}" INITIATE_VERIFICATION_URL="${INITIATE_VERIFICATION_URL}" INITIATE_ISSUANCE_URLS_AUTH_CODE_FLOW="${INITIATE_ISSUANCE_URLS_AUTH_CODE_FLOW}" xcodebuild test -workspace Runner.xcworkspace -scheme Runner  -destination 'platform=iOS Simulator,name=iPhone 15,OS=17.0.1'
