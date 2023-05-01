/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import "errors"

// IssuerCapabilities represents an issuer's self-reported capabilities.
type IssuerCapabilities struct {
	preAuthorizedCodeGrantParams *PreAuthorizedCodeGrantParams
	authorizationCodeGrantParams *authorizationCodeGrantParams
}

// PreAuthorizedCodeGrantTypeSupported indicates whether an issuer supports the pre-authorized code grant type.
func (i *IssuerCapabilities) PreAuthorizedCodeGrantTypeSupported() bool {
	return i.preAuthorizedCodeGrantParams != nil
}

// PreAuthorizedCodeGrantParams returns an object that can be used to determine an issuer's pre-authorized code grant
// parameters. The caller should call the PreAuthorizedCodeGrantTypeSupported method first and only call this method to
// get the params if PreAuthorizedCodeGrantTypeSupported returns true. This method only returns an error if
// PreAuthorizedCodeGrantTypeSupported returns false.
func (i *IssuerCapabilities) PreAuthorizedCodeGrantParams() (*PreAuthorizedCodeGrantParams, error) {
	if i.preAuthorizedCodeGrantParams == nil {
		return nil, errors.New("issuer does not support the pre-authorized code grant")
	}

	return i.preAuthorizedCodeGrantParams, nil
}

// AuthorizationCodeGrantTypeSupported indicates whether an issuer supports the authorization code grant type.
func (i *IssuerCapabilities) AuthorizationCodeGrantTypeSupported() bool {
	return i.authorizationCodeGrantParams != nil
}

func (i *IssuerCapabilities) onlyPreAuthorizedCodeGrantSupported() bool {
	return i.PreAuthorizedCodeGrantTypeSupported() && !i.AuthorizationCodeGrantTypeSupported()
}

func (i *IssuerCapabilities) onlyAuthorizationCodeGrantSupported() bool {
	return i.AuthorizationCodeGrantTypeSupported() && !i.PreAuthorizedCodeGrantTypeSupported()
}

func determineIssuerCapabilities(credentialOffer *CredentialOffer) (*IssuerCapabilities, error) {
	rawPreAuthorizedCodeGrantParams, preAuthorizedCodeGrantExists := credentialOffer.Grants[preAuthorizedGrantType]
	rawAuthorizationCodeGrantParams, authorizationCodeGrantExists := credentialOffer.Grants[authorizationCodeGrantType]

	if !preAuthorizedCodeGrantExists && !authorizationCodeGrantExists {
		return nil, errors.New("no supported grant types found")
	}

	issuerCapabilities := &IssuerCapabilities{}

	if preAuthorizedCodeGrantExists {
		var err error

		issuerCapabilities.preAuthorizedCodeGrantParams, err = processPreAuthorizedCodeGrantParams(
			rawPreAuthorizedCodeGrantParams)
		if err != nil {
			return nil, err
		}
	}

	if authorizationCodeGrantExists {
		var err error

		issuerCapabilities.authorizationCodeGrantParams, err = processAuthorizationCodeGrantParams(
			rawAuthorizationCodeGrantParams)
		if err != nil {
			return nil, err
		}
	}

	return issuerCapabilities, nil
}
