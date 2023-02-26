/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp //nolint: testpackage

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	_ "embed" //nolint:gci // required for go:embed
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/trustbloc/wallet-sdk/pkg/api"

	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/doc/jose"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/internal/testutil"
	"github.com/trustbloc/wallet-sdk/pkg/metricslogger/noop"
)

var (
	//go:embed test_data/request_object.jwt
	requestObjectJWT string

	//go:embed test_data/presentation.jsonld
	presentationJSONLD []byte
)

const (
	testSignature = "test signature"
	mockDID       = "did:example:12345"
	mockVMID      = "#key-1"
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

func TestOpenID4VP_GetQuery(t *testing.T) {
	t.Run("Inline Request Object", func(t *testing.T) {
		instance := New(requestObjectJWT, &jwtSignatureVerifierMock{}, nil, nil)

		query, err := instance.GetQuery()
		require.NoError(t, err)
		require.NotNil(t, query)
	})

	t.Run("Fetch Request Object", func(t *testing.T) {
		instance := New("openid-vc://?request_uri=https://request-object",
			&jwtSignatureVerifierMock{},
			nil,
			nil,
			WithHTTPClient(&httpClientMock{
				Response:         requestObjectJWT,
				StatusCode:       200,
				ExpectedEndpoint: "https://request-object",
			}),
		)

		query, err := instance.GetQuery()
		require.NoError(t, err)
		require.NotNil(t, query)
	})

	t.Run("Fetch Request failed", func(t *testing.T) {
		instance := New("openid-vc://?request_uri=https://request-object",
			&jwtSignatureVerifierMock{},
			nil,
			nil,
			WithHTTPClient(&httpClientMock{
				Err: errors.New("http error"),
			}),
		)

		_, err := instance.GetQuery()
		require.Contains(t, err.Error(), "http error")
	})

	t.Run("Inline Request Object", func(t *testing.T) {
		instance := New(requestObjectJWT, &jwtSignatureVerifierMock{
			err: errors.New("sig verification err"),
		}, nil, nil, nil)

		_, err := instance.GetQuery()
		require.Contains(t, err.Error(), "sig verification err")
	})

	t.Run("Fail to log retrieve request object via HTTP GET metrics event", func(t *testing.T) {
		instance := New("openid-vc://?request_uri=https://request-object",
			&jwtSignatureVerifierMock{},
			nil,
			nil,
			WithHTTPClient(&httpClientMock{
				Response:         requestObjectJWT,
				StatusCode:       200,
				ExpectedEndpoint: "https://request-object",
			}),
			WithMetricsLogger(&failingMetricsLogger{}),
		)

		query, err := instance.GetQuery()
		require.EqualError(t, err, "REQUEST_OBJECT_FETCH_FAILED(OVP1-0000):fetch request object: "+
			"failed to log event (Event=Fetch request object via an HTTP GET request to https://request-object)")
		require.Nil(t, query)
	})
}

func TestOpenID4VP_PresentCredential(t *testing.T) {
	presentation, presErr := verifiable.ParsePresentation(presentationJSONLD,
		verifiable.WithPresDisabledProofCheck(),
		verifiable.WithPresJSONLDDocumentLoader(testutil.DocumentLoader(t)))

	require.NoError(t, presErr)
	require.NotNil(t, presentation)

	mockDoc := mockResolution(t, mockDID)

	t.Run("Success", func(t *testing.T) {
		httpClient := &httpClientMock{
			StatusCode: 200,
		}

		instance := New(
			requestObjectJWT,
			&jwtSignatureVerifierMock{},
			&didResolverMock{ResolveValue: mockDoc},
			&cryptoMock{SignVal: []byte(testSignature)},
			WithHTTPClient(httpClient),
		)

		query, err := instance.GetQuery()
		require.NoError(t, err)
		require.NotNil(t, query)

		err = instance.PresentCredential([]*verifiable.Presentation{presentation})
		require.NoError(t, err)

		expectedState := "636df28459a07d50cc4b657e"
		expectedSig := base64.RawURLEncoding.EncodeToString([]byte(testSignature))

		require.Contains(t, string(httpClient.SentBody), expectedState)
		require.Contains(t, string(httpClient.SentBody), expectedSig)
	})

	t.Run("Check nonce", func(t *testing.T) {
		response, err := createAuthorizedResponse([]*verifiable.Presentation{presentation}, &requestObject{
			Nonce: "test123456",
			State: "test34566",
		}, &didResolverMock{ResolveValue: mockDoc}, &cryptoMock{SignVal: []byte(testSignature)})

		require.NoError(t, err)
		require.Equal(t, response.State, "test34566")

		idToken, err := base64.RawURLEncoding.DecodeString(strings.Split(response.IDTokenJWS, ".")[1])
		require.NoError(t, err)
		require.Contains(t, string(idToken), "test123456")

		vpToken, err := base64.RawURLEncoding.DecodeString(strings.Split(response.VPTokenJWS, ".")[1])
		require.NoError(t, err)
		require.Contains(t, string(vpToken), "test123456")
	})

	t.Run("no presentations provided", func(t *testing.T) {
		instance := New(
			requestObjectJWT,
			&jwtSignatureVerifierMock{},
			&didResolverMock{ResolveValue: mockDoc},
			&cryptoMock{SignVal: []byte(testSignature)},
			WithHTTPClient(&httpClientMock{
				StatusCode: 200,
			}),
		)

		err := instance.PresentCredential(nil)

		require.Error(t, err)
		require.Contains(t, err.Error(), "expected at least one presentation")
	})

	t.Run("no credentials provided", func(t *testing.T) {
		_, err := createAuthorizedResponse([]*verifiable.Presentation{
			{},
		}, &requestObject{
			Nonce: "test123456",
			State: "test34566",
		}, &didResolverMock{ResolveValue: mockDoc}, &cryptoMock{})

		require.Error(t, err)
		require.Contains(t, err.Error(), "presentation has no credentials")
	})

	t.Run("no subject ID found", func(t *testing.T) {
		testCases := []struct {
			vpFunc func(t *testing.T) *verifiable.Presentation
		}{
			{
				vpFunc: func(t *testing.T) *verifiable.Presentation {
					t.Helper()

					vp, err := verifiable.NewPresentation(verifiable.WithCredentials(&verifiable.Credential{
						ID: "foo",
					}))
					require.NoError(t, err)

					return vp
				},
			},
			{
				vpFunc: func(t *testing.T) *verifiable.Presentation {
					t.Helper()

					vp, err := verifiable.NewPresentation(verifiable.WithCredentials(&verifiable.Credential{
						ID: "foo",
					}))
					require.NoError(t, err)

					vp.Credentials()[0] = map[string]interface{}{}

					return vp
				},
			},
			{
				vpFunc: func(t *testing.T) *verifiable.Presentation {
					t.Helper()

					vp, err := verifiable.NewPresentation(verifiable.WithCredentials(&verifiable.Credential{
						ID: "foo",
					}))
					require.NoError(t, err)

					vp.Credentials()[0] = struct{}{}

					return vp
				},
			},
		}

		for _, testCase := range testCases {
			t.Run("", func(t *testing.T) {
				_, err := createAuthorizedResponse(
					[]*verifiable.Presentation{testCase.vpFunc(t)},
					&requestObject{
						Nonce: "test123456",
						State: "test34566",
					}, &didResolverMock{ResolveValue: mockDoc}, &cryptoMock{})

				require.Error(t, err)
				require.Contains(t, err.Error(), "presentation VC does not have a subject ID")
			})
		}
	})

	t.Run("fail to resolve signing DID", func(t *testing.T) {
		expectErr := errors.New("resolve failed")

		_, err := createAuthorizedResponse([]*verifiable.Presentation{presentation}, &requestObject{
			Nonce: "test123456",
			State: "test34566",
		}, &didResolverMock{ResolveErr: expectErr}, &cryptoMock{})

		require.ErrorIs(t, err, expectErr)
	})

	t.Run("signing DID has no signing key", func(t *testing.T) {
		_, err := createAuthorizedResponse([]*verifiable.Presentation{presentation}, &requestObject{
			Nonce: "test123456",
			State: "test34566",
		}, &didResolverMock{ResolveValue: &did.DocResolution{
			DIDDocument: &did.Doc{},
		}}, &cryptoMock{})

		require.Error(t, err)
		require.Contains(t, err.Error(), "no assertion method for signing")
	})

	t.Run("sign failed", func(t *testing.T) {
		expectErr := errors.New("sign failed")

		_, err := createAuthorizedResponse([]*verifiable.Presentation{presentation}, &requestObject{
			Nonce: "test123456",
			State: "test34566",
		}, &didResolverMock{ResolveValue: mockDoc}, &cryptoMock{SignErr: expectErr})

		require.ErrorIs(t, err, expectErr)
	})
}

func Test_doHTTPRequest(t *testing.T) {
	t.Run("Invalid http method", func(t *testing.T) {
		interaction := Interaction{httpClient: &httpClientMock{}}
		_, err := interaction.doHTTPRequest("\n\n", "url", "", nil,
			"", "")
		require.Contains(t, err.Error(), "invalid method")
	})

	t.Run("Invalid http code", func(t *testing.T) {
		interaction := Interaction{httpClient: &httpClientMock{}, metricsLogger: noop.NewMetricsLogger()}

		_, err := interaction.doHTTPRequest(http.MethodGet, "url", "", nil,
			"", "")
		require.Contains(t, err.Error(), "xpected status code 200")
	})
}

type jwtSignatureVerifierMock struct {
	err error
}

func (s *jwtSignatureVerifierMock) Verify(joseHeaders jose.Headers, payload, signingInput, signature []byte) error {
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

type httpClientMock struct {
	Response         string
	StatusCode       int
	Err              error
	ExpectedEndpoint string
	SentBody         []byte
}

func (c *httpClientMock) Do(req *http.Request) (*http.Response, error) {
	if c.ExpectedEndpoint != "" && c.ExpectedEndpoint != req.URL.String() {
		return nil, fmt.Errorf("requested endpoint %s not match %s", req.URL.String(), c.ExpectedEndpoint)
	}

	if req.Body != nil {
		respBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

		c.SentBody = respBytes
	}

	if c.Err != nil {
		return nil, c.Err
	}

	return &http.Response{
		StatusCode: c.StatusCode,
		Body:       io.NopCloser(bytes.NewBuffer([]byte(c.Response))),
	}, nil
}

func mockResolution(t *testing.T, mockDID string) *did.DocResolution {
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
		},
	}

	return docRes
}
