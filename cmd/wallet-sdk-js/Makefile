# Copyright Gen Digital Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

clean:
	@rm -rf ./dist ./node_modules ./build

wasm-test-deps:
	@mkdir -p ./build/bin
	GOBIN=$(GOBIN_PATH) go install github.com/golang/mock/mockgen@v1.5.0
	GOBIN=$(GOBIN_PATH) go install github.com/agnivade/wasmbrowsertest@v0.7.0

.PHONY: unit-test-wasm
unit-test-wasm: export GOBIN=$(GOBIN_PATH)
unit-test-wasm: wasm-test-deps
	@scripts/check_unit_wasm.sh

.PHONY: clean build-wasm
build-wasm:
	@GOOS=js GOARCH=wasm go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn" -o dist/wallet-sdk.wasm main.go

.PHONY: clean generate-js-bindings
generate-js-bindings:
	@npm install

