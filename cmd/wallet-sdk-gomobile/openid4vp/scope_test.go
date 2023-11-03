/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/openid4vp"
)

func TestNewScope(t *testing.T) {
	scope := openid4vp.NewScope([]string{"scope1", "scope2"})

	require.Equal(t, 2, scope.Length())
	require.Equal(t, "scope1", scope.AtIndex(0))
	require.Equal(t, "scope2", scope.AtIndex(1))
}
