/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credential

import (
	"errors"
	"fmt"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	"github.com/trustbloc/wallet-sdk/pkg/credentialsigner"
)

// Signer issues self-signed credentials.
type Signer struct {
	signer *credentialsigner.Signer
}

// NewSigner initializes a credential Signer for issuing self-signed credentials.
func NewSigner(didResolver api.DIDResolver, crypto api.Crypto) *Signer {
	resolverWrapper := &wrapper.VDRResolverWrapper{DIDResolver: didResolver}

	sdkSigner := credentialsigner.New(resolverWrapper, crypto)

	return &Signer{
		signer: sdkSigner,
	}
}

// Issue signs the given Verifiable Credential with the key identified by keyID, returning the signed VC.
func (s *Signer) Issue(credential *verifiable.Credential, keyID string) (*verifiable.Credential, error) {
	if credential == nil {
		return nil, errors.New("no credential specified")
	}

	signedCred, err := s.signer.Issue(credential.VC, &credentialsigner.ProofOptions{
		ProofFormat: credentialsigner.ExternalJWTProofFormat,
		KeyID:       keyID,
	})
	if err != nil {
		return nil, fmt.Errorf("signing credential: %w", err)
	}

	return verifiable.NewCredential(signedCred), nil
}
