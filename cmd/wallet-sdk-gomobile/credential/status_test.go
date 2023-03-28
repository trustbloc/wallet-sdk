/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credential_test

import (
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/credential"
)

func TestStatusVerifier(t *testing.T) {
	t.Run("test pass-through to go-sdk status verifier", func(t *testing.T) {
		sv, err := credential.NewStatusVerifier(credential.NewStatusVerifierOptionalArgs())
		require.NoError(t, err)

		err = sv.Verify(&api.VerifiableCredential{
			VC: &verifiable.Credential{},
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "status verification failed")
	})
}
