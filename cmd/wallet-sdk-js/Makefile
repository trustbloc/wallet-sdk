# Copyright Gen Digital Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

wasm-test-deps:
	@mkdir -p ./build/bin
	GOBIN=$(GOBIN_PATH) go install github.com/golang/mock/mockgen@v1.5.0
	GOBIN=$(GOBIN_PATH) go install github.com/agnivade/wasmbrowsertest@v0.7.0

.PHONY: unit-test-wasm
unit-test-wasm: export GOBIN=$(GOBIN_PATH)
unit-test-wasm: wasm-test-deps
	@scripts/check_unit_wasm.sh


.PHONY: build-wasm
build-wasm:
	@GOOS=js GOARCH=wasm go build -o assets/wallet-sdk.wasm main.go && gzip -f assets/wallet-sdk.wasm
