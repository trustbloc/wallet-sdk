/*
Copyright Avast Software. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package did //nolint:testpackage // uses internal implementation details so the HTTP client can be mocked

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/did-go/doc/did"
	"github.com/trustbloc/did-go/method/httpbinding"
)

//go:embed test_data/didconfig.json
var didCfg string

const (
	testDIDMethod           = "ion"
	testDIDMethodSpecificID = "EiCMdVLtzqqW5n6zUC3_srZxWPCseVxKXu9FqQ8LyS1mTA:eyJkZWx0YSI6eyJwYXRjaGVzIjp" +
		"beyJhY3Rpb24iOiJyZXBsYWNlIiwiZG9jdW1lbnQiOnsicHVibGljS2V5cyI6W3siaWQiOiI2NmRkNTFmZTBjYWM0ZjFhYWU" +
		"4MTJkMGFhMTA5YmMyYXZjU2lnbmluZ0tleS0yZTk3NSIsInB1YmxpY0tleUp3ayI6eyJjcnYiOiJzZWNwMjU2azEiLCJrdHk" +
		"iOiJFQyIsIngiOiJqNVQ4S1FfQ19IRGxSbXlFX1pwRjltbE1RZ3B4N19fMFJQRHhPVmM4dWt3IiwieSI6InpybDBWSllHWnh" +
		"VLXFjZWt2SlY4NGs5U2x2STQxam53NG4yTS1WMnB4MGMifSwicHVycG9zZXMiOlsiYXV0aGVudGljYXRpb24iLCJhc3NlcnR" +
		"pb25NZXRob2QiXSwidHlwZSI6IkVjZHNhU2VjcDI1NmsxVmVyaWZpY2F0aW9uS2V5MjAxOSJ9XSwic2VydmljZXMiOlt7Iml" +
		"kIjoibGlua2VkZG9tYWlucyIsInNlcnZpY2VFbmRwb2ludCI6eyJvcmlnaW5zIjpbImh0dHBzOi8vZGlkLnJvaGl0Z3VsYXR" +
		"pLmNvbS8iXX0sInR5cGUiOiJMaW5rZWREb21haW5zIn0seyJpZCI6Imh1YiIsInNlcnZpY2VFbmRwb2ludCI6eyJpbnN0YW5" +
		"jZXMiOlsiaHR0cHM6Ly9iZXRhLmh1Yi5tc2lkZW50aXR5LmNvbS92MS4wL2E0OTJjZmYyLWQ3MzMtNDA1Ny05NWE1LWE3MWZ" +
		"jMzY5NWJjOCJdfSwidHlwZSI6IklkZW50aXR5SHViIn1dfX1dLCJ1cGRhdGVDb21taXRtZW50IjoiRWlDcXRpZnUwSHg4RUV" +
		"kbGlrVnZIWGpYZzRLb0pZZUV0cDdZeGlvRzVYWmRKZyJ9LCJzdWZmaXhEYXRhIjp7ImRlbHRhSGFzaCI6IkVpQ1NVQklmYTB" +
		"XZHBXNm5oVTdNaHlSczRucTFDeEg1V1ZyUjVkUFZYV09MYmciLCJyZWNvdmVyeUNvbW1pdG1lbnQiOiJFaUF1cGoxRWZsOHd" +
		"jWlRQZTI3X0lGWEJ3MjlzOEN5SXBRX3UzVkRwUmswdkNRIn19"
	testDID = "did:" + testDIDMethod + ":" + testDIDMethodSpecificID
	doc     = `
{
  "id": "` + testDID + `",
  "@context": [
    "https://www.w3.org/ns/did/v1",
    {
      "@base": "` + testDID + `"
    }
  ],
  "service": [
    {
      "id": "#linkeddomains",
      "type": "LinkedDomains",
      "serviceEndpoint": "https://did.rohitgulati.com"
    }
  ],
  "verificationMethod": [
    {
      "id": "#66dd51fe0cac4f1aae812d0aa109bc2avcSigningKey-2e975",
      "controller": "` + testDID + `",
      "type": "EcdsaSecp256k1VerificationKey2019",
      "publicKeyJwk": {
        "kty": "EC",
        "crv": "secp256k1",
        "x": "j5T8KQ_C_HDlRmyE_ZpF9mlMQgpx7__0RPDxOVc8ukw",
        "y": "zrl0VJYGZxU-qcekvJV84k9SlvI41jnw4n2M-V2px0c"
      }
    }
  ],
  "authentication": [
    "#66dd51fe0cac4f1aae812d0aa109bc2avcSigningKey-2e975"
  ],
  "assertionMethod": [
    "#66dd51fe0cac4f1aae812d0aa109bc2avcSigningKey-2e975"
  ]
}`

	docMetadata = `
{
  "method": {
    "published": true,
    "recoveryCommitment": "EiAupj1Efl8wcZTPe27_IFXBw29s8CyIpQ_u3VDpRk0vCQ",
    "updateCommitment": "EiCqtifu0Hx8EEdlikVvHXjXg4KoJYeEtp7YxioG5XZdJg"
  },
  "equivalentId": [
    "did:ion:EiCMdVLtzqqW5n6zUC3_srZxWPCseVxKXu9FqQ8LyS1mTA"
  ],
  "canonicalId": "did:ion:EiCMdVLtzqqW5n6zUC3_srZxWPCseVxKXu9FqQ8LyS1mTA"
}`

	resolutionMetadata = `
{
  "contentType": "application/did+ld+json",
  "pattern": "^(did:ion:(?!test).+)$",
  "driverUrl": "http://driver-did-ion:8080/1.0/identifiers/",
  "duration": 403,
  "did": {
    "didString": "` + testDID + `",
    "methodSpecificId": "` + testDIDMethodSpecificID + `",
    "method": "` + testDIDMethod + `"
  }
}`

	resolutionResponse = `{
  "@context": "https://w3id.org/did-resolution/v1",
  "didDocument": ` + doc + `,
  "didDocumentMetadata": ` + docMetadata + `,
  "didResolutionMetadata": ` + resolutionMetadata + `
}`
)

func TestValidate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		httpClient := &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte(didCfg))),
				}, nil
			},
		}

		testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			assert.Equal(t, "/"+testDID, req.URL.String())
			res.Header().Add("Content-type", "application/did+ld+json")
			res.WriteHeader(http.StatusOK)
			_, err := res.Write([]byte(resolutionResponse))
			assert.NoError(t, err)
		}))

		defer func() { testServer.Close() }()

		resolver, err := httpbinding.New(testServer.URL)
		require.NoError(t, err)

		validationResult, err := validateLinkedDomains(testDID, &resolverWrapper{vdr: resolver}, httpClient)
		require.NoError(t, err)
		require.True(t, validationResult.IsValid)
		require.Equal(t, "https://did.rohitgulati.com", validationResult.ServiceURL)
	})
	t.Run("No resolver provided", func(t *testing.T) {
		opts := NewValidateLinkedDomainsOpts().SetHTTPTimeoutNanoseconds(0)

		validationResult, err := ValidateLinkedDomains(testDID, nil, opts)
		require.EqualError(t, err, "no resolver provided")
		require.Nil(t, validationResult)
	})
	t.Run("DID service validation failure", func(t *testing.T) {
		testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, _ *http.Request) {
			_, err := res.Write([]byte(didCfg))
			assert.NoError(t, err)
		}))

		defer func() { testServer.Close() }()

		validationResult, err := ValidateLinkedDomains("DID", newMockResolver(testServer.URL), nil)
		require.NoError(t, err, "DOMAIN_AND_DID_VERIFICATION_FAILED")
		require.NotNil(t, validationResult)
		require.False(t, validationResult.IsValid)
		require.NotEmpty(t, validationResult.ServiceURL)
	})
}

type resolverWrapper struct {
	vdr *httpbinding.VDR
}

func (r *resolverWrapper) Resolve(
	did string, //nolint:gocritic // no great way to avoid this import shadow (did)
) ([]byte, error) {
	didDocResolution, err := r.vdr.Read(did)
	if err != nil {
		return nil, err
	}

	return didDocResolution.JSONBytes()
}

type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

type mockResolver struct {
	serviceEndpoint string
}

func newMockResolver(serviceEndpoint string) *mockResolver {
	return &mockResolver{serviceEndpoint: serviceEndpoint}
}

// Always returns a DID doc that will fail service validation.
func (m *mockResolver) Resolve(string) ([]byte, error) {
	didDocTemplate := `{
  "@context": ["https://www.w3.org/ns/did/v1","https://identity.foundation/.well-known/did-configuration/v1"],
  "id": "did:example:123",
  "verificationMethod": [{
    "id": "did:example:123#_Qq0UL2Fq651Q0Fjd6TvnYE-faHiOpRlPVQcY_-tA4A",
    "type": "JsonWebKey2020",
    "controller": "did:example:123",
    "publicKeyJwk": {
      "kty": "OKP",
      "crv": "Ed25519",
      "x": "VCpo2LMLhn6iWku8MKvSLg2ZAoC-nlOyPVQaO3FxVeQ"
    }
  }],
  "service": [
    {
      "id":"did:example:123#foo",
      "type": "LinkedDomains",
      "serviceEndpoint": "%s"
    }
  ]
}`

	didDoc := fmt.Sprintf(didDocTemplate, m.serviceEndpoint)

	parsedDIDDoc, err := did.ParseDocument([]byte(didDoc))
	if err != nil {
		return nil, err
	}

	didDocResolution := did.DocResolution{DIDDocument: parsedDIDDoc}

	return didDocResolution.JSONBytes()
}
