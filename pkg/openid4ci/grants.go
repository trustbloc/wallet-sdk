/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import "errors"

// PreAuthorizedCodeGrantParams represents an issuer's pre-authorized code grant parameters.
type PreAuthorizedCodeGrantParams struct {
	preAuthorizedCode string
	userPINRequired   bool
}

// PINRequired indicates whether the issuer requires a PIN.
func (p *PreAuthorizedCodeGrantParams) PINRequired() bool {
	return p.userPINRequired
}

type authorizationCodeGrantParams struct {
	issuerState *string
}

func determineIssuerGrantCapabilities(
	credentialOffer *CredentialOffer,
) (*PreAuthorizedCodeGrantParams, *authorizationCodeGrantParams, error) {
	rawPreAuthorizedCodeGrantParams, preAuthorizedCodeGrantExists := credentialOffer.Grants[preAuthorizedGrantType]
	rawAuthorizationCodeGrantParams, authorizationCodeGrantExists := credentialOffer.Grants[authorizationCodeGrantType]

	if !preAuthorizedCodeGrantExists && !authorizationCodeGrantExists {
		return nil, nil, errors.New("no supported grant types found")
	}

	var preAuthorizedCodeGrantParams *PreAuthorizedCodeGrantParams

	var authorizationCodeGrantParams *authorizationCodeGrantParams

	var err error
	if preAuthorizedCodeGrantExists {
		preAuthorizedCodeGrantParams, err = processPreAuthorizedCodeGrantParams(
			rawPreAuthorizedCodeGrantParams)
		if err != nil {
			return nil, nil, err
		}
	}

	if authorizationCodeGrantExists {
		authorizationCodeGrantParams, err = processAuthorizationCodeGrantParams(
			rawAuthorizationCodeGrantParams)
		if err != nil {
			return nil, nil, err
		}
	}

	return preAuthorizedCodeGrantParams, authorizationCodeGrantParams, nil
}

func processPreAuthorizedCodeGrantParams(rawParams map[string]interface{}) (*PreAuthorizedCodeGrantParams, error) {
	preAuthorizedCodeUntyped, exists := rawParams["pre-authorized_code"]
	if !exists {
		return nil, errors.New("pre-authorized_code field value is missing")
	}

	preAuthorizedCode, ok := preAuthorizedCodeUntyped.(string)
	if !ok {
		return nil, errors.New("pre-authorized_code field value is not a bool")
	}

	var userPINRequired bool

	userPINRequiredUntyped, exists := rawParams["user_pin_required"]
	if exists { // userPINRequired is supposed to default to false if user_pin_required isn't specified.
		var ok bool

		userPINRequired, ok = userPINRequiredUntyped.(bool)
		if !ok {
			return nil, errors.New("user-pin-required field value is not a bool")
		}
	}

	return &PreAuthorizedCodeGrantParams{preAuthorizedCode: preAuthorizedCode, userPINRequired: userPINRequired}, nil
}

func processAuthorizationCodeGrantParams(rawParams map[string]interface{}) (*authorizationCodeGrantParams, error) {
	var issuerState *string

	issuerStateUntyped, exists := rawParams["issuer_state"]
	if exists {
		issuerStateAsString, ok := issuerStateUntyped.(string)
		if !ok {
			return nil, errors.New("user-pin-required field value is not a bool")
		}

		issuerState = &issuerStateAsString
	}

	return &authorizationCodeGrantParams{issuerState: issuerState}, nil
}
