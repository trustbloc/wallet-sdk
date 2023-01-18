/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credential

import (
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
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

// Issue accepts either a JSON Verifiable Credential, or a JSON-string ID for a Verifiable Credential, and signs it with
// the key identified by keyID, returning a serialized JWTVC.
func (s *Signer) Issue(credential *api.JSONObject, keyID string) ([]byte, error) {
	var credOpt credentialsigner.CredentialOpt

	if isQuoted(string(credential.Data)) {
		credOpt = credentialsigner.GivenCredentialID(unQuote(string(credential.Data)))
	} else {
		cred, err := verifiable.ParseCredential(credential.Data,
			verifiable.WithDisabledProofCheck(), verifiable.WithJSONLDDocumentLoader(s.ldLoader))
		if err != nil {
			return nil, fmt.Errorf("parsing input credential: %w", err)
		}

		credOpt = credentialsigner.GivenCredential(cred)
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

func isQuoted(s string) bool {
	return len(s) > 1 && s[0] == '"' && s[len(s)-1] == '"'
}
