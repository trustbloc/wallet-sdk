/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package did contains did related functionality.
package did

// Errors for creator modules.
const (
	Module                         = "DID"
	CreateDIDKeyFailedError        = "CREATE_DID_KEY_FAILED"
	CreateDIDIONFailedError        = "CREATE_DID_ION_FAILED"
	CreateDIDJWKFailedError        = "CREATE_DID_JWK_FAILED"
	UnsupportedDIDMethodError      = "UNSUPPORTED_DID_METHOD"
	ResolutionFailedError          = "DID_RESOLUTION_FAILED"
	ResolverInitializationFailed   = "DID_RESOLVER_INITIALIZATION_FAILED"
	WellknownInitializationFailed  = "WELLKNOWN_INITIALIZATION_FAILED"
	DomainAndDidVerificationFailed = "DOMAIN_AND_DID_VERIFICATION_FAILED"
)

// Category codes for creator module.
const (
	CreateDIDKeyFailedCode = iota
	CreateDIDIONFailedCode
	CreateDIDJWKFailedCode
	UnsupportedDIDMethodCode
	ResolutionFailedCode
	ResolverInitializationCode
	WellknownInitializationCode
	DomainAndDidVerificationCode
)
