/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ldproof

import (
	"crypto"
	"encoding/base64"
	"fmt"

	"github.com/piprate/json-gold/ld"
	diddoc "github.com/trustbloc/did-go/doc/did"
	"github.com/trustbloc/did-go/doc/ld/processor"
	"github.com/trustbloc/kms-go/spi/kms"
	"github.com/trustbloc/vc-go/presexch"
	"github.com/trustbloc/vc-go/proof"
	"github.com/trustbloc/vc-go/proof/creator"
	"github.com/trustbloc/vc-go/proof/ldproofs/ecdsasecp256k1signature2019"
	"github.com/trustbloc/vc-go/proof/ldproofs/ed25519signature2018"
	"github.com/trustbloc/vc-go/proof/ldproofs/ed25519signature2020"
	"github.com/trustbloc/vc-go/proof/ldproofs/jsonwebsignature2020"
	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/pkg/api"
)

const (
	proofPurpose = "authentication"
)

var supportedLDProofTypes = map[string]proof.LDProofDescriptor{
	ecdsasecp256k1signature2019.ProofType: ecdsasecp256k1signature2019.New(),
	ed25519signature2018.ProofType:        ed25519signature2018.New(),
	ed25519signature2020.ProofType:        ed25519signature2020.New(),
	jsonwebsignature2020.ProofType:        jsonwebsignature2020.New(),
}

// LDProof implements functionality of adding linked data proof to the verifiable presentation.
type LDProof struct {
	crypto            api.Crypto
	documentLoader    ld.DocumentLoader
	ldProofDescriptor proof.LDProofDescriptor
	keyType           kms.KeyType
}

// New returns a new instance of LDProof.
func New(
	crypto api.Crypto,
	documentLoader ld.DocumentLoader,
	ldpVPFormat *presexch.LdpType,
	keyType kms.KeyType,
) (*LDProof, error) {
	var (
		proofDesc proof.LDProofDescriptor
		found     bool
	)

	for _, proofType := range ldpVPFormat.ProofType {
		if proofDesc, found = supportedLDProofTypes[proofType]; found && isKeyTypeSupported(proofDesc, keyType) {
			break
		}
		found = false
	}

	if !found {
		return nil, fmt.Errorf("no supported linked data proof found")
	}

	return &LDProof{
		crypto:            crypto,
		documentLoader:    documentLoader,
		ldProofDescriptor: proofDesc,
		keyType:           keyType,
	}, nil
}

func isKeyTypeSupported(ldProof proof.LDProofDescriptor, keyType kms.KeyType) bool {
	for _, vm := range ldProof.SupportedVerificationMethods() {
		if vm.KMSKeyType == keyType {
			return true
		}
	}

	return false
}

// Add adds linked data proof to the verifiable presentation.
func (p *LDProof) Add(vp *verifiable.Presentation, opts ...Opt) error {
	o := &options{}

	for _, opt := range opts {
		opt(o)
	}

	if err := p.validateSigningVerificationMethod(o); err != nil {
		return err
	}

	signer, err := p.createSigner(o)
	if err != nil {
		return err
	}

	p.updatePresentationContext(vp)

	proofContext := &verifiable.LinkedDataProofContext{
		SignatureType:           p.ldProofDescriptor.ProofType(),
		ProofCreator:            creator.New(creator.WithLDProofType(p.ldProofDescriptor, signer)),
		KeyType:                 p.keyType,
		SignatureRepresentation: verifiable.SignatureProofValue,
		VerificationMethod:      o.signingVM.ID,
		Challenge:               o.nonce,
		Domain:                  o.domain,
		Purpose:                 proofPurpose,
	}

	return vp.AddLinkedDataProof(
		proofContext,
		processor.WithDocumentLoader(p.documentLoader),
	)
}

func (p *LDProof) validateSigningVerificationMethod(opts *options) error {
	if opts.signingVM == nil {
		return fmt.Errorf("missing signing verification method")
	}

	if jwk := opts.signingVM.JSONWebKey(); jwk == nil {
		return fmt.Errorf("missing jwk for %s verification method", opts.signingVM.ID)
	}

	return nil
}

func (p *LDProof) createSigner(opts *options) (*cryptoSigner, error) {
	jwk := opts.signingVM.JSONWebKey()

	tb, err := jwk.Thumbprint(crypto.SHA256)
	if err != nil {
		return nil, fmt.Errorf("create crypto thumbprint for jwk: %w", err)
	}

	return &cryptoSigner{
		crypto: p.crypto,
		keyID:  base64.RawURLEncoding.EncodeToString(tb),
	}, nil
}

func (p *LDProof) updatePresentationContext(vp *verifiable.Presentation) {
	vp.Context = append(vp.Context,
		"https://www.w3.org/ns/credentials/examples/v2",
		"https://w3c-ccg.github.io/lds-jws2020/contexts/lds-jws2020-v1.json",
	)
}

type options struct {
	signingVM *diddoc.VerificationMethod
	nonce     string
	domain    string
}

// Opt is an option for adding linked data proof.
type Opt func(opts *options)

// WithSigningVM sets signing verification method.
func WithSigningVM(vm *diddoc.VerificationMethod) Opt {
	return func(opts *options) {
		opts.signingVM = vm
	}
}

// WithNonce sets nonce.
func WithNonce(nonce string) Opt {
	return func(opts *options) {
		opts.nonce = nonce
	}
}

// WithDomain sets domain.
func WithDomain(domain string) Opt {
	return func(opts *options) {
		opts.domain = domain
	}
}

type cryptoSigner struct {
	crypto api.Crypto
	keyID  string
}

func (s *cryptoSigner) Sign(data []byte) ([]byte, error) {
	return s.crypto.Sign(data, s.keyID)
}
