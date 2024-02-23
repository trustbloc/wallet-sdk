/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package display_test

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/display"
)

const (
	sampleIssuerDisplay = `{"name":"Example University","locale":"en-US", "url": "https://server.example.com",
		"logo": {"uri": "https://exampleuniversity.com/public/logo.png"}, 
		"background_color": "#12107c", "text_color": "#FFFFFF"}`

	sampleCredentialDisplay = `{"overview":{"name":"University Credential","locale":"en-US","logo":{` +
		`"uri":"https://exampleuniversity.com/public/logo.png","alt_text":"a square logo of a university"},` +
		`"background_color":"#12107c","text_color":"#FFFFFF"},"claims":[` +
		`{"raw_id":"id","label":"ID","value_type":"string","order":0,"raw_value":"1234","locale":"en-US"},` +
		`{"raw_id":"given_name","label":"Given Name","value_type":"string","order":1,"raw_value":"Alice","locale":"en-US"},` +
		`{"raw_id":"surname","label":"Surname","value_type":"string","order":2,"raw_value":"Bowman","locale":"en-US"},` +
		`{"raw_id":"gpa","label":"GPA","value_type":"number","raw_value":"4.0","locale":"en-US"},` +
		`{"raw_id":"sensitive_id","label":"Sensitive ID","value_type":"string","value":"•••••6789",` +
		`"raw_value":"123456789","mask":"regex(^(.*).{4}$)","locale":"en-US"},` +
		`{"raw_id":"really_sensitive_id","label":"Really Sensitive ID","value_type":"string","value":"•••••••",` +
		`"raw_value":"abcdefg","mask":"regex((.*))","locale":"en-US"}` +
		`]}`

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

	claim := credentialDisplay.ClaimAtIndex(6)
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
