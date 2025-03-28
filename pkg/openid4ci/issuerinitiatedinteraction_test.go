/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/did-go/doc/did"
	"github.com/trustbloc/did-go/doc/did/endpoint"
	"github.com/trustbloc/kms-go/doc/jose"
	"github.com/trustbloc/kms-go/doc/jose/jwk"
	arieskms "github.com/trustbloc/kms-go/spi/kms"
	"github.com/trustbloc/vc-go/jwt"

	"github.com/trustbloc/wallet-sdk/internal/testutil"
	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	"github.com/trustbloc/wallet-sdk/pkg/did/resolver"
	"github.com/trustbloc/wallet-sdk/pkg/localkms"
	"github.com/trustbloc/wallet-sdk/pkg/models"
	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
	"github.com/trustbloc/wallet-sdk/pkg/openid4ci"
)

const (
	sampleTokenResponse = `{"access_token":"eyJhbGciOiJSUzI1NiIsInR5cCI6Ikp..sHQ",` +
		`"token_type":"bearer","expires_in":86400,"c_nonce":"tZignsnFbp","c_nonce_expires_in":86400}`
	mockDID              = "did:test:foo"
	mockKeyID            = "did:example:12345#testId"
	serverURLPlaceholder = "[SERVER_URL]"
)

var (
	//go:embed testdata/sample_credential_offer.json
	sampleCredentialOffer []byte

	//go:embed testdata/sample_credential_response.json
	sampleCredentialResponse []byte

	//go:embed testdata/sample_credential_response_ask.json
	sampleCredentialResponseAsk []byte

	//go:embed testdata/sample_credential_response_batch.json
	sampleCredentialResponseBatch []byte

	//go:embed testdata/sample_credential_response_jsonld.json
	sampleCredentialResponseJSONLD []byte

	//go:embed testdata/sample_issuer_metadata.json
	sampleIssuerMetadata string

	//go:embed testdata/sample_cred.jwt
	sampleCredJWT string
)

type mockIssuerServerHandler struct {
	t                                                       *testing.T
	credentialOffer                                         *openid4ci.CredentialOffer
	credentialOfferEndpointShouldFail                       bool
	credentialOfferEndpointShouldGiveUnmarshallableResponse bool
	issuerMetadata                                          string
	tokenRequestShouldFail                                  bool
	tokenRequestErrorResponse                               string
	tokenRequestShouldGiveUnmarshallableResponse            bool
	credentialRequestShouldFail                             bool
	credentialRequestErrorResponse                          string
	credentialRequestShouldGiveUnmarshallableResponse       bool
	credentialRequestShouldGiveInvalidProofResponse         bool
	batchCredentialRequestShouldFail                        bool
	credentialResponse                                      []byte
	batchCredentialResponse                                 []byte
	httpStatusCode                                          int
	ackRequestErrorResponse                                 string
	ackRequestExpectInteractionDetails                      bool
	ackRequestExpectedCalls                                 int
}

//nolint:gocyclo // test file
func (m *mockIssuerServerHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var err error

	switch request.URL.Path {
	case "/credential-offer":
		switch {
		case m.credentialOfferEndpointShouldFail:
			writer.WriteHeader(http.StatusInternalServerError)
			_, err = writer.Write([]byte("test failure"))
		case m.credentialOfferEndpointShouldGiveUnmarshallableResponse:
			_, err = writer.Write([]byte("invalid"))
		default:
			var credentialOfferBytes []byte

			credentialOfferBytes, err = json.Marshal(m.credentialOffer)
			assert.NoError(m.t, err)

			_, err = writer.Write(credentialOfferBytes)
		}
	case "/.well-known/openid-credential-issuer":
		_, err = writer.Write([]byte(m.issuerMetadata))
	case "/oidc/token":
		switch {
		case m.tokenRequestShouldFail:
			writer.WriteHeader(http.StatusInternalServerError)

			if m.tokenRequestErrorResponse != "" {
				_, err = writer.Write([]byte(m.tokenRequestErrorResponse))
			} else {
				_, err = writer.Write([]byte("test failure"))
			}
		case m.tokenRequestShouldGiveUnmarshallableResponse:
			_, err = writer.Write([]byte("invalid"))
		default:
			writer.Header().Set("Content-Type", "application/json")
			_, err = writer.Write([]byte(sampleTokenResponse))
		}
	case "/oidc/credential":
		switch {
		case m.credentialRequestShouldFail:
			writer.WriteHeader(http.StatusInternalServerError)

			if m.credentialRequestErrorResponse != "" {
				_, err = writer.Write([]byte(m.credentialRequestErrorResponse))
			} else {
				_, err = writer.Write([]byte("test failure"))
			}
		case m.credentialRequestShouldGiveUnmarshallableResponse:
			statusCode := http.StatusOK

			if m.httpStatusCode != 0 {
				statusCode = m.httpStatusCode
			}

			writer.WriteHeader(statusCode)

			_, err = writer.Write([]byte("invalid"))
		case m.credentialRequestShouldGiveInvalidProofResponse:
			writer.WriteHeader(http.StatusInternalServerError)

			_, err = writer.Write([]byte(`{"error":"invalid_proof"}`))
		default:
			_, err = writer.Write(m.credentialResponse)
		}
	case "/oidc/batch_credential":
		switch {
		case m.batchCredentialRequestShouldFail:
			writer.WriteHeader(http.StatusInternalServerError)

			_, err = writer.Write([]byte("test failure"))
		default:
			_, err = writer.Write(m.batchCredentialResponse)
		}
	case "/oidc/ack_endpoint":
		m.ackRequestExpectedCalls--

		statusCode := http.StatusNoContent

		if m.httpStatusCode != 0 {
			statusCode = m.httpStatusCode
		}

		var payload map[string]interface{}
		err = json.NewDecoder(request.Body).Decode(&payload)
		assert.NoError(m.t, err)

		_, ok := payload["interaction_details"]
		assert.Equal(m.t, m.ackRequestExpectInteractionDetails, ok)

		if m.ackRequestErrorResponse != "" {
			_, err = writer.Write([]byte(m.ackRequestErrorResponse))
		}

		writer.WriteHeader(statusCode)
	}

	assert.NoError(m.t, err)
}

type failingMetricsLogger struct {
	currentAttemptNumber int
	attemptFailNumber    int
}

func (f *failingMetricsLogger) Log(metricsEvent *api.MetricsEvent) error {
	if f.currentAttemptNumber == f.attemptFailNumber {
		return fmt.Errorf("failed to log event (Event=%s)", metricsEvent.Event)
	}

	f.currentAttemptNumber++

	return nil
}

func TestNewIssuerInitiatedInteraction(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{t: t, credentialResponse: sampleCredentialResponse}

		server := httptest.NewServer(issuerServerHandler)
		defer server.Close()

		issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

		t.Run("Credential format is jwt_vc_json", func(t *testing.T) {
			newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, false, true))
		})
		t.Run("Credential format is jwt_vc_json-ld", func(t *testing.T) {
			credentialOffer := createCredentialOffer(t, server.URL, true, true)

			credentialOfferBytes, err := json.Marshal(credentialOffer)
			require.NoError(t, err)

			credentialOfferEscaped := url.QueryEscape(string(credentialOfferBytes))

			credentialOfferIssuanceURI := "openid-credential-offer://?credential_offer=" + credentialOfferEscaped

			newIssuerInitiatedInteraction(t, credentialOfferIssuanceURI)
		})
	})
	t.Run("Fail to populate issuer metadata", func(t *testing.T) {
		requestURI := createCredentialOfferIssuanceURI(t, "invalid url", true, true)
		config := getTestClientConfig(t)

		interaction, err := openid4ci.NewIssuerInitiatedInteraction(requestURI, config)
		require.ErrorContains(t, err, "METADATA_FETCH_FAILED")
		require.Nil(t, interaction)
	})
	t.Run("Fail to parse URI", func(t *testing.T) {
		config := getTestClientConfig(t)

		interaction, err := openid4ci.NewIssuerInitiatedInteraction("%", config)
		testutil.RequireErrorContains(t, err, `parse "%": invalid URL escape "%"`)
		require.Nil(t, interaction)
	})
	t.Run("Missing client config", func(t *testing.T) {
		interaction, err := openid4ci.NewIssuerInitiatedInteraction("", nil)
		testutil.RequireErrorContains(t, err, "no client config provided")
		require.Nil(t, interaction)
	})
	t.Run("Missing DID resolver", func(t *testing.T) {
		testConfig := getTestClientConfig(t)

		testConfig.DIDResolver = nil

		interaction, err := openid4ci.NewIssuerInitiatedInteraction("", testConfig)
		testutil.RequireErrorContains(t, err, "no DID resolver provided")
		require.Nil(t, interaction)
	})
	t.Run("Fail to get credential offer", func(t *testing.T) {
		t.Run("Credential offer query parameter missing", func(t *testing.T) {
			interaction, err := openid4ci.NewIssuerInitiatedInteraction("openid-credential-offer://",
				getTestClientConfig(t))
			require.EqualError(t, err, "INVALID_ISSUANCE_URI(OCI0-0000):credential offer query "+
				"parameter missing from initiate issuance URI")
			require.Nil(t, interaction)
		})
		t.Run("Bad server URL", func(t *testing.T) {
			escapedCredentialOfferURI := url.QueryEscape("BadURL")

			credentialOfferIssuanceURI := "openid-credential-offer://?credential_offer_uri=" + escapedCredentialOfferURI

			interaction, err := openid4ci.NewIssuerInitiatedInteraction(credentialOfferIssuanceURI, getTestClientConfig(t))
			require.EqualError(t, err, "INVALID_CREDENTIAL_OFFER(OCI0-0001):failed to get credential "+
				"offer from the endpoint specified in the credential_offer_uri URL query parameter: "+
				`Get "BadURL": unsupported protocol scheme ""`)
			require.Nil(t, interaction)
		})
		t.Run("Server error", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                                 t,
				credentialOfferEndpointShouldFail: true,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			escapedCredentialOfferURI := url.QueryEscape(server.URL + "/credential-offer")

			credentialOfferIssuanceURI := "openid-credential-offer://?credential_offer_uri=" + escapedCredentialOfferURI

			interaction, err := openid4ci.NewIssuerInitiatedInteraction(credentialOfferIssuanceURI, getTestClientConfig(t))

			require.EqualError(t, err, "INVALID_CREDENTIAL_OFFER(OCI0-0001):failed to get credential offer "+
				"from the endpoint specified in the credential_offer_uri URL query parameter: "+
				"expected status code 200 but got status code 500 "+
				"with response body test failure instead")
			require.Nil(t, interaction)
		})
		t.Run("Fail to unmarshal credential offer", func(t *testing.T) {
			//nolint:gosec // false positive
			credentialOfferIssuanceURI := "openid-credential-offer://?credential_offer="

			interaction, err := openid4ci.NewIssuerInitiatedInteraction(credentialOfferIssuanceURI, getTestClientConfig(t))
			require.EqualError(t, err, "INVALID_CREDENTIAL_OFFER(OCI0-0001):failed to unmarshal "+
				"credential offer JSON into a credential offer object: unexpected end of JSON input")
			require.Nil(t, interaction)
		})
	})
	t.Run("No supported grant types found", func(t *testing.T) {
		credentialOffer := openid4ci.CredentialOffer{}

		credentialOfferBytes, err := json.Marshal(credentialOffer)
		require.NoError(t, err)

		credentialOfferEscaped := url.QueryEscape(string(credentialOfferBytes))

		credentialOfferIssuanceURI := "openid-credential-offer://?credential_offer=" + credentialOfferEscaped

		clientConfig := getTestClientConfig(t)
		clientConfig.HTTPClient = &http.Client{
			Transport: &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBufferString(sampleIssuerMetadata)),
					}, nil
				},
			},
		}

		interaction, err := openid4ci.NewIssuerInitiatedInteraction(credentialOfferIssuanceURI, clientConfig)
		require.EqualError(t, err, "no supported grant types found")
		require.Nil(t, interaction)
	})
	t.Run("Unsupported credential type", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t:                  t,
			credentialResponse: sampleCredentialResponse,
		}

		server := httptest.NewServer(issuerServerHandler)
		defer server.Close()

		issuerServerHandler.issuerMetadata = strings.ReplaceAll(`{
		  "authorization_endpoint": "[SERVER_URL]/oidc/authorize",
		  "notification_endpoint": "[SERVER_URL]/oidc/ack_endpoint",
		  "credential_configurations_supported": {
			"unsupported_configuration_id": {
			  "credential_definition": {
				"credentialSubject": {},
				"type": [
				  "VerifiableCredential",
				  "VerifiedEmployee"
				]
			  },
			  "cryptographic_binding_methods_supported": [
				"ion"
			  ],
			  "credential_signing_alg_values_supported": [
				"ED25519"
			  ],
			  "display": [
				{
				  "background_color": "#12107c",
				  "locale": "en-US",
				  "logo": {
					"alt_text": "a square logo of an employee verification",
					"uri": "https://example.com/public/logo.png"
				  },
				  "name": "Verified Employee",
				  "text_color": "#FFFFFF",
				  "url": ""
				}
			  ],
			  "format": "jwt_vc_json_unsupported",
			  "proof_types_supported": {
				"jwt": {
				  "proof_signing_alg_values_supported": [
					"ED25519"
				  ]
				}
			  }
			}
		  },
		  "credential_endpoint": "[SERVER_URL]/oidc/credential",
		  "credential_issuer": "[SERVER_URL]",
		  "display": [
			{
			  "locale": "en-US",
			  "name": "Bank Issuer",
			  "url": "http://vc-rest-echo.trustbloc.local:8075"
			}
		  ],
		  "grant_types_supported": [
			"authorization_code"
		  ]
		}`, serverURLPlaceholder, server.URL)

		var credentialOffer openid4ci.CredentialOffer

		err := json.Unmarshal(sampleCredentialOffer, &credentialOffer)
		require.NoError(t, err)

		credentialOffer.CredentialIssuer = server.URL
		credentialOffer.CredentialConfigurationIDs = []string{"unsupported_configuration_id"}

		b, err := json.Marshal(credentialOffer)
		require.NoError(t, err)

		credentialOfferEscaped := url.QueryEscape(string(b))

		requestURI := "openid-credential-offer://?credential_offer=" + credentialOfferEscaped
		config := getTestClientConfig(t)

		interaction, err := openid4ci.NewIssuerInitiatedInteraction(requestURI, config)
		require.ErrorContains(t, err, "UNSUPPORTED_CREDENTIAL_TYPE_IN_OFFER")
		require.Nil(t, interaction)
	})
	t.Run("Invalid credential configuration id", func(t *testing.T) {
		credentialOffer := createSampleCredentialOffer(t, false, true)

		credentialOffer.CredentialConfigurationIDs = []string{"invalid_configuration_id"}

		credentialOfferBytes, err := json.Marshal(credentialOffer)
		require.NoError(t, err)

		credentialOfferEscaped := url.QueryEscape(string(credentialOfferBytes))

		credentialOfferIssuanceURI := "openid-credential-offer://?credential_offer=" + credentialOfferEscaped

		clientConfig := getTestClientConfig(t)
		clientConfig.HTTPClient = &http.Client{
			Transport: &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBufferString(sampleIssuerMetadata)),
					}, nil
				},
			},
		}

		interaction, err := openid4ci.NewIssuerInitiatedInteraction(credentialOfferIssuanceURI, clientConfig)
		require.EqualError(t, err, "INVALID_CREDENTIAL_CONFIGURATION_ID(OCI0-0022):invalid credential configuration "+
			"ID (invalid_configuration_id) in credential offer")
		require.Nil(t, interaction)
	})
	t.Run("Fail to log retrieving credential offer via HTTP GET metrics event", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t:                  t,
			credentialResponse: sampleCredentialResponse,
		}

		server := httptest.NewServer(issuerServerHandler)
		defer server.Close()

		issuerServerHandler.credentialOffer = createCredentialOffer(t, server.URL, false, true)

		issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

		escapedCredentialOfferURI := url.QueryEscape(server.URL + "/credential-offer")

		credentialOfferIssuanceURI := "openid-credential-offer://?credential_offer_uri=" + escapedCredentialOfferURI

		config := getTestClientConfig(t)

		config.MetricsLogger = &failingMetricsLogger{}

		interaction, err := openid4ci.NewIssuerInitiatedInteraction(credentialOfferIssuanceURI, config)
		require.Contains(t, err.Error(),
			"failed to log event (Event=Fetch credential offer via an HTTP GET request to "+
				"http://127.0.0.1:")
		require.Nil(t, interaction)
	})
	t.Run("Issuance URL using an unsupported scheme", func(t *testing.T) {
		interaction, err := openid4ci.NewIssuerInitiatedInteraction("https://SomeCredentialOffer",
			getTestClientConfig(t))
		testutil.RequireErrorContains(t, err, "UNSUPPORTED_ISSUANCE_URI_SCHEME")
		testutil.RequireErrorContains(t, err, "https is not a supported issuance URL scheme")
		require.Nil(t, interaction)
	})
}

type mockResolver struct {
	keyWriter           api.KeyWriter
	pubJWK              *jwk.JWK
	linkedDomainsNumber *int
}

func (m *mockResolver) Resolve(string) (*did.DocResolution, error) {
	var services []did.Service

	if m.linkedDomainsNumber == nil {
		one := 1
		m.linkedDomainsNumber = &one
	}

	for range *m.linkedDomainsNumber {
		services = append(services, did.Service{
			ID:              "#LinkedDomains",
			Type:            "LinkedDomains",
			ServiceEndpoint: endpoint.NewDIDCommV1Endpoint("https://demo-issuer.trustbloc.local:8078/"),
		})
	}

	didDoc, err := makeMockDoc(m.keyWriter, m.pubJWK, services)
	if err != nil {
		return nil, err
	}

	return &did.DocResolution{DIDDocument: didDoc}, err
}

// inMemoryMetricsLogger is a simple api.inMemoryMetricsLogger implementation that saves all metrics events in memory.
type inMemoryMetricsLogger struct {
	events []*api.MetricsEvent
}

// newInMemoryMetricsLogger returns a new inMemoryMetricsLogger.
func newInMemoryMetricsLogger() *inMemoryMetricsLogger {
	return &inMemoryMetricsLogger{
		events: make([]*api.MetricsEvent, 0),
	}
}

// Log saves the given metrics events in memory.
func (m *inMemoryMetricsLogger) Log(metricsEvent *api.MetricsEvent) error {
	m.events = append(m.events, metricsEvent)

	return nil
}

func TestIssuerInitiatedInteraction_CreateAuthorizationURL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t:                  t,
			credentialResponse: sampleCredentialResponse,
		}

		server := httptest.NewServer(issuerServerHandler)
		defer server.Close()

		authorizationServerURL := fmt.Sprintf("%s/oidc/authorize", server.URL)

		issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

		t.Run("Not using any options", func(t *testing.T) {
			interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, true, true))

			authorizationURL, err := interaction.CreateAuthorizationURL("clientID", "redirectURI")
			require.NoError(t, err)
			require.Contains(t, authorizationURL, authorizationServerURL+
				"?authorization_details=%5B%7B%22credential_definition%22%3A%7B%22type%22%3A%5B%22VerifiableCredential"+
				"%22%2C%22VerifiedEmployee%22%5D%7D%2C%22format%22%3A%22jwt_vc_json%22%2C%22locations%22%3A%5B%22http%3"+
				"A%2F%2Flocalhost%3A8075%2Fissuer%2Fbank_issuer%2Fv1.0%22%5D%2C%22type%22%3A%22openid_credential%22%7D%"+
				"5D&client_id=clientID")
		})
		t.Run("Using the OAuth Discoverable Client ID Scheme", func(t *testing.T) {
			interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, true, true))

			authorizationURL, err := interaction.CreateAuthorizationURL("clientID", "redirectURI",
				openid4ci.WithOAuthDiscoverableClientIDScheme())
			require.NoError(t, err)
			require.Contains(t, authorizationURL, authorizationServerURL+
				"?authorization_details=%5B%7B%22credential_definition%22%3A%7B%22type%22%3A%5B%22VerifiableCredential"+
				"%22%2C%22VerifiedEmployee%22%5D%7D%2C%22format%22%3A%22jwt_vc_json%22%2C%22locations%22%3A%5B%22http%3"+
				"A%2F%2Flocalhost%3A8075%2Fissuer%2Fbank_issuer%2Fv1.0%22%5D%2C%22type%22%3A%22openid_credential%22%7D%"+
				"5D&client_id=clientID")
		})
		t.Run("Using custom Scopes", func(t *testing.T) {
			interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, true, true))

			authorizationURL, err := interaction.CreateAuthorizationURL("clientID", "redirectURI",
				openid4ci.WithScopes([]string{"custom_scope"}))
			require.NoError(t, err)
			require.Contains(t, authorizationURL, authorizationServerURL+
				"?authorization_details=%5B%7B%22credential_definition%22%3A%7B%22type%22%3A%5B%22VerifiableCredential"+
				"%22%2C%22VerifiedEmployee%22%5D%7D%2C%22format%22%3A%22jwt_vc_json%22%2C%22locations%22%3A%5B%22http%3"+
				"A%2F%2Flocalhost%3A8075%2Fissuer%2Fbank_issuer%2Fv1.0%22%5D%2C%22type%22%3A%22openid_credential%22%7D%"+
				"5D&client_id=clientID")
			require.Contains(t, authorizationURL, "scope=custom_scope")
		})
	})
	t.Run("Issuer does not support the authorization code grant type", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t:                  t,
			credentialResponse: sampleCredentialResponse,
		}

		server := httptest.NewServer(issuerServerHandler)
		defer server.Close()

		issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

		interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, false, false))

		authorizationURL, err := interaction.CreateAuthorizationURL("clientID", "redirectURI")
		require.ErrorContains(t, err, "issuer does not support the authorization code grant type")
		require.Empty(t, authorizationURL)
	})
}

func TestIssuerInitiatedInteraction_RequestCredential(t *testing.T) {
	t.Run("Pre-auth flow", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			t.Run("Using credential_offer", func(t *testing.T) {
				t.Run("Token endpoint defined in the OpenID configuration", func(t *testing.T) {
					issuerServerHandler := &mockIssuerServerHandler{
						t:                  t,
						credentialResponse: sampleCredentialResponse,
					}

					server := httptest.NewServer(issuerServerHandler)
					defer server.Close()

					issuerMetadata := modifyCredentialMetadata(t, sampleIssuerMetadata, func(m *issuer.Metadata) {
						m.NotificationEndpoint = ""
					})

					issuerServerHandler.issuerMetadata = strings.ReplaceAll(issuerMetadata, serverURLPlaceholder, server.URL)

					interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, false, true))

					offeredCredentialsTypes := interaction.OfferedCredentialsTypes()
					require.Len(t, offeredCredentialsTypes, 1)
					require.Contains(t, offeredCredentialsTypes[0], "VerifiedEmployee")

					credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
						keyID: mockKeyID,
					}, openid4ci.WithPIN("1234"))
					require.NoError(t, err)
					require.Len(t, credentials, 1)
					require.NotEmpty(t, credentials[0])

					metadata, err := interaction.IssuerMetadata()
					require.NoError(t, err)
					require.NotNil(t, metadata)

					requireAcknowledgment, err := interaction.RequireAcknowledgment()
					require.NoError(t, err)
					require.False(t, requireAcknowledgment)

					requestedAcknowledgment, err := interaction.Acknowledgment()
					require.Nil(t, requestedAcknowledgment)
					require.ErrorContains(t, err, "issuer not support credential acknowledgement")
				})
				t.Run("Token endpoint defined in the issuer's metadata", func(t *testing.T) {
					issuerServerHandler := &mockIssuerServerHandler{
						t:                  t,
						credentialResponse: sampleCredentialResponse,
						httpStatusCode:     http.StatusOK,
					}

					server := httptest.NewServer(issuerServerHandler)
					defer server.Close()

					issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

					interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, false, true))

					credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
						keyID: mockKeyID,
					}, openid4ci.WithPIN("1234"))
					require.NoError(t, err)
					require.Len(t, credentials, 1)
					require.NotEmpty(t, credentials[0])
				})
				t.Run("Batch credential endpoint", func(t *testing.T) {
					issuerServerHandler := &mockIssuerServerHandler{
						t:                                  t,
						batchCredentialResponse:            sampleCredentialResponseBatch,
						ackRequestExpectInteractionDetails: true,
						ackRequestExpectedCalls:            2,
					}

					server := httptest.NewServer(issuerServerHandler)
					defer server.Close()

					issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

					credentialOffer := createSampleCredentialOffer(t, false, false)
					credentialOffer.CredentialIssuer = server.URL

					credentialOffer.CredentialConfigurationIDs = append(credentialOffer.CredentialConfigurationIDs,
						"credential_configuration_id_1")

					b, err := json.Marshal(credentialOffer)
					require.NoError(t, err)

					credentialOfferURI := "openid-credential-offer://?credential_offer=" + url.QueryEscape(string(b))

					interaction := newIssuerInitiatedInteraction(t, credentialOfferURI)

					credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
						keyID: mockKeyID,
					}, openid4ci.WithPIN("1234"))
					require.NoError(t, err)
					require.Len(t, credentials, 2)
					require.NotEmpty(t, credentials[0])
					require.NotEmpty(t, credentials[1])

					requestedAcknowledgment, err := interaction.Acknowledgment()
					require.NoError(t, err)
					require.NotNil(t, requestedAcknowledgment)

					requestedAcknowledgment.InteractionDetails = map[string]interface{}{"key1": "value1"}

					err = requestedAcknowledgment.AcknowledgeIssuer(openid4ci.EventStatusCredentialAccepted, &http.Client{})
					require.NoError(t, err)

					err = requestedAcknowledgment.AcknowledgeIssuer(openid4ci.EventStatusCredentialAccepted, &http.Client{})
					require.NoError(t, err)

					require.Zero(t, issuerServerHandler.ackRequestExpectedCalls)
					require.Empty(t, requestedAcknowledgment.AckIDs)

					err = requestedAcknowledgment.AcknowledgeIssuer(openid4ci.EventStatusCredentialAccepted, &http.Client{})
					require.ErrorContains(t, err, "ack list is empty")
				})
			})
			t.Run("Issuer require acknowledgment", func(t *testing.T) {
				testCases := []struct {
					caseName string
					reject   bool
				}{
					{
						caseName: "Accepted",
						reject:   false,
					},
					{
						caseName: "Rejected",
						reject:   true,
					},
				}

				issuerServerHandler := &mockIssuerServerHandler{
					t:                                  t,
					credentialResponse:                 sampleCredentialResponseAsk,
					ackRequestExpectInteractionDetails: true,
					ackRequestExpectedCalls:            2,
				}

				server := httptest.NewServer(issuerServerHandler)
				defer server.Close()

				issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

				for _, tc := range testCases {
					interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, false, true))

					_, err := interaction.Acknowledgment()
					require.ErrorContains(t, err, "no acknowledgment data: request credentials first")

					credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
						keyID: mockKeyID,
					}, openid4ci.WithPIN("1234"))
					require.NoError(t, err)
					require.Len(t, credentials, 1)
					require.NotEmpty(t, credentials[0])

					requestedAcknowledgment, err := interaction.Acknowledgment()
					require.NoError(t, err)
					require.NotNil(t, requestedAcknowledgment)

					requestedAcknowledgment.InteractionDetails = map[string]interface{}{"key1": "value1"}

					if !tc.reject {
						err = requestedAcknowledgment.AcknowledgeIssuer(openid4ci.EventStatusCredentialAccepted, &http.Client{})
					} else {
						err = requestedAcknowledgment.AcknowledgeIssuer(openid4ci.EventStatusCredentialFailure, &http.Client{})
					}

					require.NoError(t, err)
				}

				require.Zero(t, issuerServerHandler.ackRequestExpectedCalls)
			})

			t.Run("Using credential_offer_uri", func(t *testing.T) {
				issuerServerHandler := &mockIssuerServerHandler{
					t:                  t,
					credentialResponse: sampleCredentialResponse,
				}

				server := httptest.NewServer(issuerServerHandler)
				defer server.Close()

				issuerServerHandler.credentialOffer = createCredentialOffer(t, server.URL, false, true)

				issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

				escapedCredentialOfferURI := url.QueryEscape(server.URL + "/credential-offer")

				credentialOfferIssuanceURI := "openid-credential-offer://?credential_offer_uri=" + escapedCredentialOfferURI

				config := getTestClientConfig(t)

				metricsLogger := newInMemoryMetricsLogger()

				config.MetricsLogger = metricsLogger

				interaction, err := openid4ci.NewIssuerInitiatedInteraction(credentialOfferIssuanceURI, config)
				require.NoError(t, err)
				require.NotNil(t, interaction)

				// All the other metrics event tests are done in the integration tests already.
				// However, the integration tests don't use the credential_offer_uri, so we have this test here
				// to ensure the metrics event works as expected.
				require.Len(t, metricsLogger.events, 3)
				require.Contains(t, metricsLogger.events[0].Event,
					"Fetch credential offer via an HTTP GET request to")
				require.Equal(t, "Instantiating OpenID4CI interaction object", metricsLogger.events[0].ParentEvent)
				require.Positive(t, metricsLogger.events[0].Duration.Nanoseconds())

				credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
					keyID: mockKeyID,
				}, openid4ci.WithPIN("1234"))
				require.NoError(t, err)
				require.Len(t, credentials, 1)
				require.NotEmpty(t, credentials[0])
			})
			t.Run("Token endpoint defined in the OpenID configuration - request credential HTTP 201",
				func(t *testing.T) {
					issuerServerHandler := &mockIssuerServerHandler{
						t:                  t,
						credentialResponse: sampleCredentialResponse,
						httpStatusCode:     http.StatusCreated,
					}

					server := httptest.NewServer(issuerServerHandler)
					defer server.Close()

					issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

					interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, false, true))

					credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
						keyID: mockKeyID,
					}, openid4ci.WithPIN("1234"))
					require.NoError(t, err)
					require.Len(t, credentials, 1)
					require.NotEmpty(t, credentials[0])
				})
		})
		t.Run("Missing PIN", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                  t,
				credentialResponse: sampleCredentialResponse,
				httpStatusCode:     http.StatusCreated,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, false, true))

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			})
			testutil.RequireErrorContains(t, err,
				"the credential offer requires a user PIN, but none was provided")
			require.Nil(t, credentials)
		})

		t.Run("No linked domains", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                  t,
				credentialResponse: sampleCredentialResponse,
				httpStatusCode:     http.StatusCreated,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			localKMS, err := localkms.NewLocalKMS(localkms.Config{Storage: localkms.NewMemKMSStore()})
			require.NoError(t, err)

			_, publicKey, err := localKMS.Create(arieskms.ED25519Type)
			require.NoError(t, err)

			networkDocumentLoaderHTTPTimeout := time.Second * 10

			zero := 0
			config := &openid4ci.ClientConfig{
				DIDResolver:                      &mockResolver{keyWriter: localKMS, pubJWK: publicKey, linkedDomainsNumber: &zero},
				DisableVCProofChecks:             true,
				NetworkDocumentLoaderHTTPTimeout: &networkDocumentLoaderHTTPTimeout,
			}

			issuerServerHandler.issuerMetadata = createSignedMetadata(t, localKMS, publicKey, server.URL)

			credentialOfferIssuanceURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction, err := openid4ci.NewIssuerInitiatedInteraction(credentialOfferIssuanceURI, config)
			require.NoError(t, err)
			require.NotNil(t, interaction)

			trustInfo, err := interaction.IssuerTrustInfo()
			require.NoError(t, err)
			require.Empty(t, trustInfo.Domain)
		})

		t.Run("multiple linked domains", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                  t,
				credentialResponse: sampleCredentialResponse,
				httpStatusCode:     http.StatusCreated,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			localKMS, err := localkms.NewLocalKMS(localkms.Config{Storage: localkms.NewMemKMSStore()})
			require.NoError(t, err)

			_, publicKey, err := localKMS.Create(arieskms.ED25519Type)
			require.NoError(t, err)

			networkDocumentLoaderHTTPTimeout := time.Second * 10

			zero := 2
			config := &openid4ci.ClientConfig{
				DIDResolver:                      &mockResolver{keyWriter: localKMS, pubJWK: publicKey, linkedDomainsNumber: &zero},
				DisableVCProofChecks:             true,
				NetworkDocumentLoaderHTTPTimeout: &networkDocumentLoaderHTTPTimeout,
			}

			issuerServerHandler.issuerMetadata = createSignedMetadata(t, localKMS, publicKey, server.URL)

			credentialOfferIssuanceURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction, err := openid4ci.NewIssuerInitiatedInteraction(credentialOfferIssuanceURI, config)
			require.NoError(t, err)
			require.NotNil(t, interaction)

			trustInfo, err := interaction.IssuerTrustInfo()
			require.NoError(t, err)
			require.Equal(t, "https://demo-issuer.trustbloc.local:8078", trustInfo.Domain)
		})

		t.Run("attestation VC", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                  t,
				credentialResponse: sampleCredentialResponse,
				httpStatusCode:     http.StatusCreated,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			localKMS, err := localkms.NewLocalKMS(localkms.Config{Storage: localkms.NewMemKMSStore()})
			require.NoError(t, err)

			_, publicKey, err := localKMS.Create(arieskms.ED25519Type)
			require.NoError(t, err)

			networkDocumentLoaderHTTPTimeout := time.Second * 10

			config := &openid4ci.ClientConfig{
				DIDResolver:                      &mockResolver{keyWriter: localKMS, pubJWK: publicKey},
				DisableVCProofChecks:             true,
				NetworkDocumentLoaderHTTPTimeout: &networkDocumentLoaderHTTPTimeout,
			}

			issuerServerHandler.issuerMetadata = createSignedMetadata(t, localKMS, publicKey, server.URL)

			credentialOfferIssuanceURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction, err := openid4ci.NewIssuerInitiatedInteraction(credentialOfferIssuanceURI, config)
			require.NoError(t, err)
			require.NotNil(t, interaction)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"), openid4ci.WithAttestationVC(&jwtSignerMock{
				keyID: mockKeyID,
			}, sampleCredJWT))

			require.NoError(t, err)
			require.NotNil(t, credentials)
		})

		t.Run("Invalid attestation VC", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                  t,
				credentialResponse: sampleCredentialResponse,
				httpStatusCode:     http.StatusCreated,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, false, true))

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"), openid4ci.WithAttestationVC(&jwtSignerMock{
				keyID: mockKeyID,
			}, "{}"))

			testutil.RequireErrorContains(t, err,
				"credential type of unknown structure")
			require.Nil(t, credentials)
		})

		t.Run("No token endpoint available - neither the OpenID configuration nor the issuer's metadata "+
			"specify one", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                  t,
				credentialResponse: sampleCredentialResponse,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerMetadata := strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			issuerServerHandler.issuerMetadata = modifyCredentialMetadata(t, issuerMetadata, func(m *issuer.Metadata) {
				m.TokenEndpoint = ""
			})

			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction := newIssuerInitiatedInteraction(t, requestURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			require.EqualError(t, err, "failed to get credential response: "+
				"NO_TOKEN_ENDPOINT_AVAILABLE(OCI1-0020):no token endpoint specified in issuer's metadata")
			require.Nil(t, credentials)
		})
		t.Run("Fail to reach issuer token endpoint", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t: t,
				issuerMetadata: modifyCredentialMetadata(t, sampleIssuerMetadata, func(m *issuer.Metadata) {
					m.TokenEndpoint = "http://BadURL"
				}),
			}

			server := httptest.NewServer(issuerServerHandler)

			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction := newIssuerInitiatedInteraction(t, requestURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			require.Contains(t, err.Error(), `Post `+
				`"http://BadURL": dial tcp: lookup BadURL`)
			require.Nil(t, credentials)
		})
		t.Run("Fail to get token response: server response body is not an errorResponse "+
			"object", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                      t,
				tokenRequestShouldFail: true,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction := newIssuerInitiatedInteraction(t, requestURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			testutil.RequireErrorContains(t, err, "OTHER_TOKEN_REQUEST_ERROR")
			testutil.RequireErrorContains(t, err, "received status code [500]"+
				" with body [test failure] from issuer's token endpoint")
			require.Nil(t, credentials)
		})

		t.Run("Fail to acknowledge issuer, issuer not support acknowledgment", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                  t,
				credentialResponse: sampleCredentialResponse,
				httpStatusCode:     http.StatusOK,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerMetadata := strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			issuerServerHandler.issuerMetadata = modifyCredentialMetadata(t, issuerMetadata, func(m *issuer.Metadata) {
				m.NotificationEndpoint = ""
			})

			interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, false, true))

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))

			require.NoError(t, err)
			require.Len(t, credentials, 1)
			require.NotEmpty(t, credentials[0])

			_, err = interaction.Acknowledgment()

			require.ErrorContains(t, err, "issuer not support credential acknowledgement")
		})

		t.Run("Fail to acknowledge issuer, status code not 204", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                       t,
				credentialResponse:      sampleCredentialResponse,
				httpStatusCode:          http.StatusInternalServerError,
				ackRequestErrorResponse: "{\"error\":\"expired_ack_id\"}",
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, false, true))

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			require.NoError(t, err)
			require.Len(t, credentials, 1)
			require.NotEmpty(t, credentials[0])

			ack, err := interaction.Acknowledgment()
			require.NoError(t, err)

			err = ack.AcknowledgeIssuer(openid4ci.EventStatusCredentialAccepted, &http.Client{})
			require.ErrorContains(t, err, "ACKNOWLEDGMENT_EXPIRED")
		})

		t.Run("Fail to get token response: invalid token request", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                         t,
				tokenRequestShouldFail:    true,
				tokenRequestErrorResponse: `{"error":"invalid_request"}`,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction := newIssuerInitiatedInteraction(t, requestURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			testutil.RequireErrorContains(t, err, "INVALID_TOKEN_REQUEST")
			testutil.RequireErrorContains(t, err, "received status code [500]"+
				` with body [{"error":"invalid_request"}] from issuer's token endpoint`)
			require.Nil(t, credentials)
		})
		t.Run("Fail to get token response: invalid grant", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                         t,
				tokenRequestShouldFail:    true,
				tokenRequestErrorResponse: `{"error":"invalid_grant"}`,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)
			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction := newIssuerInitiatedInteraction(t, requestURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			testutil.RequireErrorContains(t, err, "INVALID_GRANT")
			testutil.RequireErrorContains(t, err, "received status code [500]"+
				` with body [{"error":"invalid_grant"}] from issuer's token endpoint`)
			require.Nil(t, credentials)
		})
		t.Run("Fail to get token response: invalid client", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                         t,
				tokenRequestShouldFail:    true,
				tokenRequestErrorResponse: `{"error":"invalid_client"}`,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction := newIssuerInitiatedInteraction(t, requestURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			testutil.RequireErrorContains(t, err, "INVALID_CLIENT")
			testutil.RequireErrorContains(t, err, "received status code [500]"+
				` with body [{"error":"invalid_client"}] from issuer's token endpoint`)
			require.Nil(t, credentials)
		})
		t.Run("Fail to get token response: other error code", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                         t,
				tokenRequestShouldFail:    true,
				tokenRequestErrorResponse: `{"error":"someOtherErrorCode"}`,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction := newIssuerInitiatedInteraction(t, requestURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			testutil.RequireErrorContains(t, err, "OTHER_TOKEN_REQUEST_ERROR")
			testutil.RequireErrorContains(t, err, "received status code [500]"+
				` with body [{"error":"someOtherErrorCode"}] from issuer's token endpoint`)
			require.Nil(t, credentials)
		})
		t.Run("Fail to unmarshal response from issuer token endpoint", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t: t,
				tokenRequestShouldGiveUnmarshallableResponse: true,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction := newIssuerInitiatedInteraction(t, requestURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			testutil.RequireErrorContains(t, err, "failed to get token response: failed to unmarshal response from the "+
				"issuer's token endpoint: invalid character 'i' looking for beginning of value")
			require.Nil(t, credentials)
		})
		t.Run("Fail to get credential response: server response body is not an errorResponse "+
			"object", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{t: t, credentialRequestShouldFail: true}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction := newIssuerInitiatedInteraction(t, requestURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			testutil.RequireErrorContains(t, err, "received status code [500] "+
				"with body [test failure] from issuer's credential endpoint")
			require.Nil(t, credentials)
		})
		t.Run("Fail to get credential response: invalid request", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t: t, credentialRequestShouldFail: true,
				credentialRequestErrorResponse: `{"error":"invalid_request"}`,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction := newIssuerInitiatedInteraction(t, requestURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			testutil.RequireErrorContains(t, err, "INVALID_CREDENTIAL_REQUEST")
			testutil.RequireErrorContains(t, err, "received status code [500] "+
				`with body [{"error":"invalid_request"}] from issuer's credential endpoint`)
			require.Nil(t, credentials)
		})
		t.Run("Fail to get credential response: invalid proof error ", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{t: t, credentialRequestShouldGiveInvalidProofResponse: true}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction := newIssuerInitiatedInteraction(t, requestURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			testutil.RequireErrorContains(t, err, "received status code [500] "+
				`with body [{"error":"invalid_proof"}] from issuer's credential endpoint`)
			require.Nil(t, credentials)
		})
		t.Run("Fail to get credential response: invalid token", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t: t, credentialRequestShouldFail: true,
				credentialRequestErrorResponse: `{"error":"invalid_token"}`,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction := newIssuerInitiatedInteraction(t, requestURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			testutil.RequireErrorContains(t, err, "INVALID_TOKEN")
			testutil.RequireErrorContains(t, err, "received status code [500] "+
				`with body [{"error":"invalid_token"}] from issuer's credential endpoint`)
			require.Nil(t, credentials)
		})
		t.Run("Fail to get credential response: unsupported credential format", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t: t, credentialRequestShouldFail: true,
				credentialRequestErrorResponse: `{"error":"unsupported_credential_format"}`,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction := newIssuerInitiatedInteraction(t, requestURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			testutil.RequireErrorContains(t, err, "UNSUPPORTED_CREDENTIAL_FORMAT")
			testutil.RequireErrorContains(t, err, "received status code [500] "+
				`with body [{"error":"unsupported_credential_format"}] from issuer's credential endpoint`)
			require.Nil(t, credentials)
		})
		t.Run("Fail to get credential response: unsupported credential type", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t: t, credentialRequestShouldFail: true,
				credentialRequestErrorResponse: `{"error":"unsupported_credential_type"}`,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction := newIssuerInitiatedInteraction(t, requestURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			testutil.RequireErrorContains(t, err, "UNSUPPORTED_CREDENTIAL_TYPE")
			testutil.RequireErrorContains(t, err, "received status code [500] "+
				`with body [{"error":"unsupported_credential_type"}] from issuer's credential endpoint`)
			require.Nil(t, credentials)
		})
		t.Run("Fail to get credential response: invalid or missing proof", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t: t, credentialRequestShouldFail: true,
				credentialRequestErrorResponse: `{"error":"invalid_or_missing_proof"}`,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction := newIssuerInitiatedInteraction(t, requestURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			testutil.RequireErrorContains(t, err, "INVALID_OR_MISSING_PROOF")
			testutil.RequireErrorContains(t, err, "received status code [500] "+
				`with body [{"error":"invalid_or_missing_proof"}] from issuer's credential endpoint`)
			require.Nil(t, credentials)
		})
		t.Run("Fail to get credential response: other error code", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t: t, credentialRequestShouldFail: true,
				credentialRequestErrorResponse: `{"error":"someOtherErrorCode"}`,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction := newIssuerInitiatedInteraction(t, requestURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			testutil.RequireErrorContains(t, err, "OTHER_CREDENTIAL_REQUEST_ERROR")
			testutil.RequireErrorContains(t, err, "received status code [500] "+
				`with body [{"error":"someOtherErrorCode"}] from issuer's credential endpoint`)
			require.Nil(t, credentials)
		})
		t.Run("Fail to get credential response: signature error", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{t: t, credentialRequestShouldFail: true}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction := newIssuerInitiatedInteraction(t, requestURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
				Err:   errors.New("signature error"),
			}, openid4ci.WithPIN("1234"))
			testutil.RequireErrorContains(t, err, "JWT_SIGNING_FAILED")
			require.Nil(t, credentials)
		})
		t.Run("Fail to reach issuer's credential endpoint", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{t: t}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerMetadata := strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, "http://BadURL")

			issuerServerHandler.issuerMetadata = modifyCredentialMetadata(t, issuerMetadata, func(m *issuer.Metadata) {
				m.TokenEndpoint = server.URL + "/oidc/token"
			})

			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction := newIssuerInitiatedInteraction(t, requestURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))

			require.Contains(t, err.Error(), `Post "http://BadURL/oidc/credential": dial tcp: lookup BadURL`)
			require.Nil(t, credentials)
		})
		t.Run("Fail to get credential response: KID does not contain the DID part", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{t: t}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction := newIssuerInitiatedInteraction(t, requestURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: "did:example:12345",
			}, openid4ci.WithPIN("1234"))

			testutil.RequireErrorContains(t, err, "KEY_ID_MISSING_DID_PART")
			require.Nil(t, credentials)
		})
		t.Run("Fail to unmarshal response from issuer credential endpoint", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{t: t, credentialRequestShouldGiveUnmarshallableResponse: true}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction := newIssuerInitiatedInteraction(t, requestURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			testutil.RequireErrorContains(t, err, "failed to unmarshal response "+
				"from the issuer's credential endpoint: invalid character 'i' looking for beginning of value")
			require.Nil(t, credentials)
		})
		t.Run("Fail to parse VC", func(t *testing.T) {
			var credentialResponse openid4ci.CredentialResponse

			credentialResponse.Credential = ""

			credentialResponseBytes, err := json.Marshal(credentialResponse)
			require.NoError(t, err)

			issuerServerHandler := &mockIssuerServerHandler{t: t, credentialResponse: credentialResponseBytes}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction := newIssuerInitiatedInteraction(t, requestURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))

			require.Contains(t, err.Error(), "CREDENTIAL_PARSE_FAILED(OCI1-0007):failed to parse credential from "+
				"credential response at index 0: unmarshal cbor cred after hex failed\nunmarshal cbor credential: EOF")
			require.Nil(t, credentials)
		})
		t.Run("Fail VC proof check - public key not found for issuer DID", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{t: t, credentialResponse: sampleCredentialResponse}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			localKMS, err := localkms.NewLocalKMS(localkms.Config{Storage: localkms.NewMemKMSStore()})
			require.NoError(t, err)

			didResolver := &mockResolver{keyWriter: localKMS}

			config := &openid4ci.ClientConfig{
				DIDResolver: didResolver,
			}

			interaction, err := openid4ci.NewIssuerInitiatedInteraction(requestURI, config)
			require.NoError(t, err)
			require.NotNil(t, interaction)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			require.ErrorContains(t, err, "CREDENTIAL_PARSE_FAILED(OCI1-0007):failed to parse credential from "+
				"credential response at index 0: "+
				"JWS proof check: invalid public key id: public key with KID ")
			require.Nil(t, credentials)
		})
		t.Run("Fail to log fetch token via HTTP POST metrics event", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t: t,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			config := getTestClientConfig(t)
			config.MetricsLogger = &failingMetricsLogger{attemptFailNumber: 2}

			interaction, err := openid4ci.NewIssuerInitiatedInteraction(
				createCredentialOfferIssuanceURI(t, server.URL, false, true), config)
			require.NoError(t, err)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			require.Contains(t, err.Error(),
				"failed to log event (Event=Fetch token via an HTTP POST request to http://127.0.0.1:")
			require.Nil(t, credentials)
		})
		t.Run("Fail to log fetch credential via HTTP GET metrics event", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                  t,
				credentialResponse: sampleCredentialResponse,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			config := getTestClientConfig(t)
			config.MetricsLogger = &failingMetricsLogger{attemptFailNumber: 3}

			interaction, err := openid4ci.NewIssuerInitiatedInteraction(
				createCredentialOfferIssuanceURI(t, server.URL, false, true), config)
			require.NoError(t, err)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			require.Contains(t, err.Error(), "failed to get credential "+
				"response: failed to log event (Event=Fetch credential 1 of 1 via an HTTP POST request to "+
				"http://127.0.0.1:")
			require.Nil(t, credentials)
		})
		t.Run("Fail to get credential response when using batch endpoint", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                                t,
				batchCredentialRequestShouldFail: true,
				httpStatusCode:                   http.StatusOK,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			credentialOffer := createSampleCredentialOffer(t, false, false)
			credentialOffer.CredentialIssuer = server.URL

			credentialOffer.CredentialConfigurationIDs = append(credentialOffer.CredentialConfigurationIDs,
				"credential_configuration_id_1")

			b, err := json.Marshal(credentialOffer)
			require.NoError(t, err)

			credentialOfferURI := "openid-credential-offer://?credential_offer=" + url.QueryEscape(string(b))

			interaction := newIssuerInitiatedInteraction(t, credentialOfferURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			require.Contains(t, err.Error(), "failed to get credential response")
			require.Nil(t, credentials)
		})
		t.Run("Fail to unmarshal credential response from batch credential endpoint", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                       t,
				batchCredentialResponse: []byte("invalid response"),
				httpStatusCode:          http.StatusOK,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			credentialOffer := createSampleCredentialOffer(t, false, false)
			credentialOffer.CredentialIssuer = server.URL

			credentialOffer.CredentialConfigurationIDs = append(credentialOffer.CredentialConfigurationIDs,
				"credential_configuration_id_1")

			b, err := json.Marshal(credentialOffer)
			require.NoError(t, err)

			credentialOfferURI := "openid-credential-offer://?credential_offer=" + url.QueryEscape(string(b))

			interaction := newIssuerInitiatedInteraction(t, credentialOfferURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			require.Contains(t, err.Error(), "failed to get credential response")
			require.Nil(t, credentials)
		})
		t.Run("Fail batch credential endpoint", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                                  t,
				batchCredentialResponse:            sampleCredentialResponseBatch,
				ackRequestExpectInteractionDetails: true,
				ackRequestExpectedCalls:            2,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			credentialOffer := createSampleCredentialOffer(t, false, false)
			credentialOffer.CredentialIssuer = server.URL

			credentialOffer.CredentialConfigurationIDs = append(credentialOffer.CredentialConfigurationIDs,
				"credential_configuration_id_1")

			b, err := json.Marshal(credentialOffer)
			require.NoError(t, err)

			credentialOfferURI := "openid-credential-offer://?credential_offer=" + url.QueryEscape(string(b))

			interaction := newIssuerInitiatedInteraction(t, credentialOfferURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			require.NoError(t, err)
			require.Len(t, credentials, 2)
			require.NotEmpty(t, credentials[0])
			require.NotEmpty(t, credentials[1])

			requestedAcknowledgment, err := interaction.Acknowledgment()
			require.NoError(t, err)
			require.NotNil(t, requestedAcknowledgment)

			requestedAcknowledgment.InteractionDetails = map[string]any{"key1": "value1"}
			err = requestedAcknowledgment.AcknowledgeIssuer(openid4ci.EventStatusCredentialAccepted, &http.Client{
				Transport: &mockHTTPClient{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("bad_request")
					},
				},
			})
			require.Error(t, err)
			require.ErrorContains(t, err, "send acknowledge request id ack_id1")

			requestedAcknowledgment.InteractionDetails = map[string]any{"key1": "value1", "key2": func() {}}

			err = requestedAcknowledgment.AcknowledgeIssuer(openid4ci.EventStatusCredentialAccepted, &http.Client{})
			require.Error(t, err)
			require.ErrorContains(t, err, "fail to marshal acknowledgementRequest")
		})
	})
	t.Run("Auth flow", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			t.Run("Token endpoint defined in the OpenID configuration", func(t *testing.T) {
				t.Run("Issuer state specified in credential offer", func(t *testing.T) {
					issuerServerHandler := &mockIssuerServerHandler{
						t:                  t,
						credentialResponse: sampleCredentialResponse,
					}

					server := httptest.NewServer(issuerServerHandler)
					defer server.Close()

					issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

					interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, true, true))

					// Needed to create the OAuth2 config object.
					authURL, err := interaction.CreateAuthorizationURL("clientID", "redirectURI")
					require.NoError(t, err)

					redirectURIWithParams := "redirectURI?code=1234&state=" + getStateFromAuthURL(t, authURL)

					credentials, err := interaction.RequestCredentialWithAuth(&jwtSignerMock{
						keyID: mockKeyID,
					}, redirectURIWithParams)
					require.NoError(t, err)
					require.Len(t, credentials, 1)
					require.NotEmpty(t, credentials[0])
				})
				t.Run("Issuer state not specified in credential offer", func(t *testing.T) {
					issuerServerHandler := &mockIssuerServerHandler{
						t:                  t,
						credentialResponse: sampleCredentialResponse,
					}

					server := httptest.NewServer(issuerServerHandler)
					defer server.Close()

					issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

					interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, true, false))

					// Needed to create the OAuth2 config object.
					authURL, err := interaction.CreateAuthorizationURL("clientID", "redirectURI")
					require.NoError(t, err)

					redirectURIWithParams := "redirectURI?code=1234&state=" + getStateFromAuthURL(t, authURL)

					credentials, err := interaction.RequestCredentialWithAuth(&jwtSignerMock{
						keyID: mockKeyID,
					}, redirectURIWithParams)
					require.NoError(t, err)
					require.Len(t, credentials, 1)
					require.NotEmpty(t, credentials[0])
				})
				t.Run("Issuer state not specified in credential offer, but is specified by caller in "+
					"CreateAuthorizationURL call", func(t *testing.T) {
					issuerServerHandler := &mockIssuerServerHandler{
						t:                  t,
						credentialResponse: sampleCredentialResponse,
					}

					server := httptest.NewServer(issuerServerHandler)
					defer server.Close()

					issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

					interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, true, false))

					// Needed to create the OAuth2 config object.
					authURL, err := interaction.CreateAuthorizationURL("clientID", "redirectURI",
						openid4ci.WithIssuerState("IssuerState"))
					require.NoError(t, err)

					redirectURIWithParams := "redirectURI?code=1234&state=" + getStateFromAuthURL(t, authURL)

					credentials, err := interaction.RequestCredentialWithAuth(&jwtSignerMock{
						keyID: mockKeyID,
					}, redirectURIWithParams)
					require.NoError(t, err)
					require.Len(t, credentials, 1)
					require.NotEmpty(t, credentials[0])
				})
			})
			t.Run("Token endpoint defined in the OpenID configuration", func(t *testing.T) {
				issuerServerHandler := &mockIssuerServerHandler{
					t:                  t,
					credentialResponse: sampleCredentialResponse,
				}

				server := httptest.NewServer(issuerServerHandler)
				defer server.Close()

				issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

				interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, true, true))

				// Needed to create the OAuth2 config object.
				authURL, err := interaction.CreateAuthorizationURL("clientID", "redirectURI")
				require.NoError(t, err)

				redirectURIWithParams := "redirectURI?code=1234&state=" + getStateFromAuthURL(t, authURL)

				credentials, err := interaction.RequestCredentialWithAuth(&jwtSignerMock{
					keyID: mockKeyID,
				}, redirectURIWithParams)
				require.NoError(t, err)
				require.Len(t, credentials, 1)
				require.NotEmpty(t, credentials[0])
			})

			t.Run("Issuer state specified in credential offer", func(t *testing.T) {
				issuerServerHandler := &mockIssuerServerHandler{
					t:                  t,
					credentialResponse: sampleCredentialResponse,
				}

				server := httptest.NewServer(issuerServerHandler)
				defer server.Close()

				issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

				interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, true, true))

				// Needed to create the OAuth2 config object.
				authURL, err := interaction.CreateAuthorizationURL("clientID", "redirectURI")
				require.NoError(t, err)

				redirectURIWithParams := "redirectURI?code=1234&state=" + getStateFromAuthURL(t, authURL)

				credentials, err := interaction.RequestCredentialWithAuth(&jwtSignerMock{
					keyID: mockKeyID,
				}, redirectURIWithParams)
				require.NoError(t, err)
				require.Len(t, credentials, 1)
				require.NotEmpty(t, credentials[0])

				requestedAcknowledgment, err := interaction.Acknowledgment()
				require.NoError(t, err)

				require.NoError(t, requestedAcknowledgment.AcknowledgeIssuer(
					openid4ci.EventStatusCredentialAccepted, &http.Client{}))
			})
		})
		t.Run("Issuer does not support the authorization code grant type", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                  t,
				credentialResponse: sampleCredentialResponse,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, false, true))

			credentials, err := interaction.RequestCredentialWithAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, "")
			require.ErrorContains(t, err, "issuer does not support the authorization code grant type")
			require.Nil(t, credentials)
		})
		t.Run("Authorization URL not created first", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                  t,
				credentialResponse: sampleCredentialResponse,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, true, true))

			credentials, err := interaction.RequestCredentialWithAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, "")
			require.EqualError(t, err, "authorization URL must be created first")
			require.Nil(t, credentials)
		})
		t.Run("Redirect URI missing authorization code", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                  t,
				credentialResponse: sampleCredentialResponse,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, true, true))

			_, err := interaction.CreateAuthorizationURL("clientID", "redirectURI")
			require.NoError(t, err)

			credentials, err := interaction.RequestCredentialWithAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, "redirectURI?state=1234")
			require.EqualError(t, err, "redirect URI is missing an authorization code")
			require.Nil(t, credentials)
		})
		t.Run("Redirect URI missing state", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                  t,
				credentialResponse: sampleCredentialResponse,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, true, true))

			_, err := interaction.CreateAuthorizationURL("clientID", "redirectURI")
			require.NoError(t, err)

			credentials, err := interaction.RequestCredentialWithAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, "redirectURI?code=1234")
			require.EqualError(t, err, "redirect URI is missing a state value")
			require.Nil(t, credentials)
		})
		t.Run("Fail to parse redirect URI", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                  t,
				credentialResponse: sampleCredentialResponse,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, true, true))

			_, err := interaction.CreateAuthorizationURL("clientID", "redirectURI")
			require.NoError(t, err)

			credentials, err := interaction.RequestCredentialWithAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, "%")
			require.EqualError(t, err, `parse "%": invalid URL escape "%"`)
			require.Nil(t, credentials)
		})
		t.Run("Fail to create claims proof", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                  t,
				credentialResponse: sampleCredentialResponse,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

			interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, true, true))

			// Needed to create the OAuth2 config object.
			authURL, err := interaction.CreateAuthorizationURL("clientID", "redirectURI")
			require.NoError(t, err)

			redirectURIWithParams := "redirectURI?code=1234&state=" + getStateFromAuthURL(t, authURL)

			credentials, err := interaction.RequestCredentialWithAuth(&jwtSignerMock{
				keyID: mockKeyID,
				Err:   errors.New("test failure"),
			}, redirectURIWithParams)
			require.EqualError(t, err, "failed to get credential "+
				"response: JWT_SIGNING_FAILED(OCI1-0005):failed to create JWT: sign token failed: create "+
				"JWS: sign JWS: sign JWS verification data: test failure")
			require.Nil(t, credentials)
		})
	})
	t.Run("State in redirect URI does not match the state from the auth URL", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t: t,
		}

		server := httptest.NewServer(issuerServerHandler)
		defer server.Close()

		issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

		interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, true, true))

		// Needed to create the OAuth2 config object.
		_, err := interaction.CreateAuthorizationURL("clientID", "redirectURI")
		require.NoError(t, err)

		redirectURIWithParams := "redirectURI?code=1234&state=DoesNotMatch"

		credentials, err := interaction.RequestCredentialWithAuth(&jwtSignerMock{
			keyID: mockKeyID,
		}, redirectURIWithParams)
		require.EqualError(t, err, "STATE_IN_REDIRECT_URI_NOT_MATCHING_AUTH_URL(OCI1-0008):state in "+
			"redirect URI does not match the state from the authorization URL")
		require.Nil(t, credentials)
	})
	t.Run("Conflicting issuer state", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t: t,
		}

		server := httptest.NewServer(issuerServerHandler)
		defer server.Close()

		issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

		interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, true, true))

		authorizationURL, err := interaction.CreateAuthorizationURL("clientID", "redirectURI",
			openid4ci.WithIssuerState("SomeOtherState"))
		require.EqualError(t, err, "INVALID_SDK_USAGE(OCI3-0000):the credential offer already specifies "+
			"an issuer state, and a conflicting issuer state value was provided. An issuer state should only be "+
			"provided if required by the issuer and the credential offer does not specify one already")
		require.Empty(t, authorizationURL)
	})
}

func TestIssuerInitiatedInteraction_RequestCredential_NoProofFound(t *testing.T) {
	issuerServerHandler := &mockIssuerServerHandler{
		t:                  t,
		credentialResponse: sampleCredentialResponseJSONLD,
		httpStatusCode:     http.StatusOK,
	}

	server := httptest.NewServer(issuerServerHandler)
	defer server.Close()

	issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

	interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, false, true),
		enableVCProofChecks())

	credentials, err := interaction.RequestCredentialWithPreAuth(
		&jwtSignerMock{
			keyID: mockKeyID,
		},
		openid4ci.WithPIN("1234"),
	)

	require.ErrorContains(t, err, "proof not found")
	require.Empty(t, credentials)
}

func TestIssuerInitiatedInteraction_GrantTypes(t *testing.T) {
	issuerServerHandler := &mockIssuerServerHandler{
		t:                  t,
		credentialResponse: sampleCredentialResponse,
	}

	server := httptest.NewServer(issuerServerHandler)
	defer server.Close()

	issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

	requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)
	interaction := newIssuerInitiatedInteraction(t, requestURI)
	require.NotNil(t, interaction)

	require.True(t, interaction.PreAuthorizedCodeGrantTypeSupported())

	preAuthorizedCodeGrantParams, err := interaction.PreAuthorizedCodeGrantParams()
	require.NoError(t, err)
	require.NotNil(t, preAuthorizedCodeGrantParams)

	require.True(t, preAuthorizedCodeGrantParams.PINRequired())

	require.False(t, interaction.AuthorizationCodeGrantTypeSupported())

	authorizationCodeGrantParams, err := interaction.AuthorizationCodeGrantParams()
	require.EqualError(t, err,
		"INVALID_SDK_USAGE(OCI3-0000):issuer does not support the authorization code grant")
	require.Nil(t, authorizationCodeGrantParams)

	requestURI = createCredentialOfferIssuanceURI(t, server.URL, true, true)
	interaction = newIssuerInitiatedInteraction(t, requestURI)

	require.True(t, interaction.AuthorizationCodeGrantTypeSupported())

	authorizationCodeGrantParams, err = interaction.AuthorizationCodeGrantParams()
	require.NoError(t, err)
	require.NotNil(t, authorizationCodeGrantParams)

	require.NotNil(t, authorizationCodeGrantParams.IssuerState)
	require.Equal(t, "1234", *authorizationCodeGrantParams.IssuerState)
}

func TestIssuerInitiatedInteraction_DynamicClientRegistration(t *testing.T) {
	t.Run("Dynamic client registration is not supported", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t: t,
		}

		server := httptest.NewServer(issuerServerHandler)
		defer server.Close()

		issuerMetadata := strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

		issuerServerHandler.issuerMetadata = modifyCredentialMetadata(t, issuerMetadata, func(m *issuer.Metadata) {
			m.RegistrationEndpoint = nil
		})

		interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, true, true))

		supported, err := interaction.DynamicClientRegistrationSupported()
		require.NoError(t, err)
		require.False(t, supported)

		endpointResolved, err := interaction.DynamicClientRegistrationEndpoint()
		require.EqualError(t, err,
			"INVALID_SDK_USAGE(OCI3-0000):issuer does not support dynamic client registration")
		require.Empty(t, endpointResolved)
	})
	t.Run("Dynamic client registration is supported", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t: t,
		}

		server := httptest.NewServer(issuerServerHandler)
		defer server.Close()

		testEndpoint := "SomeEndpoint"

		issuerMetadata := strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

		issuerServerHandler.issuerMetadata = modifyCredentialMetadata(t, issuerMetadata, func(m *issuer.Metadata) {
			m.RegistrationEndpoint = &testEndpoint
		})

		interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, true, true))

		supported, err := interaction.DynamicClientRegistrationSupported()
		require.NoError(t, err)
		require.True(t, supported)

		endpointResolved, err := interaction.DynamicClientRegistrationEndpoint()
		require.NoError(t, err)
		require.Equal(t, testEndpoint, endpointResolved)
	})
}

func TestIssuerInitiatedInteraction_IssuerURI(t *testing.T) {
	issuerServerHandler := &mockIssuerServerHandler{
		t:                  t,
		credentialResponse: sampleCredentialResponse,
	}

	server := httptest.NewServer(issuerServerHandler)
	defer server.Close()

	issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

	interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, false, true))

	require.Equal(t, server.URL, interaction.IssuerURI())
}

func TestIssuerInitiatedInteraction_VerifyIssuer(t *testing.T) {
	t.Run("Resolved DID document has no Linked Domains services specified", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t: t,
		}

		server := httptest.NewServer(issuerServerHandler)
		defer server.Close()

		issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

		config := getTestClientConfig(t)

		didResolver, err := resolver.NewDIDResolver()
		require.NoError(t, err)

		config.DIDResolver = didResolver

		credentialOfferIssuanceURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

		interaction, err := openid4ci.NewIssuerInitiatedInteraction(credentialOfferIssuanceURI, config)
		require.NoError(t, err)
		require.NotNil(t, interaction)

		serviceURL, err := interaction.VerifyIssuer()
		require.ErrorContains(t, err, "DOMAIN_AND_DID_VERIFICATION_FAILED")
		require.Empty(t, serviceURL)
	})
}

func TestIssuerInitiatedInteraction_IssuerTrustInfo(t *testing.T) {
	issuerServerHandler := &mockIssuerServerHandler{
		t: t,
	}

	server := httptest.NewServer(issuerServerHandler)
	defer server.Close()

	t.Run("Success", func(t *testing.T) {
		t.Run("Signed metadata", func(t *testing.T) {
			localKMS, err := localkms.NewLocalKMS(localkms.Config{Storage: localkms.NewMemKMSStore()})
			require.NoError(t, err)

			_, publicKey, err := localKMS.Create(arieskms.ED25519Type)
			require.NoError(t, err)

			networkDocumentLoaderHTTPTimeout := time.Second * 10

			config := &openid4ci.ClientConfig{
				DIDResolver:                      &mockResolver{keyWriter: localKMS, pubJWK: publicKey},
				DisableVCProofChecks:             true,
				NetworkDocumentLoaderHTTPTimeout: &networkDocumentLoaderHTTPTimeout,
			}

			issuerServerHandler.issuerMetadata = createSignedMetadata(t, localKMS, publicKey, server.URL)

			credentialOfferIssuanceURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction, err := openid4ci.NewIssuerInitiatedInteraction(credentialOfferIssuanceURI, config)
			require.NoError(t, err)
			require.NotNil(t, interaction)

			trustInfo, err := interaction.IssuerTrustInfo()
			require.NoError(t, err)
			require.NotNil(t, trustInfo)
			require.Contains(t, trustInfo.Domain, "trustbloc.local")
		})

		t.Run("Origin-based trust", func(t *testing.T) {
			config := &openid4ci.ClientConfig{
				DIDResolver: &mockResolver{},
			}

			issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder,
				server.URL)

			credentialOfferIssuanceURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction, err := openid4ci.NewIssuerInitiatedInteraction(credentialOfferIssuanceURI, config)
			require.NoError(t, err)
			require.NotNil(t, interaction)

			serverURL, err := url.Parse(server.URL)
			require.NoError(t, err)

			trustInfo, err := interaction.IssuerTrustInfo()
			require.NoError(t, err)
			require.NotNil(t, trustInfo)
			require.Equal(t, serverURL.Host, trustInfo.Domain)
		})
	})
}

type clientConfigOpt func(*openid4ci.ClientConfig)

func enableVCProofChecks() clientConfigOpt {
	return func(config *openid4ci.ClientConfig) {
		config.DisableVCProofChecks = false
	}
}

func newIssuerInitiatedInteraction(t *testing.T, requestURI string,
	opts ...clientConfigOpt,
) *openid4ci.IssuerInitiatedInteraction {
	t.Helper()

	config := getTestClientConfig(t)

	for _, opt := range opts {
		opt(config)
	}

	interaction, err := openid4ci.NewIssuerInitiatedInteraction(requestURI, config)
	require.NoError(t, err)
	require.NotNil(t, interaction)

	return interaction
}

func getTestClientConfig(t *testing.T) *openid4ci.ClientConfig {
	t.Helper()

	localKMS, err := localkms.NewLocalKMS(localkms.Config{Storage: localkms.NewMemKMSStore()})
	require.NoError(t, err)

	didResolver := &mockResolver{keyWriter: localKMS}

	networkDocumentLoaderHTTPTimeout := time.Second * 10

	return &openid4ci.ClientConfig{
		DIDResolver:                      didResolver,
		DisableVCProofChecks:             true,
		NetworkDocumentLoaderHTTPTimeout: &networkDocumentLoaderHTTPTimeout,
	}
}

func createSignedMetadata(t *testing.T, localKMS *localkms.LocalKMS, publicKey *jwk.JWK, serverURL string) string {
	t.Helper()

	didResolver := &mockResolver{keyWriter: localKMS, pubJWK: publicKey}

	didDocResolution, err := didResolver.Resolve("")
	require.NoError(t, err)

	verificationMethod := didDocResolution.DIDDocument.VerificationMethod[0]

	signer, err := common.NewJWSSigner(models.VerificationMethodFromDoc(&verificationMethod), localKMS.GetCrypto())
	require.NoError(t, err)

	claims := map[string]interface{}{}

	err = json.Unmarshal([]byte(strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, serverURL)), &claims)
	require.NoError(t, err)

	token, err := jwt.NewSigned(claims, jwt.SignParameters{
		KeyID:             publicKey.KeyID,
		JWTAlg:            "",
		AdditionalHeaders: nil,
	}, signer)
	require.NoError(t, err)

	tokenSerialised, err := token.Serialize(false)
	require.NoError(t, err)

	return fmt.Sprintf(`{"signed_metadata": %q}`, tokenSerialised)
}

type mockTransport struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

// makeMockDoc creates a key in the given KMS and returns a mock DID Doc with a verification method.
func makeMockDoc(keyWriter api.KeyWriter, pubJWK *jwk.JWK, services []did.Service) (*did.Doc, error) {
	if pubJWK == nil {
		var err error

		_, pubJWK, err = keyWriter.Create(arieskms.ED25519Type)
		if err != nil {
			return nil, err
		}
	}

	pkb, err := pubJWK.PublicKeyBytes()
	if err != nil {
		return nil, err
	}

	vm := &did.VerificationMethod{
		ID:         pubJWK.KeyID,
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
		Service: services,
	}

	return newDoc, nil
}

type jwtSignerMock struct {
	keyID string
	Err   error
}

func (s *jwtSignerMock) GetKeyID() string {
	return s.keyID
}

func (s *jwtSignerMock) SignJWT(_ jwt.SignParameters, _ []byte) ([]byte, error) {
	return []byte("test signature"), s.Err
}

func (s *jwtSignerMock) CreateJWTHeaders(_ jwt.SignParameters) (jose.Headers, error) {
	return jose.Headers{
		jose.HeaderKeyID:     "KeyID",
		jose.HeaderAlgorithm: "ES384",
	}, nil
}

// includeIssuerStateParam only applies if includeAuthCodeGrant is true.
func createCredentialOfferIssuanceURI(t *testing.T, issuerURL string, includeAuthCodeGrant,
	includeIssuerStateParam bool,
) string {
	t.Helper()

	credentialOffer := createCredentialOffer(t, issuerURL, includeAuthCodeGrant, includeIssuerStateParam)

	credentialOfferBytes, err := json.Marshal(credentialOffer)
	require.NoError(t, err)

	credentialOfferEscaped := url.QueryEscape(string(credentialOfferBytes))

	return "openid-credential-offer://?credential_offer=" + credentialOfferEscaped
}

func createCredentialOffer(t *testing.T, issuerURL string, includeAuthCodeGrant,
	includeIssuerStateParam bool,
) *openid4ci.CredentialOffer {
	t.Helper()

	credentialOffer := createSampleCredentialOffer(t, includeAuthCodeGrant, includeIssuerStateParam)

	credentialOffer.CredentialIssuer = issuerURL

	return credentialOffer
}

func createSampleCredentialOffer(t *testing.T, includeAuthCodeGrant,
	includeIssuerStateParam bool,
) *openid4ci.CredentialOffer {
	t.Helper()

	var credentialOffer openid4ci.CredentialOffer

	err := json.Unmarshal(sampleCredentialOffer, &credentialOffer)
	require.NoError(t, err)

	if includeAuthCodeGrant {
		authCodeGrant := map[string]interface{}{}

		if includeIssuerStateParam {
			authCodeGrant["issuer_state"] = "1234"
		}

		credentialOffer.Grants["authorization_code"] = authCodeGrant
	}

	return &credentialOffer
}

func getStateFromAuthURL(t *testing.T, authURL string) string {
	t.Helper()

	parsedURI, err := url.Parse(authURL)
	require.NoError(t, err)

	return parsedURI.Query().Get("state")
}

func modifyCredentialMetadata(t *testing.T, metadata string, modifyFunc func(m *issuer.Metadata)) string {
	t.Helper()

	var m *issuer.Metadata
	err := json.Unmarshal([]byte(metadata), &m)
	require.NoError(t, err)

	modifyFunc(m)

	b, err := json.Marshal(m)
	require.NoError(t, err)

	return string(b)
}

type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}
