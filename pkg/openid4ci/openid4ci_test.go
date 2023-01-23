/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/crypto/tinkcrypto"
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/doc/util/didsignjwt"
	arieskms "github.com/hyperledger/aries-framework-go/pkg/kms"
	arieslocalkms "github.com/hyperledger/aries-framework-go/pkg/kms/localkms"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/internal/testutil"
	"github.com/trustbloc/wallet-sdk/pkg/did/resolver"
	"github.com/trustbloc/wallet-sdk/pkg/localkms"
	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
	"github.com/trustbloc/wallet-sdk/pkg/openid4ci"
)

//go:embed testdata/sample_credential_response.json
var sampleCredentialResponse []byte

const (
	sampleRequestURI = "openid-vc://initiate_issuance?issuer=https%3A%2F%2Fserver%2Eexample%2Ecom" +
		"&credential_type=https%3A%2F%2Fdid%2Eexample%2Eorg%2FhealthCard" +
		"&pre-authorized_code=SplxlOBeZQQYbYS6WxSbIA" +
		"&user_pin_required=false"
	sampleTokenResponse = `{"access_token":"eyJhbGciOiJSUzI1NiIsInR5cCI6Ikp..sHQ",` +
		`"token_type":"bearer","expires_in":86400,"c_nonce":"tZignsnFbp","c_nonce_expires_in":86400}`
	mockDID = "did:test:foo"
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
		config := getTestClientConfig(t)

		interaction, err := openid4ci.NewInteraction("%", config)
		testutil.RequireErrorContains(t, err, `parse "%": invalid URL escape "%"`)
		require.Nil(t, interaction)
	})
	t.Run("Fail to parse user_pin_required URL query parameter", func(t *testing.T) {
		requestURI := "openid-vc://initiate_issuance?&user_pin_required=notabool"

		config := getTestClientConfig(t)

		interaction, err := openid4ci.NewInteraction(requestURI, config)
		testutil.RequireErrorContains(t, err, `strconv.ParseBool: parsing "notabool": invalid syntax`)
		require.Nil(t, interaction)
	})
	t.Run("Missing client config", func(t *testing.T) {
		interaction, err := openid4ci.NewInteraction("", nil)
		testutil.RequireErrorContains(t, err, "no client config provided")
		require.Nil(t, interaction)
	})
	t.Run("Missing user DID", func(t *testing.T) {
		testConfig := &openid4ci.ClientConfig{}

		interaction, err := openid4ci.NewInteraction("", testConfig)
		testutil.RequireErrorContains(t, err, "no user DID provided")
		require.Nil(t, interaction)
	})
	t.Run("Missing user DID", func(t *testing.T) {
		testConfig := &openid4ci.ClientConfig{UserDID: "UserDID"}

		interaction, err := openid4ci.NewInteraction("", testConfig)
		testutil.RequireErrorContains(t, err, "no client ID provided")
		require.Nil(t, interaction)
	})
	t.Run("Missing signer provider", func(t *testing.T) {
		testConfig := &openid4ci.ClientConfig{UserDID: "UserDID", ClientID: "ClientID"}

		interaction, err := openid4ci.NewInteraction("", testConfig)
		testutil.RequireErrorContains(t, err, "no signer provider provided")
		require.Nil(t, interaction)
	})
	t.Run("Missing DID resolver", func(t *testing.T) {
		testConfig := getTestClientConfig(t)

		testConfig.DIDResolver = nil

		interaction, err := openid4ci.NewInteraction("", testConfig)
		testutil.RequireErrorContains(t, err, "no DID resolver provided")
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

		config := getTestClientConfig(t)

		interaction, err := openid4ci.NewInteraction(requestURI, config)
		require.NoError(t, err)

		result, err := interaction.Authorize()
		testutil.RequireErrorContains(t, err, "pre-authorized code is required (authorization flow not implemented)")
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

type mockResolver struct {
	keyWriter *arieslocalkms.LocalKMS
}

func (m *mockResolver) Resolve(string) (*did.DocResolution, error) {
	didDoc, err := makeMockDoc(m.keyWriter)
	if err != nil {
		return nil, err
	}

	return &did.DocResolution{DIDDocument: didDoc}, err
}

func TestInteraction_RequestCredential(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{credentialResponse: sampleCredentialResponse}
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

		credentials, err := interaction.RequestCredential(credentialRequest)
		require.NoError(t, err)
		require.Len(t, credentials, 1)
		require.NotEmpty(t, credentials[0])
	})
	t.Run("PIN required per initiation request, but none provided", func(t *testing.T) {
		requestURI := "openid-vc://initiate_issuance?&user_pin_required=true"

		config := getTestClientConfig(t)

		interaction, err := openid4ci.NewInteraction(requestURI, config)
		require.NoError(t, err)

		credentialRequest := &openid4ci.CredentialRequestOpts{}

		credentialResponses, err := interaction.RequestCredential(credentialRequest)
		testutil.RequireErrorContains(t, err, "invalid user PIN")
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
		testutil.RequireErrorContains(t, err, "failed to get token response: received status code [500] with body "+
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
		testutil.RequireErrorContains(t, err, "failed to get token response: failed to unmarshal response from the "+
			"issuer's token endpoint: invalid character 'i' looking for beginning of value")
		require.Nil(t, credentialResponses)
	})
	t.Run("Fail to create JWT", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{credentialResponse: sampleCredentialResponse}
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

		var err error
		config.DIDResolver, err = resolver.NewDIDResolver("")
		require.NoError(t, err)

		interaction, err := openid4ci.NewInteraction(requestURI, config)
		require.NoError(t, err)

		credentialRequest := &openid4ci.CredentialRequestOpts{UserPIN: "1234"}

		credentialResponses, err := interaction.RequestCredential(credentialRequest)
		testutil.RequireErrorContains(t, err, "resolve UserDID : "+
			"wrong format did input: UserDID")
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
		testutil.RequireErrorContains(t, err, "failed to get credential response: received status code [500] "+
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
		testutil.RequireErrorContains(t, err, "failed to get credential response: failed to unmarshal response "+
			"from the issuer's credential endpoint: invalid character 'i' looking for beginning of value")
		require.Nil(t, credentialResponses)
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

		vcs, err := interaction.RequestCredential(credentialRequest)
		require.EqualError(t, err, "CREDENTIAL_PARSE_FAILED(OCI1-0013):failed to parse credential "+
			"from credential response at index 0: unmarshal new credential: unexpected end of JSON input")
		require.Nil(t, vcs)
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

		err = json.Unmarshal(sampleCredentialResponse, &credentialResponse)
		require.NoError(t, err)

		credentialResponse.Credential = string(credentialUniversityDegree)

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

		resolvedDisplayData, err := interaction.ResolveDisplay("")
		require.NoError(t, err)
		require.NotNil(t, resolvedDisplayData)
	})
}

func newInteraction(t *testing.T, requestURI string) *openid4ci.Interaction {
	t.Helper()

	config := getTestClientConfig(t)

	interaction, err := openid4ci.NewInteraction(requestURI, config)
	require.NoError(t, err)
	require.NotNil(t, interaction)

	return interaction
}

func getTestClientConfig(t *testing.T) *openid4ci.ClientConfig {
	t.Helper()

	ariesLocalKMS, err := arieslocalkms.New("ThisIs://Unused", localkms.NewInMemoryStorageProvider())
	require.NoError(t, err)

	tinkCrypto, err := tinkcrypto.New()
	require.NoError(t, err)

	signerProvider := didsignjwt.UseDefaultSigner(ariesLocalKMS, tinkCrypto)

	didResolver := &mockResolver{keyWriter: ariesLocalKMS}

	return &openid4ci.ClientConfig{
		UserDID:        "UserDID",
		ClientID:       "ClientID",
		SignerProvider: signerProvider,
		DIDResolver:    didResolver,
	}
}

// makeMockDoc creates a key in the given KMS and returns a mock DID Doc with a verification method.
func makeMockDoc(keyManager *arieslocalkms.LocalKMS) (*did.Doc, error) {
	_, pkb, err := keyManager.CreateAndExportPubKeyBytes(arieskms.ED25519Type)
	if err != nil {
		return nil, err
	}

	vm := &did.VerificationMethod{
		ID:         "#key-1",
		Controller: mockDID,
		Type:       "Ed25519VerificationKey2018",
		Value:      pkb,
	}

	newDoc := &did.Doc{
		ID: mockDID,
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

	return newDoc, nil
}
