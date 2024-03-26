/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package issuer contains models for representing an issuer's metadata.
package issuer

import (
	"errors"
)

// CredentialConfigurationID is an alias for type "string" introduced to simplify understanding of the specification.
type CredentialConfigurationID = string

// Metadata represents metadata about an issuer as obtained from their .well-known OpenID configuration.
type Metadata struct {
	jwtKID *string

	// URL of the OP's OAuth 2.0 Authorization Endpoint.
	AuthorizationServer string `json:"authorization_endpoint,omitempty"`

	// URL of the Credential Issuer's Batch Credential Endpoint. This URL MUST use the https scheme and MAY contain
	// port, path and query parameter components.
	// If omitted, the Credential Issuer does not support the Batch Credential Endpoint.
	BatchCredentialEndpoint string `json:"batch_credential_endpoint,omitempty"`

	// An object that describes specifics of the Credential that the Credential Issuer supports issuance of.
	// This object contains a list of name/value pairs, where each name is a unique identifier
	// of the supported credential being described.
	CredentialConfigurationsSupported map[CredentialConfigurationID]*CredentialConfigurationSupported `json:"credential_configurations_supported,omitempty"` //nolint:lll

	// URL of the Credential Issuer's Credential Endpoint. This URL MUST use the https scheme and MAY contain
	// port, path and query parameter components.
	CredentialEndpoint string `json:"credential_endpoint,omitempty"`

	// Boolean value specifying whether the Credential Issuer supports returning credential_identifiers parameter
	// in the authorization_details Token Response parameter, with true indicating support.
	// If omitted, the default value is false.
	CredentialIdentifiersSupported *bool `json:"credential_identifiers_supported,omitempty"`

	// The Credential Issuer's identifier.
	CredentialIssuer string `json:"credential_issuer,omitempty"`

	// Object containing information about whether the Credential Issuer supports encryption of the Credential
	// and Batch Credential Response on top of TLS
	CredentialResponseEncryption *CredentialResponseEncryptionSupported `json:"credential_response_encryption,omitempty"`

	// URL of the Credential Issuer's Deferred Credential Endpoint. This URL MUST use the https scheme and MAY contain
	// port, path, and query parameter components.
	// If omitted, the Credential Issuer does not support the Deferred Credential Endpoint.
	DeferredCredentialEndpoint string `json:"deferred_credential_endpoint,omitempty"`

	// An array of objects, where each object contains display properties of a Credential Issuer for a certain language.
	LocalizedIssuerDisplays []LocalizedIssuerDisplay `json:"display,omitempty"`

	// JSON array containing a list of the OAuth 2.0 Grant Type values that this OP supports.
	GrantTypesSupported []string `json:"grant_types_supported,omitempty"`

	// URL of the Credential Issuer's Notification Endpoint. This URL MUST use the https scheme and MAY contain
	// port, path, and query parameter components.
	// If omitted, the Credential Issuer does not support the Notification Endpoint.
	NotificationEndpoint string `json:"notification_endpoint,omitempty"`

	// Boolean indicating whether the issuer accepts a Token Request with a Pre-Authorized Code but without a client id.
	// The default is false.
	PreAuthorizedGrantAnonymousAccessSupported *bool `json:"pre-authorized_grant_anonymous_access_supported,omitempty"`

	// URL of the OP's Dynamic Client Registration Endpoint.
	RegistrationEndpoint *string `json:"registration_endpoint,omitempty"`

	// JSON array containing a list of the OAuth 2.0 response_type values that this OP supports.
	ResponseTypesSupported []string `json:"response_types_supported,omitempty"`

	// JSON array containing a list of the OAuth 2.0 [RFC6749] scope values that this server supports.
	ScopesSupported []string `json:"scopes_supported,omitempty"`

	// String that is a signed JWT. This JWT contains Credential Issuer metadata parameters as claims.
	SignedMetadata string `json:"signed_metadata,omitempty"`

	// URL of the OP's OAuth 2.0 Token Endpoint.
	TokenEndpoint string `json:"token_endpoint,omitempty"`

	// JSON array containing a list of client authentication methods supported by this token endpoint. Default is "none".
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported,omitempty"`
}

// CredentialConfigurationSupported describes specifics of the Credential that
// the Credential Issuer supports issuance of.
type CredentialConfigurationSupported struct {
	// For mso_mdoc and vc+sd-jwt vc only. Object containing a list of name/value pairs,
	// where each name identifies a claim about the subject offered in the Credential.
	// The value can be another such object (nested data structures), or an array of such objects.
	Claims *map[string]interface{} `json:"claims,omitempty"`

	// Object containing the detailed description of the credential type.
	CredentialDefinition *CredentialDefinition `json:"credential_definition,omitempty"`

	// Array of case-sensitive strings that identify how the Credential is bound to the identifier of the End-User
	// who possesses the Credential.
	CryptographicBindingMethodsSupported []string `json:"cryptographic_binding_methods_supported,omitempty"`

	// Array of case-sensitive strings that identify the algorithms that the Issuer uses to sign the issued Credential.
	CredentialSigningAlgValuesSupported []string `json:"credential_signing_alg_values_supported,omitempty"`

	// An array of objects, where each object contains the display properties of the
	// supported credential for a certain language.
	LocalizedCredentialDisplays []LocalizedCredentialDisplay `json:"display,omitempty"`

	// For mso_mdoc vc only. String identifying the Credential type, as defined in [ISO.18013-5].
	Doctype string `json:"doctype,omitempty"`

	// A JSON string identifying the format of this credential, i.e., jwt_vc_json or ldp_vc.
	// Depending on the format value, the object contains further elements defining the type and (optionally)
	// particular claims the credential MAY contain and information about how to display the credential.
	Format string `json:"format"`

	// Array of the claim name values that lists them in the order they should be displayed by the Wallet.
	Order []string `json:"order,omitempty"`

	// Object that describes specifics of the key proof(s) that the Credential Issuer supports.
	ProofTypesSupported map[string]ProofTypeSupported `json:"proof_types_supported,omitempty"`

	// A JSON string identifying the scope value that this Credential Issuer supports for this particular credential.
	Scope string `json:"scope,omitempty"`

	// For vc+sd-jwt vc only. String designating the type of Credential,
	// as defined in https://datatracker.ietf.org/doc/html/draft-ietf-oauth-sd-jwt-vc-01
	Vct string `json:"vct,omitempty"`
}

// ClaimOrderAsInt returns this Claim's order value as an integer.
func (c *CredentialConfigurationSupported) ClaimOrderAsInt(claimName string) (int, error) {
	for i, cn := range c.Order {
		if cn == claimName {
			return i, nil
		}
	}

	return -1, errors.New("order is not specified")
}

// CredentialDefinition containing the detailed description of the credential type.
type CredentialDefinition struct {
	// For ldp_vc only. Array as defined in https://www.w3.org/TR/vc-data-model/#contexts.
	Context []string `json:"@context,omitempty"`

	// An object containing a list of name/value pairs, where each name identifies a claim offered in the Credential.
	// The value can be another such object (nested data structures), or an array of such objects.
	CredentialSubject map[string]*Claim `json:"credentialSubject,omitempty"`

	// Array designating the types a certain credential type supports.
	Type []string `json:"type"`
}

// CredentialResponseEncryptionSupported containing information about whether the Credential Issuer
// supports encryption of the Credential and Batch Credential Response on top of TLS.
type CredentialResponseEncryptionSupported struct {
	// Array containing a list of the JWE [RFC7516] encryption algorithms (alg values) [RFC7518] supported by the
	// Credential and Batch Credential Endpoint to encode the Credential or Batch Credential Response in a JWT [RFC7519].
	AlgValuesSupported []string `json:"alg_values_supported"`

	// Array containing a list of the JWE [RFC7516] encryption algorithms (enc values) [RFC7518]
	// supported by the Credential and Batch Credential Endpoint to encode the Credential or
	// Batch Credential Response in a JWT [RFC7519].
	EncValuesSupported []string `json:"enc_values_supported"`

	// Boolean value specifying whether the Credential Issuer requires the additional encryption on top of TLS
	// for the Credential Response. If the value is true, the Credential Issuer requires encryption for
	// every Credential Response and therefore the Wallet MUST provide encryption keys in the Credential Request.
	// If the value is false, the Wallet MAY choose whether it provides encryption keys or not.
	EncryptionRequired bool `json:"encryption_required"`
}

// ProofTypeSupported contains metadata about the proof type that the Credential Issuer supports.
type ProofTypeSupported struct {
	// Array of case-sensitive strings that identify the algorithms that the Issuer supports for this proof type.
	ProofSigningAlgValuesSupported []string `json:"proof_signing_alg_values_supported"`
}

// GetJWTKID returns the jwtKID field. This is exposed via this method instead of with an exported field because
// the linter expects all exported fields to have JSON tags, but the jwtKID field is only intended for use internally
// within Wallet-SDK.
func (m *Metadata) GetJWTKID() *string {
	return m.jwtKID
}

// SetJWTKID sets the jwtKID field.
func (m *Metadata) SetJWTKID(jwtKID string) {
	m.jwtKID = &jwtKID
}

// LocalizedCredentialDisplay represents display data for a credential as a whole for a certain locale.
// Display data for specific claims (e.g. first name, date of birth, etc.) are in SupportedCredential.CredentialSubject
// (in the parent object above).
type LocalizedCredentialDisplay struct {
	Name            string `json:"name,omitempty"`
	Locale          string `json:"locale,omitempty"`
	Logo            *Logo  `json:"logo,omitempty"`
	BackgroundColor string `json:"background_color,omitempty"`
	TextColor       string `json:"text_color,omitempty"`
}

// Claim represents display data for a specific claim in (potentially) multiple locales.
// Each ClaimDisplay represents display data for a single locale.
type Claim struct {
	LocalizedClaimDisplays []LocalizedClaimDisplay `json:"display,omitempty"`
	ValueType              string                  `json:"value_type,omitempty"`
	Pattern                string                  `json:"pattern,omitempty"`
	Mask                   string                  `json:"mask,omitempty"`
}

// Logo represents display information for a logo.
type Logo struct {
	URL     string `json:"uri,omitempty"`
	AltText string `json:"alt_text,omitempty"`
}

// LocalizedIssuerDisplay represents display information for an issuer in a specific locale.
type LocalizedIssuerDisplay struct {
	Name            string `json:"name,omitempty"`
	Locale          string `json:"locale,omitempty"`
	URL             string `json:"url,omitempty"`
	Logo            *Logo  `json:"logo,omitempty"`
	BackgroundColor string `json:"background_color,omitempty"`
	TextColor       string `json:"text_color,omitempty"`
}

// LocalizedClaimDisplay represents display information for a claim in a specific locale.
type LocalizedClaimDisplay struct {
	Name   string `json:"name,omitempty"`
	Locale string `json:"locale,omitempty"`
}
