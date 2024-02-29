/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package trustregistry implements trust registry API.
package trustregistry

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/trustbloc/wallet-sdk/pkg/trustregistry"
)

// RegistryConfig is config for trust registry API.
type RegistryConfig struct {
	EvaluateIssuanceURL        string
	EvaluatePresentationURL    string
	DisableHTTPClientTLSVerify bool
}

// Registry implements API for trust registry.
type Registry struct {
	impl *trustregistry.Registry
}

// NewRegistry creates new trust registry API.
func NewRegistry(config *RegistryConfig) *Registry {
	var httpClient *http.Client
	if config.DisableHTTPClientTLSVerify {
		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					//nolint:gosec // The ability to disable TLS is an option we provide that
					// has to be explicitly set by the user. By default, we don't disable TLS.
					// This option is only intended for testing purposes.
					InsecureSkipVerify: true,
				},
			},
		}
	}

	return &Registry{
		impl: trustregistry.New(&trustregistry.RegistryConfig{
			EvaluateIssuanceURL:     config.EvaluateIssuanceURL,
			EvaluatePresentationURL: config.EvaluatePresentationURL,
			HTTPClient:              httpClient,
		}),
	}
}

// EvaluateIssuance evaluate is issuance request by calling trust registry.
func (r *Registry) EvaluateIssuance(request *IssuanceRequest) (*EvaluationResult, error) {
	result, err := r.impl.EvaluateIssuance(&trustregistry.IssuanceRequest{
		IssuerDID:        request.IssuerDID,
		IssuerDomain:     request.IssuerDomain,
		CredentialOffers: toCredentialOffersRequest(request.credentialOffers),
	})
	if err != nil {
		return nil, err
	}

	return &EvaluationResult{
		Allowed:      result.Allowed,
		ErrorCode:    result.ErrorCode,
		ErrorMessage: result.ErrorMessage,
	}, nil
}

// EvaluatePresentation evaluate is presentation request by calling trust registry.
func (r *Registry) EvaluatePresentation(request *PresentationRequest) (*EvaluationResult, error) {
	var credentialClaims []trustregistry.CredentialClaimsToCheck

	for _, claims := range request.credentialClaims {
		credentialClaims = append(credentialClaims, trustregistry.CredentialClaimsToCheck{
			CredentialID:    claims.CredentialID,
			CredentialTypes: claims.CredentialTypes.Strings,
			IssuerID:        claims.IssuerID,
			IssuanceDate:    time.Unix(claims.IssuanceDate, 0),
			ExpirationDate:  time.Unix(claims.ExpirationDate, 0),
		})
	}

	result, err := r.impl.EvaluatePresentation(&trustregistry.PresentationRequest{
		VerifierDid:      request.VerifierDID,
		VerifierDomain:   request.VerifierDomain,
		CredentialClaims: credentialClaims,
	})
	if err != nil {
		return nil, err
	}

	return &EvaluationResult{
		Allowed:      result.Allowed,
		ErrorCode:    result.ErrorCode,
		ErrorMessage: result.ErrorMessage,
	}, nil
}

func toCredentialOffersRequest(offers []*CredentialOffer) []trustregistry.CredentialOffer {
	req := make([]trustregistry.CredentialOffer, len(offers))

	for i, offer := range offers {
		req[i] = trustregistry.CredentialOffer{
			CredentialType:             offer.CredentialType,
			CredentialFormat:           offer.CredentialFormat,
			ClientAttestationRequested: offer.ClientAttestationRequested,
		}
	}

	return req
}
