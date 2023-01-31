#!/bin/bash
#
# Copyright Avast Software. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

cd test/integration
INITIATE_ISSUANCE_URLS=$(../../build/bin/integration-cli issuance bank_issuer did_ion_issuer did_ion_issuer)
INITIATE_VERIFICATION_URLS=$(../../build/bin/integration-cli verification v_myprofile_jwt v_myprofile_jwt v_myprofile_jwt)

echo "INITIATE_ISSUANCE_URLS:${INITIATE_ISSUANCE_URLS}"
echo "INITIATE_VERIFICATION_URLS:${INITIATE_VERIFICATION_URLS}"

cd ../../demo/app
flutter test integration_test --dart-define=INITIATE_ISSUANCE_URLS="${INITIATE_ISSUANCE_URLS}" --dart-define=WALLET_DID_METHODS="ion key jwk" --dart-define=INITIATE_VERIFICATION_URLS="${INITIATE_VERIFICATION_URLS}"