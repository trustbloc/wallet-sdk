/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

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
	"strings"
	"testing"
	"time"

	"github.com/piprate/json-gold/ld"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/did-go/doc/did"
	arieskms "github.com/trustbloc/kms-go/spi/kms"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/activitylogger/mem"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/metricslogger/stderr"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/openid4ci"
	"github.com/trustbloc/wallet-sdk/internal/testutil"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/models"
	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
	goapiopenid4ci "github.com/trustbloc/wallet-sdk/pkg/openid4ci"
)

var (
	//go:embed testdata/sample_credential_response.json
	sampleCredentialResponse []byte

	//go:embed testdata/sample_credential_offer.json
	sampleCredentialOffer []byte
)

const (
	sampleTokenResponse = `{"access_token":"eyJhbGciOiJSUzI1NiIsInR5cCI6Ikp..sHQ",` +
		`"token_type":"bearer","expires_in":86400,"c_nonce":"tZignsnFbp","c_nonce_expires_in":86400}`

	mockDID   = "did:test:foo"
	mockKeyID = "did:test:foo#abcd"

	serverURLPlaceholder = "[SERVER_URL]"
)

func TestNewInteraction(t *testing.T) {
	issuerServerHandler := &mockIssuerServerHandler{
		t: t,
	}

	server := httptest.NewServer(issuerServerHandler)
	defer server.Close()

	issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

	t.Run("Success", func(t *testing.T) {
		kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
		require.NoError(t, err)

		i := createIssuerInitiatedInteraction(t, kms, nil, nil,
			createCredentialOfferIssuanceURI(t, server.URL, false),
			nil, false)
		require.NotEmpty(t, i.OTelTraceID())
	})

	t.Run("Success with disable otel", func(t *testing.T) {
		kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
		require.NoError(t, err)

		resolver := &mockResolver{keyWriter: kms}

		opts := openid4ci.NewInteractionOpts()
		opts.DisableOpenTelemetry()

		requiredArgs := openid4ci.NewIssuerInitiatedInteractionArgs(createCredentialOfferIssuanceURI(t, server.URL, false),
			kms.GetCrypto(), resolver)

		interaction, err := openid4ci.NewIssuerInitiatedInteraction(requiredArgs, opts)
		require.NoError(t, err)
		require.NotNil(t, interaction)

		require.Empty(t, interaction.OTelTraceID())
	})

	t.Run("Success with out optional args", func(t *testing.T) {
		kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
		require.NoError(t, err)

		resolver := &mockResolver{keyWriter: kms}

		requiredArgs := openid4ci.NewIssuerInitiatedInteractionArgs(createCredentialOfferIssuanceURI(t, server.URL, false),
			kms.GetCrypto(), resolver)

		interaction, err := openid4ci.NewIssuerInitiatedInteraction(requiredArgs, nil)
		require.NoError(t, err)
		require.NotNil(t, interaction)
	})

	t.Run("Success HTTP timeout", func(t *testing.T) {
		kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
		require.NoError(t, err)

		resolver := &mockResolver{keyWriter: kms}

		requiredArgs := openid4ci.NewIssuerInitiatedInteractionArgs(createCredentialOfferIssuanceURI(t, server.URL, false),
			kms.GetCrypto(), resolver)
		opts := openid4ci.NewInteractionOpts()
		opts.SetHTTPTimeoutNanoseconds((10 * time.Second).Nanoseconds())

		interaction, err := openid4ci.NewIssuerInitiatedInteraction(requiredArgs, opts)
		require.NoError(t, err)
		require.NotNil(t, interaction)
	})

	t.Run("Failed, args is nil", func(t *testing.T) {
		interaction, err := openid4ci.NewIssuerInitiatedInteraction(nil, nil)
		require.Error(t, err)
		require.Nil(t, interaction)
	})
}

type mockIssuerServerHandler struct {
	t                                                 *testing.T
	issuerMetadata                                    string
	tokenRequestShouldFail                            bool
	tokenRequestShouldGiveUnmarshallableResponse      bool
	credentialRequestShouldFail                       bool
	credentialRequestShouldGiveUnmarshallableResponse bool
	ackRequestExpectInteractionDetails                bool
	ackRequestExpectedAmount                          int
	credentialResponse                                []byte
	headersToCheck                                    *api.Headers
}

func (m *mockIssuerServerHandler) ServeHTTP(writer http.ResponseWriter, //nolint: gocyclo // test file
	request *http.Request,
) {
	var err error

	if m.headersToCheck != nil && request.URL.Path != "/oidc/ack_endpoint" {
		for _, headerToCheck := range m.headersToCheck.GetAll() {
			// Note: for these tests, we're assuming that there aren't multiple values under a single name/key.
			value := request.Header.Get(headerToCheck.Name)
			assert.Equal(m.t, headerToCheck.Value, value)
		}
	}

	switch request.URL.Path {
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
			writer.Header().Set("Content-Type", "application/json")
			_, err = writer.Write([]byte(sampleTokenResponse))
		}
	case "/oidc/credential":
		switch {
		case m.credentialRequestShouldFail:
			writer.WriteHeader(http.StatusInternalServerError)
			_, err = writer.Write([]byte("test failure"))
		case m.credentialRequestShouldGiveUnmarshallableResponse:
			_, err = writer.Write([]byte("invalid"))
		default:
			_, err = writer.Write(m.credentialResponse)
		}
	case "/oidc/ack_endpoint":
		m.ackRequestExpectedAmount--

		var payload map[string]interface{}
		err = json.NewDecoder(request.Body).Decode(&payload)
		assert.NoError(m.t, err)

		_, ok := payload["interaction_details"]
		assert.Equal(m.t, m.ackRequestExpectInteractionDetails, ok)

		writer.WriteHeader(http.StatusNoContent)
	}

	assert.NoError(m.t, err)
}

func TestIssuerInitiatedInteraction_CreateAuthorizationURL(t *testing.T) {
	issuerServerHandler := &mockIssuerServerHandler{
		t: t,
	}

	server := httptest.NewServer(issuerServerHandler)
	defer server.Close()

	issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

	t.Run("Issuer does not support the authorization code grant type", func(t *testing.T) {
		kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
		require.NoError(t, err)

		interaction := createIssuerInitiatedInteraction(t, kms, nil, nil,
			createCredentialOfferIssuanceURI(t, server.URL, false),
			nil, false)

		opts := openid4ci.NewCreateAuthorizationURLOpts().UseOAuthDiscoverableClientIDScheme()

		authorizationLink, err := interaction.CreateAuthorizationURL("clientID", "redirectURI", opts)
		requireErrorContains(t, err, "INVALID_SDK_USAGE")
		requireErrorContains(t, err, "issuer does not support the authorization code grant type")
		require.Empty(t, authorizationLink)
	})
	t.Run("Conflicting issuer state", func(t *testing.T) {
		kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
		require.NoError(t, err)

		interaction := createIssuerInitiatedInteraction(t, kms, nil, nil,
			createCredentialOfferIssuanceURI(t, server.URL, true), nil, false)

		createAuthorizationURLOpts := openid4ci.NewCreateAuthorizationURLOpts().SetIssuerState("IssuerState")

		authorizationLink, err := interaction.CreateAuthorizationURL("clientID", "redirectURI",
			createAuthorizationURLOpts)
		requireErrorContains(t, err, "INVALID_SDK_USAGE")
		requireErrorContains(t, err, "the credential offer already specifies "+
			"an issuer state, and a conflicting issuer state value was provided. An issuer state should only be "+
			"provided if required by the issuer and the credential offer does not specify one already")
		require.Empty(t, authorizationLink)
	})
	t.Run("Auth URL with custom scopes", func(t *testing.T) {
		kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
		require.NoError(t, err)

		interaction := createIssuerInitiatedInteraction(t, kms, nil, nil,
			createCredentialOfferIssuanceURI(t, server.URL, false),
			nil, false)

		opts := openid4ci.NewCreateAuthorizationURLOpts().SetScopes(api.NewStringArray().Append("custom_scope"))

		authorizationLink, err := interaction.CreateAuthorizationURL("clientID", "redirectURI", opts)
		requireErrorContains(t, err, "INVALID_SDK_USAGE")
		requireErrorContains(t, err, "issuer does not support the authorization code grant type")
		require.Empty(t, authorizationLink)
	})
}

func TestIssuerInitiatedInteraction_RequestCredential(t *testing.T) {
	issuerServerHandler := &mockIssuerServerHandler{
		t:                  t,
		credentialResponse: sampleCredentialResponse,
	}

	server := httptest.NewServer(issuerServerHandler)
	defer server.Close()

	issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

	t.Run("Success", func(t *testing.T) {
		t.Run("Without additional headers, TLS verification enabled", func(t *testing.T) {
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
		t.Run("Acknowledge reject", func(t *testing.T) {
			doRequestCredentialTestExt(t, nil, false, true, "", false)
		})
		t.Run("Acknowledge reject with code", func(t *testing.T) {
			doRequestCredentialTestExt(t, nil, false, true, "tc_declined", false)
		})
	})
	t.Run("Success with jwk public key", func(t *testing.T) {
		requestURI := createCredentialOfferIssuanceURI(t, server.URL, false)

		kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
		require.NoError(t, err)

		interactionRequiredArgs, interactionOptionalArgs := getTestArgs(t, requestURI, kms, nil, nil, nil, false)

		interaction, err := openid4ci.NewIssuerInitiatedInteraction(interactionRequiredArgs, interactionOptionalArgs)
		require.NoError(t, err)

		keyHandle, err := kms.Create(arieskms.ED25519)
		require.NoError(t, err)

		verificationMethod := &api.VerificationMethod{
			ID:   mockKeyID,
			Type: "JsonWebKey2020",
			Key:  models.VerificationKey{JSONWebKey: keyHandle.JWK},
		}

		credentials, err := interaction.RequestCredentialWithPreAuth(verificationMethod,
			openid4ci.NewRequestCredentialWithPreAuthOpts().SetPIN("1234"))

		require.NoError(t, err)
		require.NotNil(t, credentials)
	})
	t.Run("Success with jwk public key V2", func(t *testing.T) {
		requestURI := createCredentialOfferIssuanceURI(t, server.URL, false)

		kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
		require.NoError(t, err)

		interactionRequiredArgs, interactionOptionalArgs := getTestArgs(t, requestURI, kms, nil, nil, nil, false)

		interaction, err := openid4ci.NewIssuerInitiatedInteraction(interactionRequiredArgs, interactionOptionalArgs)
		require.NoError(t, err)

		keyHandle, err := kms.Create(arieskms.ED25519)
		require.NoError(t, err)

		verificationMethod := &api.VerificationMethod{
			ID:   mockKeyID,
			Type: "JsonWebKey2020",
			Key:  models.VerificationKey{JSONWebKey: keyHandle.JWK},
		}

		credentials, err := interaction.RequestCredentialWithPreAuthV2(verificationMethod,
			openid4ci.NewRequestCredentialWithPreAuthOpts().SetPIN("1234"))

		require.NoError(t, err)
		require.NotNil(t, credentials)
	})

	t.Run("attestation invalid VC", func(t *testing.T) {
		requestURI := createCredentialOfferIssuanceURI(t, server.URL, false)

		kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
		require.NoError(t, err)

		interactionRequiredArgs, interactionOptionalArgs := getTestArgs(t, requestURI, kms, nil, nil, nil, false)

		interaction, err := openid4ci.NewIssuerInitiatedInteraction(interactionRequiredArgs, interactionOptionalArgs)
		require.NoError(t, err)

		keyHandle, err := kms.Create(arieskms.ED25519)
		require.NoError(t, err)

		verificationMethod := &api.VerificationMethod{
			ID:   mockKeyID,
			Type: "JsonWebKey2020",
			Key:  models.VerificationKey{JSONWebKey: keyHandle.JWK},
		}

		_, err = interaction.RequestCredentialWithPreAuth(verificationMethod,
			openid4ci.NewRequestCredentialWithPreAuthOpts().SetPIN("1234").
				SetAttestationVC(verificationMethod, "invalidVC"))

		require.Error(t, err)
	})

	t.Run("Authorization code flow - authorization URL must be created first", func(t *testing.T) {
		kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
		require.NoError(t, err)

		interaction := createIssuerInitiatedInteraction(t, kms, nil, nil,
			createCredentialOfferIssuanceURI(t, server.URL, true),
			nil, false)

		keyHandle, err := kms.Create(arieskms.ED25519)
		require.NoError(t, err)

		pkBytes, err := keyHandle.JWK.PublicKeyBytes()
		require.NoError(t, err)

		verificationMethod := &api.VerificationMethod{
			ID:   "did:example:12345#testId",
			Type: "Ed25519VerificationKey2018",
			Key:  models.VerificationKey{Raw: pkBytes},
		}

		credentials, err := interaction.RequestCredentialWithAuth(verificationMethod,
			"redirectURIWithAuthCode", nil)
		requireErrorContains(t, err, "authorization URL must be created first")
		require.Nil(t, credentials)
	})
}

func TestIssuerInitiatedInteraction_GrantTypes(t *testing.T) {
	issuerServerHandler := &mockIssuerServerHandler{
		t: t,
	}

	server := httptest.NewServer(issuerServerHandler)
	defer server.Close()

	issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

	kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
	require.NoError(t, err)

	interaction := createIssuerInitiatedInteraction(t, kms, nil, nil,
		createCredentialOfferIssuanceURI(t, server.URL, false), nil, false)

	require.True(t, interaction.PreAuthorizedCodeGrantTypeSupported())

	preAuthorizedCodeGrantParams, err := interaction.PreAuthorizedCodeGrantParams()
	require.NoError(t, err)
	require.NotNil(t, preAuthorizedCodeGrantParams)

	require.True(t, preAuthorizedCodeGrantParams.PINRequired())

	require.False(t, interaction.AuthorizationCodeGrantTypeSupported())

	authorizationCodeGrantParams, err := interaction.AuthorizationCodeGrantParams()
	requireErrorContains(t, err, "INVALID_SDK_USAGE")
	requireErrorContains(t, err, "issuer does not support the authorization code grant")
	require.Nil(t, authorizationCodeGrantParams)

	interaction = createIssuerInitiatedInteraction(t, kms, nil, nil,
		createCredentialOfferIssuanceURI(t, server.URL, true), nil, false)

	require.True(t, interaction.AuthorizationCodeGrantTypeSupported())

	authorizationCodeGrantParams, err = interaction.AuthorizationCodeGrantParams()
	require.NoError(t, err)
	require.NotNil(t, authorizationCodeGrantParams)

	require.True(t, authorizationCodeGrantParams.HasIssuerState())

	issuerState, err := authorizationCodeGrantParams.IssuerState()
	require.NoError(t, err)
	require.Equal(t, "1234", issuerState)
}

func TestIssuerInitiatedInteraction_DynamicClientRegistration(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t: t,
		}

		server := httptest.NewServer(issuerServerHandler)
		defer server.Close()

		issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

		kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
		require.NoError(t, err)

		interaction := createIssuerInitiatedInteraction(t, kms, nil, nil,
			createCredentialOfferIssuanceURI(t, server.URL, false), nil, false)

		supported, err := interaction.DynamicClientRegistrationSupported()
		require.NoError(t, err)
		require.True(t, supported)
	})
}

func TestIssuerInitiatedInteraction_IssuerMetadata(t *testing.T) {
	t.Run("Successfully get issuer metadata", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t: t,
		}

		server := httptest.NewServer(issuerServerHandler)
		defer server.Close()

		issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

		kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
		require.NoError(t, err)

		i := createIssuerInitiatedInteraction(t, kms, nil, nil,
			createCredentialOfferIssuanceURI(t, server.URL, false),
			nil, false)
		require.NotEmpty(t, i.OTelTraceID())

		issuerMetadata, err := i.IssuerMetadata()
		require.NoError(t, err)
		require.NotNil(t, issuerMetadata)
	})
}

func TestIssuerInitiatedInteraction_VerifyIssuer(t *testing.T) {
	t.Run("Metadata not signed", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t: t,
		}

		server := httptest.NewServer(issuerServerHandler)
		defer server.Close()

		issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

		activityLogger := mem.NewActivityLogger()
		metricsLogger := stderr.NewMetricsLogger()

		kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
		require.NoError(t, err)

		interaction := createIssuerInitiatedInteraction(t, kms, activityLogger, metricsLogger,
			createCredentialOfferIssuanceURI(t, server.URL, false),
			nil, true)

		serviceURL, err := interaction.VerifyIssuer()
		requireErrorContains(t, err, "DID service validation failed")
		require.Empty(t, serviceURL)
	})
}

func TestIssuerInitiatedInteraction_IssuerTrustInfo(t *testing.T) {
	t.Run("Metadata not signed", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t: t,
		}
		server := httptest.NewServer(issuerServerHandler)

		issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

		defer server.Close()

		activityLogger := mem.NewActivityLogger()
		metricsLogger := stderr.NewMetricsLogger()

		kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
		require.NoError(t, err)

		interaction := createIssuerInitiatedInteraction(t, kms, activityLogger,
			metricsLogger, createCredentialOfferIssuanceURI(t, server.URL, false),
			nil, true)

		trustInfo, err := interaction.IssuerTrustInfo()
		require.NoError(t, err)
		require.Equal(t, 1, trustInfo.OfferLength())
		require.Equal(t, "jwt_vc_json-ld", trustInfo.OfferAtIndex(0).CredentialFormat)
		require.Equal(t, "PermanentResidentCard", trustInfo.OfferAtIndex(0).CredentialType)
	})
}

// The IssuerInitiatedInteraction alias type (Interaction) should behave the same as the
// IssuerInitiatedInteraction object, since it's just a wrapper for it.
func TestIssuerInitiatedInteractionAlias(t *testing.T) {
	issuerServerHandler := &mockIssuerServerHandler{
		t:                  t,
		credentialResponse: sampleCredentialResponse,
	}

	server := httptest.NewServer(issuerServerHandler)
	defer server.Close()

	metadata := strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

	issuerServerHandler.issuerMetadata = modifyCredentialMetadata(t, metadata, func(m *issuer.Metadata) {
		m.RegistrationEndpoint = nil
	})

	activityLogger := mem.NewActivityLogger()

	kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
	require.NoError(t, err)

	interaction := createIssuedInitiatedInteractionAlias(t, kms, activityLogger,
		createCredentialOfferIssuanceURI(t, server.URL, false),
		nil,
		false)

	keyHandle, err := kms.Create(arieskms.ED25519)
	require.NoError(t, err)

	pkBytes, err := keyHandle.JWK.PublicKeyBytes()
	require.NoError(t, err)

	credentials, err := interaction.RequestCredentialWithPreAuth(&api.VerificationMethod{
		ID:   "did:example:12345#testId",
		Type: "Ed25519VerificationKey2018",
		Key:  models.VerificationKey{Raw: pkBytes},
	}, openid4ci.NewRequestCredentialWithPreAuthOpts().SetPIN("1234"))
	require.NoError(t, err)
	require.NotNil(t, credentials)

	// The rest of these calls are not representative of how the object should be used in an OpenID4CI flow.
	// These are just here to add code coverage for the alias wrapper methods (which behave the same as the methods on
	// IssuerInitiatedInteraction) See TestIssuerInitiatedInteraction_RequestCredential or the integration tests for
	// better examples.
	authURL, err := interaction.CreateAuthorizationURL("", "", nil)
	requireErrorContains(t, err, "INVALID_SDK_USAGE")
	requireErrorContains(t, err, "issuer does not support the authorization code grant type")
	require.Empty(t, authURL)

	credentials, err = interaction.RequestCredentialWithPreAuth(nil, nil)
	requireErrorContains(t, err, "verification method must be provided")
	require.Nil(t, credentials)

	credentials, err = interaction.RequestCredentialWithAuth(nil, "", nil)
	requireErrorContains(t, err, "verification method must be provided")
	require.Nil(t, credentials)

	issuerURI := interaction.IssuerURI()
	require.NotEmpty(t, issuerURI)

	require.True(t, interaction.PreAuthorizedCodeGrantTypeSupported())

	preAuthCodeGrantParams, err := interaction.PreAuthorizedCodeGrantParams()
	require.NoError(t, err)
	require.NotNil(t, preAuthCodeGrantParams)

	require.False(t, interaction.AuthorizationCodeGrantTypeSupported())

	authCodeGrantParams, err := interaction.AuthorizationCodeGrantParams()
	requireErrorContains(t, err, "INVALID_SDK_USAGE")
	requireErrorContains(t, err, "issuer does not support the authorization code grant")
	require.Nil(t, authCodeGrantParams)

	dynamicClientRegistrationSupported, err := interaction.DynamicClientRegistrationSupported()
	require.NoError(t, err)
	require.False(t, dynamicClientRegistrationSupported)

	dynamicClientRegistrationEndpoint, err := interaction.DynamicClientRegistrationEndpoint()
	requireErrorContains(t, err, "INVALID_SDK_USAGE")
	requireErrorContains(t, err, "issuer does not support dynamic client registration")
	require.Empty(t, dynamicClientRegistrationEndpoint)

	traceID := interaction.OTelTraceID()
	require.NotEmpty(t, traceID)

	issuerMetadata, err := interaction.IssuerMetadata()
	require.NoError(t, err)
	require.NotNil(t, issuerMetadata)
}

func TestIssuerMetadataFromToGoImpl(t *testing.T) {
	goImpl := &issuer.Metadata{}
	restored := openid4ci.IssuerMetadataToGoImpl(openid4ci.IssuerMetadataFromGoImpl(goImpl))
	require.Equal(t, goImpl, restored)
}

func doRequestCredentialTest(t *testing.T, additionalHeaders *api.Headers,
	disableTLSVerification bool,
) {
	t.Helper()
	doRequestCredentialTestExt(t, additionalHeaders, disableTLSVerification, false, "", true)
}

//nolint:thelper // Not a test helper function
func doRequestCredentialTestExt(t *testing.T, additionalHeaders *api.Headers,
	disableTLSVerification bool, acknowledgeReject bool, rejectCode string, expectAckInteractionDetails bool,
) {
	issuerServerHandler := &mockIssuerServerHandler{
		t:                                  t,
		credentialResponse:                 sampleCredentialResponse,
		headersToCheck:                     additionalHeaders,
		ackRequestExpectInteractionDetails: expectAckInteractionDetails,
		ackRequestExpectedAmount:           1,
	}

	server := httptest.NewServer(issuerServerHandler)
	defer server.Close()

	issuerServerHandler.issuerMetadata = strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

	activityLogger := mem.NewActivityLogger()
	metricsLogger := stderr.NewMetricsLogger()

	kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
	require.NoError(t, err)

	interaction := createIssuerInitiatedInteraction(t, kms, activityLogger, metricsLogger,
		createCredentialOfferIssuanceURI(t, server.URL, false),
		additionalHeaders,
		disableTLSVerification)

	offeringTypes := api.StringArrayArrayToGoArray(interaction.OfferedCredentialsTypes())
	require.Len(t, offeringTypes, 1)
	require.Contains(t, offeringTypes[0], "PermanentResidentCard")

	keyHandle, err := kms.Create(arieskms.ED25519)
	require.NoError(t, err)

	pkBytes, err := keyHandle.JWK.PublicKeyBytes()
	require.NoError(t, err)

	credentials, err := interaction.RequestCredentialWithPreAuth(&api.VerificationMethod{
		ID:   "did:example:12345#testId",
		Type: "Ed25519VerificationKey2018",
		Key:  models.VerificationKey{Raw: pkBytes},
	}, openid4ci.NewRequestCredentialWithPreAuthOpts().SetPIN("1234"))
	require.NoError(t, err)
	require.NotNil(t, credentials)

	requireAcknowledgment, err := interaction.RequireAcknowledgment()
	require.NoError(t, err)
	require.True(t, requireAcknowledgment)

	acknowledgment, err := interaction.Acknowledgment()
	require.NotNil(t, acknowledgment)
	require.NoError(t, err)

	acknowledgmentData, err := acknowledgment.Serialize()
	require.NotEmpty(t, acknowledgmentData)
	require.NoError(t, err)

	acknowledgmentRestored, err := openid4ci.NewAcknowledgment(acknowledgmentData)
	require.NotEmpty(t, acknowledgmentRestored)
	require.NoError(t, err)

	if expectAckInteractionDetails {
		err = acknowledgmentRestored.SetInteractionDetails(`{"key1": "value1"}`)
		require.NoError(t, err)
	}

	switch {
	case acknowledgeReject && rejectCode != "":
		err = acknowledgmentRestored.RejectWithCode(rejectCode)
		require.NoError(t, err)

		err = acknowledgmentRestored.RejectWithCode(rejectCode)
		require.ErrorContains(t, err, "ack list is empty")
	case acknowledgeReject:
		err = acknowledgmentRestored.Reject()
		require.NoError(t, err)

		err = acknowledgmentRestored.Reject()
		require.ErrorContains(t, err, "ack list is empty")
	default:
		err = acknowledgmentRestored.Success()
		require.NoError(t, err)

		err = acknowledgmentRestored.Success()
		require.ErrorContains(t, err, "ack list is empty")
	}

	require.Zero(t, issuerServerHandler.ackRequestExpectedAmount, 0)

	numberOfActivitiesLogged := activityLogger.Length()
	require.Equal(t, 1, numberOfActivitiesLogged)

	activity := activityLogger.AtIndex(0)

	require.NotEmpty(t, activity.ID)
	require.Equal(t, goapi.LogTypeCredentialActivity, activity.Type())
	require.NotEmpty(t, activity.UnixTimestamp())
	require.Equal(t, server.URL, activity.Client())
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

	issuerMetadata, err := interaction.IssuerMetadata()
	require.NoError(t, err)
	require.NotNil(t, issuerMetadata)
}

func TestNewRequestedAcknowledgment(t *testing.T) {
	_, err := openid4ci.NewAcknowledgment("[")
	require.Error(t, err)
}

func createIssuerInitiatedInteraction(t *testing.T, kms *localkms.KMS, activityLogger api.ActivityLogger,
	metricsLogger api.MetricsLogger, requestURI string, additionalHeaders *api.Headers, disableTLSVerification bool,
) *openid4ci.IssuerInitiatedInteraction {
	t.Helper()

	requiredArgs, opts := getTestArgs(t, requestURI, kms, activityLogger, metricsLogger, additionalHeaders,
		disableTLSVerification)

	interaction, err := openid4ci.NewIssuerInitiatedInteraction(requiredArgs, opts)
	require.NoError(t, err)
	require.NotNil(t, interaction)

	return interaction
}

func createIssuedInitiatedInteractionAlias(t *testing.T, kms *localkms.KMS,
	activityLogger api.ActivityLogger, requestURI string,
	additionalHeaders *api.Headers, disableTLSVerification bool,
) *openid4ci.IssuerInitiatedInteraction {
	t.Helper()

	requiredArgs, opts := getTestArgsAlias(t, requestURI, kms, activityLogger, additionalHeaders,
		disableTLSVerification)

	interaction, err := openid4ci.NewIssuerInitiatedInteraction(requiredArgs, opts)
	require.NoError(t, err)
	require.NotNil(t, interaction)

	return interaction
}

// getTestArgs accepts an optional activityLogger and also one optional mockVMCreator.
func getTestArgs(t *testing.T, initiateIssuanceURI string, kms *localkms.KMS,
	activityLogger api.ActivityLogger, metricsLogger api.MetricsLogger, additionalHeaders *api.Headers,
	disableTLSVerification bool,
) (*openid4ci.IssuerInitiatedInteractionArgs, *openid4ci.InteractionOpts) {
	t.Helper()

	resolver := &mockResolver{keyWriter: kms}

	opts := openid4ci.NewInteractionOpts()
	opts.DisableVCProofChecks()
	opts.SetDocumentLoader(&documentLoaderWrapper{DocumentLoader: testutil.DocumentLoader(t)})

	timeout := time.Second * 10
	opts.SetHTTPTimeoutNanoseconds(timeout.Nanoseconds())

	opts.EnableDIProofChecks(kms)

	if activityLogger != nil {
		opts.SetActivityLogger(activityLogger)
	}

	if metricsLogger != nil {
		opts.SetMetricsLogger(metricsLogger)
	}

	if additionalHeaders != nil {
		opts.AddHeaders(additionalHeaders)
	}

	if disableTLSVerification {
		opts.DisableHTTPClientTLSVerify()
	}

	requiredArgs := openid4ci.NewIssuerInitiatedInteractionArgs(initiateIssuanceURI, kms.GetCrypto(), resolver)

	return requiredArgs, opts
}

// getTestArgsAlias accepts an optional activityLogger and also one optional mockVMCreator.
func getTestArgsAlias(t *testing.T, initiateIssuanceURI string, kms *localkms.KMS,
	activityLogger api.ActivityLogger, additionalHeaders *api.Headers,
	disableTLSVerification bool,
) (*openid4ci.IssuerInitiatedInteractionArgs, *openid4ci.InteractionOpts) {
	t.Helper()

	resolver := &mockResolver{keyWriter: kms}

	opts := openid4ci.NewInteractionOpts()
	opts.DisableVCProofChecks()
	opts.SetDocumentLoader(&documentLoaderWrapper{DocumentLoader: testutil.DocumentLoader(t)})

	timeout := time.Second * 10
	opts.SetHTTPTimeoutNanoseconds(timeout.Nanoseconds())

	if activityLogger != nil {
		opts.SetActivityLogger(activityLogger)
	}

	if additionalHeaders != nil {
		opts.AddHeaders(additionalHeaders)
	}

	if disableTLSVerification {
		opts.DisableHTTPClientTLSVerify()
	}

	requiredArgs := openid4ci.NewIssuerInitiatedInteractionArgs(initiateIssuanceURI, kms.GetCrypto(), resolver)

	return requiredArgs, opts
}

type documentLoaderWrapper struct {
	DocumentLoader ld.DocumentLoader
}

func (l *documentLoaderWrapper) LoadDocument(u string) (*api.LDDocument, error) {
	doc, err := l.DocumentLoader.LoadDocument(u)
	if err != nil {
		return nil, err
	}

	wrappedDoc := &api.LDDocument{
		DocumentURL: doc.DocumentURL,
		ContextURL:  doc.ContextURL,
	}

	documentBytes, err := json.Marshal(doc.Document)
	if err != nil {
		return nil, fmt.Errorf("fail to unmarshal ld document bytes: %w", err)
	}

	wrappedDoc.Document = string(documentBytes)

	return wrappedDoc, nil
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

func createCredentialOfferIssuanceURI(t *testing.T, issuerURL string, includeAuthCodeGrant bool) string {
	t.Helper()

	credentialOffer := createCredentialOffer(t, issuerURL, includeAuthCodeGrant)

	credentialOfferBytes, err := json.Marshal(credentialOffer)
	require.NoError(t, err)

	credentialOfferEscaped := url.QueryEscape(string(credentialOfferBytes))

	return "openid-credential-offer://?credential_offer=" + credentialOfferEscaped
}

func createCredentialOffer(t *testing.T, issuerURL string, includeAuthCodeGrant bool) *goapiopenid4ci.CredentialOffer {
	t.Helper()

	credentialOffer := createSampleCredentialOffer(t, includeAuthCodeGrant)

	credentialOffer.CredentialIssuer = issuerURL

	return credentialOffer
}

func createSampleCredentialOffer(t *testing.T, includeAuthCodeGrant bool) *goapiopenid4ci.CredentialOffer {
	t.Helper()

	var credentialOffer goapiopenid4ci.CredentialOffer

	err := json.Unmarshal(sampleCredentialOffer, &credentialOffer)
	require.NoError(t, err)

	if includeAuthCodeGrant {
		credentialOffer.Grants["authorization_code"] = map[string]interface{}{"issuer_state": "1234"}
	}

	return &credentialOffer
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
