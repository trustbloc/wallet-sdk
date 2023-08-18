/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package localkms_test

import (
	"errors"
	"testing"

	"github.com/google/tink/go/keyset"
	"github.com/hyperledger/aries-framework-go/component/kmscrypto/mock/crypto"
	"github.com/hyperledger/aries-framework-go/component/kmscrypto/mock/kms"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/pkg/localkms"
)

func TestAriesCryptoWrapper(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		wrapper := localkms.NewAriesCryptoWrapper(
			&kms.KeyManager{
				GetKeyValue: &keyset.Handle{},
				GetKeyErr:   nil,
			},
			&crypto.Crypto{
				SignValue: []byte("test signature"),
				SignErr:   nil,
				VerifyErr: nil,
			},
		)

		signature, err := wrapper.Sign([]byte("test data"), "testID")
		require.NoError(t, err)
		require.Equal(t, []byte("test signature"), signature)
	})

	t.Run("Success did key id", func(t *testing.T) {
		wrapper := localkms.NewAriesCryptoWrapper(
			&kms.KeyManager{
				GetKeyValue: &keyset.Handle{},
				GetKeyErr:   nil,
			},
			&crypto.Crypto{
				SignValue: []byte("test signature"),
				SignErr:   nil,
				VerifyErr: nil,
			},
		)

		signature, err := wrapper.Sign([]byte("test data"), "did:some:example#testID")
		require.NoError(t, err)
		require.Equal(t, []byte("test signature"), signature)
	})

	t.Run("Invalid key id", func(t *testing.T) {
		wrapper := localkms.NewAriesCryptoWrapper(
			&kms.KeyManager{
				GetKeyErr: errors.New("invalid key id"),
			},
			&crypto.Crypto{},
		)

		_, err := wrapper.Sign([]byte("test data"), "testID")
		require.Error(t, err)
	})
}
