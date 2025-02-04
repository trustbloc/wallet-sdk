/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import (
	"encoding/json"
	"time"

	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

// EventStatus used to acknowledge issuer that client accepts or rejects credentials.
type EventStatus string

const (
	// EventStatusCredentialAccepted is to be used when the Credential was successfully stored in the Wallet,
	// with or without user action.
	EventStatusCredentialAccepted EventStatus = "credential_accepted" //nolint:gosec,nolintlint
	// EventStatusCredentialFailure acknowledge issuer that client rejects credentials.
	EventStatusCredentialFailure EventStatus = "credential_failure" //nolint:gosec,nolintlint
	// EventStatusCredentialDeleted is to be used when the unsuccessful Credential issuance was caused by a user action.
	EventStatusCredentialDeleted EventStatus = "credential_deleted" //nolint:gosec,nolintlint
)

// CredentialOffer represents the Credential Offer object as defined in
// https://openid.github.io/OpenID4VCI/openid-4-verifiable-credential-issuance-wg-draft.html#section-4.1.1.
type CredentialOffer struct {
	CredentialIssuer           string                            `json:"credential_issuer,omitempty"`
	CredentialConfigurationIDs []string                          `json:"credential_configuration_ids"`
	Grants                     map[string]map[string]interface{} `json:"grants,omitempty"`
}

// AuthorizeResult is the object returned from the Client.Authorize method.
// An empty/missing AuthorizationRedirectEndpoint indicates that the wallet is pre-authorized.
type AuthorizeResult struct {
	AuthorizationRedirectEndpoint string
	UserPINRequired               bool
}

// authorizationDetails is a model to convey the details about the Credentials the Client wants to obtain.
type authorizationDetails struct {
	// REQUIRED when Format parameter is not present.
	// String specifying a unique identifier of the Credential being described in the
	// credential_configurations_supported map in the Credential Issuer Metadata.
	// The referenced object in the credential_configurations_supported map conveys the details,
	// such as the format, for issuance of the requested Credential.
	// It MUST NOT be present if format parameter is present.
	CredentialConfigurationID string `json:"credential_configuration_id,omitempty"`

	// Object containing the detailed description of the credential type.
	CredentialDefinition *issuer.CredentialDefinition `json:"credential_definition,omitempty"`

	// REQUIRED when CredentialConfigurationID parameter is not present.
	// String identifying the format of the Credential the Wallet needs.
	// This Credential format identifier determines further claims in the authorization details object needed
	// to identify the Credential type in the requested format.
	// It MUST NOT be present if credential_configuration_id parameter is present.
	Format string `json:"format,omitempty"`

	// An array of strings that allows a client to specify the location of the resource server(s)
	// allowing the Authorization Server to mint audience restricted access tokens.
	Locations []string `json:"locations,omitempty"`

	// String that determines the authorization details type. MUST be set to "openid_credential" for OIDC4VC.
	Type string `json:"type"`
}

// CredentialResponse is the object returned from the Client.Callback method.
// It contains the issued credential and the credential's format.
type CredentialResponse struct {
	// OPTIONAL. Contains issued Credential.
	// It MUST be present when transaction_id is not returned.
	// It MAY be a string or an object, depending on the Credential format.
	// Deprecated: Use Credentials instead.
	Credential interface{} `json:"credential,omitempty"`
	// OPTIONAL. String identifying a Deferred Issuance transaction.
	// This claim is contained in the response if the Credential Issuer was unable to immediately issue the Credential.
	// Deprecated.
	TransactionID string `json:"transaction_id"`
	// OPTIONAL. String containing a nonce to be used to create a proof of possession of key material
	// when requesting a Credential.
	// Deprecated.
	CNonce string `json:"c_nonce"`
	// OPTIONAL. Number denoting the lifetime in seconds of the c_nonce.
	// Deprecated.
	CNonceExpiresIn int `json:"c_nonce_expires_in"`
	// OPTIONAL. String identifying an issued Credential that the Wallet includes in the Notification Request.
	AckID string `json:"notification_id"`
	// Contains an array of one or more issued Credentials.
	Credentials []CredentialResponseCredentialObject `json:"credentials"`
}

// CredentialResponseCredentialObject is a model for credentials field from credential response.
type CredentialResponseCredentialObject struct {
	Credential interface{} `json:"credential"`
}

// SerializeToCredentialsBytes serializes underlying credential to proper bytes representation depending on
// credential format.
func (r *CredentialResponse) SerializeToCredentialsBytes() ([]byte, error) {
	// TODO: https://github.com/trustbloc/wallet-sdk/issues/456 check response.Format after
	// VCS starts return valid value.
	switch cred := r.Credential.(type) {
	case string:
		return []byte(cred), nil
	default:
		return json.Marshal(cred)
	}
}

type batchCredentialResponse struct {
	CNonce              *string              `json:"c_nonce,omitempty"`
	CNonceExpiresIn     *int                 `json:"c_nonce_expires_in,omitempty"`
	CredentialResponses []CredentialResponse `json:"credential_responses"`
}

type preAuthTokenResponse struct {
	AccessToken     string `json:"access_token,omitempty"`
	TokenType       string `json:"token_type,omitempty"`
	ExpiresIn       int    `json:"expires_in,omitempty"`
	RefreshToken    string `json:"refresh_token,omitempty"`
	CNonce          string `json:"c_nonce,omitempty"`
	CNonceExpiresIn int    `json:"c_nonce_expires_in,omitempty"`
}

type universalAuthToken struct {
	AccessToken  string    `json:"access_token,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
}

type credentialRequest struct {
	CredentialDefinition *credentialDefinition `json:"credential_definition,omitempty"`
	Format               string                `json:"format,omitempty"`
	Proof                proof                 `json:"proof,omitempty"`
}

type credentialDefinition struct {
	Context           *[]string               `json:"@context,omitempty"`
	CredentialSubject *map[string]interface{} `json:"credentialSubject,omitempty"`
	Type              []string                `json:"type"`
}

type proof struct {
	ProofType       string `json:"proof_type,omitempty"`
	JWT             string `json:"jwt,omitempty"`
	CNonce          string `json:"c_nonce,omitempty"`
	CNonceExpiresIn int    `json:"c_nonce_expires_in,omitempty"`
}

type batchCredentialRequest struct {
	CredentialRequests []credentialRequest `json:"credential_requests"`
}

type errorResponse struct {
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
	// containing a nonce to be used to create a proof of possession of key material
	// when requesting a Credential.
	CNonce string `json:"c_nonce"`
	// number denoting the lifetime in seconds of the c_nonce.
	CNonceExpiresIn int `json:"c_nonce_expires_in"`
}

type acknowledgementRequest struct {
	// Type of the notification event.
	// It MUST be a case-sensitive string whose value is either `credential_accepted`, `credential_failure`,
	// or `credential_deleted`.
	//
	//  `credential_accepted` is to be used when the Credential was successfully stored in the Wallet,
	// with or without user action.
	//  `credential_deleted` is to be used when the unsuccessful Credential issuance was caused by a user action.
	//
	// In all other unsuccessful cases, `credential_failure` is to be used.
	Event EventStatus `json:"event"`

	// Human-readable ASCII text providing additional information, used to assist the Credential Issuer
	// developer in understanding the event that occurred.
	EventDescription *string `json:"event_description,omitempty"`

	// Optional issuer identifier.
	IssuerIdentifier string `json:"issuer_identifier"`

	// Ack ID.
	// String received in the Credential Response or the Batch Credential Response.
	NotificationID string `json:"notification_id"`

	InteractionDetails map[string]interface{} `json:"interaction_details,omitempty"`
}

// InvalidProofError -- special type of error to handle case described in specification
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0.html#section-7.3.2
//
//nolint:recvcheck
type InvalidProofError struct {
	ParentError     *walleterror.Error
	CNonce          string
	CNonceExpiresIn int
}

func NewInvalidProofError(parentError *walleterror.Error, cNonce string, cNonceExpiresIn int) *InvalidProofError {
	return &InvalidProofError{
		ParentError:     parentError,
		CNonce:          cNonce,
		CNonceExpiresIn: cNonceExpiresIn,
	}
}

func (e InvalidProofError) Error() string {
	if e.ParentError == nil {
		return ""
	}

	return e.ParentError.Error()
}

func (e *InvalidProofError) Unwrap() error {
	return e.ParentError
}
