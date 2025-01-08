/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package ldproof contains a function for adding linked data proof to a verifiable presentation.
package ldproof

import (
	"fmt"
	"strings"

	"github.com/piprate/json-gold/ld"
	"github.com/samber/lo"
	"github.com/trustbloc/did-go/doc/did"
	"github.com/trustbloc/did-go/doc/ld/processor"
	vdrapi "github.com/trustbloc/did-go/vdr/api"
	"github.com/trustbloc/kms-go/doc/util/jwkkid"
	"github.com/trustbloc/kms-go/spi/kms"
	"github.com/trustbloc/vc-go/dataintegrity"
	"github.com/trustbloc/vc-go/dataintegrity/suite"
	"github.com/trustbloc/vc-go/dataintegrity/suite/ecdsa2019"
	"github.com/trustbloc/vc-go/dataintegrity/suite/eddsa2022"
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

//nolint:gochecknoglobals
var supportedLDProofTypes = map[string]proof.LDProofDescriptor{
	ecdsasecp256k1signature2019.ProofType: ecdsasecp256k1signature2019.New(),
	ed25519signature2018.ProofType:        ed25519signature2018.New(),
	ed25519signature2020.ProofType:        ed25519signature2020.New(),
	jsonwebsignature2020.ProofType:        jsonwebsignature2020.New(),
}

//nolint:gochecknoglobals
var supportedDIKeyTypes = map[string][]kms.KeyType{
	ecdsa2019.SuiteTypeNew: {kms.ECDSAP256TypeIEEEP1363, kms.ECDSAP384TypeIEEEP1363},
	eddsa2022.SuiteType:    {kms.ED25519Type},
}

// LDProof implements functionality of adding linked data proof to the verifiable presentation.
type LDProof struct {
	crypto         api.Crypto
	documentLoader ld.DocumentLoader
	didResolver    api.DIDResolver
}

// New returns a new instance of LDProof.
func New(
	crypto api.Crypto,
	documentLoader ld.DocumentLoader,
	didResolver api.DIDResolver,
) *LDProof {
	return &LDProof{
		crypto:         crypto,
		documentLoader: documentLoader,
		didResolver:    didResolver,
	}
}

// Add adds linked data proof to the verifiable presentation.
func (p *LDProof) Add(vp *verifiable.Presentation, opts ...Opt) error {
	o := &options{}

	for _, opt := range opts {
		opt(o)
	}

	keyID, keyType, err := getKeyIDAndType(o.verificationMethod)
	if err != nil {
		return fmt.Errorf("get key ID and type: %w", err)
	}

	for _, proofType := range o.ldpType.ProofType {
		// data integrity proof
		if proofType == ecdsa2019.SuiteTypeNew || proofType == eddsa2022.SuiteType {
			if keyTypes, found := supportedDIKeyTypes[proofType]; found && lo.Contains(keyTypes, keyType) {
				return p.addDataIntegrityProof(vp, proofType, keyID, o)
			}
		}

		// linked data proof
		if proofDesc, found := supportedLDProofTypes[proofType]; found && isKeyTypeSupported(proofDesc, keyType) {
			return p.addLinkedDataProof(vp, proofDesc, keyID, keyType, o)
		}
	}

	return fmt.Errorf("no supported ldp types found")
}

func getKeyIDAndType(vm *did.VerificationMethod) (string, kms.KeyType, error) {
	if vm == nil {
		return "", "", fmt.Errorf("missing verification method")
	}

	if jwk := vm.JSONWebKey(); jwk != nil {
		keyType, err := jwk.KeyType()
		if err != nil {
			return "", "", fmt.Errorf("get key type from jwk: %w", err)
		}

		return jwk.KeyID, keyType, nil
	}

	if len(vm.Value) == 0 {
		return "", "", fmt.Errorf("missing key value for %s verification method", vm.ID)
	}

	switch vm.Type {
	case ed25519signature2018.VerificationMethodType,
		ed25519signature2020.VerificationMethodType:
		kid, err := jwkkid.CreateKID(vm.Value, kms.ED25519Type)
		if err != nil {
			return "", "", fmt.Errorf("failed to generate key ID for ed25519 key: %w", err)
		}

		return kid, kms.ED25519Type, nil
	}

	return "", "", fmt.Errorf("unsupported verification method type: %s", vm.Type)
}

func fullVMID(id, vmID string) string {
	if vmID == "" {
		return id
	}

	if vmID[0] == '#' {
		return id + vmID
	}

	if strings.HasPrefix(vmID, "did:") {
		return vmID
	}

	return id + "#" + vmID
}

func isKeyTypeSupported(ldProof proof.LDProofDescriptor, keyType kms.KeyType) bool {
	for _, vm := range ldProof.SupportedVerificationMethods() {
		if vm.KMSKeyType == keyType {
			return true
		}
	}

	return false
}

func (p *LDProof) addDataIntegrityProof(
	vp *verifiable.Presentation,
	dataIntegritySuite, keyID string,
	o *options,
) error {
	var initializer suite.SignerInitializer

	switch dataIntegritySuite {
	case ecdsa2019.SuiteTypeNew:
		initializer = ecdsa2019.NewSignerInitializer(
			&ecdsa2019.SignerInitializerOptions{
				LDDocumentLoader: p.documentLoader,
				SignerGetter:     ecdsa2019.WithStaticSigner(p.createSigner(keyID)),
			},
		)
	case eddsa2022.SuiteType:
		initializer = eddsa2022.NewSignerInitializer(
			&eddsa2022.SignerInitializerOptions{
				LDDocumentLoader: p.documentLoader,
				SignerGetter:     eddsa2022.WithStaticSigner(p.createSigner(keyID)),
			},
		)
	default:
		return fmt.Errorf("unsupported data integrity suite: %s", dataIntegritySuite)
	}

	dataIntegritySigner, err := dataintegrity.NewSigner(
		&dataintegrity.Options{
			DIDResolver: &didResolverWrapper{
				didResolver: p.didResolver,
			},
		},
		initializer,
	)
	if err != nil {
		return err
	}

	proofContext := &verifiable.DataIntegrityProofContext{
		SigningKeyID: fullVMID(o.did, o.verificationMethod.ID),
		CryptoSuite:  dataIntegritySuite,
		ProofPurpose: proofPurpose,
		Challenge:    o.challenge,
		Domain:       o.domain,
	}

	return vp.AddDataIntegrityProof(proofContext, dataIntegritySigner)
}

func (p *LDProof) addLinkedDataProof(
	vp *verifiable.Presentation,
	proofDesc proof.LDProofDescriptor,
	keyID string,
	keyType kms.KeyType,
	o *options,
) error {
	proofContext := &verifiable.LinkedDataProofContext{
		SignatureType:           proofDesc.ProofType(),
		ProofCreator:            creator.New(creator.WithLDProofType(proofDesc, p.createSigner(keyID))),
		KeyType:                 keyType,
		SignatureRepresentation: verifiable.SignatureProofValue,
		VerificationMethod:      o.verificationMethod.ID,
		Challenge:               o.challenge,
		Domain:                  o.domain,
		Purpose:                 proofPurpose,
	}

	p.updatePresentationContext(vp)

	return vp.AddLinkedDataProof(proofContext,
		processor.WithDocumentLoader(p.documentLoader))
}

func (p *LDProof) updatePresentationContext(vp *verifiable.Presentation) {
	vp.Context = append(vp.Context,
		"https://www.w3.org/ns/credentials/examples/v2",
		"https://w3c-ccg.github.io/lds-jws2020/contexts/lds-jws2020-v1.json",
	)
}

type signer struct {
	crypto api.Crypto
	keyID  string
}

func (s *signer) Sign(msg []byte) ([]byte, error) {
	return s.crypto.Sign(msg, s.keyID)
}

func (p *LDProof) createSigner(keyID string) *signer {
	return &signer{
		crypto: p.crypto,
		keyID:  keyID,
	}
}

type didResolverWrapper struct {
	didResolver api.DIDResolver
}

func (d *didResolverWrapper) Resolve(id string, _ ...vdrapi.DIDMethodOption) (*did.DocResolution, error) {
	return d.didResolver.Resolve(id)
}
