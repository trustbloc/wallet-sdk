package openid4ci

import openid4cigoapi "github.com/trustbloc/wallet-sdk/pkg/openid4ci"

// IssuerCapabilities represents an issuer's self-reported capabilities.
type IssuerCapabilities struct {
	goAPIIssuerCapabilities *openid4cigoapi.IssuerCapabilities
}

// PreAuthorizedCodeGrantTypeSupported indicates whether an issuer supports the pre-authorized code grant type.
func (i *IssuerCapabilities) PreAuthorizedCodeGrantTypeSupported() bool {
	return i.goAPIIssuerCapabilities.PreAuthorizedCodeGrantTypeSupported()
}

// PreAuthorizedCodeGrantParams returns an object that can be used to determine an issuer's pre-authorized code grant
// parameters. The caller should call the PreAuthorizedCodeGrantTypeSupported method first and only call this method to
// get the params if PreAuthorizedCodeGrantTypeSupported returns true. This method only returns an error if
// PreAuthorizedCodeGrantTypeSupported returns false, so the error return here can be safely ignored if
// PreAuthorizedCodeGrantTypeSupported returns true.
func (i *IssuerCapabilities) PreAuthorizedCodeGrantParams() (*PreAuthorizedCodeGrantParams, error) {
	goAPIPreAuthorizedCodeGrantParams, err := i.goAPIIssuerCapabilities.PreAuthorizedCodeGrantParams()
	if err != nil {
		return nil, err
	}

	return &PreAuthorizedCodeGrantParams{
		goAPIPreAuthorizedCodeGrantParams: goAPIPreAuthorizedCodeGrantParams,
	}, nil
}

// AuthorizationCodeGrantTypeSupported indicates whether an issuer supports the authorization code grant type.
func (i *IssuerCapabilities) AuthorizationCodeGrantTypeSupported() bool {
	return i.goAPIIssuerCapabilities.AuthorizationCodeGrantTypeSupported()
}
