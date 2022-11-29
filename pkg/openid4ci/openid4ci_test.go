/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci_test

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
	"github.com/trustbloc/wallet-sdk/pkg/openid4ci"
)

const (
	sampleRequestURI = "openid-vc://initiate_issuance?issuer=https%3A%2F%2Fserver%2Eexample%2Ecom" +
		"&credential_type=https%3A%2F%2Fdid%2Eexample%2Eorg%2FhealthCard" +
		"&pre-authorized_code=SplxlOBeZQQYbYS6WxSbIA" +
		"&user_pin_required=false"
	sampleTokenResponse = `{"access_token":"eyJhbGciOiJSUzI1NiIsInR5cCI6Ikp..sHQ",` +
		`"token_type":"bearer","expires_in":86400,"c_nonce":"tZignsnFbp","c_nonce_expires_in":86400}`
	sampleCredentialResponse = `{"format":"jwt_vc","credential":"LUpixVCWJk0eOt4CXQe1NXK....WZwmhmn9OQp6YxX0a2L",` +
		`"c_nonce":"fGFF7UkhLa","c_nonce_expires_in":86400}`
)

var (
	//go:embed testdata/sample_issuer_metadata.json
	sampleIssuerMetadata []byte

	//go:embed testdata/credential_university_degree.jsonld
	credentialUniversityDegree []byte
)

func TestNewInteraction(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		newInteraction(t, sampleRequestURI)
	})
	t.Run("Fail to parse URI", func(t *testing.T) {
		interaction, err := openid4ci.NewInteraction("%")
		require.EqualError(t, err, `parse "%": invalid URL escape "%"`)
		require.Nil(t, interaction)
	})
	t.Run("Fail to parse user_pin_required URL query parameter", func(t *testing.T) {
		requestURI := "openid-vc://initiate_issuance?&user_pin_required=notabool"

		interaction, err := openid4ci.NewInteraction(requestURI)
		require.EqualError(t, err, `strconv.ParseBool: parsing "notabool": invalid syntax`)
		require.Nil(t, interaction)
	})
}

func TestInteraction_Authorize(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		interaction := newInteraction(t, sampleRequestURI)

		result, err := interaction.Authorize()
		require.NoError(t, err)
		require.NotNil(t, result)
	})
	t.Run("Pre-authorized code not provided", func(t *testing.T) {
		requestURI := "openid-vc://initiate_issuance"

		interaction, err := openid4ci.NewInteraction(requestURI)
		require.NoError(t, err)

		result, err := interaction.Authorize()
		require.EqualError(t, err, "pre-authorized code is required (authorization flow not implemented)")
		require.Nil(t, result)
	})
}

type mockIssuerServerHandler struct {
	issuerMetadata                                    string
	tokenRequestShouldFail                            bool
	tokenRequestShouldGiveUnmarshallableResponse      bool
	credentialRequestShouldFail                       bool
	credentialRequestShouldGiveUnmarshallableResponse bool
	credentialResponse                                []byte
}

func (m *mockIssuerServerHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var err error

	switch request.URL.Path {
	case "/.well-known/openid-configuration":
		_, err = writer.Write([]byte(m.issuerMetadata))
	case "/connect/token":
		switch {
		case m.tokenRequestShouldFail:
			writer.WriteHeader(http.StatusInternalServerError)
			_, err = writer.Write([]byte("test failure"))
		case m.tokenRequestShouldGiveUnmarshallableResponse:
			_, err = writer.Write([]byte("invalid"))
		default:
			_, err = writer.Write([]byte(sampleTokenResponse))
		}
	case "/credential":
		switch {
		case m.credentialRequestShouldFail:
			writer.WriteHeader(http.StatusInternalServerError)
			_, err = writer.Write([]byte("test failure"))
		case m.credentialRequestShouldGiveUnmarshallableResponse:
			_, err = writer.Write([]byte("invalid"))
		default:
			_, err = writer.Write(m.credentialResponse)
		}
	}

	if err != nil {
		println(err.Error())
	}
}

func TestInteraction_RequestCredential(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{credentialResponse: []byte(sampleCredentialResponse)}
		server := httptest.NewServer(issuerServerHandler)

		issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"issuer":"https://server.example.com",`+
			`"authorization_endpoint":"https://server.example.com/connect/authorize",`+
			`"token_endpoint":"%s/connect/token",`+
			`"pushed_authorization_request_endpoint":"https://server.example.com/connect/par-authorize",`+
			`"require_pushed_authorization_requests":false,`+
			`"credential_endpoint":"%s/credential"}`, server.URL, server.URL)

		defer server.Close()

		serverURLEscaped := url.QueryEscape(server.URL)

		requestURI := "openid-vc://initiate_issuance?issuer=" + serverURLEscaped +
			"&credential_type=https%3A%2F%2Fdid%2Eexample%2Eorg%2FhealthCard" +
			"&pre-authorized_code=SplxlOBeZQQYbYS6WxSbIA" +
			"&user_pin_required=false"

		interaction := newInteraction(t, requestURI)

		credentialRequest := &openid4ci.CredentialRequestOpts{UserPIN: "1234"}

		credentialResponses, err := interaction.RequestCredential(credentialRequest)
		require.NoError(t, err)
		require.Len(t, credentialResponses, 1)
		require.NotEmpty(t, credentialResponses[0])
	})
	t.Run("PIN required per initiation request, but none provided", func(t *testing.T) {
		requestURI := "openid-vc://initiate_issuance?&user_pin_required=true"

		interaction, err := openid4ci.NewInteraction(requestURI)
		require.NoError(t, err)

		credentialRequest := &openid4ci.CredentialRequestOpts{}

		credentialResponses, err := interaction.RequestCredential(credentialRequest)
		require.EqualError(t, err, "invalid user PIN")
		require.Nil(t, credentialResponses)
	})
	t.Run("Fail to get issuer metadata", func(t *testing.T) {
		requestURI := "openid-vc://initiate_issuance?issuer=http://BadURL" +
			"&credential_type=https%3A%2F%2Fdid%2Eexample%2Eorg%2FhealthCard" +
			"&pre-authorized_code=SplxlOBeZQQYbYS6WxSbIA" +
			"&user_pin_required=false"

		interaction := newInteraction(t, requestURI)

		credentialRequest := &openid4ci.CredentialRequestOpts{}

		credentialResponses, err := interaction.RequestCredential(credentialRequest)
		require.Contains(t, err.Error(), `failed to get issuer metadata: Get `+
			`"http://BadURL/.well-known/openid-configuration": dial tcp: lookup BadURL:`)
		require.Nil(t, credentialResponses)
	})
	t.Run("Fail to reach issuer token endpoint", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{}
		server := httptest.NewServer(issuerServerHandler)

		issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"token_endpoint":"%s/connect/token"}`,
			"http://BadURL")

		defer server.Close()

		serverURLEscaped := url.QueryEscape(server.URL)

		requestURI := "openid-vc://initiate_issuance?issuer=" + serverURLEscaped +
			"&credential_type=https%3A%2F%2Fdid%2Eexample%2Eorg%2FhealthCard" +
			"&pre-authorized_code=SplxlOBeZQQYbYS6WxSbIA" +
			"&user_pin_required=false"

		interaction := newInteraction(t, requestURI)

		credentialRequest := &openid4ci.CredentialRequestOpts{}

		credentialResponses, err := interaction.RequestCredential(credentialRequest)
		require.Contains(t, err.Error(), `failed to get token response: Post `+
			`"http://BadURL/connect/token": dial tcp: lookup BadURL:`)
		require.Nil(t, credentialResponses)
	})
	t.Run("Fail to get token response: server failure", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{tokenRequestShouldFail: true}
		server := httptest.NewServer(issuerServerHandler)

		issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"issuer":"https://server.example.com",`+
			`"authorization_endpoint":"https://server.example.com/connect/authorize",`+
			`"token_endpoint":"%s/connect/token",`+
			`"pushed_authorization_request_endpoint":"https://server.example.com/connect/par-authorize",`+
			`"require_pushed_authorization_requests":false,`+
			`"credential_endpoint":"%s/credential"}`, server.URL, server.URL)

		defer server.Close()

		serverURLEscaped := url.QueryEscape(server.URL)

		requestURI := "openid-vc://initiate_issuance?issuer=" + serverURLEscaped +
			"&credential_type=https%3A%2F%2Fdid%2Eexample%2Eorg%2FhealthCard" +
			"&pre-authorized_code=SplxlOBeZQQYbYS6WxSbIA" +
			"&user_pin_required=false"

		interaction := newInteraction(t, requestURI)

		credentialRequest := &openid4ci.CredentialRequestOpts{}

		credentialResponses, err := interaction.RequestCredential(credentialRequest)
		require.EqualError(t, err, "failed to get token response: received status code [500] with body "+
			"[test failure] from issuer's token endpoint")
		require.Nil(t, credentialResponses)
	})
	t.Run("Fail to unmarshal response from issuer token endpoint", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{tokenRequestShouldGiveUnmarshallableResponse: true}
		server := httptest.NewServer(issuerServerHandler)

		issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"issuer":"https://server.example.com",`+
			`"authorization_endpoint":"https://server.example.com/connect/authorize",`+
			`"token_endpoint":"%s/connect/token",`+
			`"pushed_authorization_request_endpoint":"https://server.example.com/connect/par-authorize",`+
			`"require_pushed_authorization_requests":false,`+
			`"credential_endpoint":"%s/credential"}`, server.URL, server.URL)

		defer server.Close()

		serverURLEscaped := url.QueryEscape(server.URL)

		requestURI := "openid-vc://initiate_issuance?issuer=" + serverURLEscaped +
			"&credential_type=https%3A%2F%2Fdid%2Eexample%2Eorg%2FhealthCard" +
			"&pre-authorized_code=SplxlOBeZQQYbYS6WxSbIA" +
			"&user_pin_required=false"

		interaction := newInteraction(t, requestURI)

		credentialRequest := &openid4ci.CredentialRequestOpts{}

		credentialResponses, err := interaction.RequestCredential(credentialRequest)
		require.EqualError(t, err, "failed to get token response: failed to unmarshal response from the "+
			"issuer's token endpoint: invalid character 'i' looking for beginning of value")
		require.Nil(t, credentialResponses)
	})
	t.Run("Fail to get credential response: server failure", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{credentialRequestShouldFail: true}
		server := httptest.NewServer(issuerServerHandler)

		issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"issuer":"https://server.example.com",`+
			`"authorization_endpoint":"https://server.example.com/connect/authorize",`+
			`"token_endpoint":"%s/connect/token",`+
			`"pushed_authorization_request_endpoint":"https://server.example.com/connect/par-authorize",`+
			`"require_pushed_authorization_requests":false,`+
			`"credential_endpoint":"%s/credential"}`, server.URL, server.URL)

		defer server.Close()

		serverURLEscaped := url.QueryEscape(server.URL)

		requestURI := "openid-vc://initiate_issuance?issuer=" + serverURLEscaped +
			"&credential_type=https%3A%2F%2Fdid%2Eexample%2Eorg%2FhealthCard" +
			"&pre-authorized_code=SplxlOBeZQQYbYS6WxSbIA" +
			"&user_pin_required=false"

		interaction := newInteraction(t, requestURI)

		credentialRequest := &openid4ci.CredentialRequestOpts{}

		credentialResponses, err := interaction.RequestCredential(credentialRequest)
		require.EqualError(t, err, "failed to get credential response: received status code [500] "+
			"with body [test failure] from issuer's credential endpoint")
		require.Nil(t, credentialResponses)
	})
	t.Run("Fail to reach issuer's credential endpoint", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{}
		server := httptest.NewServer(issuerServerHandler)

		issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"token_endpoint":"%s/connect/token",`+
			`"credential_endpoint":"%s/credential"}`, server.URL, "http://BadURL")

		defer server.Close()

		serverURLEscaped := url.QueryEscape(server.URL)

		requestURI := "openid-vc://initiate_issuance?issuer=" + serverURLEscaped +
			"&credential_type=https%3A%2F%2Fdid%2Eexample%2Eorg%2FhealthCard" +
			"&pre-authorized_code=SplxlOBeZQQYbYS6WxSbIA" +
			"&user_pin_required=false"

		interaction := newInteraction(t, requestURI)

		credentialRequest := &openid4ci.CredentialRequestOpts{}

		credentialResponses, err := interaction.RequestCredential(credentialRequest)
		require.Contains(t, err.Error(), `failed to get credential response: `+
			`Post "http://BadURL/credential": dial tcp: lookup BadURL:`)
		require.Nil(t, credentialResponses)
	})
	t.Run("Fail to unmarshal response from issuer credential endpoint", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{credentialRequestShouldGiveUnmarshallableResponse: true}
		server := httptest.NewServer(issuerServerHandler)

		issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"issuer":"https://server.example.com",`+
			`"authorization_endpoint":"https://server.example.com/connect/authorize",`+
			`"token_endpoint":"%s/connect/token",`+
			`"pushed_authorization_request_endpoint":"https://server.example.com/connect/par-authorize",`+
			`"require_pushed_authorization_requests":false,`+
			`"credential_endpoint":"%s/credential"}`, server.URL, server.URL)

		defer server.Close()

		serverURLEscaped := url.QueryEscape(server.URL)

		requestURI := "openid-vc://initiate_issuance?issuer=" + serverURLEscaped +
			"&credential_type=https%3A%2F%2Fdid%2Eexample%2Eorg%2FhealthCard" +
			"&pre-authorized_code=SplxlOBeZQQYbYS6WxSbIA" +
			"&user_pin_required=false"

		interaction := newInteraction(t, requestURI)

		credentialRequest := &openid4ci.CredentialRequestOpts{}

		credentialResponses, err := interaction.RequestCredential(credentialRequest)
		require.EqualError(t, err, "failed to get credential response: failed to unmarshal response "+
			"from the issuer's credential endpoint: invalid character 'i' looking for beginning of value")
		require.Nil(t, credentialResponses)
	})
}

func TestInteraction_ResolveDisplay(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{}
		server := httptest.NewServer(issuerServerHandler)

		defer server.Close()

		var issuerMetadata issuer.Metadata

		err := json.Unmarshal(sampleIssuerMetadata, &issuerMetadata)
		require.NoError(t, err)

		issuerMetadata.CredentialEndpoint = fmt.Sprintf("%s/credential", server.URL)
		issuerMetadata.TokenEndpoint = fmt.Sprintf("%s/connect/token", server.URL)

		issuerMetadataBytes, err := json.Marshal(issuerMetadata)
		require.NoError(t, err)

		issuerServerHandler.issuerMetadata = string(issuerMetadataBytes)

		var credentialResponse openid4ci.CredentialResponse

		err = json.Unmarshal([]byte(sampleCredentialResponse), &credentialResponse)
		require.NoError(t, err)

		credentialResponse.Credential = base64.URLEncoding.EncodeToString(credentialUniversityDegree)

		credentialResponseBytes, err := json.Marshal(credentialResponse)
		require.NoError(t, err)

		issuerServerHandler.credentialResponse = credentialResponseBytes

		serverURLEscaped := url.QueryEscape(server.URL)

		requestURI := "openid-vc://initiate_issuance?issuer=" + serverURLEscaped +
			"&credential_type=https%3A%2F%2Fdid%2Eexample%2Eorg%2FhealthCard" +
			"&pre-authorized_code=SplxlOBeZQQYbYS6WxSbIA" +
			"&user_pin_required=false"

		interaction := newInteraction(t, requestURI)

		credentialRequest := &openid4ci.CredentialRequestOpts{UserPIN: "1234"}

		_, err = interaction.RequestCredential(credentialRequest)
		require.NoError(t, err)

		resolvedDisplayData, err := interaction.ResolveDisplay()
		require.NoError(t, err)
		require.NotNil(t, resolvedDisplayData)
	})
	t.Run("Fail to decode VC", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{}
		server := httptest.NewServer(issuerServerHandler)

		defer server.Close()

		var issuerMetadata issuer.Metadata

		err := json.Unmarshal(sampleIssuerMetadata, &issuerMetadata)
		require.NoError(t, err)

		issuerMetadata.CredentialEndpoint = fmt.Sprintf("%s/credential", server.URL)
		issuerMetadata.TokenEndpoint = fmt.Sprintf("%s/connect/token", server.URL)

		issuerMetadataBytes, err := json.Marshal(issuerMetadata)
		require.NoError(t, err)

		issuerServerHandler.issuerMetadata = string(issuerMetadataBytes)

		var credentialResponse openid4ci.CredentialResponse

		credentialResponse.Credential = "*NotBase64*"

		credentialResponseBytes, err := json.Marshal(credentialResponse)
		require.NoError(t, err)

		issuerServerHandler.credentialResponse = credentialResponseBytes

		serverURLEscaped := url.QueryEscape(server.URL)

		requestURI := "openid-vc://initiate_issuance?issuer=" + serverURLEscaped +
			"&credential_type=https%3A%2F%2Fdid%2Eexample%2Eorg%2FhealthCard" +
			"&pre-authorized_code=SplxlOBeZQQYbYS6WxSbIA" +
			"&user_pin_required=false"

		interaction := newInteraction(t, requestURI)

		credentialRequest := &openid4ci.CredentialRequestOpts{UserPIN: "1234"}

		_, err = interaction.RequestCredential(credentialRequest)
		require.NoError(t, err)

		resolvedDisplayData, err := interaction.ResolveDisplay()
		require.EqualError(t, err, "illegal base64 data at input byte 0")
		require.Nil(t, resolvedDisplayData)
	})
	t.Run("Fail to parse VC", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{}
		server := httptest.NewServer(issuerServerHandler)

		defer server.Close()

		var issuerMetadata issuer.Metadata

		err := json.Unmarshal(sampleIssuerMetadata, &issuerMetadata)
		require.NoError(t, err)

		issuerMetadata.CredentialEndpoint = fmt.Sprintf("%s/credential", server.URL)
		issuerMetadata.TokenEndpoint = fmt.Sprintf("%s/connect/token", server.URL)

		issuerMetadataBytes, err := json.Marshal(issuerMetadata)
		require.NoError(t, err)

		issuerServerHandler.issuerMetadata = string(issuerMetadataBytes)

		var credentialResponse openid4ci.CredentialResponse

		credentialResponse.Credential = ""

		credentialResponseBytes, err := json.Marshal(credentialResponse)
		require.NoError(t, err)

		issuerServerHandler.credentialResponse = credentialResponseBytes

		serverURLEscaped := url.QueryEscape(server.URL)

		requestURI := "openid-vc://initiate_issuance?issuer=" + serverURLEscaped +
			"&credential_type=https%3A%2F%2Fdid%2Eexample%2Eorg%2FhealthCard" +
			"&pre-authorized_code=SplxlOBeZQQYbYS6WxSbIA" +
			"&user_pin_required=false"

		interaction := newInteraction(t, requestURI)

		credentialRequest := &openid4ci.CredentialRequestOpts{UserPIN: "1234"}

		_, err = interaction.RequestCredential(credentialRequest)
		require.NoError(t, err)

		resolvedDisplayData, err := interaction.ResolveDisplay()
		require.EqualError(t, err, "unmarshal new credential: unexpected end of JSON input")
		require.Nil(t, resolvedDisplayData)
	})
}

func newInteraction(t *testing.T, requestURI string) *openid4ci.Interaction {
	t.Helper()

	interaction, err := openid4ci.NewInteraction(requestURI)
	require.NoError(t, err)
	require.NotNil(t, interaction)

	return interaction
}
