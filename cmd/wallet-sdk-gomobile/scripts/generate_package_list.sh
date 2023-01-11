#!/bin/bash
#
# Copyright Avast Software. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script needs to be called from the same folder that the wallet-sdk-gomobile makefile is in.

all_packages=$(go list ./...)

# Depending on the environment, all_packages may use newlines to separate packages instead of spaces.
# The line below converts the newlines to spaces for compatibility with the gomobile command.
# If all_packages already uses spaces, then this command won't make any changes.
all_packages_space_separated="${all_packages//$'\n'/ }"

# The wrapper package is only used internally. Since this package contains exported functions that are not gomobile
# compatible, and we don't want or need this to be exposed to the user of the SDK, we exclude it from the gomobile
# bindings.
package_to_remove="github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"

packages_for_bindings="${all_packages_space_separated//$package_to_remove/}"

echo ${packages_for_bindings}
