# Copyright Avast Software.
#
# SPDX-License-Identifier: Apache-2.0

GOBIN_PATH=$(abspath .)/.build/bin
GOPROXY ?= https://proxy.golang.org

OS := $(shell uname)
ifeq  ($(OS),$(filter $(OS),Darwin Linux))
	PATH:=$(PATH):$(GOBIN_PATH)
else
	PATH:=$(PATH);$(subst /,\\,$(GOBIN_PATH))
endif

ALPINE_VER ?= 3.20
GO_ALPINE_VER ?= 3.20
GO_VER ?= 1.23

NEW_VERSION ?= $(shell git describe --tags --always `git rev-list --tags --max-count=1`)-SNAPSHOT-$(shell git rev-parse --short=7 HEAD)
GIT_REV ?= $(shell git rev-parse HEAD)
BUILD_TIME ?= $(shell date)

# Namespace for the images
DOCKER_OUTPUT_NS         ?= ghcr.io
REPO_IMAGE_NAME          ?= trustbloc

export TERM := xterm-256color

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
	@GIT_REV="${GIT_REV}" NEW_VERSION="${NEW_VERSION}" BUILD_TIME="${BUILD_TIME}" make generate-android-bindings -C ./cmd/wallet-sdk-gomobile

.PHONY: generate-ios-bindings
generate-ios-bindings:
	@GIT_REV="${GIT_REV}" NEW_VERSION="${NEW_VERSION}" BUILD_TIME="${BUILD_TIME}" make generate-ios-bindings -C ./cmd/wallet-sdk-gomobile

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
	@go build -modfile test/integration/webhook/go.mod -o ./build/bin/webhook-server test/integration/webhook/main.go

.PHONY: sample-webhook-docker
sample-webhook-docker:
	@echo "Building sample webhook server docker image"
	@docker build -f ./images/mocks/webhook/Dockerfile -t wallet-sdk/sample-webhook:latest \
	--build-arg GO_VER=$(GO_VER) \
	--build-arg ALPINE_VER=$(GO_ALPINE_VER) \
	--build-arg GO_PROXY=$(GOPROXY) \
	--build-arg GO_IMAGE=$(GO_IMAGE) .

.PHONY: mock-login-consent-docker
mock-login-consent-docker:
	@echo "Building mock login consent server"
	@docker build -f ./images/mocks/loginconsent/Dockerfile --no-cache -t  wallet-sdk/mock-login-consent:latest \
	--build-arg GO_VER=$(GO_VER) \
	--build-arg ALPINE_VER=$(GO_ALPINE_VER) \
	--build-arg GO_IMAGE=$(GO_IMAGE) .

.PHONY: mock-trust-registry-docker
mock-trust-registry-docker:
	@echo "Building mock trust registry server"
	@docker build -f ./images/mocks/trustregistry/Dockerfile --no-cache -t  wallet-sdk/mock-trust-registry:latest \
	--build-arg GO_VER=$(GO_VER) \
	--build-arg ALPINE_VER=$(GO_ALPINE_VER) \
	--build-arg GO_IMAGE=$(GO_IMAGE) .

.PHONY: mock-attestation-docker
mock-attestation-docker:
	@echo "Building mock attestation server"
	@docker build -f ./images/mocks/attestation/Dockerfile --no-cache -t  wallet-sdk/mock-attestation:latest \
	--build-arg GO_VER=$(GO_VER) \
	--build-arg ALPINE_VER=$(GO_ALPINE_VER) \
	--build-arg GO_PROXY=$(GOPROXY) \
	--build-arg GO_IMAGE=$(GO_IMAGE) .

.PHONY: integration-test
integration-test: mock-login-consent-docker mock-trust-registry-docker mock-attestation-docker sample-webhook-docker generate-test-keys
	@cd test/integration && go mod tidy && ENABLE_COMPOSITION=true go test -count=1 -v -cover . -p 1 -timeout=10m -race

.PHONY: build-integration-cli
build-integration-cli:
	@echo "Building integration cli"
	@mkdir -p ./build/bin
	@cd test/integration/cli && go build -o ../../../build/bin/integration-cli main.go

.PHONY: prepare-integration-test-flutter
prepare-integration-test-flutter: build-integration-cli mock-login-consent-docker mock-trust-registry-docker mock-attestation-docker sample-webhook-docker generate-test-keys start-integration-env-flutter

.PHONY: start-integration-env-flutter
start-integration-env-flutter:
	@scripts/prepare_integration_test_flutter.sh

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

.PHONY: dart-format
dart-format:
	@dart format demo/app/ -l 120

.PHONY: clean
clean:
	@rm -rf ./.build
	@rm -rf coverage*.out
	@rm -Rf ./test/bdd/docker-compose.log

.PHONY: tidy-modules
tidy-modules:
	@find . -type d \( -name build -prune \) -o -name go.mod -print | while read -r gomod_path; do \
		dir_path=$$(dirname "$$gomod_path"); \
		echo "Executing 'go mod tidy' in directory: $$dir_path"; \
		(cd "$$dir_path" && go mod tidy) || exit 1; \
	done
