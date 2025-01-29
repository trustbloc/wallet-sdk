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

// SetInteractionDetails extends next acknowledgment request with serializedInteractionDetails.
func (a *Acknowledgment) SetInteractionDetails(serializedInteractionDetails string) error {
	var interactionDetails map[string]interface{}

	if err := json.Unmarshal([]byte(serializedInteractionDetails), &interactionDetails); err != nil {
		return fmt.Errorf("decode ci ack interaction details: %w", err)
	}

	a.acknowledgment.InteractionDetails = interactionDetails

	return nil
}

// Success acknowledges the client's acceptance of credentials. Each call to this function
// acknowledges the client's acceptance of the next credential in the list of issued credentials.
//
// The first call acknowledges the first credential, the second call acknowledges the second credential,
// the third call acknowledges the third, and so on. If the number of function calls exceeds the number
// of credentials issued in the current session, the function returns an error "ack list is empty".
//
// Between the calls caller might set different interaction details using SetInteractionDetails.
func (a *Acknowledgment) Success() error {
	return a.acknowledgment.AcknowledgeIssuer(openid4cigoapi.EventStatusCredentialAccepted, &http.Client{})
}

// Reject acknowledges the client's rejection of credentials. Each call to this function
// acknowledges the client's rejection of the next credential in the list of issued credentials.
//
// The first call acknowledges the first credential, the second call acknowledges the second credential,
// the third call acknowledges the third, and so on. If the number of function calls exceeds the number
// of credentials issued in the current session, the function returns an error "ack list is empty".
//
// Between the calls caller might set different interaction details using SetInteractionDetails.
func (a *Acknowledgment) Reject() error {
	return a.acknowledgment.AcknowledgeIssuer(openid4cigoapi.EventStatusCredentialFailure, &http.Client{})
}

// RejectWithCode acknowledges the client's rejection of credentials with specific code.
// See Reject for details.
func (a *Acknowledgment) RejectWithCode(code string) error {
	return a.acknowledgment.AcknowledgeIssuer(openid4cigoapi.EventStatus(code), &http.Client{})
}
