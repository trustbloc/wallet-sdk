/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package attestation

import attestationgoapi "github.com/trustbloc/wallet-sdk/pkg/attestation"

// AttestRequest is a request for attestation.
type AttestRequest struct {
	wrappedRequest attestationgoapi.AttestWalletInitRequest
}

// NewAttestRequest creates a new attestation request.
func NewAttestRequest() *AttestRequest {
	return &AttestRequest{}
}

// AddAssertion adds assertion to the request.
func (a *AttestRequest) AddAssertion(assertion string) *AttestRequest {
	a.wrappedRequest.Assertions = append(a.wrappedRequest.Assertions, assertion)

	return a
}

// AddClientAssertionType adds client assertion type to the request.
func (a *AttestRequest) AddClientAssertionType(clientAssertionType string) *AttestRequest {
	a.wrappedRequest.ClientAssertionType = append(a.wrappedRequest.ClientAssertionType, clientAssertionType)

	return a
}

// AddWalletAuthentication adds wallet authentication.
func (a *AttestRequest) AddWalletAuthentication(key, value string) *AttestRequest {
	if a.wrappedRequest.WalletAuthentication == nil {
		a.wrappedRequest.WalletAuthentication = make(map[string]interface{})
	}

	a.wrappedRequest.WalletAuthentication[key] = value

	return a
}

// AddWalletMetadata adds wallet metadata to the request.
func (a *AttestRequest) AddWalletMetadata(key, value string) *AttestRequest {
	if a.wrappedRequest.WalletMetadata == nil {
		a.wrappedRequest.WalletMetadata = make(map[string]interface{})
	}

	a.wrappedRequest.WalletMetadata[key] = value

	return a
}
