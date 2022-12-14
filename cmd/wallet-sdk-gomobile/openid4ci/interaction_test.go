/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/doc/jose/jwk/jwksupport"
	arieskms "github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
	"github.com/trustbloc/wallet-sdk/pkg/did/creator"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/openid4ci"
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
	mockDID = "did:test:foo"
)

func TestNewInteraction(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		createInteraction(t, sampleRequestURI)
	})
	t.Run("Fail to parse user_pin_required URL query parameter", func(t *testing.T) {
		requestURI := "openid-vc:///initiate_issuance?&user_pin_required=notabool"

		config := getTestClientConfig(t)

		interaction, err := openid4ci.NewInteraction(requestURI, config)
		require.EqualError(t, err, `strconv.ParseBool: parsing "notabool": invalid syntax`)
		require.Nil(t, interaction)
	})
}

func TestInteraction_Authorize(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		interaction := createInteraction(t, sampleRequestURI)

		result, err := interaction.Authorize()
		require.NoError(t, err)
		require.NotNil(t, result)
	})
	t.Run("Pre-authorized code not provided", func(t *testing.T) {
		requestURI := "openid-vc:///initiate_issuance"

		config := getTestClientConfig(t)

		interaction, err := openid4ci.NewInteraction(requestURI, config)
		require.NoError(t, err)

		result, err := interaction.Authorize()
		require.EqualError(t, err, "pre-authorized code is required (authorization flow not implemented)")
		require.Nil(t, result)
	})
}

type mockIssuerServerHandler struct {
	issuerMetadata string
}

func (m *mockIssuerServerHandler) ServeHTTP(writer http.ResponseWriter, reader *http.Request) {
	var err error

	switch reader.URL.Path {
	case "/.well-known/openid-configuration":
		_, err = writer.Write([]byte(m.issuerMetadata))
	case "/connect/token":
		_, err = writer.Write([]byte(sampleTokenResponse))
	case "/credential":
		_, err = writer.Write([]byte(sampleCredentialResponse))
	}

	if err != nil {
		println(err.Error())
	}
}

type failingSignerCreator struct{}

func (f *failingSignerCreator) Create(*api.JSONObject) (api.Signer, error) {
	return nil, errors.New("test failure")
}

func TestInteraction_RequestCredential(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{}
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

		interaction := createInteraction(t, requestURI)

		credentialRequest := openid4ci.NewCredentialRequestOpts("")

		result, err := interaction.RequestCredential(credentialRequest)
		require.NoError(t, err)
		require.NotNil(t, result)
	})
	t.Run("Success with jwk public key", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{}
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

		config := getTestClientConfig(t, func(handle *api.KeyHandle, kt string) (*did.VerificationMethod, error) {
			jwk, err := jwksupport.PubKeyBytesToJWK(handle.PubKey, arieskms.KeyType(kt))
			require.NoError(t, err)

			return did.NewVerificationMethodFromJWK(handle.KeyID, creator.JSONWebKey2020, mockDID, jwk)
		})

		interaction, err := openid4ci.NewInteraction(requestURI, config)
		require.NoError(t, err)

		credentialRequest := openid4ci.NewCredentialRequestOpts("")

		result, err := interaction.RequestCredential(credentialRequest)
		require.NoError(t, err)
		require.NotNil(t, result)
	})
	t.Run("Fail to create gomobile signer", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{}
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

		config := getTestClientConfig(t)

		config.SignerCreator = &failingSignerCreator{}

		interaction, err := openid4ci.NewInteraction(requestURI, config)
		require.NoError(t, err)

		credentialRequest := openid4ci.NewCredentialRequestOpts("")

		result, err := interaction.RequestCredential(credentialRequest)
		require.EqualError(t, err, "failed to create JWT: failed to create gomobile signer: test failure")
		require.Nil(t, result)
	})
}

func createInteraction(t *testing.T, requestURI string) *openid4ci.Interaction {
	t.Helper()

	config := getTestClientConfig(t)

	interaction, err := openid4ci.NewInteraction(requestURI, config)
	require.NoError(t, err)
	require.NotNil(t, interaction)

	return interaction
}

// getTestClientConfig accepts one optional mockVMCreator.
func getTestClientConfig(t *testing.T, useMockVM ...mockVMCreator) *openid4ci.ClientConfig {
	t.Helper()

	kms, err := localkms.NewKMS(nil)
	require.NoError(t, err)

	signerCreator, err := localkms.CreateSignerCreator(kms)
	require.NoError(t, err)

	resolver := &mockResolver{keyWriter: kms}

	if len(useMockVM) > 0 {
		resolver.makeVM = useMockVM[0]
	}

	clientConfig := openid4ci.NewClientConfig("UserDID", "ClientID", signerCreator, resolver)

	return clientConfig
}

type mockVMCreator func(handle *api.KeyHandle, keyType string) (*did.VerificationMethod, error)

type mockResolver struct {
	keyWriter *localkms.KMS
	makeVM    mockVMCreator
}

func (m *mockResolver) Resolve(string) ([]byte, error) {
	keyHandle, err := m.keyWriter.Create(localkms.KeyTypeED25519)
	if err != nil {
		return nil, err
	}

	if m.makeVM == nil {
		m.makeVM = func(handle *api.KeyHandle, _ string) (*did.VerificationMethod, error) {
			return &did.VerificationMethod{
				ID:         "#key-1",
				Controller: mockDID,
				Type:       "Ed25519VerificationKey2018",
				Value:      handle.PubKey,
			}, nil
		}
	}

	vm, err := m.makeVM(keyHandle, localkms.KeyTypeED25519)
	if err != nil {
		return nil, err
	}

	return mockDocResolution(vm)
}

// mockDocResolution returns a mock DID Doc Resolution with the given verification method.
func mockDocResolution(vm *did.VerificationMethod) ([]byte, error) {
	newDoc := &did.Doc{
		Context: "https://w3id.org/did/v1",
		ID:      mockDID,
		AssertionMethod: []did.Verification{
			{
				VerificationMethod: *vm,
				Relationship:       did.AssertionMethod,
			},
		},
		VerificationMethod: []did.VerificationMethod{
			*vm,
		},
	}

	didDocResolution := &did.DocResolution{DIDDocument: newDoc}

	return didDocResolution.JSONBytes()
}
