/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import "errors"

// PreAuthorizedCodeGrantParams represents an issuer's pre-authorized code grant parameters.
type PreAuthorizedCodeGrantParams struct {
	preAuthorizedCode string
	txCode            *TxCode
}

// TxCode is a code intended to bind the pre-authorized code to a certain transaction to prevent replay attack.
type TxCode struct {
	inputMode   string
	length      int
	description string
}

// PINRequired indicates whether the issuer requires a PIN.
func (p *PreAuthorizedCodeGrantParams) PINRequired() bool {
	return p.txCode != nil
}

// AuthorizationCodeGrantParams represents an issuer's authorization code grant parameters.
type AuthorizationCodeGrantParams struct {
	IssuerState *string
}

func determineIssuerGrantCapabilities(
	credentialOffer *CredentialOffer,
) (*PreAuthorizedCodeGrantParams, *AuthorizationCodeGrantParams, error) {
	rawPreAuthorizedCodeGrantParams, preAuthorizedCodeGrantExists := credentialOffer.Grants[preAuthorizedGrantType]
	rawAuthorizationCodeGrantParams, authorizationCodeGrantExists := credentialOffer.Grants[authorizationCodeGrantType]

	if !preAuthorizedCodeGrantExists && !authorizationCodeGrantExists {
		return nil, nil, errors.New("no supported grant types found")
	}

	var preAuthorizedCodeGrantParams *PreAuthorizedCodeGrantParams

	var authorizationCodeGrantParams *AuthorizationCodeGrantParams

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

	var txCode *TxCode

	txCodeUntyped, exists := rawParams["tx_code"]
	if exists {
		var m map[string]interface{}

		if m, ok = txCodeUntyped.(map[string]interface{}); !ok {
			return nil, errors.New("tx_code is not a valid json object")
		}

		txCode = &TxCode{}

		var (
			inputMode, description string
			length                 float64
		)

		if inputMode, ok = m["input_mode"].(string); ok {
			txCode.inputMode = inputMode
		}

		if length, ok = m["length"].(float64); ok {
			txCode.length = int(length)
		}

		if description, ok = m["description"].(string); ok {
			txCode.description = description
		}
	}

	return &PreAuthorizedCodeGrantParams{preAuthorizedCode: preAuthorizedCode, txCode: txCode}, nil
}

func processAuthorizationCodeGrantParams(rawParams map[string]interface{}) (*AuthorizationCodeGrantParams, error) {
	var issuerState *string

	issuerStateUntyped, exists := rawParams["issuer_state"]
	if exists {
		issuerStateAsString, ok := issuerStateUntyped.(string)
		if !ok {
			return nil, errors.New("user-pin-required field value is not a bool")
		}

		issuerState = &issuerStateAsString
	}

	return &AuthorizationCodeGrantParams{IssuerState: issuerState}, nil
}
