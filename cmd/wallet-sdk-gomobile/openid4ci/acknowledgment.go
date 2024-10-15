package openid4ci

import (
	"encoding/json"
	"fmt"
	"net/http"

	openid4cigoapi "github.com/trustbloc/wallet-sdk/pkg/openid4ci"
)

// Acknowledgment represents an object that allows to acknowledge the issuer the user's accepted or rejected credential.
type Acknowledgment struct {
	acknowledgment *openid4cigoapi.Acknowledgment
}

// NewAcknowledgment recreates acknowledgment object from serialized state.
func NewAcknowledgment(serialized string) (*Acknowledgment, error) {
	acknowledgment := &openid4cigoapi.Acknowledgment{}

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
		return fmt.Errorf("decode ci ack interaction details: %w", err)
	}

	return nil
}

// Success acknowledge issuer that client accepts credentials.
func (a *Acknowledgment) Success() error {
	return a.acknowledgment.AcknowledgeIssuer(openid4cigoapi.EventStatusCredentialAccepted, &http.Client{})
}

// Reject acknowledge issuer that client rejects credentials.
func (a *Acknowledgment) Reject() error {
	return a.acknowledgment.AcknowledgeIssuer(openid4cigoapi.EventStatusCredentialFailure, &http.Client{})
}

// RejectWithCode sends rejection message to issuer with a reject code.
func (a *Acknowledgment) RejectWithCode(code string) error {
	return a.acknowledgment.AcknowledgeIssuer(openid4cigoapi.EventStatus(code), &http.Client{})
}
