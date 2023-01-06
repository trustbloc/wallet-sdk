/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credential

import (
	"errors"
	"fmt"

	"github.com/piprate/json-gold/ld"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	"github.com/trustbloc/wallet-sdk/pkg/credentialsigner"
)

// Signer issues self-signed credentials.
type Signer struct {
	signer   *credentialsigner.Signer
	ldLoader ld.DocumentLoader
}

// NewSigner initializes a credential Signer for issuing self-signed credentials.
func NewSigner(
	credReader api.CredentialReader,
	didResolver api.DIDResolver,
	crypto api.Crypto,
	ldLoader api.LDDocumentLoader,
) (*Signer, error) {
	ldLoaderWrapper := &wrapper.DocumentLoaderWrapper{DocumentLoader: ldLoader}
	readerWrapper := &wrapper.CredentialReaderWrapper{CredentialReader: credReader, DocumentLoader: ldLoaderWrapper}
	resolverWrapper := &wrapper.VDRResolverWrapper{DIDResolver: didResolver}

	sdkSigner := credentialsigner.New(readerWrapper, resolverWrapper, crypto)

	return &Signer{
		signer:   sdkSigner,
		ldLoader: ldLoaderWrapper,
	}, nil
}

// Issue signs the given Verifiable Credential with the key identified by keyID, returning a serialized JWT VC.
// The Verifiable Credential can either be provided directly or it can be specified by credID, in which case it will be
// retrieved from this Signer's CredentialReader.
func (s *Signer) Issue(credential *api.VerifiableCredential, credID, keyID string) ([]byte, error) {
	var credOpt credentialsigner.CredentialOpt

	if credential == nil && credID == "" {
		return nil, errors.New("no credential specified")
	}

	if credential != nil {
		credOpt = credentialsigner.GivenCredential(credential.VC)
	} else {
		credOpt = credentialsigner.GivenCredentialID(credID)
	}

	signedCred, err := s.signer.Issue(credOpt, &credentialsigner.ProofOptions{
		ProofFormat: credentialsigner.ExternalJWTProofFormat,
		KeyID:       keyID,
	})
	if err != nil {
		return nil, fmt.Errorf("signing credential: %w", err)
	}

	marshalledCred, err := signedCred.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshalling signed credential: %w", err)
	}

	return marshalledCred, nil
}
