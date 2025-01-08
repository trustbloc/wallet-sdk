/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/did-go/doc/did"
	vdrapi "github.com/trustbloc/did-go/vdr/api"
	mockvdr "github.com/trustbloc/did-go/vdr/mock"

	"github.com/trustbloc/wallet-sdk/pkg/common"
)

func TestDIDKeyResolver_Resolve(t *testing.T) {
	req := require.New(t)

	didDoc := createDIDDoc()
	publicKey := didDoc.VerificationMethod[0]

	vdrRegistry := &mockvdr.VDRegistry{
		ResolveValue: didDoc,
	}

	resolver := common.NewVDRKeyResolver(&vdrResolverAdapter{vdr: vdrRegistry})
	req.NotNil(resolver)

	pubKey, err := resolver.ResolveVerificationMethod(publicKey.ID, didDoc.ID)
	req.NoError(err)
	req.Equal(publicKey.Value, pubKey.Value)
	req.Equal("Ed25519VerificationKey2018", pubKey.Type)
	req.NotNil(pubKey.JWK)
	req.Equal("EdDSA", pubKey.JWK.Algorithm)
}

type vdrResolverAdapter struct {
	vdr vdrapi.Registry
}

func (a *vdrResolverAdapter) Resolve(didID string) (*did.DocResolution, error) {
	return a.vdr.Resolve(didID)
}

//nolint:lll
func createDIDDoc() *did.Doc {
	didDocJSON := `{
  "@context": [
    "https://w3id.org/did/v1"
  ],
  "id": "did:test:2WxUJa8nVjXr5yS69JWoKZ",
  "verificationMethod": [
    {
      "controller": "did:test:8STcrCQFzFxKey7YSbj62A",
      "id": "did:test:8STcrCQFzFxKey7YSbj62A#keys-1",
      "publicKeyJwk": {
        "kty": "OKP",
        "crv": "Ed25519",
        "alg": "EdDSA",
        "x": "PD34BecP4G7UcAj2u1ygB9MX31jJnqtkJFvkR1o8nIE"
      },
      "type": "Ed25519VerificationKey2018"
    }
  ],
  "service": [
    {
      "id": "did:test:8STcrCQFzFxKey7YSbj62A#endpoint-1",
      "priority": 0,
      "recipientKeys": [
        "did:test:8STcrCQFzFxKey7YSbj62A#keys-1"
      ],
      "routingKeys": null,
      "serviceEndpoint": "http://localhost:47582",
      "type": "did-communication"
    }
  ],
  "authentication": [
    {
      "controller": "did:test:2WxUJa8nVjXr5yS69JWoKZ",
      "id": "did:test:2WxUJa8nVjXr5yS69JWoKZ#keys-2",
      "publicKeyJwk": {
        "kty": "OKP",
        "crv": "Ed25519",
        "alg": "EdDSA",
        "x": "DEfkntM3vCV5WtS-1G9cBMmkNJSPlVdjwSdHmHbirTg"
      },
      "type": "Ed25519VerificationKey2018"
    }
  ],
  "assertionMethod": [
    {
      "id": "did:v1:test:nym:z6MkfG5HTrBXzsAP8AbayNpG3ZaoyM4PCqNPrdWQRSpHDV6J#z6MkqfvdBsFw4QdGrZrnx7L1EKfY5zh9tT4gumUGsMMEZHY3",
      "type": "Ed25519VerificationKey2018",
      "controller": "did:v1:test:nym:z6MkfG5HTrBXzsAP8AbayNpG3ZaoyM4PCqNPrdWQRSpHDV6J",
      "publicKeyBase58": "CDfabd1Vis8ok526GYNAPE7YGRRJUZpLDkZM35PDe4kf"
    }
  ],
  "created": "2020-04-13T12:51:08.274813+03:00",
  "updated": "2020-04-13T12:51:08.274813+03:00"
}`

	didDoc, err := did.ParseDocument([]byte(didDocJSON))
	if err != nil {
		panic(err)
	}

	return didDoc
}
