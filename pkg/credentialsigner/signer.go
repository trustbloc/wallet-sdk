/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package credentialsigner implements verifiable credential signing for self-issued credentials.
package credentialsigner

import (
	"errors"
	"fmt"
	"strings"

	diddoc "github.com/trustbloc/did-go/doc/did"
	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	"github.com/trustbloc/wallet-sdk/pkg/models"
)

// Signer signs credentials.
type Signer struct {
	didResolver api.DIDResolver
	crypto      api.Crypto
}

// New initializes a credential Signer.
func New(didResolver api.DIDResolver, crypto api.Crypto) *Signer {
	return &Signer{
		didResolver: didResolver,
		crypto:      crypto,
	}
}

// ProofFormat determines whether a credential or presentation should be signed with an external JWT proof
// (wrapping the credential to form a JWT-VC) or with an embedded LD proof.
type ProofFormat string

const (
	// ExternalJWTProofFormat indicates that a credential or presentation should be signed with an external JWT proof.
	ExternalJWTProofFormat = "ExternalJWTProofFormat"
	// EmbeddedLDProofFormat indicates that a credential or presentation should be signed with an embedded LD proof.
	EmbeddedLDProofFormat = "EmbeddedLDProofFormat"
	// number of sections in verification method.
	vmSectionCount = 2
)

// ProofOptions contains options for issuing a credential.
type ProofOptions struct {
	// ProofFormat determines the format of the issued credential,
	// either ExternalJWTProofFormat or EmbeddedLDProofFormat.
	ProofFormat ProofFormat
	// KeyID is the DID-url key identifier for the signing key to use to issue the credential.
	KeyID string
}

// Issue signs the given credential.
func (s *Signer) Issue(credential *verifiable.Credential, proofOptions *ProofOptions) (*verifiable.Credential, error) {
	if credential == nil {
		return nil, errors.New("no credential provided")
	}

	switch proofOptions.ProofFormat {
	case ExternalJWTProofFormat:
		return s.issueJWTVC(credential, proofOptions)
	case EmbeddedLDProofFormat:
		return nil, fmt.Errorf("JSON-LD proof format not currently supported")
	default:
		return nil, fmt.Errorf("proof format not recognized")
	}
}

func (s *Signer) issueJWTVC(unsignedVC *verifiable.Credential, proofOptions *ProofOptions,
) (*verifiable.Credential, error) {
	docVM, fullKID, _, err := resolveSigningVMWithRelationship(proofOptions.KeyID, s.didResolver)
	if err != nil {
		return nil, fmt.Errorf("resolving verification method for signing key: %w", err)
	}

	vm := models.VerificationMethodFromDoc(docVM)

	jwtSigner, err := common.NewJWSSigner(vm, s.crypto)
	if err != nil {
		return nil, fmt.Errorf("initializing jwt signer: %w", err)
	}

	vcAlg, err := algByName(jwtSigner.Algorithm())
	if err != nil {
		return nil, err
	}

	vc, err := unsignedVC.CreateSignedJWTVC(false, vcAlg, jwtSigner, fullKID)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT VC: %w", err)
	}

	return vc, nil
}

func algByName(alg string) (verifiable.JWSAlgorithm, error) {
	switch alg {
	case "RS256":
		return verifiable.RS256, nil
	case "PS256":
		return verifiable.PS256, nil
	case "EdDSA":
		return verifiable.EdDSA, nil
	case "ES256K":
		return verifiable.ECDSASecp256k1, nil
	case "ES256":
		return verifiable.ECDSASecp256r1, nil
	case "ES384":
		return verifiable.ECDSASecp384r1, nil
	case "ES521":
		return verifiable.ECDSASecp521r1, nil
	default:
		return -1, fmt.Errorf("unsupported algorithm: %v", alg)
	}
}

// resolveSigningVMWithRelationship resolves a DID KeyID using the given did resolver, and returns either:
//
//   - the Verification Method identified by the given key ID, or
//   - the first Assertion Method in the DID doc, if the DID provided has no fragment component.
//
// Returns:
//   - a verification method suitable for signing.
//   - the full DID#KID identifier of the returned verification method.
//   - the name of the signing-supporting verification relationship found for this verification method.
func resolveSigningVMWithRelationship(
	kid string,
	didResolver api.DIDResolver,
) (*diddoc.VerificationMethod, string, string, error) {
	vmSplit := strings.Split(kid, "#")

	if len(vmSplit) > vmSectionCount {
		return nil, "", "", errors.New("invalid verification method format")
	}

	signingDID := vmSplit[0]

	docRes, err := didResolver.Resolve(signingDID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to resolve signing DID: %w", err)
	}

	if len(vmSplit) == 1 {
		// look for assertionmethod
		verificationMethods := docRes.DIDDocument.VerificationMethods(diddoc.AssertionMethod)

		if len(verificationMethods[diddoc.AssertionMethod]) > 0 {
			vm := verificationMethods[diddoc.AssertionMethod][0].VerificationMethod

			return &vm, fullVMID(signingDID, vm.ID), "assertionMethod", nil
		}

		return nil, "", "", fmt.Errorf("DID provided has no assertion method to use as a default signing key")
	}

	vmID := vmSplit[vmSectionCount-1]

	for _, verifications := range docRes.DIDDocument.VerificationMethods() {
		for _, verification := range verifications {
			if isSigningKey(verification.Relationship) && vmIDFragmentOnly(verification.VerificationMethod.ID) == vmID {
				vm := verification.VerificationMethod

				return &vm, kid, verificationRelationshipName(verification.Relationship), nil
			}
		}
	}

	return nil, "", "", fmt.Errorf("did document has no verification method with given ID")
}

func fullVMID(did, vmID string) string {
	vmIDSplit := strings.Split(vmID, "#")

	if len(vmIDSplit) == 1 {
		return did + "#" + vmIDSplit[0]
	} else if vmIDSplit[0] == "" {
		return did + "#" + vmIDSplit[1]
	}

	return vmID
}

func verificationRelationshipName(rel diddoc.VerificationRelationship) string {
	switch rel { //nolint:exhaustive
	case diddoc.VerificationRelationshipGeneral:
		return ""
	case diddoc.AssertionMethod:
		return "assertionMethod"
	case diddoc.Authentication:
		return "authentication"
	}

	return ""
}

func vmIDFragmentOnly(vmID string) string {
	vmSplit := strings.Split(vmID, "#")
	if len(vmSplit) == 1 {
		return vmSplit[0]
	}

	return vmSplit[1]
}

func isSigningKey(vr diddoc.VerificationRelationship) bool {
	switch vr { //nolint:exhaustive
	case diddoc.AssertionMethod, diddoc.Authentication, diddoc.VerificationRelationshipGeneral:
		return true
	}

	return false
}
