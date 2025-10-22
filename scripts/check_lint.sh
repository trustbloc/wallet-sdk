#!/bin/bash
#
# Copyright Avast Software. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -e

echo "Running $0"

DOCKER_CMD=${DOCKER_CMD:-docker}
GOLANGCI_LINT_IMAGE="golangci/golangci-lint:v2.5"

if [ ! $(command -v ${DOCKER_CMD}) ]; then
    exit 0
fi

echo "Linting top-level module..."
${DOCKER_CMD} run --rm -e GOPROXY=${GOPROXY} -v $(pwd):/opt/workspace -w /opt/workspace/ ${GOLANGCI_LINT_IMAGE} golangci-lint run --timeout 5m -c /opt/workspace/.golangci.yml
echo "Linting wallet-sdk-gomobile..."
${DOCKER_CMD} run --rm -e GOPROXY=${GOPROXY} -v $(pwd):/opt/workspace -w /opt/workspace/cmd/wallet-sdk-gomobile ${GOLANGCI_LINT_IMAGE} golangci-lint run --timeout 5m -c /opt/workspace/.golangci.yml
