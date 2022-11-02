/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package creator_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/did/creator"
)

func TestDIDCreator_Create(t *testing.T) {
	didCreator := creator.NewDIDCreator(nil)

	createDIDOpts := &api.CreateDIDOpts{}

	didDocResolution, err := didCreator.Create(creator.DIDMethodKey, createDIDOpts)
	require.NoError(t, err)
	require.NotEmpty(t, didDocResolution)
}
