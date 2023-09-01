/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package display

import (
	"errors"

	afgoverifiable "github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	goapicredentialschema "github.com/trustbloc/wallet-sdk/pkg/credentialschema"
)

// Resolve resolves display information for issued credentials based on an issuer's metadata, which is fetched
// using the issuer's (base) URI.
// The CredentialDisplays in the returned Data object correspond to the VCs passed in and are in the
// same order.
// This method requires one or more VCs and the issuer's base URI.
func Resolve(vcs *verifiable.CredentialsArray, issuerURI string, opts *Opts) (*Data, error) {
	goAPIOpts, err := generateGoAPIOpts(vcs, issuerURI, opts)
	if err != nil {
		return nil, err
	}

	resolvedDisplayData, err := goapicredentialschema.Resolve(goAPIOpts...)
	if err != nil {
		return nil, err
	}

	return &Data{resolvedDisplayData: resolvedDisplayData}, nil
}

func generateGoAPIOpts(vcs *verifiable.CredentialsArray, issuerURI string,
	opts *Opts,
) ([]goapicredentialschema.ResolveOpt, error) {
	if vcs == nil {
		return nil, errors.New("no credentials specified")
	}

	if issuerURI == "" {
		return nil, errors.New("no issuer URI specified")
	}

	if opts == nil {
		opts = NewOpts()
	}

	httpClient := wrapper.NewHTTPClient(opts.httpTimeout, opts.additionalHeaders, opts.disableHTTPClientTLSVerification)

	goAPIOpts := []goapicredentialschema.ResolveOpt{
		goapicredentialschema.WithCredentials(mobileVCsArrayToGoAPIVCsArray(vcs)),
		goapicredentialschema.WithIssuerURI(issuerURI),
		goapicredentialschema.WithPreferredLocale(opts.preferredLocale),
		goapicredentialschema.WithHTTPClient(httpClient),
	}

	if opts.metricsLogger != nil {
		goAPIOpt := goapicredentialschema.WithMetricsLogger(
			&wrapper.MobileMetricsLoggerWrapper{MobileAPIMetricsLogger: opts.metricsLogger})

		goAPIOpts = append(goAPIOpts, goAPIOpt)
	}

	if opts.maskingString != nil {
		goAPIOpt := goapicredentialschema.WithMaskingString(*opts.maskingString)

		goAPIOpts = append(goAPIOpts, goAPIOpt)
	}

	return goAPIOpts, nil
}

func mobileVCsArrayToGoAPIVCsArray(vcs *verifiable.CredentialsArray) []*afgoverifiable.Credential {
	goAPIVCs := make([]*afgoverifiable.Credential, vcs.Length())

	for i := 0; i < vcs.Length(); i++ {
		goAPIVCs[i] = vcs.AtIndex(i).VC
	}

	return goAPIVCs
}
