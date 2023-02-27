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
var sampleDIDDocResolution string

//go:embed testdata/doc_with_jwk.jsonld
var jwkDIDDocResolution string

//go:embed testdata/doc_missing_assertionmethod.jsonld
var invalidDIDDocResolution string

func TestDIDDocResolution_ID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		didDocResolution := api.NewDIDDocResolution(sampleDIDDocResolution)

		id, err := didDocResolution.ID()
		require.NoError(t, err)
		require.NotEmpty(t, id)
	})
	t.Run("Fail to parse DID document resolution content", func(t *testing.T) {
		didDocResolution := api.NewDIDDocResolution("")

		id, err := didDocResolution.ID()
		require.EqualError(t, err, "failed to parse DID document resolution content: "+
			"unexpected end of JSON input")
		require.Empty(t, id)
	})
}

func TestDIDDocResolution_AssertionMethod(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		didDocResolution := api.NewDIDDocResolution(sampleDIDDocResolution)

		vm, err := didDocResolution.AssertionMethod()
		require.NoError(t, err)
		require.NotEmpty(t, vm)
	})
	t.Run("Success with JWK", func(t *testing.T) {
		didDocResolution := api.NewDIDDocResolution(jwkDIDDocResolution)

		vm, err := didDocResolution.AssertionMethod()
		require.NoError(t, err)
		require.NotEmpty(t, vm)
	})
	t.Run("Fail to parse DID document resolution content", func(t *testing.T) {
		didDocResolution := api.NewDIDDocResolution("")

		vm, err := didDocResolution.AssertionMethod()
		require.EqualError(t, err, "failed to parse DID document resolution content: "+
			"unexpected end of JSON input")
		require.Empty(t, vm)
	})
	t.Run("DID document missing assertion method", func(t *testing.T) {
		didDocResolution := api.NewDIDDocResolution(invalidDIDDocResolution)

		vm, err := didDocResolution.AssertionMethod()
		require.EqualError(t, err, "DID provided has no assertion method to use as a default signing key")
		require.Empty(t, vm)
	})
}
