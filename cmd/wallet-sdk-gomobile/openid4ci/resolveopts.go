/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import (
	"errors"

	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	goapicredentialschema "github.com/trustbloc/wallet-sdk/pkg/credentialschema"
)

func prepareOpts(vcs *api.VerifiableCredentialsArray,
	issuerURI, preferredLocale string,
) ([]goapicredentialschema.ResolveOpt, error) {
	if vcs == nil {
		return nil, errors.New("no credentials specified")
	}

	if issuerURI == "" {
		return nil, errors.New("no issuer URI specified")
	}

	const minimumNumberOfOpts = 2

	opts := make([]goapicredentialschema.ResolveOpt, minimumNumberOfOpts)

	opts[0] = goapicredentialschema.WithCredentials(mobileVCsArrayToGoAPIVCsArray(vcs))
	opts[1] = goapicredentialschema.WithIssuerURI(issuerURI)

	if preferredLocale != "" {
		opt := goapicredentialschema.WithPreferredLocale(preferredLocale)

		opts = append(opts, opt) //nolint:makezero // false positive
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
