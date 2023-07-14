#!/bin/bash
#
# Copyright Gen Digital Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

GOOS=js GOARCH=wasm go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn" -o dist/wallet-sdk.wasm main.go
# TODO: add support to instantiate wasm streaming with compressed wasm build
# gzip -f dist/wallet-sdk.wasm

cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" dist/
