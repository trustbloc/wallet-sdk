package openid4ci

import (
	"errors"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/otel"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"

	openid4cigoapi "github.com/trustbloc/wallet-sdk/pkg/openid4ci"
)

// PreAuthorizedCodeGrantParams represents an issuer's pre-authorized code grant parameters.
type PreAuthorizedCodeGrantParams struct {
	goAPIPreAuthorizedCodeGrantParams *openid4cigoapi.PreAuthorizedCodeGrantParams
}

// PINRequired indicates whether the issuer requires a PIN.
func (p *PreAuthorizedCodeGrantParams) PINRequired() bool {
	return p.goAPIPreAuthorizedCodeGrantParams.PINRequired()
}

// AuthorizationCodeGrantParams represents an issuer's authorization code grant parameters.
type AuthorizationCodeGrantParams struct {
	goAPIAuthorizationCodeGrantParams *openid4cigoapi.AuthorizationCodeGrantParams
	oTel                              *otel.Trace
}

// HasIssuerState indicates whether this AuthorizationCodeGrantParams has an issuer state string.
func (a *AuthorizationCodeGrantParams) HasIssuerState() bool {
	return a.goAPIAuthorizationCodeGrantParams.IssuerState != nil
}

// IssuerState returns the issuer state string. The HasIssuerState method should be called first to
// ensure this AuthorizationCodeGrantParams object has an issuer state string first before calling this method.
// This method returns an error if (and only if) HasIssuerState returns false.
func (a *AuthorizationCodeGrantParams) IssuerState() (string, error) {
	if a.goAPIAuthorizationCodeGrantParams.IssuerState != nil {
		return *a.goAPIAuthorizationCodeGrantParams.IssuerState, nil
	}

	return "", wrapper.ToMobileErrorWithTrace(walleterror.NewInvalidSDKUsageError(openid4cigoapi.ErrorModule,
		errors.New("authorization code grant params does not specify an issuer state")), a.oTel)
}
