/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci_test

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/did-go/doc/did"
	"github.com/trustbloc/kms-go/doc/jose"
	arieskms "github.com/trustbloc/kms-go/spi/kms"

	"github.com/trustbloc/wallet-sdk/internal/testutil"
	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/localkms"
	"github.com/trustbloc/wallet-sdk/pkg/openid4ci"
)

const (
	sampleTokenResponse = `{"access_token":"eyJhbGciOiJSUzI1NiIsInR5cCI6Ikp..sHQ",` +
		`"token_type":"bearer","expires_in":86400,"c_nonce":"tZignsnFbp","c_nonce_expires_in":86400}`
	mockDID   = "did:test:foo"
	mockKeyID = "did:example:12345#testId"
)

var (
	//go:embed testdata/sample_credential_offer.json
	sampleCredentialOffer []byte

	//go:embed testdata/sample_credential_response.json
	sampleCredentialResponse []byte
)

type mockIssuerServerHandler struct {
	t                                                       *testing.T
	credentialOffer                                         *openid4ci.CredentialOffer
	credentialOfferEndpointShouldFail                       bool
	credentialOfferEndpointShouldGiveUnmarshallableResponse bool
	openIDConfig                                            *openid4ci.OpenIDConfig
	openIDConfigEndpointShouldFail                          bool
	issuerMetadata                                          string
	tokenRequestShouldFail                                  bool
	tokenRequestErrorResponse                               string
	tokenRequestShouldGiveUnmarshallableResponse            bool
	credentialRequestShouldFail                             bool
	credentialRequestErrorResponse                          string
	credentialRequestShouldGiveUnmarshallableResponse       bool
	credentialResponse                                      []byte
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
			require.NoError(m.t, err)

			_, err = writer.Write(credentialOfferBytes)
		}
	case "/.well-known/openid-configuration":
		if m.openIDConfigEndpointShouldFail {
			writer.WriteHeader(http.StatusInternalServerError)
			_, err = writer.Write([]byte("test failure"))
			require.NoError(m.t, err)

			return
		}

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
	case "/credential":
		switch {
		case m.credentialRequestShouldFail:
			writer.WriteHeader(http.StatusInternalServerError)

			if m.credentialRequestErrorResponse != "" {
				_, err = writer.Write([]byte(m.credentialRequestErrorResponse))
			} else {
				_, err = writer.Write([]byte("test failure"))
			}
		case m.credentialRequestShouldGiveUnmarshallableResponse:
			_, err = writer.Write([]byte("invalid"))
		default:
			_, err = writer.Write(m.credentialResponse)
		}
	}

	require.NoError(m.t, err)
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
		t.Run("Credential format is jwt_vc_json", func(t *testing.T) {
			newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, "example.com", false, true))
		})
		t.Run("Credential format is jwt_vc_json-ld", func(t *testing.T) {
			credentialOffer := createSampleCredentialOffer(t, true, true)

			credentialOffer.Credentials[0].Format = "jwt_vc_json-ld"

			credentialOfferBytes, err := json.Marshal(credentialOffer)
			require.NoError(t, err)

			credentialOfferEscaped := url.QueryEscape(string(credentialOfferBytes))

			credentialOfferIssuanceURI := "openid-credential-offer://?credential_offer=" + credentialOfferEscaped

			newIssuerInitiatedInteraction(t, credentialOfferIssuanceURI)
		})
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

		interaction, err := openid4ci.NewIssuerInitiatedInteraction(credentialOfferIssuanceURI, getTestClientConfig(t))
		require.EqualError(t, err, "no supported grant types found")
		require.Nil(t, interaction)
	})
	t.Run("Unsupported credential type", func(t *testing.T) {
		credentialOffer := createSampleCredentialOffer(t, false, true)

		credentialOffer.Credentials[0].Format = "UnsupportedType"

		credentialOfferBytes, err := json.Marshal(credentialOffer)
		require.NoError(t, err)

		credentialOfferEscaped := url.QueryEscape(string(credentialOfferBytes))

		credentialOfferIssuanceURI := "openid-credential-offer://?credential_offer=" + credentialOfferEscaped

		interaction, err := openid4ci.NewIssuerInitiatedInteraction(credentialOfferIssuanceURI, getTestClientConfig(t))
		require.EqualError(t, err, "UNSUPPORTED_CREDENTIAL_TYPE_IN_OFFER(OCI0-0002):unsupported "+
			"credential type (UnsupportedType) in credential offer at index 0 of credentials object "+
			"(must be jwt_vc_json or jwt_vc_json-ld)")
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

		issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential"}`,
			server.URL)

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
	keyWriter api.KeyWriter
}

func (m *mockResolver) Resolve(string) (*did.DocResolution, error) {
	didDoc, err := makeMockDoc(m.keyWriter)
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

		issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
			TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
		}

		authorizationServerURL := fmt.Sprintf("%s/auth", server.URL)

		issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential", `+
			`"authorization_server":"%s"}`,
			server.URL, authorizationServerURL)

		t.Run("Not using any options", func(t *testing.T) {
			interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, true, true))

			authorizationURL, err := interaction.CreateAuthorizationURL("clientID", "redirectURI")
			require.NoError(t, err)
			require.Contains(t, authorizationURL, authorizationServerURL+
				"?authorization_details=%7B%22type%22%3A%22openid_credential%22%2C%22locations"+
				"%22%3A%5B%22%22%5D%2C%22types%22%3A%5B%22VerifiableCredential%22%2C%22VerifiedEmployee%22%5D%2C%22"+
				"format%22%3A%22jwt_vc_json%22%7D&client_id=clientID")
		})
		t.Run("Using the OAuth Discoverable Client ID Scheme", func(t *testing.T) {
			interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, true, true))

			authorizationURL, err := interaction.CreateAuthorizationURL("clientID", "redirectURI",
				openid4ci.WithOAuthDiscoverableClientIDScheme())
			require.NoError(t, err)
			require.Contains(t, authorizationURL, authorizationServerURL+
				"?authorization_details=%7B%22type%22%3A%22openid_credential%22%2C%22locations"+
				"%22%3A%5B%22%22%5D%2C%22types%22%3A%5B%22VerifiableCredential%22%2C%22VerifiedEmployee%22%5D%2C%22"+
				"format%22%3A%22jwt_vc_json%22%7D&client_id=clientID")
		})
	})
	t.Run("Fail to get issuer metadata", func(t *testing.T) {
		interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, "example.com", true, true))

		authorizationURL, err := interaction.CreateAuthorizationURL("clientID", "redirectURI")
		require.EqualError(t, err, "METADATA_FETCH_FAILED(OCI1-0004):failed to get issuer metadata: openid "+
			`configuration endpoint: Get "example.com/.well-known/openid-credential-issuer": unsupported protocol scheme ""`)
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

					issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
						TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
					}

					issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential"}`,
						server.URL)

					interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, false, true))

					credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
						keyID: mockKeyID,
					}, openid4ci.WithPIN("1234"))
					require.NoError(t, err)
					require.Len(t, credentials, 1)
					require.NotEmpty(t, credentials[0])
				})
				t.Run("Token endpoint defined in the issuer's metadata", func(t *testing.T) {
					issuerServerHandler := &mockIssuerServerHandler{
						t:                  t,
						credentialResponse: sampleCredentialResponse,
						openIDConfig:       &openid4ci.OpenIDConfig{},
					}

					server := httptest.NewServer(issuerServerHandler)
					defer server.Close()

					issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential",`+
						`"token_endpoint":"%s/oidc/token"}`, server.URL, server.URL)

					interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, false, true))

					credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
						keyID: mockKeyID,
					}, openid4ci.WithPIN("1234"))
					require.NoError(t, err)
					require.Len(t, credentials, 1)
					require.NotEmpty(t, credentials[0])
				})
			})
			t.Run("Using credential_offer_uri", func(t *testing.T) {
				issuerServerHandler := &mockIssuerServerHandler{
					t:                  t,
					credentialResponse: sampleCredentialResponse,
				}

				server := httptest.NewServer(issuerServerHandler)
				defer server.Close()

				issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
					TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
				}

				issuerServerHandler.credentialOffer = createCredentialOffer(t, server.URL, false, true)

				issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential"}`,
					server.URL)

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
				require.Len(t, metricsLogger.events, 2)
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
		})
		t.Run("Missing PIN", func(t *testing.T) {
			config := getTestClientConfig(t)

			interaction, err := openid4ci.NewIssuerInitiatedInteraction(
				createCredentialOfferIssuanceURI(t, "example.com", false, true), config)
			require.NoError(t, err)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			})
			testutil.RequireErrorContains(t, err,
				"the credential offer requires a user PIN, but none was provided")
			require.Nil(t, credentials)
		})
		t.Run("No token endpoint available - neither the OpenID configuration nor the issuer's metadata "+
			"specify one", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:              t,
				issuerMetadata: "{}",
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction := newIssuerInitiatedInteraction(t, requestURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			require.EqualError(t, err, "failed to get credential response: "+
				"NO_TOKEN_ENDPOINT_AVAILABLE(OCI1-0019):no token endpoint available. Neither the OpenID "+
				"configuration nor the issuer's metadata specify one")
			require.Nil(t, credentials)
		})
		t.Run("Fail to reach issuer token endpoint", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:              t,
				openIDConfig:   &openid4ci.OpenIDConfig{TokenEndpoint: "http://BadURL"},
				issuerMetadata: "{}",
			}
			server := httptest.NewServer(issuerServerHandler)

			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction := newIssuerInitiatedInteraction(t, requestURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			require.Contains(t, err.Error(), `failed to get token response: issuer's token endpoint: Post `+
				`"http://BadURL": dial tcp: lookup BadURL`)
			require.Nil(t, credentials)
		})
		t.Run("Fail to get token response: server response body is not an errorResponse "+
			"object", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                      t,
				tokenRequestShouldFail: true,
				issuerMetadata:         "{}",
			}
			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

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
		t.Run("Fail to get token response: invalid token request", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                         t,
				tokenRequestShouldFail:    true,
				tokenRequestErrorResponse: `{"error":"invalid_request"}`,
				issuerMetadata:            "{}",
			}
			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

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
				issuerMetadata:            "{}",
			}
			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

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
				issuerMetadata:            "{}",
			}
			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

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
				issuerMetadata:            "{}",
			}
			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

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
				issuerMetadata: "{}",
			}
			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

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

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

			issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential"}`, server.URL)

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

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

			issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential"}`, server.URL)

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
		t.Run("Fail to get credential response: invalid token", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t: t, credentialRequestShouldFail: true,
				credentialRequestErrorResponse: `{"error":"invalid_token"}`,
			}
			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

			issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential"}`, server.URL)

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

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

			issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential"}`, server.URL)

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

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

			issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential"}`, server.URL)

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

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

			issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential"}`, server.URL)

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

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

			issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential"}`, server.URL)

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

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

			issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential"}`, server.URL)

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

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

			issuerServerHandler.issuerMetadata = `{"credential_endpoint":"http://BadURL"}`

			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction := newIssuerInitiatedInteraction(t, requestURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			require.Contains(t, err.Error(), `Post "http://BadURL": dial tcp: lookup BadURL`)
			require.Nil(t, credentials)
		})
		t.Run("Fail to get credential response: KID does not contain the DID part", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{t: t}
			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

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

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

			issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential"}`, server.URL)

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

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

			issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential"}`, server.URL)

			requestURI := createCredentialOfferIssuanceURI(t, server.URL, false, true)

			interaction := newIssuerInitiatedInteraction(t, requestURI)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			require.EqualError(t, err, "CREDENTIAL_PARSE_FAILED(OCI1-0007):failed to parse credential from "+
				"credential response at index 0: unmarshal new credential: unexpected end of JSON input")
			require.Nil(t, credentials)
		})
		t.Run("Fail VC proof check - public key not found for issuer DID", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{t: t, credentialResponse: sampleCredentialResponse}
			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

			issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential"}`, server.URL)

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
			require.EqualError(t, err, "CREDENTIAL_PARSE_FAILED(OCI1-0007):failed to parse credential from "+
				"credential response at index 0: "+
				"JWS proof check: unmarshal VC JWT claims: parse JWT: "+
				"parse JWT from compact JWS: public key with KID d3cfd36b-4f75-4041-b416-f0a7a3c6b9f6 is not "+
				"found for DID did:orb:uAAA:EiDpzs0hy0q0If4ZfJA1kxBQd9ed6FoBFhhqDWSiBeKaIg")
			require.Nil(t, credentials)
		})
		t.Run("Fail to log fetch OpenID config metrics event", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:              t,
				issuerMetadata: "{}",
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

			config := getTestClientConfig(t)
			config.MetricsLogger = &failingMetricsLogger{attemptFailNumber: 2}

			interaction, err := openid4ci.NewIssuerInitiatedInteraction(
				createCredentialOfferIssuanceURI(t, server.URL, false, true), config)
			require.NoError(t, err)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			require.Contains(t, err.Error(),
				"failed to log event (Event=Fetch issuer's OpenID configuration via an HTTP GET request "+
					"to http://127.0.0.1:")
			require.Nil(t, credentials)
		})
		t.Run("Fail to log fetch token via HTTP POST metrics event", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:              t,
				issuerMetadata: "{}",
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

			config := getTestClientConfig(t)
			config.MetricsLogger = &failingMetricsLogger{attemptFailNumber: 3}

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
		t.Run("Fail to log fetch metadata via HTTP GET metrics event", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:              t,
				issuerMetadata: "{}",
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

			config := getTestClientConfig(t)
			config.MetricsLogger = &failingMetricsLogger{attemptFailNumber: 1}

			interaction, err := openid4ci.NewIssuerInitiatedInteraction(
				createCredentialOfferIssuanceURI(t, server.URL, false, true), config)
			require.NoError(t, err)

			credentials, err := interaction.RequestCredentialWithPreAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, openid4ci.WithPIN("1234"))
			require.Contains(t, err.Error(), "METADATA_FETCH_FAILED(OCI1-0004):failed to get issuer metadata: "+
				"openid configuration endpoint: "+
				"failed to log event (Event=Fetch issuer metadata via an HTTP GET request to http://127.0.0.1:")
			require.Nil(t, credentials)
		})
		t.Run("Fail to log fetch credential via HTTP GET metrics event", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                  t,
				credentialResponse: sampleCredentialResponse,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

			issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential"}`,
				server.URL)

			config := getTestClientConfig(t)
			config.MetricsLogger = &failingMetricsLogger{attemptFailNumber: 4}

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

					issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
						TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
					}

					issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential"}`,
						server.URL)

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

					issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
						TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
					}

					issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential"}`,
						server.URL)

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

					issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
						TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
					}

					issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential"}`,
						server.URL)

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
					openIDConfig:       &openid4ci.OpenIDConfig{},
				}

				server := httptest.NewServer(issuerServerHandler)
				defer server.Close()

				issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential",`+
					`"token_endpoint":"%s/oidc/token"}`, server.URL, server.URL)

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
		})
		t.Run("Authorization URL not created first", func(t *testing.T) {
			interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, "example.com", true, true))

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

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

			authorizationServerURL := fmt.Sprintf("%s/auth", server.URL)

			issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential", `+
				`"authorization_server":"%s"}`,
				server.URL, authorizationServerURL)

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

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

			authorizationServerURL := fmt.Sprintf("%s/auth", server.URL)

			issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential", `+
				`"authorization_server":"%s"}`,
				server.URL, authorizationServerURL)

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

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

			authorizationServerURL := fmt.Sprintf("%s/auth", server.URL)

			issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential", `+
				`"authorization_server":"%s"}`,
				server.URL, authorizationServerURL)

			interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, true, true))

			_, err := interaction.CreateAuthorizationURL("clientID", "redirectURI")
			require.NoError(t, err)

			credentials, err := interaction.RequestCredentialWithAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, "%")
			require.EqualError(t, err, `parse "%": invalid URL escape "%"`)
			require.Nil(t, credentials)
		})
		t.Run("No token endpoint available - couldn't fetch OpenID configuration and the "+
			"issuer's metadata (the fallback) doesn't specify one", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                              t,
				credentialResponse:             sampleCredentialResponse,
				openIDConfigEndpointShouldFail: true,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

			issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential"}`,
				server.URL)

			interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, true, true))

			// Needed to create the OAuth2 config object.
			authURL, err := interaction.CreateAuthorizationURL("clientID", "redirectURI")
			require.NoError(t, err)

			redirectURIWithParams := "redirectURI?code=1234&state=" + getStateFromAuthURL(t, authURL)

			credentials, err := interaction.RequestCredentialWithAuth(&jwtSignerMock{
				keyID: mockKeyID,
			}, redirectURIWithParams)
			require.EqualError(t, err, "NO_TOKEN_ENDPOINT_AVAILABLE(OCI1-0019):no token endpoint available. "+
				"An OpenID configuration couldn't be fetched, and the issuer's metadata doesn't specify a token "+
				"endpoint. OpenID configuration fetch error: openid configuration endpoint: expected status code 200 "+
				"but got status code 500 with response body test failure instead")
			require.Nil(t, credentials)
		})
		t.Run("Fail to create claims proof", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{
				t:                  t,
				credentialResponse: sampleCredentialResponse,
			}

			server := httptest.NewServer(issuerServerHandler)
			defer server.Close()

			issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
				TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
			}

			issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential"}`,
				server.URL)

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

		issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
			TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
		}

		issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential"}`,
			server.URL)

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

		issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
			TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
		}

		issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential"}`,
			server.URL)

		interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, true, true))

		authorizationURL, err := interaction.CreateAuthorizationURL("clientID", "redirectURI",
			openid4ci.WithIssuerState("SomeOtherState"))
		require.EqualError(t, err, "INVALID_SDK_USAGE(OCI3-0000):the credential offer already specifies "+
			"an issuer state, and a conflicting issuer state value was provided. An issuer state should only be "+
			"provided if required by the issuer and the credential offer does not specify one already")
		require.Empty(t, authorizationURL)
	})
}

func TestIssuerInitiatedInteraction_GrantTypes(t *testing.T) {
	interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, "example.com", false, true))

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

	interaction = newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, "example.com", true, true))

	require.True(t, interaction.AuthorizationCodeGrantTypeSupported())

	authorizationCodeGrantParams, err = interaction.AuthorizationCodeGrantParams()
	require.NoError(t, err)
	require.NotNil(t, authorizationCodeGrantParams)

	require.NotNil(t, authorizationCodeGrantParams.IssuerState)
	require.Equal(t, "1234", *authorizationCodeGrantParams.IssuerState)
}

func TestIssuerInitiatedInteraction_DynamicClientRegistration(t *testing.T) {
	t.Run("Fail to get OpenID configuration", func(t *testing.T) {
		interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, "example.com", false, true))

		supported, err := interaction.DynamicClientRegistrationSupported()
		require.EqualError(t, err, "ISSUER_OPENID_CONFIG_FETCH_FAILED(OCI1-0003):failed to fetch issuer's "+
			"OpenID configuration: "+`openid configuration endpoint: Get "example.com/.well-known/openid-configuration"`+
			`: unsupported protocol scheme ""`)
		require.False(t, supported)

		endpoint, err := interaction.DynamicClientRegistrationEndpoint()
		require.EqualError(t, err, "ISSUER_OPENID_CONFIG_FETCH_FAILED(OCI1-0003):failed to fetch issuer's "+
			"OpenID configuration: "+`openid configuration endpoint: Get "example.com/.well-known/openid-configuration"`+
			`: unsupported protocol scheme ""`)
		require.Empty(t, endpoint)
	})
	t.Run("Dynamic client registration is not supported", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t: t,
		}

		server := httptest.NewServer(issuerServerHandler)
		defer server.Close()

		issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{}

		interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, true, true))

		supported, err := interaction.DynamicClientRegistrationSupported()
		require.NoError(t, err)
		require.False(t, supported)

		endpoint, err := interaction.DynamicClientRegistrationEndpoint()
		require.EqualError(t, err,
			"INVALID_SDK_USAGE(OCI3-0000):issuer does not support dynamic client registration")
		require.Empty(t, endpoint)
	})
	t.Run("Dynamic client registration is supported", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t: t,
		}

		server := httptest.NewServer(issuerServerHandler)
		defer server.Close()

		testEndpoint := "SomeEndpoint"

		issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{RegistrationEndpoint: &testEndpoint}

		interaction := newIssuerInitiatedInteraction(t, createCredentialOfferIssuanceURI(t, server.URL, true, true))

		supported, err := interaction.DynamicClientRegistrationSupported()
		require.NoError(t, err)
		require.True(t, supported)

		endpoint, err := interaction.DynamicClientRegistrationEndpoint()
		require.NoError(t, err)
		require.Equal(t, testEndpoint, endpoint)
	})
}

func TestIssuerInitiatedInteraction_Issuer_URI(t *testing.T) {
	testIssuerURI := "https://example.com"
	requestURI := createCredentialOfferIssuanceURI(t, testIssuerURI, false, true)

	interaction := newIssuerInitiatedInteraction(t, requestURI)

	require.Equal(t, testIssuerURI, interaction.IssuerURI())
}

func newIssuerInitiatedInteraction(t *testing.T, requestURI string) *openid4ci.IssuerInitiatedInteraction {
	t.Helper()

	config := getTestClientConfig(t)

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

// makeMockDoc creates a key in the given KMS and returns a mock DID Doc with a verification method.
func makeMockDoc(keyWriter api.KeyWriter) (*did.Doc, error) {
	_, pkJWK, err := keyWriter.Create(arieskms.ED25519Type)
	if err != nil {
		return nil, err
	}

	pkb, err := pkJWK.PublicKeyBytes()
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

type jwtSignerMock struct {
	keyID string
	Err   error
}

func (s *jwtSignerMock) GetKeyID() string {
	return s.keyID
}

func (s *jwtSignerMock) Sign([]byte) ([]byte, error) {
	return []byte("test signature"), s.Err
}

func (s *jwtSignerMock) Headers() jose.Headers {
	return jose.Headers{
		jose.HeaderKeyID:     "KeyID",
		jose.HeaderAlgorithm: "ES384",
	}
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
