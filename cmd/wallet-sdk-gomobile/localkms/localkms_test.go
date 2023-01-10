/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package localkms_test

import (
	"fmt"
	"testing"

	"github.com/btcsuite/btcutil/base58"
	"github.com/hyperledger/aries-framework-go/component/storageutil/mem"
	"github.com/hyperledger/aries-framework-go/pkg/doc/jose/jwk/jwksupport"
	arieskms "github.com/hyperledger/aries-framework-go/pkg/kms"
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
	t.Run("fail to parse verification method JWK", func(t *testing.T) {
		signerCreator := newSignerCreator(t)

		signer, err := signerCreator.Create(&api.JSONObject{
			Data: []byte(`{
				"id": "foo",
				"type": "JsonWebKey2020",
				"publicKeyJwk": {}
			}`),
		})
		require.EqualError(t, err, "failed to unmarshal verification method JSON into a did.VerificationMethod")
		require.Nil(t, signer)
	})
	t.Run("Failed to create Aries signer", func(t *testing.T) {
		signerCreator := newSignerCreator(t)

		signer, err := signerCreator.Create(&api.JSONObject{Data: []byte("{}")})
		require.EqualError(t, err, "failed to create Aries signer: parsing verification method: vm.Type '' not supported")
		require.Nil(t, signer)
	})
	t.Run("success - verification method with raw key bytes", func(t *testing.T) {
		kmsStore, err := arieskms.NewAriesProviderWrapper(mem.NewProvider())
		require.NoError(t, err)

		keyManager, err := localkms.NewKMS(kmsStore)
		require.NoError(t, err)

		key, err := keyManager.Create(localkms.KeyTypeED25519)
		require.NoError(t, err)

		signerCreator, err := localkms.NewSignerCreator(keyManager)
		require.NoError(t, err)
		require.NotNil(t, signerCreator)

		signer, err := signerCreator.Create(&api.JSONObject{
			Data: []byte(fmt.Sprintf(`{
				"id": "%s",
				"type": "Ed25519VerificationKey2018",
				"publicKeyBase58": "%s"
			}`, key.KeyID, base58.Encode(key.PubKey))),
		})
		require.NoError(t, err)
		require.NotNil(t, signer)
	})
	t.Run("success - JWK verification method", func(t *testing.T) {
		kmsStore, err := arieskms.NewAriesProviderWrapper(mem.NewProvider())
		require.NoError(t, err)

		keyManager, err := localkms.NewKMS(kmsStore)
		require.NoError(t, err)

		key, err := keyManager.Create(localkms.KeyTypeED25519)
		require.NoError(t, err)

		jwk, err := jwksupport.PubKeyBytesToJWK(key.PubKey, arieskms.ED25519Type)
		require.NoError(t, err)

		jwkBytes, err := jwk.MarshalJSON()
		require.NoError(t, err)

		signerCreator, err := localkms.NewSignerCreator(keyManager)
		require.NoError(t, err)
		require.NotNil(t, signerCreator)

		signer, err := signerCreator.Create(&api.JSONObject{
			Data: []byte(fmt.Sprintf(`{
				"id": "%s",
				"type": "JsonWebKey2020",
				"publicKeyJwk": %s
			}`, key.KeyID, string(jwkBytes))),
		})
		require.NoError(t, err)
		require.NotNil(t, signer)
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
