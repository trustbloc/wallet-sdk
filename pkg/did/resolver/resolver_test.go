/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resolver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/pkg/did/resolver"
)

const (
	docID = "did:ion:test"
	doc   = `{
  "@context": ["https://w3id.org/did/v1","https://w3id.org/did/v2"],
  "id": "did:ion:test"
}`
)

func TestDIDResolver(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		didResolver, err := resolver.NewDIDResolver(resolver.WithHTTPTimeout(time.Second * 10))
		require.NoError(t, err)

		testcases := []struct {
			name string
			did  string
		}{
			{
				name: "did:key",
				did:  "did:key:z6MkjfbzWitsSUyFMTbBUSWNsJBHR7BefFp1WmABE3kRw8Qr",
			},
			{
				name: "did:jwk",
				did:  "did:jwk:eyJjcnYiOiJFZDI1NTE5Iiwia3R5IjoiT0tQIiwieCI6IndVQWp5UmFqRDhSLTJ2Zm1oZU1lRzNPUTViY0F4OFZKRHhUdDl5SDRVbDgifQ", //nolint:lll
			},
		}

		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				didDocResolution, e := didResolver.Resolve(tc.did)
				require.NoError(t, e)
				require.NotEmpty(t, didDocResolution)
			})
		}
	})

	t.Run("httpbinding initialization error", func(t *testing.T) {
		didResolver, err := resolver.NewDIDResolver(resolver.WithResolverServerURI("not a uri"))
		require.Error(t, err)
		require.Nil(t, didResolver)
		require.Contains(t, err.Error(), "failed to initialize client for DID resolution server")
	})

	t.Run("httpbinding resolve", func(t *testing.T) {
		testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.Header().Add("Content-type", "application/did+ld+json")
			res.WriteHeader(http.StatusOK)

			_, err := res.Write([]byte(doc))
			assert.NoError(t, err)
		}))

		defer func() { testServer.Close() }()

		didResolver, err := resolver.NewDIDResolver(resolver.WithResolverServerURI(testServer.URL))
		require.NoError(t, err)

		didDocResolution, err := didResolver.Resolve(docID)
		require.NoError(t, err)
		require.NotNil(t, didDocResolution)
		require.NotNil(t, didDocResolution.DIDDocument)
		require.Equal(t, docID, didDocResolution.DIDDocument.ID)
	})
}

func TestDIDResolver_InvalidDID(t *testing.T) {
	didResolver, err := resolver.NewDIDResolver()
	require.NoError(t, err)

	didDocResolution, err := didResolver.Resolve("did:example:abc")
	require.Error(t, err)
	requireErrorContains(t, err, "resolve did:example:abc : did method example not supported for vdr")
	require.Empty(t, didDocResolution)
}

func requireErrorContains(t *testing.T, err error, errString string) { //nolint:thelper
	require.Error(t, err)
	require.Contains(t, err.Error(), errString)
}
