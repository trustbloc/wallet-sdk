# Copyright SecureKey Technologies Inc.
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

.PHONY: test-reference-app-web
test-reference-app-web:
	@cd test/referenceapp && npm install && ionic serve

.PHONY: test-reference-app-ios
test-reference-app-ios:
	@cd test/referenceapp && ionic cap open ios

.PHONY: test-reference-app-android
test-reference-app-android:
	@cd test/referenceapp && ionic cap open android
