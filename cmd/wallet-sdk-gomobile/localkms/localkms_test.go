/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package localkms_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
)

func TestLocalKMS_Create(t *testing.T) {
	localKMS, err := localkms.NewKMS(nil)
	require.NoError(t, err)

	keyHandle, err := localKMS.Create(localkms.KeyTypeED25519)
	require.NoError(t, err)
	require.NotEmpty(t, keyHandle.PubKey)
	require.NotEmpty(t, keyHandle.KeyID)
}

func TestLocalKMS_ExportPubKey(t *testing.T) {
	localKMS, err := localkms.NewKMS(nil)
	require.NoError(t, err)

	keyHandle, err := localKMS.Create(localkms.KeyTypeED25519)
	require.NoError(t, err)
	require.NotEmpty(t, keyHandle.PubKey)
	require.NotEmpty(t, keyHandle.KeyID)

	_, err = localKMS.ExportPubKey(keyHandle.KeyID)
	require.Contains(t, err.Error(), "not implemented")
}

func TestLocalKMS_GetCrypto(t *testing.T) {
	localKMS, err := localkms.NewKMS(nil)
	require.NoError(t, err)

	crypto := localKMS.GetCrypto()
	require.NotNil(t, crypto)
}

func TestGetDefaultSignerCreator(t *testing.T) {
	newSignerCreator(t)
}

func TestSignerCreator_Create(t *testing.T) {
	t.Run("Unmarshal failure", func(t *testing.T) {
		signerCreator := newSignerCreator(t)

		signer, err := signerCreator.Create(&api.JSONObject{})
		require.EqualError(t, err, "failed to unmarshal verification method JSON into a did.VerificationMethod")
		require.Nil(t, signer)
	})
	t.Run("Failed to create Aries signer", func(t *testing.T) {
		signerCreator := newSignerCreator(t)

		signer, err := signerCreator.Create(&api.JSONObject{Data: []byte("{}")})
		require.EqualError(t, err, "failed to create Aries signer: parsing verification method: vm.Type '' not supported")
		require.Nil(t, signer)
	})
}

func newSignerCreator(t *testing.T) *localkms.SignerCreator {
	t.Helper()

	kms, err := localkms.NewKMS(nil)
	require.NoError(t, err)

	signerCreator, err := localkms.NewSignerCreator(kms)
	require.NoError(t, err)
	require.NotNil(t, signerCreator)

	return signerCreator
}
