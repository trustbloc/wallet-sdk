/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package credentialschema contains a function that can be used to resolve display values per the OpenID4CI spec.
// This implementation follows the 27 October 2022 revision of
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0.html#section-11.2
package credentialschema

import (
	"encoding/json"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	goapicredentialschema "github.com/trustbloc/wallet-sdk/pkg/credentialschema"
)

// Credentials represents the different ways that credentials can be passed in to the Resolve function.
// At most one out of VCs and Reader can be used for a given call to Resolve.
// If reader is specified, then IDs must also be specified. The corresponding credentials will be
// retrieved from the credentialReader.
type Credentials struct {
	// VCs is an array of Verifiable Credentials.
	VCs *api.JSONArray
	// Reader allows for access to VCs stored via some storage mechanism. This is ignored if VCs is set.
	Reader api.CredentialReader
	// IDs specifies which credentials should be retrieved from the reader as a JSON array of strings.
	IDs *api.JSONArray
}

// IssuerMetadata represents the different ways that issuer metadata can be specified in the Resolve function.
// At most one out of issuerURI and metadata can be used for a given call to Resolve.
// Setting issuerURI will cause the Resolve function to fetch an issuer's metadata by doing a lookup on its
// OpenID configuration endpoint. issuerURI is expected to be the base URL for the issuer.
// Alternatively, if metadata is set, then it will be used directly.
type IssuerMetadata struct {
	IssuerURI string
	Metadata  *api.JSONObject
}

// Resolve resolves display information for some issued credentials based on an issuer's metadata.
// The CredentialDisplays in the returned ResolvedDisplayData object correspond to the VCs passed in and are in the
// same order.
// This method requires one VC source and one issuer metadata source.
func Resolve(credentials *Credentials, issuerMetadata *IssuerMetadata,
	preferredLocale string,
) (*api.JSONObject, error) {
	opts, err := prepareOpts(credentials, issuerMetadata, preferredLocale)
	if err != nil {
		return nil, err
	}

	resolvedDisplayData, err := goapicredentialschema.Resolve(opts...)
	if err != nil {
		return nil, err
	}

	resolvedDisplayDataBytes, err := json.Marshal(resolvedDisplayData)
	if err != nil {
		return nil, err
	}

	return &api.JSONObject{Data: resolvedDisplayDataBytes}, nil
}
