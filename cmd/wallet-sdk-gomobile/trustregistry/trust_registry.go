/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package trustregistry implements trust registry API.
package trustregistry

import (
	"time"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	"github.com/trustbloc/wallet-sdk/pkg/trustregistry"
)

// RegistryConfig is config for trust registry API.
type RegistryConfig struct {
	EvaluateIssuanceURL        string
	EvaluatePresentationURL    string
	DisableHTTPClientTLSVerify bool
	additionalHeaders          api.Headers
}

// AddHeader adds the given HTTP header to all REST calls made to the trust registry during evaluation flow.
func (r *RegistryConfig) AddHeader(header *api.Header) {
	r.additionalHeaders.Add(header)
}

// Registry implements API for trust registry.
type Registry struct {
	impl *trustregistry.Registry
}

// NewRegistry creates new trust registry API.
func NewRegistry(config *RegistryConfig) *Registry {
	httpClient := wrapper.NewHTTPClient(nil, config.additionalHeaders, config.DisableHTTPClientTLSVerify)

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

	var attestationsRequired []string
	if result.Data != nil {
		attestationsRequired = result.Data.AttestationsRequired
	}

	return &EvaluationResult{
		Allowed:              result.Allowed,
		ErrorCode:            result.ErrorCode,
		ErrorMessage:         result.ErrorMessage,
		attestationsRequired: attestationsRequired,
		denyReasons:          result.DenyReasons,
	}, nil
}

// EvaluatePresentation evaluate is presentation request by calling trust registry.
func (r *Registry) EvaluatePresentation(request *PresentationRequest) (*EvaluationResult, error) {
	var credentialClaims []trustregistry.CredentialClaimsToCheck

	for _, claims := range request.credentialClaims {
		var contentJSON interface{}
		if claims.CredentialClaimKeys != nil {
			contentJSON = claims.CredentialClaimKeys.ContentJSON
		}

		credentialClaims = append(credentialClaims, trustregistry.CredentialClaimsToCheck{
			CredentialID:        claims.CredentialID,
			CredentialTypes:     claims.CredentialTypes.Strings,
			IssuerID:            claims.IssuerID,
			IssuanceDate:        time.Unix(claims.IssuanceDate, 0),
			ExpirationDate:      time.Unix(claims.ExpirationDate, 0),
			CredentialClaimKeys: contentJSON,
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

	var attestationsRequired []string

	var multipleCredentialAllowed bool

	if result.Data != nil {
		attestationsRequired = result.Data.AttestationsRequired
		multipleCredentialAllowed = result.Data.MultipleCredentialAllowed
	}

	return &EvaluationResult{
		Allowed:                   result.Allowed,
		ErrorCode:                 result.ErrorCode,
		ErrorMessage:              result.ErrorMessage,
		attestationsRequired:      attestationsRequired,
		MultipleCredentialAllowed: multipleCredentialAllowed,
		denyReasons:               result.DenyReasons,
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
