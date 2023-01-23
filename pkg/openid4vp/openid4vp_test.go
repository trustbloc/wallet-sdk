/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp //nolint: testpackage

import (
	"bytes"
	_ "embed" //nolint:gci // required for go:embed
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/doc/jose"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/internal/testutil"
)

var (
	//go:embed test_data/request_object.jwt
	requestObjectJWT string

	//go:embed test_data/presentation.jsonld
	presentationJSONLD []byte
)

func TestOpenID4VP_GetQuery(t *testing.T) {
	t.Run("Inline Request Object", func(t *testing.T) {
		instance := New(requestObjectJWT, &jwtSignatureVerifierMock{})

		query, err := instance.GetQuery()
		require.NoError(t, err)
		require.NotNil(t, query)
	})

	t.Run("Fetch Request Object", func(t *testing.T) {
		instance := New("openid-vc://?request_uri=https://request-object",
			&jwtSignatureVerifierMock{},
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
		}, nil)

		_, err := instance.GetQuery()
		require.Contains(t, err.Error(), "sig verification err")
	})
}

func TestOpenID4VP_PresentCredential(t *testing.T) {
	presentation, presErr := verifiable.ParsePresentation(presentationJSONLD,
		verifiable.WithPresDisabledProofCheck(),
		verifiable.WithPresJSONLDDocumentLoader(testutil.DocumentLoader(t)))

	require.NoError(t, presErr)
	require.NotNil(t, presentation)

	t.Run("Success", func(t *testing.T) {
		httpClient := &httpClientMock{
			StatusCode: 200,
		}

		instance := New(requestObjectJWT, &jwtSignatureVerifierMock{}, WithHTTPClient(httpClient))

		query, err := instance.GetQuery()
		require.NoError(t, err)
		require.NotNil(t, query)

		err = instance.PresentCredential(presentation, &jwtSignerMock{keyID: "did:example:12345#testId"})
		require.NoError(t, err)

		expectedState := "636df28459a07d50cc4b657e"
		expectedSig := base64.RawURLEncoding.EncodeToString([]byte("test signature"))

		require.Contains(t, string(httpClient.SentBody), expectedState)
		require.Contains(t, string(httpClient.SentBody), expectedSig)
	})

	t.Run("Check nonce", func(t *testing.T) {
		response, err := createAuthorizedResponse(presentation,
			&requestObject{
				Nonce: "test123456",
				State: "test34566",
			}, &jwtSignerMock{keyID: "did:example:12345#testId"})

		require.NoError(t, err)
		require.Equal(t, response.State, "test34566")

		idToken, err := base64.RawURLEncoding.DecodeString(strings.Split(response.IDTokenJWS, ".")[1])
		require.NoError(t, err)
		require.Contains(t, string(idToken), "test123456")

		vpToken, err := base64.RawURLEncoding.DecodeString(strings.Split(response.VPTokenJWS, ".")[1])
		require.NoError(t, err)
		require.Contains(t, string(vpToken), "test123456")
	})

	t.Run("Invalid kid", func(t *testing.T) {
		_, err := createAuthorizedResponse(presentation,
			&requestObject{}, &jwtSignerMock{keyID: "did:example:12345"})

		require.Contains(t, err.Error(), "kid not containing did part did:example:12345")
	})

	t.Run("sign failed", func(t *testing.T) {
		_, err := createAuthorizedResponse(presentation,
			&requestObject{
				Nonce: "test123456",
				State: "test34566",
			}, &jwtSignerMock{keyID: "did:example:12345#testId", Err: errors.New("sign failed")})

		require.Contains(t, err.Error(), "sign failed")
	})

	t.Run("send authorized response failed", func(t *testing.T) {
		err := sendAuthorizedResponse(&httpClientMock{}, "response", "redirectURI")
		require.Contains(t, err.Error(), "expected status code 200")
	})
}

func Test_doHTTPRequest(t *testing.T) {
	t.Run("Invalid http method", func(t *testing.T) {
		_, err := doHTTPRequest(&httpClientMock{}, "\n\n", "url", "", nil)
		require.Contains(t, err.Error(), "invalid method")
	})

	t.Run("Invalid http code", func(t *testing.T) {
		_, err := doHTTPRequest(&httpClientMock{}, "GEt", "url", "", nil)
		require.Contains(t, err.Error(), "xpected status code 200")
	})
}

type jwtSignatureVerifierMock struct {
	err error
}

func (s *jwtSignatureVerifierMock) Verify(joseHeaders jose.Headers, payload, signingInput, signature []byte) error {
	return s.err
}

type jwtSignerMock struct {
	keyID string
	Err   error
}

func (s *jwtSignerMock) GetKeyID() string {
	return s.keyID
}

func (s *jwtSignerMock) Sign(data []byte) ([]byte, error) {
	return []byte("test signature"), s.Err
}

func (s *jwtSignerMock) Headers() jose.Headers {
	return jose.Headers{
		jose.HeaderKeyID:     "KeyID",
		jose.HeaderAlgorithm: "ES384",
	}
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
