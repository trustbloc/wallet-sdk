/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package oauth2_test

import (
	"testing"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/oauth2"
)

func TestClientMetadata(t *testing.T) {
	clientMetadata := oauth2.NewClientMetadata()

	clientMetadata.SetRedirectURIs(nil)
	require.Equal(t, 0, clientMetadata.RedirectURIs().Length())

	redirectURIs := &api.StringArray{Strings: []string{"RedirectURI1"}}
	clientMetadata.SetRedirectURIs(redirectURIs)
	require.Equal(t, 1, clientMetadata.RedirectURIs().Length())
	require.Equal(t, "RedirectURI1", clientMetadata.RedirectURIs().AtIndex(0))

	clientMetadata.SetTokenEndpointAuthMethod("TokenEndpointAuthMethod")
	require.Equal(t, "TokenEndpointAuthMethod", clientMetadata.TokenEndpointAuthMethod())

	clientMetadata.SetGrantTypes(nil)
	require.Equal(t, 0, clientMetadata.GrantTypes().Length())

	grantTypes := &api.StringArray{Strings: []string{"GrantType1"}}
	clientMetadata.SetGrantTypes(grantTypes)
	require.Equal(t, 1, clientMetadata.GrantTypes().Length())
	require.Equal(t, "GrantType1", clientMetadata.GrantTypes().AtIndex(0))

	clientMetadata.SetResponseTypes(nil)
	require.Equal(t, 0, clientMetadata.ResponseTypes().Length())

	responseTypes := &api.StringArray{Strings: []string{"ResponseType1"}}
	clientMetadata.SetResponseTypes(responseTypes)
	require.Equal(t, 1, clientMetadata.ResponseTypes().Length())
	require.Equal(t, "ResponseType1", clientMetadata.ResponseTypes().AtIndex(0))

	clientMetadata.SetClientName("ClientName")
	require.Equal(t, "ClientName", clientMetadata.ClientName())

	clientMetadata.SetClientURI("ClientURI")
	require.Equal(t, "ClientURI", clientMetadata.ClientURI())

	clientMetadata.SetLogoURI("LogoURI")
	require.Equal(t, "LogoURI", clientMetadata.LogoURI())

	clientMetadata.SetScope("Scope")
	require.Equal(t, "Scope", clientMetadata.Scope())

	clientMetadata.SetContacts(nil)
	require.Equal(t, 0, clientMetadata.Contacts().Length())

	contacts := &api.StringArray{Strings: []string{"Contact1"}}
	clientMetadata.SetContacts(contacts)
	require.Equal(t, 1, clientMetadata.Contacts().Length())
	require.Equal(t, "Contact1", clientMetadata.Contacts().AtIndex(0))

	clientMetadata.SetTOSURI("TOSURI")
	require.Equal(t, "TOSURI", clientMetadata.TOSURI())

	clientMetadata.SetPolicyURI("PolicyURI")
	require.Equal(t, "PolicyURI", clientMetadata.PolicyURI())

	clientMetadata.SetJWKSetURI("JWKSetURI")
	require.Equal(t, "JWKSetURI", clientMetadata.JWKSetURI())

	jwks := api.NewJSONWebKeySet()
	clientMetadata.SetJWKSet(jwks)
	require.NotNil(t, clientMetadata.JWKSet())

	clientMetadata.SetSoftwareID("SoftwareID")
	require.Equal(t, "SoftwareID", clientMetadata.SoftwareID())

	clientMetadata.SetSoftwareVersion("SoftwareVersion")
	require.Equal(t, "SoftwareVersion", clientMetadata.SoftwareVersion())

	clientMetadata.SetIssuerState("IssuerState")
	require.Equal(t, "IssuerState", clientMetadata.IssuerState())
}
