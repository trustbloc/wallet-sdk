/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci_test

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/openid4ci"
)

const (
	sampleIssuerDisplay = `{"name":"Example University","locale":"en-US"}`

	sampleCredentialDisplay = `{"overview":{"name":"University Credential","locale":"en-US","logo":{` +
		`"url":"https://exampleuniversity.com/public/logo.png","alternative_text":"a square logo of a university"},` +
		`"background_color":"#12107c","text_color":"#FFFFFF"},"claims":[` +
		`{"label":"ID","value_type":"string","value":"1234","locale":"en-US"},` +
		`{"label":"Given Name","value_type":"string","value":"Alice","locale":"en-US"},` +
		`{"label":"Surname","value_type":"string","value":"Bowman","locale":"en-US"},` +
		`{"label":"GPA","value_type":"number","value":"4.0","locale":"en-US"}]}`

	sampleDisplayData = `{"issuer_display":` + sampleIssuerDisplay +
		`,"credential_displays":[` + sampleCredentialDisplay + `]}`
)

func TestDisplayDataParseAndSerialize(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		displayData, err := openid4ci.ParseDisplayData(sampleDisplayData)
		require.NoError(t, err)
		checkResolvedDisplayData(t, displayData)

		serializedDisplayData, err := displayData.Serialize()
		require.NoError(t, err)
		require.NotEmpty(t, serializedDisplayData)
	})
	t.Run("Fail to unmarshal", func(t *testing.T) {
		displayData, err := openid4ci.ParseDisplayData("")
		require.EqualError(t, err, "unexpected end of JSON input")
		require.Nil(t, displayData)
	})
}

func TestIssuerDisplayParseAndSerialize(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		issuerDisplay, err := openid4ci.ParseIssuerDisplay(sampleIssuerDisplay)
		require.NoError(t, err)
		checkIssuerDisplay(t, issuerDisplay)

		serializedIssuerDisplay, err := issuerDisplay.Serialize()
		require.NoError(t, err)
		require.NotEmpty(t, serializedIssuerDisplay)
	})
	t.Run("Fail to unmarshal", func(t *testing.T) {
		issuerDisplay, err := openid4ci.ParseIssuerDisplay("")
		require.EqualError(t, err, "unexpected end of JSON input")
		require.Nil(t, issuerDisplay)
	})
}

func TestCredentialDisplayParseAndSerialize(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		credentialDisplay, err := openid4ci.ParseCredentialDisplay(sampleCredentialDisplay)
		require.NoError(t, err)
		checkCredentialDisplay(t, credentialDisplay)

		serializedCredentialDisplay, err := credentialDisplay.Serialize()
		require.NoError(t, err)
		require.NotEmpty(t, serializedCredentialDisplay)
	})
	t.Run("Fail to unmarshal", func(t *testing.T) {
		credentialDisplay, err := openid4ci.ParseCredentialDisplay("")
		require.EqualError(t, err, "unexpected end of JSON input")
		require.Nil(t, credentialDisplay)
	})
}

func TestDisplayData_CredentialDisplayAtIndex_IndexOutOfBounds(t *testing.T) {
	displayData, err := openid4ci.ParseDisplayData(sampleDisplayData)
	require.NoError(t, err)

	credentialDisplay := displayData.CredentialDisplayAtIndex(1)
	require.Nil(t, credentialDisplay)

	credentialDisplay = displayData.CredentialDisplayAtIndex(-1)
	require.Nil(t, credentialDisplay)
}

func TestCredentialDisplay_ClaimAtIndex_IndexOutOfBounds(t *testing.T) {
	credentialDisplay, err := openid4ci.ParseCredentialDisplay(sampleCredentialDisplay)
	require.NoError(t, err)

	claim := credentialDisplay.ClaimAtIndex(4)
	require.Nil(t, claim)

	claim = credentialDisplay.ClaimAtIndex(-1)
	require.Nil(t, claim)
}
