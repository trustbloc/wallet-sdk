/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package common implements common functionality like jwt sign and did public key resolve.
package common

import (
	"fmt"
	"strings"

	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/verifier"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"

	"github.com/trustbloc/wallet-sdk/pkg/api"
)

// VDRKeyResolver resolves DID in order to find public keys for VC verification using vdr.Registry.
// A source of DID could be issuer of VC or holder of VP. It can be also obtained from
// JWS "issuer" claim or "verificationMethod" of Linked Data Proof.
type VDRKeyResolver struct {
	resolver api.DIDResolver
}

// NewVDRKeyResolver creates VDRKeyResolver.
func NewVDRKeyResolver(resolver api.DIDResolver) *VDRKeyResolver {
	return &VDRKeyResolver{resolver: resolver}
}

func (r *VDRKeyResolver) resolvePublicKey(issuerDID, keyID string) (*verifier.PublicKey, error) {
	docResolution, err := r.resolver.Resolve(issuerDID)
	if err != nil {
		return nil, fmt.Errorf("resolve DID %s: %w", issuerDID, err)
	}

	for _, verifications := range docResolution.DIDDocument.VerificationMethods() {
		for _, verification := range verifications {
			if strings.Contains(verification.VerificationMethod.ID, keyID) {
				return &verifier.PublicKey{
					Type:  verification.VerificationMethod.Type,
					Value: verification.VerificationMethod.Value,
					JWK:   verification.VerificationMethod.JSONWebKey(),
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("public key with KID %s is not found for DID %s", keyID, issuerDID)
}

// PublicKeyFetcher returns Public Key Fetcher via DID resolution mechanism.
func (r *VDRKeyResolver) PublicKeyFetcher() verifiable.PublicKeyFetcher {
	return r.resolvePublicKey
}
