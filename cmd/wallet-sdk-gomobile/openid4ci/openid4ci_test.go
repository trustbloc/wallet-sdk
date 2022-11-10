/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/openid4ci"
)

func TestNewInteraction(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		createInteraction(t)
	})
	t.Run("Fail to parse user_pin_required URL query parameter", func(t *testing.T) {
		requestURI := "openid-vc:///initiate_issuance?&user_pin_required=notabool"

		interaction, err := openid4ci.NewInteraction(requestURI)
		require.EqualError(t, err, `strconv.ParseBool: parsing "notabool": invalid syntax`)
		require.Nil(t, interaction)
	})
}

func TestInteraction_Authorize(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		interaction := createInteraction(t)

		result, err := interaction.Authorize()
		require.NoError(t, err)
		require.NotNil(t, result)
	})
	t.Run("Pre-authorized code not provided", func(t *testing.T) {
		requestURI := "openid-vc:///initiate_issuance"

		interaction, err := openid4ci.NewInteraction(requestURI)
		require.NoError(t, err)

		result, err := interaction.Authorize()
		require.EqualError(t, err, "pre-authorized code is required (authorization flow not implemented)")
		require.Nil(t, result)
	})
	t.Run("PIN required per initiation request, but none provided", func(t *testing.T) {
		requestURI := "openid-vc:///initiate_issuance?&user_pin_required=true"

		interaction, err := openid4ci.NewInteraction(requestURI)
		require.NoError(t, err)

		credentialRequest := &openid4ci.CredentialRequest{}

		result, err := interaction.RequestCredential(credentialRequest)
		require.EqualError(t, err, "PIN required (per initiation request)")
		require.Nil(t, result)
	})
}

func TestInteraction_RequestCredential(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		interaction := createInteraction(t)

		credentialRequest := &openid4ci.CredentialRequest{}

		result, err := interaction.RequestCredential(credentialRequest)
		require.NoError(t, err)
		require.NotNil(t, result)
	})
}

func createInteraction(t *testing.T) *openid4ci.Interaction {
	t.Helper()

	requestURI := "openid-vc:///initiate_issuance?issuer=https%3A%2F%2Fserver%2Eexample%2Ecom" +
		"&credential_type=https%3A%2F%2Fdid%2Eexample%2Eorg%2FhealthCard" +
		"&pre-authorized_code=SplxlOBeZQQYbYS6WxSbIA"

	interaction, err := openid4ci.NewInteraction(requestURI)
	require.NoError(t, err)
	require.NotNil(t, interaction)

	return interaction
}
