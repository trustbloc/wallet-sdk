/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

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
		opts := did.NewResolverOpts().SetHTTPTimeoutNanoseconds(0).DisableHTTPClientTLSVerify()

		didResolver, err := did.NewResolver(opts)
		require.NoError(t, err)

		didDocResolution, err := didResolver.Resolve("did:key:z6MkjfbzWitsSUyFMTbBUSWNsJBHR7BefFp1WmABE3kRw8Qr")
		require.NoError(t, err)
		require.NotEmpty(t, didDocResolution)
	})

	t.Run("fail to initialize with invalid resolver server URI", func(t *testing.T) {
		opts := did.NewResolverOpts()
		opts.SetResolverServerURI("not a uri")

		didResolver, err := did.NewResolver(opts)
		require.Error(t, err)
		require.Nil(t, didResolver)
		require.Contains(t, err.Error(), "failed to initialize client for DID resolution server")
	})
}

func TestDIDResolver_InvalidDID(t *testing.T) {
	didResolver, err := did.NewResolver(nil)
	require.NoError(t, err)

	didDocResolution, err := didResolver.Resolve("did:example:abc")
	require.Error(t, err)
	requireErrorContains(t, err, "DID_RESOLUTION_FAILED")
	require.Empty(t, didDocResolution)
}

func requireErrorContains(t *testing.T, err error, errString string) { //nolint:thelper
	require.Error(t, err)
	require.Contains(t, err.Error(), errString)
}
