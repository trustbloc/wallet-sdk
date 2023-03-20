/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package display

import (
	"errors"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	"github.com/trustbloc/wallet-sdk/pkg/common"

	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	goapicredentialschema "github.com/trustbloc/wallet-sdk/pkg/credentialschema"
)

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

	const minimumNumberOfOpts = 2

	opts := make([]goapicredentialschema.ResolveOpt, minimumNumberOfOpts)

	opts[0] = goapicredentialschema.WithCredentials(mobileVCsArrayToGoAPIVCsArray(resolveDisplayOpts.VCs))
	opts[1] = goapicredentialschema.WithIssuerURI(resolveDisplayOpts.IssuerURI)

	if resolveDisplayOpts.PreferredLocale != "" {
		opt := goapicredentialschema.WithPreferredLocale(resolveDisplayOpts.PreferredLocale)

		opts = append(opts, opt) //nolint:makezero // false positive
	}

	if resolveDisplayOpts.MetricsLogger != nil {
		opt := goapicredentialschema.WithMetricsLogger(
			&wrapper.MobileMetricsLoggerWrapper{MobileAPIMetricsLogger: resolveDisplayOpts.MetricsLogger})

		opts = append(opts, opt) //nolint:makezero // false positive
	}

	//nolint:makezero // false positive
	opts = append(opts, goapicredentialschema.WithHTTPClient(common.DefaultHTTPClient()))

	return opts, nil
}

func mobileVCsArrayToGoAPIVCsArray(vcs *api.VerifiableCredentialsArray) []*verifiable.Credential {
	goAPIVCs := make([]*verifiable.Credential, vcs.Length())

	for i := 0; i < vcs.Length(); i++ {
		goAPIVCs[i] = vcs.AtIndex(i).VC
	}

	return goAPIVCs
}
