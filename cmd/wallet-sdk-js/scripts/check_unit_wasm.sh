#!/bin/bash
#
# Copyright Gen Digital Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
set -e

echo "Running $0"

# Running wasm unit test

cd jsinterop
PKGS="github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-js/jsinterop"
PATH="$GOBIN:$PATH" GOOS=js GOARCH=wasm go test $PKGS -count=1 -exec=wasmbrowsertest -timeout=10m -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn"
cd -
