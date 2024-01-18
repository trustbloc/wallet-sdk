/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package trustregistry

import "time"

// EvaluationResult result of policy evaluation.
type EvaluationResult struct {
	Allowed      bool
	ErrorCode    string
	ErrorMessage string
}

// IssuanceRequest  contains data about the issuance request, that is sent to the trust registry API for evaluation.
type IssuanceRequest struct {
	IssuerDID                  string
	IssuerDomain               string
	CredentialType             string
	CredentialFormat           string
	ClientAttestationRequested bool
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
	CredentialID    string
	credentialTypes []string
	IssuerID        string
	IssuanceDate    time.Time
	ExpirationDate  time.Time
}

// AddType adds credential type.
func (c *CredentialClaimsToCheck) AddType(t string) *CredentialClaimsToCheck {
	c.credentialTypes = append(c.credentialTypes, t)

	return c
}
