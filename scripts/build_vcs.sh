#!/bin/bash
#
# Copyright Avast Software. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -e

echo "Copying vcs..."
pwd=$(pwd)
rm -rf ./.build/vcs
git clone -b main https://github.com/trustbloc/vcs ./.build/vcs
cd ./.build/vcs || exit

git checkout ${VCS_COMMIT}

cp -rf ../../test/integration/fixtures ./test/bdd/

make generate

cd cmd/vc-rest || exit

go mod tidy

cd $pwd || exit
