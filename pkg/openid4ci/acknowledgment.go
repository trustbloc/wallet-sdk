/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
)

// Acknowledgment represents an object that allows to acknowledge the issuer the user's accepted or rejected credential.
type Acknowledgment struct {
	AckIDs                []string               `json:"ack_ids,omitempty"`
	CredentialAckEndpoint string                 `json:"credential_ack_endpoint,omitempty"`
	IssuerURI             string                 `json:"issuer_uri,omitempty"`
	AuthToken             *universalAuthToken    `json:"auth_token,omitempty"`
	InteractionDetails    map[string]interface{} `json:"interaction_details,omitempty"`
}

// AcknowledgeIssuer acknowledge issuer that client accepts or rejects credentials using first existing AckIDs.
func (a *Acknowledgment) AcknowledgeIssuer(
	eventStatus EventStatus, httpClient *http.Client,
) error {
	if len(a.AckIDs) == 0 {
		return errors.New("ack list is empty")
	}

	// Pull first ackID
	ackID := a.AckIDs[0]

	// Reduce slice size
	a.AckIDs = a.AckIDs[1:]

	return a.sendAcknowledge(ackID, eventStatus, httpClient)
}

func (a *Acknowledgment) sendAcknowledge(
	ackID string, eventStatus EventStatus, httpClient *http.Client,
) error {
	ackRequest := acknowledgementRequest{
		Event:              eventStatus,
		EventDescription:   nil,
		IssuerIdentifier:   a.IssuerURI,
		NotificationID:     ackID,
		InteractionDetails: a.InteractionDetails,
	}

	err := a.sendAcknowledgeRequest(ackRequest, httpClient)
	if err != nil {
		return fmt.Errorf("send acknowledge request id %s: %w", ackID, err)
	}

	return nil
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
			fmt.Printf("failed to close response body: %s\n", errClose.Error())
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
