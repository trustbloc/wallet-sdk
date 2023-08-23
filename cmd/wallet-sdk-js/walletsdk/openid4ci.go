/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package walletsdk

import (
	"errors"

	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	"github.com/trustbloc/wallet-sdk/pkg/models"
	"github.com/trustbloc/wallet-sdk/pkg/openid4ci"
)

// OpenID4CIIssuerInitiatedInteraction wraps openid4ci.IssuerInitiatedInteraction and necessary dependencies.
type OpenID4CIIssuerInitiatedInteraction struct {
	Interaction *openid4ci.IssuerInitiatedInteraction
	crypto      api.Crypto
}

// RequestCredentialWithPreAuth requests credential(s) from the issuer. This method can only be used for the
// pre-authorized code flow, where it acts as the final step in the Interaction with the issuer.
// For the equivalent method for the authorization code flow, see RequestCredentialWithAuth instead.
// If a PIN is required (which can be checked via the Capabilities method), then it must be passed
// into this method via the WithPIN option.
func (i *OpenID4CIIssuerInitiatedInteraction) RequestCredentialWithPreAuth(vm *models.VerificationMethod, pin string,
) ([]*verifiable.Credential, error) {
	signer, err := i.createSigner(vm)
	if err != nil {
		return nil, err
	}

	return i.Interaction.RequestCredentialWithPreAuth(signer, openid4ci.WithPIN(pin))
}

func (i *OpenID4CIIssuerInitiatedInteraction) createSigner(vm *models.VerificationMethod) (*common.JWSSigner, error) {
	if vm == nil {
		return nil, errors.New("verification method must be provided")
	}

	signer, err := common.NewJWSSigner(vm, i.crypto)
	if err != nil {
		return nil, err
	}

	return signer, nil
}
