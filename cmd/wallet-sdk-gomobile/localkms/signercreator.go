/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package localkms

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/crypto/tinkcrypto"
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/doc/util/didsignjwt"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

// SignerCreator is an api.DIDJWTSignerCreator implementation that uses an in-memory KMS and the aries-framework-go
// Tink crypto implementation.
type SignerCreator struct {
	didJWTSignerCreate func(verificationMethod *api.JSONObject) (api.Signer, error)
}

// Create returns a corresponding Signer type for the given DID doc verificationMethod object.
func (l *SignerCreator) Create(verificationMethod *api.JSONObject) (api.Signer, error) {
	return l.didJWTSignerCreate(verificationMethod)
}

// CreateSignerCreator returns a type that can be passed in to an OpenID4CI interaction to facilitate JWT signing.
// It uses an in-memory KMS and the aries-framework-go Tink crypto implementation for signing operations.
func CreateSignerCreator(kms *KMS) (*SignerCreator, error) {
	tinkCrypto, err := tinkcrypto.New()
	if err != nil {
		return nil, err
	}

	ariesKMS := kms.goAPILocalKMS.GetAriesKMS() //nolint: staticcheck // Will be removed in a future version

	ariesSignerCreator := didsignjwt.UseDefaultSigner(ariesKMS, tinkCrypto)

	gomobileSignerCreator := func(verificationMethod *api.JSONObject) (api.Signer, error) {
		var vm did.VerificationMethod

		err := json.Unmarshal(verificationMethod.Data, &vm)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal verification method JSON into a did.VerificationMethod")
		}

		ariesSigner, err := ariesSignerCreator(&vm)
		if err != nil {
			return nil, fmt.Errorf("failed to create Aries signer: %w", err)
		}

		return ariesSigner, nil
	}

	return &SignerCreator{didJWTSignerCreate: gomobileSignerCreator}, nil
}
