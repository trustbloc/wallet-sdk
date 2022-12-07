/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package api_test

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

//go:embed testdata/valid_doc_resolution.jsonld
var sampleDIDDocResolution []byte

func TestDIDDocResolution_ID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		didDocResolution := api.NewDIDDocResolution(sampleDIDDocResolution)

		id, err := didDocResolution.ID()
		require.NoError(t, err)
		require.NotEmpty(t, id)
	})
	t.Run("Fail to parse DID document resolution content", func(t *testing.T) {
		didDocResolution := api.NewDIDDocResolution([]byte{})

		id, err := didDocResolution.ID()
		require.EqualError(t, err, "failed to parse DID document resolution content: "+
			"unexpected end of JSON input")
		require.Empty(t, id)
	})
}
