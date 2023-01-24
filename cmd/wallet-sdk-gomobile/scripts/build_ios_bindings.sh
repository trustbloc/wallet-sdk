#!/bin/bash
#
# Copyright Avast Software. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script needs to be called from the same folder that the wallet-sdk-gomobile makefile is in.

packages_for_bindings=$(. scripts/generate_package_list.sh)

gomobile bind -ldflags '-w -s' -target=ios -o bindings/ios/walletsdk.xcframework ${packages_for_bindings}
