#!/bin/bash
#
# Copyright Avast Software. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script needs to be called from the same folder that the wallet-sdk-gomobile makefile is in.

packages_for_bindings=$(. scripts/generate_package_list.sh)

version_package="github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/version"
x_flags="-X '$version_package.version=$NEW_VERSION' -X '$version_package.gitRev=$GIT_REV' -X '$version_package.buildTime=$BUILD_TIME'"
echo "x_flags: $x_flags"

gomobile bind -ldflags "-w -s $x_flags" -target=ios -o bindings/ios/walletsdk.xcframework ${packages_for_bindings}
