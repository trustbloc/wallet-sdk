#!/bin/bash
#
# Copyright Gen Digital Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

ROOT=`pwd`

echo "starting containers..."
cd $ROOT/test/integration/fixtures
(source .env && docker-compose -f docker-compose.yml up --force-recreate -d)

cd $ROOT/test/integration
INITIATE_ISSUANCE_URL="$(../../build/bin/integration-cli issuance bank_issuer)"
INITIATE_VERIFICATION_URL="$(../../build/bin/integration-cli verification v_myprofile_jwt_verified_employee)"

echo "INITIATE_ISSUANCE_URL:${INITIATE_ISSUANCE_URL}"
echo "INITIATE_VERIFICATION_URL:${INITIATE_VERIFICATION_URL}"

cd $ROOT/cmd/wallet-sdk-js

echo "build wasm..."
make build-wasm generate-js-bindings

INITIATE_ISSUANCE_URL="${INITIATE_ISSUANCE_URL}" INITIATE_VERIFICATION_URL="${INITIATE_VERIFICATION_URL}" \
npm run test

