/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// RequireErrorContains custom require function to check that error contains a substring.
// nolint:thelper,nolintlint //it isn't helper function it is require function.
func RequireErrorContains(t *testing.T, err error, errString string) {
	require.Error(t, err)
	require.Contains(t, err.Error(), errString)
}
