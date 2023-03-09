/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package version_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/version"
)

func TestGetVersion(t *testing.T) {
	require.Equal(t, "version-not-set", version.GetVersion())
}

func TestGetGitRevision(t *testing.T) {
	require.Equal(t, "git-rev-not-set", version.GetGitRevision())
}

func TestGetBuildTime(t *testing.T) {
	require.Equal(t, "build-time-not-set", version.GetBuildTime())
}
