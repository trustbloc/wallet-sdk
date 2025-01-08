/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package display

import (
	"errors"

	"github.com/trustbloc/vc-go/proof/defaults"
	afgoverifiable "github.com/trustbloc/vc-go/verifiable"
	goapicredentialschema "github.com/trustbloc/wallet-sdk/pkg/credentialschema"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/openid4ci"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	"github.com/trustbloc/wallet-sdk/pkg/common"
)

// Resolve resolves display information for issued credentials based on an issuer's metadata, which is fetched
// using the issuer's (base) URI.
// The CredentialDisplays in the returned Data object correspond to the VCs passed in and are in the
// same order.
// This method requires one or more VCs and the issuer's base URI.
// Deprecated: Use ResolveCredential function instead, which would give data for all locales. The consumer of the SDK
// can run logic to display exact locale data to the user.
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

func ResolveCredential(credentialsArray *verifiable.CredentialsArray, issuerURI string, opts *Opts) (*Resolved, error) {
	goAPIOpts, err := generateGoAPIOpts(credentialsArray, issuerURI, opts)
	if err != nil {
		return nil, err
	}

	resolvedDisplayData, err := goapicredentialschema.ResolveCredential(goAPIOpts...)
	if err != nil {
		return nil, err
	}

	return &Resolved{resolvedDisplayData: resolvedDisplayData}, nil
}

func ResolveCredentialV2(
	credentialsArray *verifiable.CredentialsArrayV2,
	issuerURI string,
	opts *Opts,
) (*Resolved, error) {
	credentials := &verifiable.CredentialsArray{}
	opts.credentialConfigIDs = make([]string, credentialsArray.Length())

	for i := range credentialsArray.Length() {
		credentials.Add(credentialsArray.AtIndex(i))
		opts.credentialConfigIDs[i] = credentialsArray.ConfigIDAtIndex(i)
	}

	goAPIOpts, err := generateGoAPIOpts(credentials, issuerURI, opts)
	if err != nil {
		return nil, err
	}

	resolvedDisplayData, err := goapicredentialschema.ResolveCredential(goAPIOpts...)
	if err != nil {
		return nil, err
	}

	return &Resolved{resolvedDisplayData: resolvedDisplayData}, nil
}

// ResolveCredentialOffer resolves display information for some offered credentials based on an issuer's metadata.
// The CredentialDisplays in the returned ResolvedDisplayData object correspond to the offered credential types
// passed in and are in the same order.
func ResolveCredentialOffer(
	issuerMetadata *openid4ci.IssuerMetadata, offeredTypes *api.StringArrayArray, preferredLocale string,
) *Data {
	resolvedDisplayData := goapicredentialschema.ResolveCredentialOffer(openid4ci.IssuerMetadataToGoImpl(issuerMetadata),
		api.StringArrayArrayToGoArray(offeredTypes),
		preferredLocale)

	return &Data{resolvedDisplayData: resolvedDisplayData}
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
		goapicredentialschema.WithCredentials(mobileVCsArrayToGoAPIVCsArray(vcs), opts.credentialConfigIDs...),
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

	if opts.skipNonClaimData {
		goAPIOpt := goapicredentialschema.WithSkipNonClaimData()

		goAPIOpts = append(goAPIOpts, goAPIOpt)
	}

	if opts.didResolver != nil {
		jwtVerifier := defaults.NewDefaultProofChecker(
			common.NewVDRKeyResolver(&wrapper.VDRResolverWrapper{
				DIDResolver: opts.didResolver,
			}))

		goAPIOpt := goapicredentialschema.WithJWTSignatureVerifier(jwtVerifier)

		goAPIOpts = append(goAPIOpts, goAPIOpt)
	}

	return goAPIOpts, nil
}

func mobileVCsArrayToGoAPIVCsArray(vcs *verifiable.CredentialsArray) []*afgoverifiable.Credential {
	goAPIVCs := make([]*afgoverifiable.Credential, vcs.Length())

	for i := range vcs.Length() {
		goAPIVCs[i] = vcs.AtIndex(i).VC
	}

	return goAPIVCs
}
