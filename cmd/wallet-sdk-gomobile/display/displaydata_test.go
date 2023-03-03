/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package display_test

import (
	_ "embed"
	"testing"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/display"

	"github.com/stretchr/testify/require"
)

const (
	sampleIssuerDisplay = `{"name":"Example University","locale":"en-US"}`

	sampleCredentialDisplay = `{"overview":{"name":"University Credential","locale":"en-US","logo":{` +
		`"url":"https://exampleuniversity.com/public/logo.png","alt_text":"a square logo of a university"},` +
		`"background_color":"#12107c","text_color":"#FFFFFF"},"claims":[` +
		`{"label":"ID","value_type":"string","order":0,"value":"1234","locale":"en-US"},` +
		`{"label":"Given Name","value_type":"string","order":1,"value":"Alice","locale":"en-US"},` +
		`{"label":"Surname","value_type":"string","order":2,"value":"Bowman","locale":"en-US"},` +
		`{"label":"GPA","value_type":"number","value":"4.0","locale":"en-US"}]}`

	sampleDisplayData = `{"issuer_display":` + sampleIssuerDisplay +
		`,"credential_displays":[` + sampleCredentialDisplay + `]}`
)

func TestDisplayDataParseAndSerialize(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		displayData, err := display.ParseData(sampleDisplayData)
		require.NoError(t, err)
		checkResolvedDisplayData(t, displayData)

		serializedDisplayData, err := displayData.Serialize()
		require.NoError(t, err)
		require.NotEmpty(t, serializedDisplayData)
	})
	t.Run("Fail to unmarshal", func(t *testing.T) {
		displayData, err := display.ParseData("")
		require.EqualError(t, err, "unexpected end of JSON input")
		require.Nil(t, displayData)
	})
}

func TestIssuerDisplayParseAndSerialize(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		issuerDisplay, err := display.ParseIssuerDisplay(sampleIssuerDisplay)
		require.NoError(t, err)
		checkIssuerDisplay(t, issuerDisplay)

		serializedIssuerDisplay, err := issuerDisplay.Serialize()
		require.NoError(t, err)
		require.NotEmpty(t, serializedIssuerDisplay)
	})
	t.Run("Fail to unmarshal", func(t *testing.T) {
		issuerDisplay, err := display.ParseIssuerDisplay("")
		require.EqualError(t, err, "unexpected end of JSON input")
		require.Nil(t, issuerDisplay)
	})
}

func TestCredentialDisplayParseAndSerialize(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		credentialDisplay, err := display.ParseCredentialDisplay(sampleCredentialDisplay)
		require.NoError(t, err)
		checkCredentialDisplay(t, credentialDisplay)

		serializedCredentialDisplay, err := credentialDisplay.Serialize()
		require.NoError(t, err)
		require.NotEmpty(t, serializedCredentialDisplay)
	})
	t.Run("Fail to unmarshal", func(t *testing.T) {
		credentialDisplay, err := display.ParseCredentialDisplay("")
		require.EqualError(t, err, "unexpected end of JSON input")
		require.Nil(t, credentialDisplay)
	})
}

func TestDisplayData_CredentialDisplayAtIndex_IndexOutOfBounds(t *testing.T) {
	displayData, err := display.ParseData(sampleDisplayData)
	require.NoError(t, err)

	credentialDisplay := displayData.CredentialDisplayAtIndex(1)
	require.Nil(t, credentialDisplay)

	credentialDisplay = displayData.CredentialDisplayAtIndex(-1)
	require.Nil(t, credentialDisplay)
}

func TestCredentialDisplay_ClaimAtIndex_IndexOutOfBounds(t *testing.T) {
	credentialDisplay, err := display.ParseCredentialDisplay(sampleCredentialDisplay)
	require.NoError(t, err)

	claim := credentialDisplay.ClaimAtIndex(4)
	require.Nil(t, claim)

	claim = credentialDisplay.ClaimAtIndex(-1)
	require.Nil(t, claim)
}

func TestClaim_Order_NoSpecifiedOrder(t *testing.T) {
	displayData, err := display.ParseData(sampleDisplayData)
	require.NoError(t, err)

	credentialDisplay := displayData.CredentialDisplayAtIndex(0)

	claim := credentialDisplay.ClaimAtIndex(3)

	require.False(t, claim.HasOrder())

	order, err := claim.Order()
	require.EqualError(t, err, "claim has no specified order")
	require.Equal(t, order, -1)
}
