package openid4ci

import openid4cigoapi "github.com/trustbloc/wallet-sdk/pkg/openid4ci"

// PreAuthorizedCodeGrantParams represents an issuer's pre-authorized code grant parameters.
type PreAuthorizedCodeGrantParams struct {
	goAPIPreAuthorizedCodeGrantParams *openid4cigoapi.PreAuthorizedCodeGrantParams
}

// PINRequired indicates whether the issuer requires a PIN.
func (p *PreAuthorizedCodeGrantParams) PINRequired() bool {
	return p.goAPIPreAuthorizedCodeGrantParams.PINRequired()
}
