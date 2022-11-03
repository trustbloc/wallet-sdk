/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package localkms_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
)

func TestLocalKMS_Create(t *testing.T) {
	localKMS, err := localkms.NewKMS()
	require.NoError(t, err)

	keyHandle, err := localKMS.Create(localkms.KeyTypeED25519)
	require.NoError(t, err)
	require.NotEmpty(t, keyHandle.Key)
	require.NotEmpty(t, keyHandle.KeyID)
}
