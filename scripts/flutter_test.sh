#!/bin/bash
#
# Copyright Avast Software. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

cd test/integration
INITIATE_ISSUANCE_URL=$(../../build/bin/integration-cli issuance)
INITIATE_VERIFICATION_URL=$(../../build/bin/integration-cli verification)

echo "INITIATE_ISSUANCE_URL:${INITIATE_ISSUANCE_URL}"
echo "INITIATE_VERIFICATION_URL:${INITIATE_VERIFICATION_URL}"

cd ../../demo/app
flutter test integration_test --dart-define=INITIATE_ISSUANCE_URL="${INITIATE_ISSUANCE_URL}" --dart-define=INITIATE_VERIFICATION_URL="${INITIATE_VERIFICATION_URL}"