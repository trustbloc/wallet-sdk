/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package did_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/did"
)

func TestDIDResolver(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		didResolver, err := did.NewResolver("")
		require.NoError(t, err)

		didDocResolution, err := didResolver.Resolve("did:key:z6MkjfbzWitsSUyFMTbBUSWNsJBHR7BefFp1WmABE3kRw8Qr")
		require.NoError(t, err)
		require.NotEmpty(t, didDocResolution)
	})

	t.Run("fail to initialize with invalid resolver server URI", func(t *testing.T) {
		didResolver, err := did.NewResolver("not a uri")
		require.Error(t, err)
		require.Nil(t, didResolver)
		require.Contains(t, err.Error(), "failed to initialize client for DID resolution server")
	})
}

func TestDIDResolver_InvalidDID(t *testing.T) {
	didResolver, err := did.NewResolver("")
	require.NoError(t, err)

	didDocResolution, err := didResolver.Resolve("did:example:abc")
	require.Error(t, err)
	require.EqualError(t, err, "resolve did:example:abc : did method example not supported for vdr")
	require.Empty(t, didDocResolution)
}
