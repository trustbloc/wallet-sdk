/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package trustregistry

import (
	"strings"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/openid4vp"
)

// EvaluationResult result of policy evaluation.
type EvaluationResult struct {
	Allowed                   bool
	ErrorCode                 string
	ErrorMessage              string
	MultipleCredentialAllowed bool
	attestationsRequired      []string
	denyReasons               []string
}

// DenyReason check the reasons when the transaction is not allowed (=false).
func (e *EvaluationResult) DenyReason() string {
	return strings.Join(e.denyReasons, ", ")
}

// RequestedAttestationLength returns the number attestation requested.
func (e *EvaluationResult) RequestedAttestationLength() int {
	return len(e.attestationsRequired)
}

// RequestedAttestationAtIndex returns requested attestation by index.
func (e *EvaluationResult) RequestedAttestationAtIndex(index int) string {
	return e.attestationsRequired[index]
}

// CredentialOffer contains data related to a credential type being offered in an issuance request.
type CredentialOffer struct {
	CredentialType             string
	CredentialFormat           string
	ClientAttestationRequested bool
}

// IssuanceRequest  contains data about the issuance request, that is sent to the trust registry API for evaluation.
type IssuanceRequest struct {
	IssuerDID        string
	IssuerDomain     string
	credentialOffers []*CredentialOffer
}

// AddCredentialOffers adds credential offer to evaluate during issuance evaluations.
func (i *IssuanceRequest) AddCredentialOffers(c *CredentialOffer) *IssuanceRequest {
	i.credentialOffers = append(i.credentialOffers, c)

	return i
}

// PresentationRequest contains data about the presentation request,
// that is sent to the trust registry API for evaluation.
type PresentationRequest struct {
	VerifierDID      string
	VerifierDomain   string
	credentialClaims []*CredentialClaimsToCheck
}

// AddCredentialClaims adds credential data to evaluate during presentation evaluations.
func (p *PresentationRequest) AddCredentialClaims(c *CredentialClaimsToCheck) *PresentationRequest {
	p.credentialClaims = append(p.credentialClaims, c)

	return p
}

// CredentialClaimsToCheck contains data about credentials in the presentation request,
// that is sent to the trust registry API for evaluation.
type CredentialClaimsToCheck struct {
	CredentialID        string
	CredentialTypes     *api.StringArray
	IssuerID            string
	IssuanceDate        int64
	ExpirationDate      int64
	CredentialClaimKeys *openid4vp.CredentialClaimKeys
}

// LegacyNewCredentialClaimsToCheck create new CredentialClaimsToCheck object.
func LegacyNewCredentialClaimsToCheck(
	credentialID string, credentialTypes *api.StringArray, issuerID string,
	issuanceDate int64, expirationDate int64,
) *CredentialClaimsToCheck {
	return &CredentialClaimsToCheck{
		CredentialID:    credentialID,
		CredentialTypes: credentialTypes,
		IssuerID:        issuerID,
		IssuanceDate:    issuanceDate,
		ExpirationDate:  expirationDate,
	}
}
