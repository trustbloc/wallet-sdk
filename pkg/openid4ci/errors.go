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
	ClientConfigNoDIDResolverProvidedError    = "CLIENT_CONFIG_NO_DID_RESOLVER_PROVIDED"
	InvalidIssuanceURIError                   = "INVALID_ISSUANCE_URI"
	InvalidCredentialOfferError               = "INVALID_CREDENTIAL_OFFER" //nolint:gosec //false positive
	UnsupportedCredentialTypeInOfferError     = "UNSUPPORTED_CREDENTIAL_TYPE_IN_OFFER"
	PINRequiredError                          = "PIN_REQUIRED"
	IssuerOpenIDConfigFetchFailedError        = "ISSUER_OPENID_FETCH_FAILED"
	MetadataFetchFailedError                  = "METADATA_FETCH_FAILED"
	TokenFetchFailedError                     = "TOKEN_FETCH_FAILED" //nolint:gosec //false positive
	JWTSigningFailedError                     = "JWT_SIGNING_FAILED"
	CredentialFetchFailedError                = "CREDENTIAL_FETCH_FAILED" //nolint:gosec //false positive
	KeyIDNotContainDIDPartError               = "KEY_ID_NOT_CONTAIN_DID_PART"
	CredentialParseError                      = "CREDENTIAL_PARSE_FAILED"                     //nolint:gosec,lll //false positive, can't shorten
	StateInRedirectURINotMatchingAuthURLError = "STATE_IN_REDIRECT_URI_NOT_MATCHING_AUTH_URL" //nolint:gosec,lll //false positive, can't shorten
)

// Constants' names and reasons are obvious so they do not require additional comments.
// nolint:golint,nolintlint
const (
	NoClientConfigProvidedCode = iota
	ClientConfigNoDIDResolverProvidedCode
	InvalidIssuanceURICode
	InvalidCredentialOfferCode
	UnsupportedCredentialTypeInOfferCode
	PINRequiredCode
	IssuerOpenIDConfigFetchFailedCode
	MetadataFetchFailedCode
	TokenFetchFailedCode
	JWTSigningFailedCode
	CredentialFetchFailedCode
	KeyIDNotContainDIDPartCode
	CredentialParseFailedCode
	StateInRedirectURINotMatchingAuthURLCode
)
