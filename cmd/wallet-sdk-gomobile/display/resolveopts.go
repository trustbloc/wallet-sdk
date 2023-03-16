/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package display

import (
	"errors"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"

	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	goapicredentialschema "github.com/trustbloc/wallet-sdk/pkg/credentialschema"
)

// ResolveOpts contains the various parameters for the Resolve function.
type ResolveOpts struct {
	VCs               *api.VerifiableCredentialsArray // Required
	IssuerURI         string                          // Required
	PreferredLocale   string                          // Optional
	MetricsLogger     api.MetricsLogger               // Optional
	additionalHeaders api.Headers                     // Optional, must use the AddHeaders method to modify this
}

// NewResolveOpts creates a new ResolveOpts object. This function only takes in required parameters. Optional parameters
// can be set by setting the fields on the ResolveOpts object that you get back from this function directly.
func NewResolveOpts(vcs *api.VerifiableCredentialsArray, issuerURI string) *ResolveOpts {
	return &ResolveOpts{
		VCs:       vcs,
		IssuerURI: issuerURI,
	}
}

// AddHeaders adds the given HTTP headers to all REST calls made to the issuer during display resolution.
func (r *ResolveOpts) AddHeaders(headers *api.Headers) {
	headersAsArray := headers.GetAll()

	for i := range headersAsArray {
		r.additionalHeaders.Add(&headersAsArray[i])
	}
}

func prepareOpts(resolveDisplayOpts *ResolveOpts) ([]goapicredentialschema.ResolveOpt, error) {
	if resolveDisplayOpts == nil {
		return nil, errors.New("resolve display opts object cannot be nil")
	}

	if resolveDisplayOpts.VCs == nil {
		return nil, errors.New("no credentials specified")
	}

	if resolveDisplayOpts.IssuerURI == "" {
		return nil, errors.New("no issuer URI specified")
	}

	httpClient := wrapper.NewHTTPClient()
	httpClient.AddHeaders(&resolveDisplayOpts.additionalHeaders)

	opts := []goapicredentialschema.ResolveOpt{
		goapicredentialschema.WithCredentials(mobileVCsArrayToGoAPIVCsArray(resolveDisplayOpts.VCs)),
		goapicredentialschema.WithIssuerURI(resolveDisplayOpts.IssuerURI),
		goapicredentialschema.WithPreferredLocale(resolveDisplayOpts.PreferredLocale),
		goapicredentialschema.WithHTTPClient(httpClient),
	}

	if resolveDisplayOpts.MetricsLogger != nil {
		opt := goapicredentialschema.WithMetricsLogger(
			&wrapper.MobileMetricsLoggerWrapper{MobileAPIMetricsLogger: resolveDisplayOpts.MetricsLogger})

		opts = append(opts, opt)
	}

	return opts, nil
}

func mobileVCsArrayToGoAPIVCsArray(vcs *api.VerifiableCredentialsArray) []*verifiable.Credential {
	goAPIVCs := make([]*verifiable.Credential, vcs.Length())

	for i := 0; i < vcs.Length(); i++ {
		goAPIVCs[i] = vcs.AtIndex(i).VC
	}

	return goAPIVCs
}
