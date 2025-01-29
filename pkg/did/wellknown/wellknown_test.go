/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package wellknown_test

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

	"github.com/trustbloc/wallet-sdk/internal/testutil"
	"github.com/trustbloc/wallet-sdk/pkg/did/resolver"
	"github.com/trustbloc/wallet-sdk/pkg/did/wellknown"
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

		didResolver, err := httpbinding.New(testServer.URL)
		require.NoError(t, err)

		valid, domain, err := wellknown.ValidateLinkedDomains(testDID, &resolverWrapper{vdr: didResolver}, httpClient)
		require.NoError(t, err)
		require.True(t, valid)
		require.Equal(t, "https://did.rohitgulati.com", domain)
	})
	t.Run("No resolver provided", func(t *testing.T) {
		valid, domain, err := wellknown.ValidateLinkedDomains(testDID, nil, nil)
		testutil.RequireErrorContains(t, err, "no resolver provided")
		require.False(t, valid)
		require.Empty(t, domain)
	})
	t.Run("Fail to resolve DID (invalid DID format)", func(t *testing.T) {
		didResolver, err := resolver.NewDIDResolver()
		require.NoError(t, err)

		valid, domain, err := wellknown.ValidateLinkedDomains("InvalidDID", didResolver, nil)
		testutil.RequireErrorContains(t, err, "WELLKNOWN_INITIALIZATION_FAILED")
		require.False(t, valid)
		require.Empty(t, domain)
	})
	t.Run("Resolved DID document has no services", func(t *testing.T) {
		didResolver, err := resolver.NewDIDResolver()
		require.NoError(t, err)

		sampleDIDWithoutServices := "did:key:z6MkoTHsgNNrby8JzCNQ1iRLyW5QQ6R8Xuu6AA8igGrMVPUM"

		valid, domain, err := wellknown.ValidateLinkedDomains(sampleDIDWithoutServices, didResolver, nil)
		require.NoError(t, err)
		require.False(t, valid)
		require.Empty(t, domain)
	})

	t.Run("Resolved DID document has more than one service", func(t *testing.T) {
		didDoc := `{
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
      "serviceEndpoint": {
        "origins": ["https://did.rohitgulati.com"]
      }
    },
	{
      "id":"did:example:123#foo",
      "type": "LinkedDomains",
      "serviceEndpoint": {
        "origins": ["https://did.rohitgulati.com"]
      }
    }
  ]
}`

		_, domain, err := wellknown.ValidateLinkedDomains("DID", newMockResolver(didDoc), nil)
		require.NoError(t, err)
		require.Equal(t, "https://did.rohitgulati.com", domain)
	})
	t.Run("DID service validation failure", func(t *testing.T) {
		testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, _ *http.Request) {
			_, err := res.Write([]byte(didCfg))
			assert.NoError(t, err)
		}))

		defer func() { testServer.Close() }()

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

		didDoc := fmt.Sprintf(didDocTemplate, testServer.URL)

		valid, domain, err := wellknown.ValidateLinkedDomains("DID", newMockResolver(didDoc), nil)
		require.NoError(t, err)
		require.False(t, valid)
		require.NotEmpty(t, domain)
	})
	t.Run("Service type is not a string", func(t *testing.T) {
		didDoc := `{
  "@context": ["https://www.w3.org/ns/did/v1","https://identity.foundation/.well-known/did-configuration/v1"],
  "id": "did:example:123",
  "service": [
    {
      "id":"did:example:123#foo",
      "type": ["Type1", "Type2"],
      "serviceEndpoint": "https://identity.foundation"
    }
  ]
}`
		valid, domain, err := wellknown.ValidateLinkedDomains("DID", newMockResolver(didDoc), nil)
		testutil.RequireErrorContains(t, err, "resolved DID document is not supported since it contains a service type "+
			"at index 0 that is not a simple string")
		require.False(t, valid)
		require.Empty(t, domain)
	})
	t.Run("Resolved DID document has a service, but it's not of the Linked Domains type", func(t *testing.T) {
		didDoc := `{
  "@context": ["https://www.w3.org/ns/did/v1","https://identity.foundation/.well-known/did-configuration/v1"],
  "id": "did:example:123",
  "service": [
    {
      "id":"did:example:123#foo",
      "type": "SomeOtherServiceType",
      "serviceEndpoint": "https://identity.foundation"
    }
  ]
}`
		valid, domain, err := wellknown.ValidateLinkedDomains("DID", newMockResolver(didDoc), nil)
		require.NoError(t, err)
		require.False(t, valid)
		require.Empty(t, domain)
	})
}

type resolverWrapper struct {
	vdr *httpbinding.VDR
}

func (r *resolverWrapper) Resolve(
	did string, //nolint:gocritic // no great way to avoid this import shadow (did)
) (*did.DocResolution, error) {
	return r.vdr.Read(did)
}

type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

type mockResolver struct {
	didDocToReturn string
}

func newMockResolver(didDocToReturn string) *mockResolver {
	return &mockResolver{didDocToReturn}
}

func (m *mockResolver) Resolve(string) (*did.DocResolution, error) {
	parsedDIDDoc, err := did.ParseDocument([]byte(m.didDocToReturn))
	if err != nil {
		return nil, err
	}

	return &did.DocResolution{DIDDocument: parsedDIDDoc}, err
}
