#!/bin/bash
#
# Copyright Avast Software. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script needs to be called from the same folder that the wallet-sdk-gomobile makefile is in.

packages_for_bindings=$(. scripts/generate_package_list.sh)

android_java_pkg="dev.trustbloc.wallet.sdk"

gomobile bind -ldflags '-w -s' -androidapi 22 -o bindings/android/walletsdk.aar -javapkg=${android_java_pkg-pkg} -target=android ${packages_for_bindings}
