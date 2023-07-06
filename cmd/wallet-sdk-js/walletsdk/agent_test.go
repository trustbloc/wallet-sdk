/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package walletsdk_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-js/walletsdk"
	"github.com/trustbloc/wallet-sdk/pkg/localkms"
)

func TestNewAgent(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		kmsStore := localkms.NewMemKMSStore()

		agent, err := walletsdk.NewAgent("", kmsStore)

		require.NoError(t, err)
		require.NotNil(t, agent)
	})
}
