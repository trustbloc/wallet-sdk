/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credential_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	afgoverifiable "github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/credential"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
)

func TestStatusVerifier(t *testing.T) {
	t.Run("test pass-through to go-sdk status verifier", func(t *testing.T) {
		opts := credential.NewStatusVerifierOpts().SetHTTPTimeoutNanoseconds(0)

		sv, err := credential.NewStatusVerifier(opts)
		require.NoError(t, err)

		err = sv.Verify(&verifiable.Credential{
			VC: &afgoverifiable.Credential{},
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "status verification failed")
	})
	t.Run("NewStatusVerifierWithDIDResolver called with a nil DID resolver", func(t *testing.T) {
		sv, err := credential.NewStatusVerifierWithDIDResolver(nil, nil)
		require.EqualError(t, err, "DID resolver must be provided. If support for DID-URL "+
			"resolution of status credentials is not needed, then use NewStatusVerifier instead")
		require.Nil(t, sv)
	})
}
