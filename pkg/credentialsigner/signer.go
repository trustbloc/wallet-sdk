/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package credentialsigner implements verifiable credential signing for self-issued credentials.
package credentialsigner

import (
	"fmt"

	diddoc "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/doc/util/didsignjwt"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdr"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	"github.com/trustbloc/wallet-sdk/pkg/models"
)

// Signer signs credentials.
type Signer struct {
	credReader  api.CredentialReader
	didResolver api.DIDResolver
	crypto      api.Crypto
}

// New initializes a credential Signer.
func New(credReader api.CredentialReader, didResolver api.DIDResolver, crypto api.Crypto) *Signer {
	return &Signer{
		credReader:  credReader,
		didResolver: didResolver,
		crypto:      crypto,
	}
}

type credOptData struct {
	vc   *verifiable.Credential
	vcID string
}

// CredentialOpt provides the credential for Signer to sign.
type CredentialOpt func(data *credOptData)

// GivenCredential provides a verifiable.Credential to Signer.
func GivenCredential(credential *verifiable.Credential) CredentialOpt {
	return func(data *credOptData) {
		data.vc = credential
	}
}

// GivenCredentialID provides a credential ID for a credential for Signer to load.
func GivenCredentialID(credID string) CredentialOpt {
	return func(data *credOptData) {
		data.vcID = credID
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
func (s *Signer) Issue(credential CredentialOpt, proofOptions *ProofOptions) (*verifiable.Credential, error) {
	vc, err := s.readCredential(credential)
	if err != nil {
		return nil, err
	}

	switch proofOptions.ProofFormat {
	case ExternalJWTProofFormat:
		return s.issueJWTVC(vc, proofOptions)
	case EmbeddedLDProofFormat:
		return nil, fmt.Errorf("JSON-LD proof format not currently supported")
	default:
		return nil, fmt.Errorf("proof format not recognized")
	}
}

func (s *Signer) readCredential(credential CredentialOpt) (*verifiable.Credential, error) {
	input := &credOptData{}
	credential(input)

	if input.vc != nil {
		return input.vc, nil
	}

	if input.vcID == "" {
		return nil, fmt.Errorf("no Credential provided")
	}

	if s.credReader == nil {
		return nil, fmt.Errorf("credential ID provided for fetching, but Signer instance does not have a CredentialReader")
	}

	vc, err := s.credReader.Get(input.vcID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch credential: %w", err)
	}

	return vc, nil
}

func (s *Signer) issueJWTVC(vc *verifiable.Credential, proofOptions *ProofOptions) (*verifiable.Credential, error) {
	docVM, fullKID, err := didsignjwt.ResolveSigningVM(proofOptions.KeyID, &didResolverWrapper{didResolver: s.didResolver})
	if err != nil {
		return nil, fmt.Errorf("resolving verification method for signing key: %w", err)
	}

	vm := models.VerificationMethodFromDoc(docVM)

	jwtSigner, err := common.NewJWSSigner(vm, s.crypto)
	if err != nil {
		return nil, fmt.Errorf("initializing jwt signer: %w", err)
	}

	claims, err := vc.JWTClaims(false)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT claims for VC: %w", err)
	}

	alg, hasAlg := jwtSigner.Headers().Algorithm()
	if !hasAlg {
		return nil, fmt.Errorf("signer missing algorithm header")
	}

	vcAlg, err := algByName(alg)
	if err != nil {
		return nil, err
	}

	jws, err := claims.MarshalJWS(vcAlg, &signerWrapper{jwtSigner}, fullKID)
	if err != nil {
		return nil, fmt.Errorf("failed to sign JWT VC: %w", err)
	}

	vc.JWT = jws

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

type signerWrapper struct {
	signer api.JWTSigner
}

// Sign wraps api.JWTSigner.
func (s *signerWrapper) Sign(data []byte) ([]byte, error) {
	return s.signer.Sign(data)
}

// Alg returns the alg field from api.JWTSigner Headers().
func (s *signerWrapper) Alg() string {
	alg, _ := s.signer.Headers().Algorithm()

	return alg
}

type didResolverWrapper struct {
	didResolver api.DIDResolver
}

func (d *didResolverWrapper) Resolve(did string, _ ...vdr.DIDMethodOption) (*diddoc.DocResolution, error) {
	return d.didResolver.Resolve(did)
}
