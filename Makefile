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

ALPINE_VER ?= 3.16
GO_VER ?= 1.19

ANDROID_EMULATOR_NAME ?= WalletSDKDeviceEmulator

.PHONY: all
all: checks unit-test integration-test

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
	@rm -rf demo/app/ios/Runner/walletsdk.xcframework && cp -R cmd/wallet-sdk-gomobile/bindings/ios/walletsdk.xcframework demo/app/ios/Runner

.PHONY: demo-app-ios
demo-app-ios:generate-ios-bindings copy-ios-bindings
	@cd demo/app && flutter doctor  && flutter clean && npm install -g ios-sim && ios-sim start --devicetypeid "iPhone-14" && flutter devices && flutter run

.PHONY: demo-app-android
demo-app-android: generate-android-bindings copy-android-bindings
	@cd demo/app && flutter doctor && flutter clean && flutter run && flutter emulators --launch  Pixel_3a_API_33_arm64-v8a  && flutter run -d Pixel_3a_API_33_arm64-v8a

.PHONY: sample-webhook
sample-webhook:
	@echo "Building sample webhook server"
	@mkdir -p ./build/bin
	@go build -o ./build/bin/webhook-server test/integration/webhook/main.go

.PHONY: integration-test
integration-test: generate-test-keys
	@cd test/integration $$ go mod tidy && ENABLE_COMPOSITION=true go test -count=1 -v -cover . -p 1 -timeout=10m -race

.PHONY: build-integration-cli
build-integration-cli:
	@echo "Building integration cli"
	@mkdir -p ./build/bin
	@cd test/integration/cli && go build -o ../../../build/bin/integration-cli main.go

.PHONY: prepare-integration-test-flutter
prepare-integration-test-flutter: build-integration-cli generate-test-keys
	@cd test/integration/fixtures && docker-compose -f docker-compose.yml up --force-recreate -d

.PHONY: integration-test-flutter
integration-test-flutter:
	@scripts/flutter_test.sh

.PHONY: integration-test-android
integration-test-android:
	@scripts/android_test.sh

.PHONY: integration-test-ios
integration-test-ios:
	@scripts/ios_test.sh

.PHONY: install-flutter-dependencies
install-flutter-dependencies:
	@cd demo/app && flutter pub get

.PHONY: start-android-emulator
start-android-emulator:
	@emulator -avd $(ANDROID_EMULATOR_NAME) -writable-system -no-snapshot-load -no-cache

# TODO (#264): frapsoft/openssl only has an amd64 version. While this does work under amd64 and arm64 Mac OS currently,
#               we should add an arm64 version for systems that can only run arm64 code.
.PHONY: generate-test-keys
generate-test-keys:
	@mkdir -p -p test/integration/fixtures/keys/tls
	@docker run -i --platform linux/amd64 --rm \
		-v $(abspath .):/opt/workspace/wallet-sdk \
		--entrypoint /opt/workspace/wallet-sdk/scripts/generate_test_keys.sh \
		frapsoft/openssl