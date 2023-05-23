/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hyperledger/aries-framework-go/component/models/did"
	vdrapi "github.com/hyperledger/aries-framework-go/component/vdr/api"
	mockvdr "github.com/hyperledger/aries-framework-go/component/vdr/mock"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/pkg/common"
)

func TestDIDKeyResolver_Resolve(t *testing.T) {
	req := require.New(t)

	didDoc := createDIDDoc()
	publicKey := didDoc.VerificationMethod[0]
	authentication := didDoc.Authentication[0]
	assertionMethod := didDoc.AssertionMethod[0]

	vdrRegistry := &mockvdr.VDRegistry{
		ResolveValue: didDoc,
	}

	resolver := common.NewVDRKeyResolver(&vdrResolverAdapter{vdr: vdrRegistry})
	req.NotNil(resolver)

	pubKey, err := resolver.PublicKeyFetcher()(didDoc.ID, publicKey.ID)
	req.NoError(err)
	req.Equal(publicKey.Value, pubKey.Value)
	req.Equal("Ed25519VerificationKey2018", pubKey.Type)
	req.NotNil(pubKey.JWK)
	req.Equal(pubKey.JWK.Algorithm, "EdDSA")

	authPubKey, err := resolver.PublicKeyFetcher()(didDoc.ID, authentication.VerificationMethod.ID)
	req.NoError(err)
	req.Equal(authentication.VerificationMethod.Value, authPubKey.Value)
	req.Equal("Ed25519VerificationKey2018", authPubKey.Type)
	req.NotNil(authPubKey.JWK)
	req.Equal(authPubKey.JWK.Algorithm, "EdDSA")

	assertMethPubKey, err := resolver.PublicKeyFetcher()(didDoc.ID, assertionMethod.VerificationMethod.ID)
	req.NoError(err)
	req.Equal(assertionMethod.VerificationMethod.Value, assertMethPubKey.Value)
	req.Equal("Ed25519VerificationKey2018", assertMethPubKey.Type)

	pubKey, err = resolver.PublicKeyFetcher()(didDoc.ID, "invalid key")
	req.Error(err)
	req.EqualError(err, fmt.Sprintf("public key with KID invalid key is not found for DID %s", didDoc.ID))
	req.Nil(pubKey)

	vdrRegistry.ResolveErr = errors.New("resolver error")
	pubKey, err = resolver.PublicKeyFetcher()(didDoc.ID, "")
	req.Error(err)
	req.EqualError(err, fmt.Sprintf("resolve DID %s: resolver error", didDoc.ID))
	req.Nil(pubKey)
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
      "id": "did:test:2WxUJa8nVjXr5yS69JWoKZ#keys-1",
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
