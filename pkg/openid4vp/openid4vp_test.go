/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp //nolint: testpackage

import (
	"crypto/ed25519"
	"crypto/rand"
	_ "embed" //nolint:gci // required for go:embed
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/did-go/doc/did"
	"github.com/trustbloc/did-go/doc/did/endpoint"
	"github.com/trustbloc/kms-go/doc/jose"
	"github.com/trustbloc/kms-go/doc/jose/jwk"
	"github.com/trustbloc/kms-go/doc/util/jwkkid"
	"github.com/trustbloc/kms-go/spi/kms"
	"github.com/trustbloc/vc-go/jwt"
	"github.com/trustbloc/vc-go/presexch"
	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/internal/testutil"
	"github.com/trustbloc/wallet-sdk/pkg/activitylogger/noop"
	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	"github.com/trustbloc/wallet-sdk/pkg/internal/mock"
	"github.com/trustbloc/wallet-sdk/pkg/models"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

var (
	//go:embed test_data/request_object.jwt
	requestObjectJWT string

	//go:embed test_data/request_object_ldp_vp.jwt
	requestObjectJWTLdpVP string

	//go:embed test_data/presentation.jsonld
	presentationJSONLD []byte

	//go:embed test_data/credentials.jsonld
	credentialsJSONLD []byte

	//go:embed test_data/attestation_cred.jwt
	attestationCredJWT string
)

const (
	testSignature = "test signature"
	mockDID       = "did:example:12345"
	mockVMID      = "#key-1"
	verifierDID   = "did:test:acde"
)

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

func TestNewInteraction(t *testing.T) {
	t.Run("Inline Request Object", func(t *testing.T) {
		interaction, err := NewInteraction(requestObjectJWT, &jwtSignatureVerifierMock{}, nil, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, interaction)
	})

	t.Run("Fetch Request Object", func(t *testing.T) {
		t.Run("request_uri is not URL-encoded", func(t *testing.T) {
			interaction, err := NewInteraction(
				"openid-vc://?request_uri=https://request-object",
				&jwtSignatureVerifierMock{},
				nil,
				nil,
				nil,
				WithHTTPClient(&mock.HTTPClientMock{
					Response:         requestObjectJWT,
					StatusCode:       200,
					ExpectedEndpoint: "https://request-object",
				}),
			)
			require.NoError(t, err)
			require.NotNil(t, interaction)
		})
		t.Run("request_uri is URL-encoded", func(t *testing.T) {
			interaction, err := NewInteraction(
				"openid-vc://?request_uri=https%3A%2F%2Frequest-object",
				&jwtSignatureVerifierMock{},
				nil,
				nil,
				nil,
				WithHTTPClient(&mock.HTTPClientMock{
					Response:         requestObjectJWT,
					StatusCode:       200,
					ExpectedEndpoint: "https://request-object",
				}),
			)
			require.NoError(t, err)
			require.NotNil(t, interaction)
		})
		t.Run("openid4vp protocol with redirect_uri client id scheme", func(t *testing.T) {
			reqObject := &requestObject{
				ClientIDScheme: redirectURIScheme,
				ResponseURI:    "https://example.com/redirect?query=param",
			}

			token, err := jwt.NewUnsecured(reqObject)
			require.NoError(t, err)

			reqObjectJWT, err := token.Serialize(false)
			require.NoError(t, err)

			interaction, err := NewInteraction("openid4vp://authorize?client_id=https://example.com/redirect&"+
				"request_uri=https://example.com/request-object",
				&jwtSignatureVerifierMock{},
				nil,
				nil,
				nil,
				WithHTTPClient(&mock.HTTPClientMock{
					Response:         reqObjectJWT,
					StatusCode:       200,
					ExpectedEndpoint: "https://example.com/request-object",
				}),
			)
			require.NoError(t, err)
			require.NotNil(t, interaction)
		})
		t.Run("openid4vp protocol with unsupported client id scheme", func(t *testing.T) {
			reqObject := &requestObject{
				ClientIDScheme: "unsupported_scheme",
				ResponseURI:    "https://example.com/redirect?query=param",
			}

			token, err := jwt.NewUnsecured(reqObject)
			require.NoError(t, err)

			reqObjectJWT, err := token.Serialize(false)
			require.NoError(t, err)

			interaction, err := NewInteraction("openid4vp://authorize?client_id=https://example.com/redirect&"+
				"request_uri=https://example.com/request-object",
				&jwtSignatureVerifierMock{},
				nil,
				nil,
				nil,
				WithHTTPClient(&mock.HTTPClientMock{
					Response:         reqObjectJWT,
					StatusCode:       200,
					ExpectedEndpoint: "https://example.com/request-object",
				}),
			)
			require.Error(t, err)
			require.ErrorContains(t, err, "unsupported client_id_scheme")
			require.Nil(t, interaction)
		})
		t.Run("client_id mismatch between authorization request and request object", func(t *testing.T) {
			reqObject := &requestObject{
				ClientIDScheme: redirectURIScheme,
				ResponseURI:    "https://invalid.example.com/redirect",
			}

			token, err := jwt.NewUnsecured(reqObject)
			require.NoError(t, err)

			reqObjectJWT, err := token.Serialize(false)
			require.NoError(t, err)

			interaction, err := NewInteraction("openid4vp://authorize?client_id=https://example.com/redirect&"+
				"request_uri=https://example.com/request-object",
				&jwtSignatureVerifierMock{},
				nil,
				nil,
				nil,
				WithHTTPClient(&mock.HTTPClientMock{
					Response:         reqObjectJWT,
					StatusCode:       200,
					ExpectedEndpoint: "https://example.com/request-object",
				}),
			)
			require.ErrorContains(t, err, "client_id mismatch between authorization request and request object")
			require.Nil(t, interaction)
		})
	})

	t.Run("Fetch Request failed", func(t *testing.T) {
		t.Run("HTTP call error", func(t *testing.T) {
			interaction, err := NewInteraction(
				"openid-vc://?request_uri=https://request-object",
				&jwtSignatureVerifierMock{},
				nil,
				nil,
				nil,
				WithHTTPClient(&mock.HTTPClientMock{
					Err: errors.New("http error"),
				}),
			)
			require.Contains(t, err.Error(), "http error")
			require.Nil(t, interaction)
		})
		t.Run("URL parsing error", func(t *testing.T) {
			interaction, err := NewInteraction(
				"openid-vc://%",
				&jwtSignatureVerifierMock{},
				nil,
				nil,
				nil,
				nil,
			)
			testutil.RequireErrorContains(t, err, "INVALID_AUTHORIZATION_REQUEST")
			testutil.RequireErrorContains(t, err, "invalid URL escape")
			require.Nil(t, interaction)
		})
		t.Run("URI missing request_uri parameter", func(t *testing.T) {
			interaction, err := NewInteraction(
				"openid-vc://",
				&jwtSignatureVerifierMock{},
				nil,
				nil,
				nil,
				nil,
			)
			testutil.RequireErrorContains(t, err, "INVALID_AUTHORIZATION_REQUEST")
			testutil.RequireErrorContains(t, err, "request_uri missing from authorization request URI")
			require.Nil(t, interaction)
		})
	})

	t.Run("Inline Request Object", func(t *testing.T) {
		interaction, err := NewInteraction(requestObjectJWT, &jwtSignatureVerifierMock{
			err: errors.New("sig verification err"),
		}, nil, nil, nil, nil)
		require.Contains(t, err.Error(), "sig verification err")
		require.Nil(t, interaction)
	})

	t.Run("Fail to log retrieve request object via HTTP GET metrics event", func(t *testing.T) {
		interaction, err := NewInteraction("openid-vc://?request_uri=https://request-object",
			&jwtSignatureVerifierMock{},
			nil,
			nil,
			testutil.DocumentLoader(t),
			WithHTTPClient(&mock.HTTPClientMock{
				Response:         requestObjectJWT,
				StatusCode:       200,
				ExpectedEndpoint: "https://request-object",
			}),
			WithMetricsLogger(&failingMetricsLogger{}),
		)
		require.EqualError(t, err, "REQUEST_OBJECT_FETCH_FAILED(OVP1-0001):fetch request object: "+
			"failed to log event (Event=Fetch request object via an HTTP GET request to https://request-object)")
		require.Nil(t, interaction)
	})
}

func TestOpenID4VP_PresentCredential(t *testing.T) {
	lddl := testutil.DocumentLoader(t)

	presentation, presErr := verifiable.ParsePresentation(presentationJSONLD,
		verifiable.WithPresDisabledProofCheck(),
		verifiable.WithPresJSONLDDocumentLoader(lddl))

	require.NoError(t, presErr)
	require.NotNil(t, presentation)

	var credentials, singleCred []*verifiable.Credential

	var rawCreds []json.RawMessage

	require.NoError(t, json.Unmarshal(credentialsJSONLD, &rawCreds))

	for _, credBytes := range rawCreds {
		cred, credErr := verifiable.ParseCredential(
			credBytes,
			verifiable.WithDisabledProofCheck(),
			verifiable.WithJSONLDDocumentLoader(lddl),
		)
		require.NoError(t, credErr)

		credentials = append(credentials, cred)
	}

	singleCred = append(singleCred, credentials[0])

	mockDoc := mockResolution(t, mockDID, false)

	mockPresentationDefinition := &presexch.PresentationDefinition{
		ID: uuid.NewString(),
		InputDescriptors: []*presexch.InputDescriptor{
			{
				ID: uuid.NewString(),
				Schema: []*presexch.Schema{
					{
						URI: fmt.Sprintf("%s#%s", verifiable.V1ContextID, verifiable.VCType),
					},
				},
			},
		},
	}

	mockRequestObject := &requestObject{
		Nonce:                  "test123456",
		State:                  "test34566",
		PresentationDefinition: mockPresentationDefinition,
		ResponseType:           "vp_token id_token",
	}

	t.Run("Success", func(t *testing.T) {
		httpClient := &mock.HTTPClientMock{
			StatusCode: 200,
		}

		interaction, err := NewInteraction(
			requestObjectJWT,
			&jwtSignatureVerifierMock{},
			&didResolverMock{ResolveValue: mockDoc},
			&cryptoMock{SignVal: []byte(testSignature)},
			lddl,
			WithHTTPClient(httpClient),
			WithActivityLogger(noop.NewActivityLogger()),
		)
		require.NoError(t, err)

		query := interaction.GetQuery()
		require.NotNil(t, query)

		displayData := interaction.VerifierDisplayData()

		require.Equal(t, verifierDID, displayData.DID)
		require.Equal(t, "v_myprofile_jwt", displayData.Name)
		require.Equal(t, "test verifier", displayData.Purpose)
		require.Equal(t, "https://example.com/verifier/logo", displayData.LogoURI)

		err = interaction.PresentCredential(credentials, CustomClaims{})

		require.NoError(t, err)

		expectedState := "636df28459a07d50cc4b657e"
		expectedSig := base64.RawURLEncoding.EncodeToString([]byte(testSignature))

		require.Contains(t, string(httpClient.SentBody), expectedState)
		require.Contains(t, string(httpClient.SentBody), expectedSig)

		// TODO: https://github.com/trustbloc/wallet-sdk/issues/459 refactor this into validation helper functions
		data, err := url.ParseQuery(string(httpClient.SentBody))
		require.NoError(t, err)

		require.Contains(t, data, "id_token")
		require.NotEmpty(t, data["id_token"])

		require.Contains(t, data, "vp_token")
		require.NotEmpty(t, data["vp_token"])

		require.Contains(t, data, "presentation_submission")
		require.NotEmpty(t, data["presentation_submission"])

		require.NotContains(t, data, "interaction_details")

		var presentationSubmission *presexch.PresentationSubmission

		require.NoError(t, json.Unmarshal([]byte(data["presentation_submission"][0]), &presentationSubmission))

		var vpTokenList []string

		require.NoError(t, json.Unmarshal([]byte(data["vp_token"][0]), &vpTokenList))

		var presentations []*verifiable.Presentation

		for _, s := range vpTokenList {
			parsedPresentation, e := verifiable.ParsePresentation(
				[]byte(s),
				verifiable.WithPresDisabledProofCheck(),
				verifiable.WithDisabledJSONLDChecks())
			require.NoError(t, e)

			presentations = append(presentations, parsedPresentation)
		}

		_, err = interaction.requestObject.PresentationDefinition.Match(
			presentations,
			lddl,
			presexch.WithDisableSchemaValidation(),
			presexch.WithMergedSubmission(presentationSubmission),
			presexch.WithCredentialOptions(verifiable.WithDisabledProofCheck(), verifiable.WithJSONLDDocumentLoader(lddl)),
		)
		require.NoError(t, err)

		ack := interaction.Acknowledgment()
		require.Equal(t, "test://response", ack.ResponseURI)
	})

	t.Run("Success - Unsafe", func(t *testing.T) {
		httpClient := &mock.HTTPClientMock{
			StatusCode: 200,
		}

		interaction, err := NewInteraction(
			requestObjectJWT,
			&jwtSignatureVerifierMock{},
			&didResolverMock{ResolveValue: mockDoc},
			&cryptoMock{SignVal: []byte(testSignature)},
			lddl,
			WithHTTPClient(httpClient),
		)
		require.NoError(t, err)

		query := interaction.GetQuery()
		require.NotNil(t, query)

		err = interaction.PresentCredentialUnsafe(singleCred[0], CustomClaims{})
		require.NoError(t, err)
	})

	t.Run("Success - with opts", func(t *testing.T) {
		httpClient := &mock.HTTPClientMock{
			StatusCode: 200,
		}

		crypto := &cryptoMock{SignVal: []byte(testSignature)}

		interaction, err := NewInteraction(
			requestObjectJWT,
			&jwtSignatureVerifierMock{},
			&didResolverMock{ResolveValue: mockDoc},
			crypto,
			lddl,
			WithHTTPClient(httpClient),
		)
		require.NoError(t, err)

		verificationMethod := mockDoc.DIDDocument.VerificationMethod[0]

		attestationSigner, err := common.NewJWSSigner(models.VerificationMethodFromDoc(&verificationMethod), crypto)
		require.NoError(t, err)

		query := interaction.GetQuery()
		require.NotNil(t, query)

		err = interaction.PresentCredential(singleCred, CustomClaims{},
			WithAttestationVC(attestationSigner, attestationCredJWT),
			WithInteractionDetails(map[string]interface{}{"key1": "value1", "key2": "value2"}),
		)
		require.NoError(t, err)

		// TODO: https://github.com/trustbloc/wallet-sdk/issues/459 refactor this into validation helper functions
		data, err := url.ParseQuery(string(httpClient.SentBody))
		require.NoError(t, err)

		require.Contains(t, data, "id_token")
		require.NotEmpty(t, data["id_token"])

		require.Contains(t, data, "vp_token")
		require.NotEmpty(t, data["vp_token"])

		require.Contains(t, data, "presentation_submission")
		require.NotEmpty(t, data["presentation_submission"])

		require.Contains(t, data, "interaction_details")

		interactionDetailsRaw, err := base64.StdEncoding.DecodeString(data["interaction_details"][0])
		require.NoError(t, err)

		var interactionDetails map[string]interface{}
		err = json.Unmarshal(interactionDetailsRaw, &interactionDetails)
		require.NoError(t, err)

		require.Equal(t, map[string]interface{}{"key1": "value1", "key2": "value2"}, interactionDetails)
	})

	t.Run("Failure - with opts", func(t *testing.T) {
		httpClient := &mock.HTTPClientMock{
			StatusCode: 200,
		}

		crypto := &cryptoMock{SignVal: []byte(testSignature)}

		interaction, err := NewInteraction(
			requestObjectJWT,
			&jwtSignatureVerifierMock{},
			&didResolverMock{ResolveValue: mockDoc},
			crypto,
			lddl,
			WithHTTPClient(httpClient),
		)
		require.NoError(t, err)

		verificationMethod := mockDoc.DIDDocument.VerificationMethod[0]

		attestationSigner, err := common.NewJWSSigner(models.VerificationMethodFromDoc(&verificationMethod), crypto)
		require.NoError(t, err)

		query := interaction.GetQuery()
		require.NotNil(t, query)

		err = interaction.PresentCredential(singleCred, CustomClaims{},
			WithAttestationVC(attestationSigner, "invalid_cred"),
			WithInteractionDetails(map[string]any{"key1": "value1", "key2": "value2"}),
		)
		require.Error(t, err)
		require.ErrorContains(t, err, "unsupported credential format")
	})

	t.Run("Success - with interaction details opts", func(t *testing.T) {
		httpClient := &mock.HTTPClientMock{
			StatusCode: 200,
		}

		crypto := &cryptoMock{SignVal: []byte(testSignature)}

		interaction, err := NewInteraction(
			requestObjectJWT,
			&jwtSignatureVerifierMock{},
			&didResolverMock{ResolveValue: mockDoc},
			crypto,
			lddl,
			WithHTTPClient(httpClient),
		)
		require.NoError(t, err)

		verificationMethod := mockDoc.DIDDocument.VerificationMethod[0]

		attestationSigner, err := common.NewJWSSigner(models.VerificationMethodFromDoc(&verificationMethod), crypto)
		require.NoError(t, err)

		query := interaction.GetQuery()
		require.NotNil(t, query)

		err = interaction.PresentCredential(singleCred, CustomClaims{},
			WithAttestationVC(attestationSigner, attestationCredJWT),
			WithInteractionDetails(map[string]interface{}{"key1": "value1", "key2": func() {}}),
		)
		require.Error(t, err)
		require.ErrorContains(t, err, "encode interaction details")
	})

	t.Run("Success - with opts, multi cred", func(t *testing.T) {
		httpClient := &mock.HTTPClientMock{
			StatusCode: 200,
		}

		crypto := &cryptoMock{SignVal: []byte(testSignature)}

		interaction, err := NewInteraction(
			requestObjectJWT,
			&jwtSignatureVerifierMock{},
			&didResolverMock{ResolveValue: mockDoc},
			crypto,
			lddl,
			WithHTTPClient(httpClient),
		)
		require.NoError(t, err)

		verificationMethod := mockDoc.DIDDocument.VerificationMethod[0]

		attestationSigner, err := common.NewJWSSigner(models.VerificationMethodFromDoc(&verificationMethod), crypto)
		require.NoError(t, err)

		query := interaction.GetQuery()
		require.NotNil(t, query)

		err = interaction.PresentCredential(credentials, CustomClaims{},
			WithAttestationVC(attestationSigner, attestationCredJWT))
		require.NoError(t, err)
	})

	t.Run("Success - with ldp_vp, single cred", func(t *testing.T) {
		mockHTTPClient := &mock.HTTPClientMock{
			StatusCode: 200,
		}

		crypto := &cryptoMock{SignVal: []byte(testSignature)}

		interaction, err := NewInteraction(
			requestObjectJWTLdpVP,
			&jwtSignatureVerifierMock{},
			&didResolverMock{ResolveValue: mockResolution(t, mockDID, true)},
			crypto,
			lddl,
			WithHTTPClient(mockHTTPClient),
		)
		require.NoError(t, err)

		query := interaction.GetQuery()
		require.NotNil(t, query)

		err = interaction.PresentCredential(singleCred, CustomClaims{})
		require.NoError(t, err)

		data, err := url.ParseQuery(string(mockHTTPClient.SentBody))
		require.NoError(t, err)

		require.Contains(t, data, "id_token")
		require.NotEmpty(t, data["id_token"])

		require.Contains(t, data, "vp_token")
		require.NotEmpty(t, data["vp_token"])

		require.Contains(t, data, "presentation_submission")
		require.NotEmpty(t, data["presentation_submission"])
	})

	t.Run("Failure - with metrics logger error", func(t *testing.T) {
		mockHTTPClient := &mock.HTTPClientMock{
			StatusCode: 200,
		}

		crypto := &cryptoMock{SignVal: []byte(testSignature)}

		interaction, err := NewInteraction(
			requestObjectJWTLdpVP,
			&jwtSignatureVerifierMock{},
			&didResolverMock{ResolveValue: mockResolution(t, mockDID, true)},
			crypto,
			lddl,
			WithHTTPClient(mockHTTPClient),
			WithMetricsLogger(&failingMetricsLogger{attemptFailNumber: 1}),
		)
		require.NoError(t, err)

		query := interaction.GetQuery()
		require.NotNil(t, query)

		err = interaction.PresentCredential(singleCred, CustomClaims{})
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to log event")
	})

	t.Run("Success - with ldp_vp, multi cred", func(t *testing.T) {
		mockHTTPClient := &mock.HTTPClientMock{
			StatusCode: 200,
		}

		crypto := &cryptoMock{SignVal: []byte(testSignature)}

		var ldpCredentials []*verifiable.Credential

		for _, cred := range credentials {
			if cred.IsJWT() {
				continue
			}

			ldpCredentials = append(ldpCredentials, cred)
		}

		interaction, err := NewInteraction(
			requestObjectJWTLdpVP,
			&jwtSignatureVerifierMock{},
			&didResolverMock{ResolveValue: mockResolution(t, mockDID, true)},
			crypto,
			lddl,
			WithHTTPClient(mockHTTPClient),
		)
		require.NoError(t, err)

		query := interaction.GetQuery()
		require.NotNil(t, query)

		err = interaction.PresentCredential(ldpCredentials, CustomClaims{})
		require.NoError(t, err)

		data, err := url.ParseQuery(string(mockHTTPClient.SentBody))
		require.NoError(t, err)

		require.Contains(t, data, "id_token")
		require.NotEmpty(t, data["id_token"])

		require.Contains(t, data, "vp_token")
		require.NotEmpty(t, data["vp_token"])

		require.Contains(t, data, "presentation_submission")
		require.NotEmpty(t, data["presentation_submission"])
	})

	t.Run("Check custom claims", func(t *testing.T) {
		response, err := createAuthorizedResponse(
			singleCred,
			mockRequestObject,
			CustomClaims{ScopeClaims: map[string]interface{}{
				"customClaimName": "customClaimValue",
			}},
			&didResolverMock{ResolveValue: mockDoc},
			&cryptoMock{SignVal: []byte(testSignature)},
			lddl,
			nil,
		)
		require.NoError(t, err)

		body, err := base64.RawURLEncoding.DecodeString(strings.Split(response.IDTokenJWS, ".")[1])
		require.NoError(t, err)

		require.Contains(t, string(body), "customClaimName")
		require.Contains(t, string(body), "customClaimValue")
	})

	t.Run("Check nonce", func(t *testing.T) {
		response, err := createAuthorizedResponse(
			singleCred,
			mockRequestObject,
			CustomClaims{},
			&didResolverMock{ResolveValue: mockDoc},
			&cryptoMock{SignVal: []byte(testSignature)},
			lddl,
			nil,
		)

		require.NoError(t, err)
		require.Equal(t, "test34566", response.State)

		idToken, err := base64.RawURLEncoding.DecodeString(strings.Split(response.IDTokenJWS, ".")[1])
		require.NoError(t, err)
		require.Contains(t, string(idToken), "test123456")

		vpToken, err := base64.RawURLEncoding.DecodeString(strings.Split(response.VPToken, ".")[1])
		require.NoError(t, err)
		require.Contains(t, string(vpToken), "test123456")
	})

	t.Run("unsafe skip constraint validation", func(t *testing.T) {
		strType := "string"
		required := presexch.Required

		pd := &presexch.PresentationDefinition{
			ID: uuid.NewString(),
			InputDescriptors: []*presexch.InputDescriptor{{
				ID: uuid.NewString(),
				Schema: []*presexch.Schema{{
					URI: fmt.Sprintf("%s#%s", verifiable.V1ContextID, verifiable.VCType),
				}},
				// These constraints aren't satisfied by the provided VC...
				Constraints: &presexch.Constraints{
					LimitDisclosure: &required,
					Fields: []*presexch.Field{{
						Path: []string{"$.credentialSubject.taxResidency", "$.vc.credentialSubject.taxResidency"},
						Filter: &presexch.Filter{
							FilterItem: presexch.FilterItem{
								Type: &strType,
							},
						},
					}},
				},
			}},
		}

		req := &requestObject{
			Nonce:                  "test123456",
			State:                  "test34566",
			PresentationDefinition: pd,
		}

		// ...so creating a VP fails...
		response, err := createAuthorizedResponse(
			singleCred,
			req,
			CustomClaims{},
			&didResolverMock{ResolveValue: mockDoc},
			&cryptoMock{SignVal: []byte(testSignature)},
			lddl,
			nil,
		)

		require.Nil(t, response)
		require.ErrorIs(t, err, presexch.ErrNoCredentials)

		// ...but creating a VP without constraint validation succeeds...
		response, err = createAuthorizedResponse(
			singleCred,
			req,
			CustomClaims{},
			&didResolverMock{ResolveValue: mockDoc},
			&cryptoMock{SignVal: []byte(testSignature)},
			lddl,
			&presentOpts{
				ignoreConstraints: true,
			},
		)

		require.NoError(t, err)
		require.Equal(t, "test34566", response.State)
	})

	t.Run("no credentials provided", func(t *testing.T) {
		interaction, err := NewInteraction(
			requestObjectJWT,
			&jwtSignatureVerifierMock{},
			&didResolverMock{ResolveValue: mockDoc},
			&cryptoMock{SignVal: []byte(testSignature)},
			lddl,
			WithHTTPClient(&mock.HTTPClientMock{
				StatusCode: 200,
			}),
		)
		require.NoError(t, err)

		query := interaction.GetQuery()
		require.NotNil(t, query)

		err = interaction.PresentCredential(nil, CustomClaims{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "expected at least one credential")
	})

	t.Run("no subject ID found", func(t *testing.T) {
		testCases := []struct {
			vc []verifiable.CredentialContents
		}{
			{
				vc: []verifiable.CredentialContents{
					{
						ID:      "foo",
						Context: []string{verifiable.V1ContextURI},
						Types:   []string{verifiable.VCType},
					},
				},
			},
			{
				vc: []verifiable.CredentialContents{
					{
						ID:      "foo",
						Context: []string{verifiable.V1ContextURI},
						Types:   []string{verifiable.VCType},
					},
					{
						ID:      "bar",
						Context: []string{verifiable.V1ContextURI},
						Types:   []string{verifiable.VCType},
					},
				},
			},
		}

		for _, testCase := range testCases {
			t.Run("", func(t *testing.T) {
				var vcs []*verifiable.Credential

				for _, vcc := range testCase.vc {
					vc, err := verifiable.CreateCredential(vcc, nil)
					require.NoError(t, err)

					vcs = append(vcs, vc)
				}

				_, err := createAuthorizedResponse(
					vcs,
					mockRequestObject,
					CustomClaims{},
					&didResolverMock{ResolveValue: mockDoc},
					&cryptoMock{},
					lddl,
					&presentOpts{},
				)

				require.Error(t, err)
				require.Contains(t, err.Error(), "VC does not have a subject ID")
			})
		}
	})

	t.Run("fail to resolve signing DID", func(t *testing.T) {
		expectErr := errors.New("resolve failed")

		_, err := createAuthorizedResponse(singleCred, mockRequestObject, CustomClaims{},
			&didResolverMock{ResolveErr: expectErr}, &cryptoMock{}, lddl, &presentOpts{})
		require.ErrorIs(t, err, expectErr)
	})

	t.Run("signing DID has no signing key", func(t *testing.T) {
		_, err := createAuthorizedResponse(singleCred, mockRequestObject, CustomClaims{},
			&didResolverMock{ResolveValue: &did.DocResolution{
				DIDDocument: &did.Doc{},
			}}, &cryptoMock{}, lddl, nil)

		require.Error(t, err)
		require.Contains(t, err.Error(), "no assertion method for signing")
	})

	t.Run("sign failed", func(t *testing.T) {
		expectErr := errors.New("sign failed")

		_, err := createAuthorizedResponse(
			singleCred,
			mockRequestObject,
			CustomClaims{},
			&didResolverMock{ResolveValue: mockDoc},
			&cryptoMock{SignErr: expectErr},
			lddl,
			nil,
		)

		require.ErrorIs(t, err, expectErr)
	})

	t.Run("fail to add data integrity proof", func(t *testing.T) {
		reqObject := &requestObject{
			Nonce:                  "test123456",
			State:                  "test34566",
			PresentationDefinition: mockPresentationDefinition,
			ResponseType:           "vp_token id_token",
			ClientMetadata:         clientMetadata{VPFormats: &presexch.Format{LdpVP: &presexch.LdpType{}}},
		}

		t.Run("single credential", func(t *testing.T) {
			_, err := createAuthorizedResponse(
				singleCred,
				reqObject,
				CustomClaims{},
				&didResolverMock{ResolveValue: mockDoc},
				&cryptoMock{},
				lddl,
				&presentOpts{},
			)
			require.ErrorContains(t, err, "no supported ldp types found")
		})
	})

	t.Run("fail to send authorized response", func(t *testing.T) {
		httpClient := &mock.HTTPClientMock{
			StatusCode: http.StatusInternalServerError,
		}

		interaction, err := NewInteraction(
			requestObjectJWT,
			&jwtSignatureVerifierMock{},
			&didResolverMock{ResolveValue: mockDoc},
			&cryptoMock{SignVal: []byte(testSignature)},
			lddl,
			WithHTTPClient(httpClient),
		)
		require.NoError(t, err)

		query := interaction.GetQuery()
		require.NotNil(t, query)

		t.Run("Invalid scope", func(t *testing.T) {
			errResponse := &errorResponse{Error: "invalid_scope"}

			errResponseBytes, errMarshal := json.Marshal(errResponse)
			require.NoError(t, errMarshal)

			httpClient.Response = string(errResponseBytes)

			err = interaction.PresentCredential(credentials, CustomClaims{})
			testutil.RequireErrorContains(t, err, InvalidScopeError)
		})

		t.Run("Invalid request", func(t *testing.T) {
			errResponse := &errorResponse{Error: "invalid_request"}

			errResponseBytes, errMarshal := json.Marshal(errResponse)
			require.NoError(t, errMarshal)

			httpClient.Response = string(errResponseBytes)

			err = interaction.PresentCredential(credentials, CustomClaims{})
			testutil.RequireErrorContains(t, err, InvalidRequestError)
		})

		t.Run("Invalid client", func(t *testing.T) {
			errResponse := &errorResponse{Error: "invalid_client"}

			errResponseBytes, errMarshal := json.Marshal(errResponse)
			require.NoError(t, errMarshal)

			httpClient.Response = string(errResponseBytes)

			err = interaction.PresentCredential(credentials, CustomClaims{})
			testutil.RequireErrorContains(t, err, InvalidClientError)
		})

		t.Run("VP formats not supported", func(t *testing.T) {
			errResponse := &errorResponse{Error: "vp_formats_not_supported"}

			errResponseBytes, errMarshal := json.Marshal(errResponse)
			require.NoError(t, errMarshal)

			httpClient.Response = string(errResponseBytes)

			err = interaction.PresentCredential(credentials, CustomClaims{})
			testutil.RequireErrorContains(t, err, VPFormatsNotSupportedError)
		})

		t.Run("Invalid presentation definition URI", func(t *testing.T) {
			errResponse := &errorResponse{Error: "invalid_presentation_definition_uri"}

			errResponseBytes, errMarshal := json.Marshal(errResponse)
			require.NoError(t, errMarshal)

			httpClient.Response = string(errResponseBytes)

			err = interaction.PresentCredential(credentials, CustomClaims{})
			testutil.RequireErrorContains(t, err, InvalidPresentationDefinitionURIError)
		})

		t.Run("Invalid presentation definition reference", func(t *testing.T) {
			errResponse := &errorResponse{Error: "invalid_presentation_definition_reference"}

			errResponseBytes, errMarshal := json.Marshal(errResponse)
			require.NoError(t, errMarshal)

			httpClient.Response = string(errResponseBytes)

			err = interaction.PresentCredential(credentials, CustomClaims{})
			testutil.RequireErrorContains(t, err, InvalidPresentationDefinitionReferenceError)
		})

		t.Run("Unknown/other error type in errorResponse object", func(t *testing.T) {
			errResponse := &errorResponse{Error: "other"}

			errResponseBytes, errMarshal := json.Marshal(errResponse)
			require.NoError(t, errMarshal)

			httpClient.Response = string(errResponseBytes)

			err = interaction.PresentCredential(credentials, CustomClaims{})
			testutil.RequireErrorContains(t, err, OtherAuthorizationResponseError)
		})

		t.Run("MS Entra error response format", func(t *testing.T) {
			t.Run("Bad or missing field", func(t *testing.T) {
				errResponse := &msEntraErrorResponse{Error: errorInfo{InnerError: innerError{
					Code:    "badOrMissingField",
					Message: "message",
				}}}

				errResponseBytes, errMarshal := json.Marshal(errResponse)
				require.NoError(t, errMarshal)

				httpClient.Response = string(errResponseBytes)

				err = interaction.PresentCredential(credentials, CustomClaims{})
				require.Error(t, err)

				var walletError *walleterror.Error

				require.ErrorAs(t, err, &walletError)
				require.Equal(t, MSEntraBadOrMissingFieldsError, walletError.Category)
				require.Equal(t, "message", walletError.Message)
			})
			t.Run("Not found", func(t *testing.T) {
				errResponse := &msEntraErrorResponse{Error: errorInfo{InnerError: innerError{Code: "notFound"}}}

				errResponseBytes, errMarshal := json.Marshal(errResponse)
				require.NoError(t, errMarshal)

				httpClient.Response = string(errResponseBytes)

				err = interaction.PresentCredential(credentials, CustomClaims{})
				testutil.RequireErrorContains(t, err, MSEntraNotFoundError)
			})
			t.Run("Token error", func(t *testing.T) {
				errResponse := &msEntraErrorResponse{Error: errorInfo{InnerError: innerError{Code: "tokenError"}}}

				errResponseBytes, errMarshal := json.Marshal(errResponse)
				require.NoError(t, errMarshal)

				httpClient.Response = string(errResponseBytes)

				err = interaction.PresentCredential(credentials, CustomClaims{})
				testutil.RequireErrorContains(t, err, MSEntraTokenError)
			})
			t.Run("Transient error", func(t *testing.T) {
				errResponse := &msEntraErrorResponse{Error: errorInfo{InnerError: innerError{Code: "transientError"}}}

				errResponseBytes, errMarshal := json.Marshal(errResponse)
				require.NoError(t, errMarshal)

				httpClient.Response = string(errResponseBytes)

				err = interaction.PresentCredential(credentials, CustomClaims{})
				testutil.RequireErrorContains(t, err, MSEntraTransientError)
			})
			t.Run("Unknown/other error type in msEntraErrorResponse object", func(t *testing.T) {
				errResponse := &msEntraErrorResponse{Error: errorInfo{InnerError: innerError{Code: "other"}}}

				errResponseBytes, errMarshal := json.Marshal(errResponse)
				require.NoError(t, errMarshal)

				httpClient.Response = string(errResponseBytes)

				err = interaction.PresentCredential(credentials, CustomClaims{})
				testutil.RequireErrorContains(t, err, OtherAuthorizationResponseError)
			})
		})

		t.Run("Response body is neither an errorResponse nor an msEntraErrorResponse object", func(t *testing.T) {
			httpClient.Response = ""

			err = interaction.PresentCredential(credentials, CustomClaims{})
			testutil.RequireErrorContains(t, err, OtherAuthorizationResponseError)
		})
	})
}

func TestOpenID4VP_TrustInfo(t *testing.T) {
	lddl := testutil.DocumentLoader(t)

	t.Run("Success", func(t *testing.T) {
		mockDoc := mockResolutionWithServices(t, mockDID)
		httpClient := &mock.HTTPClientMock{
			StatusCode: 200,
		}

		interaction, err := NewInteraction(
			requestObjectJWT,
			&jwtSignatureVerifierMock{},
			&didResolverMock{ResolveValue: mockDoc},
			&cryptoMock{SignVal: []byte(testSignature)},
			lddl,
			WithHTTPClient(httpClient),
		)
		require.NoError(t, err)

		info, err := interaction.TrustInfo()
		require.NoError(t, err)
		require.NotNil(t, info)
		require.Equal(t, "mock-uri", info.Domain)
	})

	t.Run("Success: origin-based trust info", func(t *testing.T) {
		reqObject := &requestObject{
			ClientIDScheme: redirectURIScheme,
			ResponseURI:    "https://example.com/redirect",
		}

		token, err := jwt.NewUnsecured(reqObject)
		require.NoError(t, err)

		reqObjectJWT, err := token.Serialize(false)
		require.NoError(t, err)

		interaction, err := NewInteraction("openid4vp://authorize?client_id=https://example.com/redirect&"+
			"request_uri=https://example.com/request-object",
			&jwtSignatureVerifierMock{},
			&didResolverMock{ResolveValue: mockResolutionWithServices(t, mockDID)},
			&cryptoMock{SignVal: []byte(testSignature)},
			lddl,
			WithHTTPClient(&mock.HTTPClientMock{
				Response:         reqObjectJWT,
				StatusCode:       200,
				ExpectedEndpoint: "https://example.com/request-object",
			}),
		)
		require.NoError(t, err)

		info, err := interaction.TrustInfo()
		require.NoError(t, err)
		require.NotNil(t, info)
		require.Equal(t, "example.com", info.Domain)
	})

	t.Run("Failure", func(t *testing.T) {
		httpClient := &mock.HTTPClientMock{
			StatusCode: 200,
		}

		interaction, err := NewInteraction(
			requestObjectJWT,
			&jwtSignatureVerifierMock{},
			nil,
			&cryptoMock{SignVal: []byte(testSignature)},
			lddl,
			WithHTTPClient(httpClient),
		)
		require.NoError(t, err)

		info, err := interaction.TrustInfo()
		require.ErrorContains(t, err, "no resolver provided")
		require.Nil(t, info)
	})

	t.Run("Failure: invalid response URI", func(t *testing.T) {
		reqObject := &requestObject{
			ClientIDScheme: redirectURIScheme,
			ResponseURI:    "https://example.com/redirect",
		}

		token, err := jwt.NewUnsecured(reqObject)
		require.NoError(t, err)

		reqObjectJWT, err := token.Serialize(false)
		require.NoError(t, err)

		interaction, err := NewInteraction("openid4vp://authorize?client_id=https://example.com/redirect&"+
			"request_uri=https://example.com/request-object",
			&jwtSignatureVerifierMock{},
			&didResolverMock{ResolveValue: mockResolutionWithServices(t, mockDID)},
			&cryptoMock{SignVal: []byte(testSignature)},
			lddl,
			WithHTTPClient(&mock.HTTPClientMock{
				Response:         reqObjectJWT,
				StatusCode:       200,
				ExpectedEndpoint: "https://example.com/request-object",
			}),
		)
		require.NoError(t, err)

		interaction.requestObject.ResponseURI = "http://:invalid.uri"

		info, err := interaction.TrustInfo()
		require.Error(t, err)
		require.Nil(t, info)
		require.ErrorContains(t, err, `parse "http://:invalid.uri"`)
	})
}

func TestInteraction_Scope(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		interaction := &Interaction{requestObject: &requestObject{
			Scope: "openid+msregistration",
		}}

		require.Len(t, interaction.CustomScope(), 1)
		require.NotContains(t, interaction.CustomScope(), "openid")
		require.Contains(t, interaction.CustomScope(), "msregistration")
	})
}

func TestResolverAdapter(t *testing.T) {
	mockDoc := mockResolution(t, mockDID, false)
	adapter := wrapResolver(&didResolverMock{ResolveValue: mockDoc})

	doc, err := adapter.Resolve(mockDID)
	require.NoError(t, err)
	require.Equal(t, mockDoc.DIDDocument.ID, doc.DIDDocument.ID)
}

func TestOpenID4VP_PresentedClaims(t *testing.T) {
	lddl := testutil.DocumentLoader(t)

	presentation, presErr := verifiable.ParsePresentation(presentationJSONLD,
		verifiable.WithPresDisabledProofCheck(),
		verifiable.WithPresJSONLDDocumentLoader(lddl))

	require.NoError(t, presErr)
	require.NotNil(t, presentation)

	var credentials []*verifiable.Credential

	var rawCreds []json.RawMessage

	require.NoError(t, json.Unmarshal(credentialsJSONLD, &rawCreds))

	for _, credBytes := range rawCreds {
		cred, credErr := verifiable.ParseCredential(
			credBytes,
			verifiable.WithDisabledProofCheck(),
			verifiable.WithJSONLDDocumentLoader(lddl),
		)
		require.NoError(t, credErr)

		credentials = append(credentials, cred)
	}

	mockDoc := mockResolution(t, mockDID, false)

	t.Run("Success", func(t *testing.T) {
		httpClient := &mock.HTTPClientMock{
			StatusCode: 200,
		}

		interaction, err := NewInteraction(
			requestObjectJWT,
			&jwtSignatureVerifierMock{},
			&didResolverMock{ResolveValue: mockDoc},
			&cryptoMock{SignVal: []byte(testSignature)},
			lddl,
			WithHTTPClient(httpClient),
		)
		require.NoError(t, err)

		query := interaction.GetQuery()
		require.NotNil(t, query)

		displayData := interaction.VerifierDisplayData()

		require.Equal(t, verifierDID, displayData.DID)
		require.Equal(t, "v_myprofile_jwt", displayData.Name)
		require.Equal(t, "test verifier", displayData.Purpose)
		require.Equal(t, "https://example.com/verifier/logo", displayData.LogoURI)

		claims, err := interaction.PresentedClaims(credentials[0])
		require.NoError(t, err)
		require.NotNil(t, claims)

		claimsJSON, err := json.Marshal(copyJSONKeysOnly(claims))
		require.NoError(t, err)
		require.JSONEq(t, `
			{
				"name":{},
				"spouse":{},
				"degree":{
					"degree":{},
					"type":{}
				}
			}
			`, string(claimsJSON))
	})
}

func TestAcknowledgment_AcknowledgeVerifier(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ack := &Acknowledgment{
			ResponseURI: "https://verifier/present",
			State:       "98822a39-9178-4742-a2dc-aba49879fc7b",
		}

		client := &mock.HTTPClientMock{
			ExpectedEndpoint: "https://verifier/present",
			StatusCode:       200,
		}

		err := ack.AcknowledgeVerifier("error", "desc", client)
		require.NoError(t, err)

		data, err := url.ParseQuery(string(client.SentBody))
		require.NoError(t, err)

		require.Equal(t, "error", data["error"][0])
		require.Equal(t, "desc", data["error_description"][0])
		require.Equal(t, "98822a39-9178-4742-a2dc-aba49879fc7b", data["state"][0])

		require.NotContains(t, data, "interaction_details")
	})

	t.Run("Success: with interaction details", func(t *testing.T) {
		ack := &Acknowledgment{
			ResponseURI:        "https://verifier/present",
			State:              "98822a39-9178-4742-a2dc-aba49879fc7b",
			InteractionDetails: map[string]interface{}{"key1": "value1"},
		}

		client := &mock.HTTPClientMock{
			ExpectedEndpoint: "https://verifier/present",
			StatusCode:       200,
		}

		err := ack.AcknowledgeVerifier("error", "desc", client)
		require.NoError(t, err)

		data, err := url.ParseQuery(string(client.SentBody))
		require.NoError(t, err)

		require.Equal(t, "error", data["error"][0])
		require.Equal(t, "desc", data["error_description"][0])
		require.Equal(t, "98822a39-9178-4742-a2dc-aba49879fc7b", data["state"][0])

		require.Contains(t, data, "interaction_details")

		interactionDetailsRaw, err := base64.StdEncoding.DecodeString(data["interaction_details"][0])
		require.NoError(t, err)

		var interactionDetails map[string]interface{}
		err = json.Unmarshal(interactionDetailsRaw, &interactionDetails)
		require.NoError(t, err)

		require.Equal(t, map[string]interface{}{"key1": "value1"}, interactionDetails)
	})

	t.Run("Fail to do request", func(t *testing.T) {
		ack := &Acknowledgment{
			ResponseURI: "https://verifier/present",
			State:       "98822a39-9178-4742-a2dc-aba49879fc7b",
		}

		err := ack.AcknowledgeVerifier("error", "desc",
			&mock.HTTPClientMock{
				Err: errors.New("http error"),
			},
		)

		require.ErrorContains(t, err, "http error")
	})

	t.Run("Unexpected status code", func(t *testing.T) {
		ack := &Acknowledgment{
			ResponseURI: "https://verifier/present",
			State:       "98822a39-9178-4742-a2dc-aba49879fc7b",
		}

		err := ack.AcknowledgeVerifier("error", "desc",
			&mock.HTTPClientMock{
				ExpectedEndpoint: "https://verifier/present",
				StatusCode:       500,
			},
		)

		require.ErrorContains(t, err, "but got status code 500")
	})
}

func TestCopyJSONKeysOnly(t *testing.T) {
	t.Run("Success simple", func(t *testing.T) {
		src := map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		}

		result, err := json.Marshal(copyJSONKeysOnly(src))
		require.NoError(t, err)
		require.JSONEq(t, `{"key1":{},"key2":{}}`, string(result))
	})

	t.Run("Success address", func(t *testing.T) {
		src := map[string]interface{}{
			"name": "test name",
			"address": map[string]interface{}{
				"streetAddress":      "test street",
				"addressCountryCode": "test country",
			},
		}

		result, err := json.Marshal(copyJSONKeysOnly(src))
		require.NoError(t, err)
		require.JSONEq(t, `{"name": {},
        "address": {
          "streetAddress": {},
          "addressCountryCode": {}
        }}`, string(result))
	})

	t.Run("Success array", func(t *testing.T) {
		src := map[string]interface{}{
			"key1": []interface{}{"value1"},
			"key2": "value2",
		}

		result, err := json.Marshal(copyJSONKeysOnly(src))
		require.NoError(t, err)
		require.JSONEq(t, `{"key1":[{}],"key2":{}}`, string(result))
	})
}

type jwtSignatureVerifierMock struct {
	err error
}

func (s *jwtSignatureVerifierMock) CheckJWTProof(jose.Headers, string, []byte, []byte) error {
	return s.err
}

type didResolverMock struct {
	ResolveValue *did.DocResolution
	ResolveErr   error
}

func (d *didResolverMock) Resolve(string) (*did.DocResolution, error) {
	return d.ResolveValue, d.ResolveErr
}

type cryptoMock struct {
	SignVal   []byte
	SignErr   error
	VerifyErr error
}

func (c *cryptoMock) Sign([]byte, string) ([]byte, error) {
	return c.SignVal, c.SignErr
}

func (c *cryptoMock) Verify(_, _ []byte, _ string) error {
	return c.VerifyErr
}

func mockResolution(t *testing.T, mockDID string, useJWK bool) *did.DocResolution { //nolint:unparam
	t.Helper()

	edPub, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	var mockVM *did.VerificationMethod

	if useJWK {
		var key *jwk.JWK

		key, err = jwkkid.BuildJWK(edPub, kms.ED25519Type)
		require.NoError(t, err)

		mockVM, err = did.NewVerificationMethodFromJWK(mockVMID, "JsonWebKey2020", mockDID, key)
		require.NoError(t, err)
	} else {
		mockVM = did.NewVerificationMethodFromBytes(mockVMID, "Ed25519VerificationKey2018", mockDID, edPub)
	}

	docRes := &did.DocResolution{
		DIDDocument: &did.Doc{
			ID:      mockDID,
			Context: []string{did.ContextV1},
			VerificationMethod: []did.VerificationMethod{
				*mockVM,
			},
			AssertionMethod: []did.Verification{
				{
					VerificationMethod: *mockVM,
				},
			},
		},
	}

	return docRes
}

func mockResolutionWithServices(t *testing.T, mockDID string) *did.DocResolution {
	t.Helper()

	edPub, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	mockVM := did.NewVerificationMethodFromBytes(mockVMID, "Ed25519VerificationKey2018", mockDID, edPub)

	docRes := &did.DocResolution{
		DIDDocument: &did.Doc{
			ID:      mockDID,
			Context: []string{did.ContextV1},
			VerificationMethod: []did.VerificationMethod{
				*mockVM,
			},
			AssertionMethod: []did.Verification{
				{
					VerificationMethod: *mockVM,
				},
			},
			Service: []did.Service{{
				Type:            "LinkedDomains",
				ServiceEndpoint: endpoint.NewDIDCommV1Endpoint("mock-uri"),
			}},
		},
	}

	return docRes
}
