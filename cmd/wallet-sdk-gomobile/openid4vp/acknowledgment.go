/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/trustbloc/wallet-sdk/pkg/openid4vp"
)

// Acknowledgment represents an object that allows to acknowledge the verifier on presentation request status.
type Acknowledgment struct {
	acknowledgment *openid4vp.Acknowledgment
}

// NewAcknowledgment recreates acknowledgment object from serialized state.
func NewAcknowledgment(serialized string) (*Acknowledgment, error) {
	acknowledgment := &openid4vp.Acknowledgment{}

	err := json.Unmarshal([]byte(serialized), acknowledgment)
	if err != nil {
		return nil, fmt.Errorf("invalid requested acknowledgment json structure: %w", err)
	}

	return &Acknowledgment{acknowledgment: acknowledgment}, nil
}

// Serialize the acknowledgment object so it can be restored later.
func (a *Acknowledgment) Serialize() (string, error) {
	data, err := json.Marshal(a.acknowledgment)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// SetInteractionDetails extends acknowledgment request with serializedInteractionDetails.
func (a *Acknowledgment) SetInteractionDetails(serializedInteractionDetails string) error {
	if err := json.Unmarshal([]byte(serializedInteractionDetails), &a.acknowledgment.InteractionDetails); err != nil {
		return fmt.Errorf("decode vp ack interaction details: %w", err)
	}

	return nil
}

// NoConsent acknowledge verifier that user does not consent to the presentation request.
func (a *Acknowledgment) NoConsent() error {
	return a.acknowledgment.AcknowledgeVerifier(openid4vp.AccessDeniedErrorResponse,
		openid4vp.NoConsentErrorDescription,
		&http.Client{},
	)
}

// NoMatchingCredential acknowledge verifier that no matching credential was found.
func (a *Acknowledgment) NoMatchingCredential() error {
	return a.acknowledgment.AcknowledgeVerifier(openid4vp.AccessDeniedErrorResponse,
		openid4vp.NoMatchFoundErrorDescription,
		&http.Client{},
	)
}

// WithCode sends acknowledgment message to verifier with the custom error code and description.
func (a *Acknowledgment) WithCode(code, desc string) error {
	return a.acknowledgment.AcknowledgeVerifier(code, desc, &http.Client{})
}
