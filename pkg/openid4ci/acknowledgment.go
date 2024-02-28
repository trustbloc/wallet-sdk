package openid4ci

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
)

// Acknowledgment represents an object that allows to acknowledge the issuer the user's accepted or rejected credential.
type Acknowledgment struct {
	AckIDs                []string            `json:"ack_ids,omitempty"`
	CredentialAckEndpoint string              `json:"credential_ack_endpoint,omitempty"`
	IssuerURI             string              `json:"issuer_uri,omitempty"`
	AuthToken             *universalAuthToken `json:"auth_token,omitempty"`
}

// AcknowledgeIssuer acknowledge issuer that client accepts or rejects credentials.
func (a *Acknowledgment) AcknowledgeIssuer(
	eventStatus EventStatus, httpClient *http.Client,
) error {
	var ackRequest acknowledgementRequest

	for _, ackID := range a.AckIDs {
		ackRequest.Credentials = append(ackRequest.Credentials, credentialAcknowledgement{
			NotificationID:   ackID,
			Event:            eventStatus,
			EventDescription: nil,
			IssuerIdentifier: a.IssuerURI,
		})
	}

	return a.sendAcknowledgeRequest(ackRequest, httpClient)
}

func (a *Acknowledgment) sendAcknowledgeRequest(
	acknowledgementRequest acknowledgementRequest, httpClient *http.Client,
) error {
	askEndpointURL := a.CredentialAckEndpoint

	requestBytes, err := json.Marshal(acknowledgementRequest)
	if err != nil {
		return fmt.Errorf("fail to marshal acknowledgementRequest: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodPost, askEndpointURL, bytes.NewBuffer(requestBytes))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	httpClient = createOAuthHTTPClient(&oauth2.Config{}, a.AuthToken, httpClient)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		errClose := resp.Body.Close()
		if errClose != nil {
			println(fmt.Sprintf("failed to close response body: %s", errClose.Error()))
		}
	}()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return processCredentialErrorResponse(resp.StatusCode, respBytes)
	}

	return nil
}
