# Copyright Avast Software.
#
# SPDX-License-Identifier: Apache-2.0

GOBIN_PATH=$(abspath .)/.build/bin

OS := $(shell uname)
ifeq  ($(OS),$(filter $(OS),Darwin Linux))
	PATH:=$(PATH):$(GOBIN_PATH)
else
	PATH:=$(PATH);$(subst /,\\,$(GOBIN_PATH))
endif

.PHONY: all
all: checks unit-test

.PHONY: checks
checks: license lint

.PHONY: lint
lint: 
	@scripts/check_lint.sh

.PHONY: license
license:
	@scripts/check_license.sh

.PHONY: unit-test
unit-test: 
	@scripts/check_unit.sh

.PHONY: generate-android-bindings
generate-android-bindings:
	@make generate-android-bindings -C ./cmd/wallet-sdk-gomobile

.PHONY: generate-ios-bindings
generate-ios-bindings:
	@make generate-ios-bindings -C ./cmd/wallet-sdk-gomobile

.PHONY: copy-android-bindings
copy-android-bindings:
	@mkdir -p "demo/app/android/app/libs" && cp -R cmd/wallet-sdk-gomobile/bindings/android/walletsdk.aar demo/app/android/app/libs

.PHONY: copy-ios-bindings
copy-ios-bindings:
	@cp -R cmd/wallet-sdk-gomobile/bindings/ios/walletsdk.xcframework demo/app/ios/Runner

.PHONY: demo-app-ios
demo-app-ios:generate-ios-bindings copy-ios-bindings
	@cd demo/app && flutter doctor  && flutter clean && npm install -g ios-sim && ios-sim start --devicetypeid "iPhone-14" && flutter devices && flutter run

.PHONY: demo-app-android
demo-app-android: generate-android-bindings copy-android-bindings
	@cd demo/app && flutter doctor && flutter clean && flutter run && flutter emulators --launch  Pixel_3a_API_33_arm64-v8a  && flutter run -d Pixel_3a_API_33_arm64-v8a
