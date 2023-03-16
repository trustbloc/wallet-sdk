/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package display

import (
	goapicredentialschema "github.com/trustbloc/wallet-sdk/pkg/credentialschema"
)

// Resolve resolves display information for issued credentials based on an issuer's metadata, which is fetched
// using the issuer's (base) URI.
// The CredentialDisplays in the returned Data object correspond to the VCs passed in and are in the
// same order.
// This method requires one or more VCs and the issuer's base URI.
// PreferredLocale is optional parameter that allows the caller to specify their preferred locale to look for while
// resolving VC display data. If the preferred locale is not available (or the parameter is not specified),
// then the first locale specified by the issuer's metadata will be used during resolution. The actual locales used
// for various pieces of display information are available in the Data object.
// MetricsLogger is optional parameter that, if set, will enable performance metrics logging. Metrics events will
// be pushed to the provided implementation.
func Resolve(resolveDisplayOpts *ResolveOpts) (*Data, error) {
	opts, err := prepareOpts(resolveDisplayOpts)
	if err != nil {
		return nil, err
	}

	resolvedDisplayData, err := goapicredentialschema.Resolve(opts...)
	if err != nil {
		return nil, err
	}

	return &Data{resolvedDisplayData: resolvedDisplayData}, nil
}
