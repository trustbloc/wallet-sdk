/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

// Constants' names and reasons are obvious so they do not require additional comments.
// nolint:golint,nolintlint
const (
	module                                    = "OCI"
	NoClientConfigProvidedError               = "NO_CLIENT_CONFIG_PROVIDED"
	ClientConfigNoUserDidProvidedError        = "CLIENT_CONFIG_NO_USER_DID_PROVIDED"
	ClientConfigNoClientIDProvidedError       = "CLIENT_CONFIG_NO_CLIENT_ID_PROVIDED"
	ClientConfigNoSignerProviderProvidedError = "CLIENT_CONFIG_NO_SIGNER_PROVIDER_PROVIDED"
	ClientConfigNoDIDResolverProvidedError    = "CLIENT_CONFIG_DID_RESOLVER_PROVIDED"
	PreAuthorizedCodeRequiredError            = "PRE_AUTHORIZED_CODE_REQUIRED"
	InvalidIssuanceURIError                   = "INVALID_ISSUANCE_URI"
	UserPINRequiredParseFailedError           = "USER_PIN_REQUIRED_PARSE_FAILED"
	PinCodeRequiredError                      = "PIN_CODE_REQUIRED"
	MetadataFetchFailedError                  = "META_DATA_FETCH_FAILED"
	TokenFetchFailedError                     = "TOKEN_FETCH_FAILED" //nolint:gosec //false positive
	JWTSigningFailedError                     = "JWT_SIGNING_FAILED"
	CredentialFetchFailedError                = "CREDENTIAL_FETCH_FAILED"          //nolint:gosec //false positive
	CredentialParseFailedError                = "CREDENTIAL_PARSE_FAILED"          //nolint:gosec //false positive
	CredentialSchemaResolveFailedError        = "CREDENTIAL_SCHEMA_RESOLVE_FAILED" //nolint:gosec //false positive
)

// Constants' names and reasons are obvious so they do not require additional comments.
// nolint:golint,nolintlint
const (
	NoClientConfigProvidedCode = iota
	ClientConfigNoUserDidProvidedCode
	ClientConfigNoClientIDProvidedCode
	ClientConfigNoSignerProviderProvidedCode
	ClientConfigNoDIDResolverProvidedCode
	PreAuthorizedCodeRequiredCode
	InvalidIssuanceURICode
	UserPINRequiredParseFailedCode
	PinCodeRequiredCode
	MetadataFetchFailedCode
	TokenFetchFailedCode
	JWTSigningFailedCode
	CredentialFetchFailedCode
	CredentialParseFailedCode
	CredentialSchemaResolveFailedCode
)
