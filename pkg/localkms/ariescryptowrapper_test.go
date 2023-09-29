/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package localkms_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/kms-go/mock/suite"
	mockwrapper "github.com/trustbloc/kms-go/mock/wrapper"

	"github.com/trustbloc/wallet-sdk/pkg/localkms"
)

func TestAriesCryptoWrapper(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		wrapper := localkms.NewAriesCryptoWrapper(&suite.MockSuite{
			FixedKeyCryptoVal: &mockwrapper.MockFixedKeyCrypto{
				SignVal: []byte("test signature"),
			},
			SignVal: []byte("test signature"),
		})

		signature, err := wrapper.Sign([]byte("test data"), "testID")
		require.NoError(t, err)
		require.Equal(t, []byte("test signature"), signature)
	})

	t.Run("Success did key id", func(t *testing.T) {
		wrapper := localkms.NewAriesCryptoWrapper(&suite.MockSuite{
			FixedKeyCryptoVal: &mockwrapper.MockFixedKeyCrypto{
				SignVal: []byte("test signature"),
			},
			SignVal: []byte("test signature"),
		})

		signature, err := wrapper.Sign([]byte("test data"), "did:some:example#testID")
		require.NoError(t, err)
		require.Equal(t, []byte("test signature"), signature)
	})

	t.Run("unknown key", func(t *testing.T) {
		wrapper := localkms.NewAriesCryptoWrapper(&suite.MockSuite{
			FixedKeyCryptoErr: errors.New("suite does not have signing key with this ID"),
		})

		_, err := wrapper.Sign([]byte("test data"), "testID")
		require.Error(t, err)
	})

	t.Run("signing error", func(t *testing.T) {
		wrapper := localkms.NewAriesCryptoWrapper(&suite.MockSuite{
			SignErr: errors.New("sign error"),
		})

		_, err := wrapper.Sign([]byte("test data"), "testID")
		require.Error(t, err)
	})
}
