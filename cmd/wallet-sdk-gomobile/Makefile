# Copyright Avast Software.
#
# SPDX-License-Identifier: Apache-2.0

.PHONY: generate-all-bindings
generate-all-bindings: dependencies gomobile-android-bind gomobile-ios-bind
	@make tidy

.PHONY: generate-android-bindings
generate-android-bindings: gomobile-android-bind
	@make tidy

.PHONY: generate-ios-bindings
generate-ios-bindings: gomobile-ios-bind
	@make tidy

.PHONY: dependencies
dependencies:
	@echo 'Adding gomobile dependencies to go.mod...'
	@go get -u golang.org/x/mobile/cmd/gomobile

.PHONY: gomobile-android-bind
gomobile-android-bind: dependencies tidy
	@make dependencies
	@mkdir -p bindings/android
	@echo 'Generating Android bindings...'
	@scripts/build_android_bindings.sh

.PHONY: gomobile-ios-bind
gomobile-ios-bind: dependencies tidy
	@make dependencies
	@mkdir -p bindings/ios
	@echo 'Generating iOS bindings...'
	@scripts/build_ios_bindings.sh

.PHONY: tidy
tidy:
# Adding the gomobile dependencies causes some new entries to be added to go.mod, however, these dependencies
# are only required for the gomobile command itself, and not the actual Go code. What this means is that they show
# up as unused, and so will reappear after generating the bindings and disappear after running go mod tidy.
# This can be annoying for commits, as it causes the go.mod file to change more times than needed. To make this
# cleaner/more consistent, we call go mod tidy after we generate the bindings (see the bindings targets above).
	@echo 'Tidying up go.mod...'
	@go mod tidy

