/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package localkms

import (
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcutil/base58"
	"github.com/hyperledger/aries-framework-go/pkg/crypto/tinkcrypto"
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/doc/jose/jwk"
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

// NewSignerCreator used to generate constructor for SignerCreator.
func NewSignerCreator(kms *KMS) (*SignerCreator, error) {
	return CreateSignerCreator(kms)
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
		vm, err := workaroundUnmarshalVM(verificationMethod.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal verification method JSON into a did.VerificationMethod")
		}

		ariesSigner, err := ariesSignerCreator(vm)
		if err != nil {
			return nil, fmt.Errorf("failed to create Aries signer: %w", err)
		}

		return ariesSigner, nil
	}

	return &SignerCreator{didJWTSignerCreate: gomobileSignerCreator}, nil
}

func workaroundUnmarshalVM(vmBytes []byte) (*did.VerificationMethod, error) {
	vmMap := map[string]json.RawMessage{}

	err := json.Unmarshal(vmBytes, &vmMap)
	if err != nil {
		return nil, err
	}

	var (
		jsonKey    jwk.JWK
		id         string
		typ        string
		controller string
		value      []byte
	)

	id = stringEntry(vmMap["id"])
	typ = stringEntry(vmMap["type"])
	controller = stringEntry(vmMap["controller"])

	rawValue := stringEntry(vmMap["publicKeyBase58"])
	if rawValue != "" {
		value = base58.Decode(rawValue)
	}

	if jwkBytes, ok := vmMap["publicKeyJwk"]; ok {
		err = json.Unmarshal(jwkBytes, &jsonKey)
		if err != nil {
			return nil, err
		}
	}

	if jsonKey.Valid() {
		return did.NewVerificationMethodFromJWK(id, typ, controller, &jsonKey)
	}

	return did.NewVerificationMethodFromBytes(id, typ, controller, value), nil
}

func stringEntry(entry json.RawMessage) string {
	if len(entry) > 1 {
		return string(entry[1 : len(entry)-1])
	}

	return ""
}
