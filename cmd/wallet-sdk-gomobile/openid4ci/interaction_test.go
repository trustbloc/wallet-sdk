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

	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	arieskms "github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/stretchr/testify/require"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/did/creator"
	"github.com/trustbloc/wallet-sdk/pkg/models"
	goapiopenid4ci "github.com/trustbloc/wallet-sdk/pkg/openid4ci"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/activitylogger/mem"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/openid4ci"
)

//go:embed testdata/sample_credential_response.json
var sampleCredentialResponse []byte

const (
	sampleTokenResponse = `{"access_token":"eyJhbGciOiJSUzI1NiIsInR5cCI6Ikp..sHQ",` +
		`"token_type":"bearer","expires_in":86400,"c_nonce":"tZignsnFbp","c_nonce_expires_in":86400}`
	mockDID = "did:test:foo"

	mockKeyID = "did:test:foo#abcd"
)

func TestNewInteraction(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
		require.NoError(t, err)

		createInteraction(t, kms, nil, createTestRequestURI("example.com"), nil, false)
	})
}

type mockIssuerServerHandler struct {
	t                                                 *testing.T
	openIDConfig                                      *goapiopenid4ci.OpenIDConfig
	issuerMetadata                                    string
	tokenRequestShouldFail                            bool
	tokenRequestShouldGiveUnmarshallableResponse      bool
	credentialRequestShouldFail                       bool
	credentialRequestShouldGiveUnmarshallableResponse bool
	credentialResponse                                []byte
	headersToCheck                                    *api.Headers
}

func (m *mockIssuerServerHandler) ServeHTTP(writer http.ResponseWriter, //nolint: gocyclo // test file
	request *http.Request,
) {
	var err error

	if m.headersToCheck != nil {
		for _, headerToCheck := range m.headersToCheck.GetAll() {
			// Note: for these tests, we're assuming that there aren't multiple values under a single name/key.
			value := request.Header.Get(headerToCheck.Name)
			require.Equal(m.t, headerToCheck.Value, value)
		}
	}

	switch request.URL.Path {
	case "/.well-known/openid-configuration":
		var openIDConfigBytes []byte

		openIDConfigBytes, err = json.Marshal(m.openIDConfig)
		if err != nil {
			break
		}

		_, err = writer.Write(openIDConfigBytes)
	case "/.well-known/openid-credential-issuer":
		_, err = writer.Write([]byte(m.issuerMetadata))
	case "/oidc/token":
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

	require.NoError(m.t, err)
}

func TestInteraction_RequestCredential(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Using default options", func(t *testing.T) {
			doRequestCredentialTest(t, nil, false)
		})
		t.Run("With additional headers", func(t *testing.T) {
			additionalHeaders := api.NewHeaders()

			additionalHeaders.Add(api.NewHeader("header-name-1", "header-value-1"))
			additionalHeaders.Add(api.NewHeader("header-name-2", "header-value-2"))

			doRequestCredentialTest(t, additionalHeaders, false)

			t.Run("With TLS verification disabled", func(t *testing.T) {
				doRequestCredentialTest(t, additionalHeaders, true)
			})
		})
		t.Run("With TLS verification disabled", func(t *testing.T) {
			doRequestCredentialTest(t, nil, true)
		})
	})
	t.Run("Success with jwk public key", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t:                  t,
			credentialResponse: sampleCredentialResponse,
		}
		server := httptest.NewServer(issuerServerHandler)

		issuerServerHandler.openIDConfig = &goapiopenid4ci.OpenIDConfig{
			TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
		}

		issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential",`+
			`"credential_issuer":"https://server.example.com"}`, server.URL)

		defer server.Close()

		requestURI := createTestRequestURI(server.URL)

		kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
		require.NoError(t, err)

		config := getTestClientConfig(t, kms, nil, nil, false)

		interaction, err := openid4ci.NewInteraction(requestURI, config)
		require.NoError(t, err)

		credentialRequest := openid4ci.NewCredentialRequestOpts("1234")

		keyHandle, err := kms.Create(arieskms.ED25519)
		require.NoError(t, err)

		result, err := interaction.RequestCredential(credentialRequest, &api.VerificationMethod{
			ID:   mockKeyID,
			Type: creator.JSONWebKey2020,
			Key:  models.VerificationKey{JSONWebKey: keyHandle.JWK},
		})

		require.NoError(t, err)
		require.NotNil(t, result)
	})
	t.Run("Fail to sign", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{t: t}
		server := httptest.NewServer(issuerServerHandler)

		defer server.Close()

		kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
		require.NoError(t, err)

		config := getTestClientConfig(t, kms, nil, nil, false)

		requestURI := createTestRequestURI(server.URL)

		interaction, err := openid4ci.NewInteraction(requestURI, config)
		require.NoError(t, err)

		credentialRequest := openid4ci.NewCredentialRequestOpts("")

		result, err := interaction.RequestCredential(credentialRequest, &api.VerificationMethod{
			ID: "did:example:12345#testId", Type: "Invalid",
		})
		requireErrorContains(t, err, "UNSUPPORTED_ALGORITHM")
		require.Nil(t, result)
	})
	t.Run("Missing user PIN", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t:                  t,
			credentialResponse: sampleCredentialResponse,
		}
		server := httptest.NewServer(issuerServerHandler)

		issuerServerHandler.openIDConfig = &goapiopenid4ci.OpenIDConfig{
			TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
		}

		issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential",`+
			`"credential_issuer":"https://server.example.com"}`, server.URL)

		defer server.Close()

		activityLogger := mem.NewActivityLogger()

		kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
		require.NoError(t, err)

		interaction := createInteraction(t, kms, activityLogger, createTestRequestURI(server.URL), nil, false)

		keyHandle, err := kms.Create(arieskms.ED25519)
		require.NoError(t, err)

		pkBytes, err := keyHandle.JWK.PublicKeyBytes()
		require.NoError(t, err)

		credentials, err := interaction.RequestCredential(nil, &api.VerificationMethod{
			ID:   "did:example:12345#testId",
			Type: "Ed25519VerificationKey2018",
			Key:  models.VerificationKey{Raw: pkBytes},
		})
		requireErrorContains(t, err, "the credential offer requires a user PIN, but it was not provided")
		require.Nil(t, credentials)
	})
}

//nolint:thelper // Not a test helper function
func doRequestCredentialTest(t *testing.T, additionalHeaders *api.Headers,
	disableTLSVerification bool,
) {
	issuerServerHandler := &mockIssuerServerHandler{
		t:                  t,
		credentialResponse: sampleCredentialResponse,
		headersToCheck:     additionalHeaders,
	}
	server := httptest.NewServer(issuerServerHandler)

	issuerServerHandler.openIDConfig = &goapiopenid4ci.OpenIDConfig{
		TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
	}

	issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential",`+
		`"credential_issuer":"https://server.example.com"}`, server.URL)

	defer server.Close()

	activityLogger := mem.NewActivityLogger()

	kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
	require.NoError(t, err)

	interaction := createInteraction(t, kms, activityLogger, createTestRequestURI(server.URL), additionalHeaders,
		disableTLSVerification)

	credentialRequest := openid4ci.NewCredentialRequestOpts("1234")

	keyHandle, err := kms.Create(arieskms.ED25519)
	require.NoError(t, err)

	pkBytes, err := keyHandle.JWK.PublicKeyBytes()
	require.NoError(t, err)

	result, err := interaction.RequestCredential(credentialRequest, &api.VerificationMethod{
		ID:   "did:example:12345#testId",
		Type: "Ed25519VerificationKey2018",
		Key:  models.VerificationKey{Raw: pkBytes},
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	numberOfActivitiesLogged := activityLogger.Length()
	require.Equal(t, 1, numberOfActivitiesLogged)

	activity := activityLogger.AtIndex(0)

	require.NotEmpty(t, activity.ID)
	require.Equal(t, goapi.LogTypeCredentialActivity, activity.Type())
	require.NotEmpty(t, activity.UnixTimestamp())
	require.Equal(t, "https://server.example.com", activity.Client())
	require.Equal(t, "oidc-issuance", activity.Operation())
	require.Equal(t, goapi.ActivityLogStatusSuccess, activity.Status())

	params := activity.Params()
	require.NotNil(t, params)

	keyValuePairs := params.AllKeyValuePairs()

	numberOfKeyValuePairs := keyValuePairs.Length()

	require.Equal(t, 1, numberOfKeyValuePairs)

	keyValuePair := keyValuePairs.AtIndex(0)

	key := keyValuePair.Key()
	require.Equal(t, "subjectIDs", key)

	subjectIDs, err := keyValuePair.ValueStringArray()
	require.NoError(t, err)

	numberOfSubjectIDs := subjectIDs.Length()
	require.Equal(t, 1, numberOfSubjectIDs)

	subjectID := subjectIDs.AtIndex(0)
	require.Equal(t, "did:orb:uAAA:EiARTvvCsWFTSCc35447YpI2MJpFAaJZtFlceVz9lcMYVw", subjectID)
}

func createInteraction(t *testing.T, kms *localkms.KMS, activityLogger api.ActivityLogger, requestURI string,
	additionalHeaders *api.Headers, disableTLSVerification bool,
) *openid4ci.Interaction {
	t.Helper()

	config := getTestClientConfig(t, kms, activityLogger, additionalHeaders, disableTLSVerification)

	interaction, err := openid4ci.NewInteraction(requestURI, config)
	require.NoError(t, err)
	require.NotNil(t, interaction)

	return interaction
}

// getTestClientConfig accepts an optional activityLogger and also one optional mockVMCreator.
func getTestClientConfig(t *testing.T, kms *localkms.KMS, activityLogger api.ActivityLogger,
	additionalHeaders *api.Headers, disableTLSVerification bool,
) *openid4ci.ClientConfig {
	t.Helper()

	resolver := &mockResolver{keyWriter: kms}

	clientConfig := openid4ci.NewClientConfig("ClientID", kms.GetCrypto(), resolver, activityLogger)
	clientConfig.DisableVCProofChecks()

	if additionalHeaders != nil {
		clientConfig.AddHeaders(additionalHeaders)
	}

	if disableTLSVerification {
		clientConfig.DisableHTTPClientTLSVerify()
	}

	return clientConfig
}

type mockVMCreator func(key *api.JSONWebKey, keyType string) (*did.VerificationMethod, error)

type mockResolver struct {
	keyWriter *localkms.KMS
	makeVM    mockVMCreator
}

func (m *mockResolver) Resolve(string) ([]byte, error) {
	newKey, err := m.keyWriter.Create(localkms.KeyTypeED25519)
	if err != nil {
		return nil, err
	}

	if m.makeVM == nil {
		m.makeVM = func(key *api.JSONWebKey, _ string) (*did.VerificationMethod, error) {
			if key.JWK == nil {
				return nil, fmt.Errorf("nil key")
			}

			if key.JWK.Kty != "OKP" || key.JWK.Crv != "Ed25519" {
				return nil, fmt.Errorf("default test resolver only supports ed25519 key")
			}

			pkb, e := key.JWK.PublicKeyBytes()
			if e != nil {
				return nil, e
			}

			return &did.VerificationMethod{
				ID:         "#key-1",
				Controller: mockDID,
				Type:       "Ed25519VerificationKey2018",
				Value:      pkb,
			}, nil
		}
	}

	vm, err := m.makeVM(newKey, localkms.KeyTypeED25519)
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

func requireErrorContains(t *testing.T, err error, errString string) { //nolint:thelper
	require.Error(t, err)
	require.Contains(t, err.Error(), errString)
}

func createTestRequestURI(issuerURL string) string {
	issuerURLEscaped := url.QueryEscape(issuerURL)

	return "openid-vc://?credential_offer=%7B%22credential_issuer%22%3A%22" + issuerURLEscaped +
		"%22%2C%22credentials%22%3A%5B%7B%22format%22%3A%22jwt_vc_json%22%2C%22types%22%3A%5B%22Verifiable" +
		"Credential%22%2C%22VerifiedEmployee%22%5D%7D%5D%2C%22grants%22%3A%7B%22urn%3Aietf%3Aparams%3Aoaut" +
		"h%3Agrant-type%3Apre-authorized_code%22%3A%7B%22pre-authorized_code%22%3A%228e557518-bbb1-4483-94" +
		"90-d80f4f54f3361677012959367644351%22%2C%22user_pin_required%22%3Atrue%7D%7D%7D"
}
