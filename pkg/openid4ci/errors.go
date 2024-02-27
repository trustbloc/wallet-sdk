/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

// Constants' names and reasons are obvious, so they do not require additional comments.
// nolint:golint,nolintlint
const (
	ErrorModule                               = "OCI"
	InvalidIssuanceURIError                   = "INVALID_ISSUANCE_URI"
	InvalidCredentialOfferError               = "INVALID_CREDENTIAL_OFFER"            //nolint:gosec //false positive
	InvalidCredentialConfigurationIDError     = "INVALID_CREDENTIAL_CONFIGURATION_ID" //nolint:gosec //false positive
	UnsupportedCredentialTypeInOfferError     = "UNSUPPORTED_CREDENTIAL_TYPE_IN_OFFER"
	MetadataFetchFailedError                  = "METADATA_FETCH_FAILED"
	JWTSigningFailedError                     = "JWT_SIGNING_FAILED"
	KeyIDMissingDIDPartError                  = "KEY_ID_MISSING_DID_PART"
	CredentialParseError                      = "CREDENTIAL_PARSE_FAILED"                     //nolint:gosec,lll //false positive, can't shorten
	StateInRedirectURINotMatchingAuthURLError = "STATE_IN_REDIRECT_URI_NOT_MATCHING_AUTH_URL" //nolint:gosec,lll //false positive, can't shorten
	InvalidTokenRequestError                  = "INVALID_TOKEN_REQUEST"                       //nolint:gosec,lll //false positive, can't shorten
	InvalidGrantError                         = "INVALID_GRANT"
	InvalidClientError                        = "INVALID_CLIENT"
	OtherTokenRequestError                    = "OTHER_TOKEN_REQUEST_ERROR"      //nolint:gosec,lll //false positive, can't shorten
	OtherCredentialRequestError               = "OTHER_CREDENTIAL_REQUEST_ERROR" //nolint:gosec,lll //false positive, can't shorten
	InvalidCredentialRequestError             = "INVALID_CREDENTIAL_REQUEST"     //nolint:gosec,lll //false positive, can't shorten
	InvalidTokenError                         = "INVALID_TOKEN"
	UnsupportedCredentialFormatError          = "UNSUPPORTED_CREDENTIAL_FORMAT"
	UnsupportedCredentialTypeError            = "UNSUPPORTED_CREDENTIAL_TYPE"
	InvalidOrMissingProofError                = "INVALID_OR_MISSING_PROOF"
	UnsupportedIssuanceURISchemeError         = "UNSUPPORTED_ISSUANCE_URI_SCHEME"
	NoTokenEndpointAvailableError             = "NO_TOKEN_ENDPOINT_AVAILABLE" //nolint:gosec //false positive
	AcknowledgmentExpiredError                = "ACKNOWLEDGMENT_EXPIRED"
)

// Constants' names and reasons are obvious, so they do not require additional comments.
// nolint:golint,nolintlint
const (
	InvalidIssuanceURICode                   = 0
	InvalidCredentialOfferCode               = 1
	UnsupportedCredentialTypeInOfferCode     = 2
	IssuerOpenIDConfigFetchFailedCode        = 3
	MetadataFetchFailedCode                  = 4
	JWTSigningFailedCode                     = 5
	KeyIDMissingDIDPartCode                  = 6
	CredentialParseFailedCode                = 7
	StateInRedirectURINotMatchingAuthURLCode = 8
	InvalidTokenRequestErrorCode             = 9
	InvalidGrantErrorCode                    = 10
	InvalidClientErrorCode                   = 11
	OtherTokenResponseErrorCode              = 12
	OtherCredentialRequestErrorCode          = 13
	InvalidCredentialRequestErrorCode        = 14
	InvalidTokenErrorCode                    = 15
	UnsupportedCredentialFormatErrorCode     = 16
	UnsupportedCredentialTypeErrorCode       = 17
	InvalidOrMissingProofErrorCode           = 18
	UnsupportedIssuanceURISchemeCode         = 19
	NoTokenEndpointAvailableErrorCode        = 20
	AcknowledgmentExpiredErrorCode           = 21
	InvalidCredentialConfigurationIDCode     = 22
)
