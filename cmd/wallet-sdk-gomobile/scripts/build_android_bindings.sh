#!/bin/bash
#
# Copyright Avast Software. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script needs to be called from the same folder that the wallet-sdk-gomobile makefile is in.

packages_for_bindings=$(. scripts/generate_package_list.sh)

android_java_pkg="dev.trustbloc.wallet.sdk"
version_package="github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/version"
x_flags="-X '$version_package.version=$NEW_VERSION' -X '$version_package.gitRev=$GIT_REV' -X '$version_package.buildTime=$BUILD_TIME'"
echo "x_flags: $x_flags"

# --- 16K page size alignment enforcement ---
# Some newer ARM64 Android devices require shared objects to have segment alignment compatible with 16K pages.
# Older/internal Go linker outputs 0x1000 alignment; forcing external linking via clang + max-page-size=16384 fixes this.

set -euo pipefail

export CGO_ENABLED=1

# Attempt to auto-detect ANDROID_NDK_HOME if not provided.
if [ -z "${ANDROID_NDK_HOME:-}" ]; then
    if [ -n "${ANDROID_HOME:-}" ] && [ -d "${ANDROID_HOME}/ndk" ]; then
        latest_ndk_dir=$(ls -1 "${ANDROID_HOME}/ndk" 2>/dev/null | sort -V | tail -1 || true)
        if [ -n "$latest_ndk_dir" ]; then
            export ANDROID_NDK_HOME="${ANDROID_HOME}/ndk/${latest_ndk_dir}"
            echo "Detected ANDROID_NDK_HOME=$ANDROID_NDK_HOME"
        fi
    fi
fi

# Pick the aarch64 clang (android API 24+ to cover modern devices) if NDK found.
if [ -n "${ANDROID_NDK_HOME:-}" ] && [ -d "$ANDROID_NDK_HOME" ]; then
    clang_aarch64=$(echo "$ANDROID_NDK_HOME"/toolchains/llvm/prebuilt/*/bin/aarch64-linux-android24-clang 2>/dev/null | awk 'NR==1{print $1}')
    if [ -x "$clang_aarch64" ]; then
        export CC="$clang_aarch64"
        echo "Using external linker toolchain CC=$CC"
        external_link_flags="-linkmode=external -extldflags '-Wl,-z,max-page-size=16384'"
    else
        echo "WARNING: aarch64 clang not found; proceeding without external link flags (may retain 0x1000 alignment)" >&2
        external_link_flags=""
    fi
else
    echo "WARNING: ANDROID_NDK_HOME not set or invalid; cannot enforce 16K alignment" >&2
    external_link_flags=""
fi

ldflags="-w -s $x_flags $external_link_flags"
echo "Final ldflags: $ldflags"

# Raise androidapi to 24 (Android 7.0) which is widely supported and matches the clang we selected.
ANDROID_API_LEVEL=${ANDROID_API_LEVEL:-24}

echo "Building AAR (androidapi ${ANDROID_API_LEVEL}) with forced external linking (if available) for 16K alignment..."

gomobile bind -ldflags "$ldflags" -androidapi ${ANDROID_API_LEVEL} -o bindings/android/walletsdk.aar -javapkg=${android_java_pkg-pkg} -target=android/arm64,android/arm ${packages_for_bindings}

echo "AAR generated at bindings/android/walletsdk.aar"

