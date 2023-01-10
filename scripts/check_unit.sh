#!/bin/bash
#
# Copyright Avast Software. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
set -e

echo "Running $0"

pwd=`pwd`
touch "$pwd"/coverage.out

amend_coverage_file () {
if [ -f profile.out ]; then
    cat profile.out | grep -v ".gen.go" >> "$pwd"/coverage.out
    rm profile.out
fi
}

# Running wallet-sdk unit tests
PKGS=`go list github.com/trustbloc/wallet-sdk/... 2> /dev/null | \
                                                  grep -v /mocks`
go test $PKGS -count=1 -race -coverprofile=profile.out -covermode=atomic -timeout=10m
amend_coverage_file

# Running wallet-sdk-gomobile unit tests
cd cmd/wallet-sdk-gomobile
PKGS=`go list github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile... 2> /dev/null | \
                                                  grep -v /mocks`
go test $PKGS -count=1 -race -coverprofile=profile.out -covermode=atomic -timeout=10m
amend_coverage_file

cd "$pwd"