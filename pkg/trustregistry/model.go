/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package trustregistry

import "time"

// EvaluationResult result of policy evaluation.
type EvaluationResult struct {
	Allowed      bool            `json:"allowed,omitempty"`
	ErrorCode    string          `json:"errorCode,omitempty"`
	ErrorMessage string          `json:"errorMessage,omitempty"`
	DenyReasons  []string        `json:"deny_reasons,omitempty"`
	Data         *EvaluationData `json:"payload,omitempty"`
}

// EvaluationData data from policy evaluation.
type EvaluationData struct {
	AttestationsRequired      []string `json:"attestations_required,omitempty"`
	MultipleCredentialAllowed bool     `json:"multiple_credentials_allowed,omitempty"`
}

// CredentialOffer contains data related to a credential type being offered in an issuance request.
type CredentialOffer struct {
	CredentialType             string `json:"credential_type"`
	CredentialFormat           string `json:"credential_format"`
	ClientAttestationRequested bool   `json:"client_attestation_requested"`
}

// IssuanceRequest  contains data about the issuance request, that is sent to the trust registry API for evaluation.
type IssuanceRequest struct {
	IssuerDID        string            `json:"issuer_did"`
	IssuerDomain     string            `json:"issuer_domain"`
	CredentialOffers []CredentialOffer `json:"credential_offers"`
}

// PresentationRequest contains data about the presentation request,
// that is sent to the trust registry API for evaluation.
type PresentationRequest struct {
	VerifierDid      string                    `json:"verifier_did"`
	VerifierDomain   string                    `json:"verifier_domain"`
	CredentialClaims []CredentialClaimsToCheck `json:"credential_metadata"`
}

// CredentialClaimsToCheck  contains data about credentials in the presentation request,
// that is sent to the trust registry API for evaluation.
type CredentialClaimsToCheck struct {
	CredentialID        string      `json:"credential_id"`
	CredentialTypes     []string    `json:"credential_types"`
	IssuerID            string      `json:"issuer_id"`
	IssuanceDate        time.Time   `json:"issuance_date"`
	ExpirationDate      time.Time   `json:"expiration_date"`
	CredentialClaimKeys interface{} `json:"credential_claim_keys"`
}
