/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/piprate/json-gold/ld"
	"github.com/trustbloc/wallet-sdk/pkg/common"

	"github.com/trustbloc/wallet-sdk/cmd/utilities/gomobilewrappers"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	goapicredentialschema "github.com/trustbloc/wallet-sdk/pkg/credentialschema"
	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
)

// We don't check here to see if multiple conflicting options are used - that's left up to the
// goapicredentialschema.Resolve method that get called after this function returns.
func prepareOpts(credentials *Credentials, issuerMetadata *IssuerMetadata,
	preferredLocale string,
) ([]goapicredentialschema.ResolveOpt, error) {
	if credentials == nil {
		return nil, errors.New("no credentials specified")
	}

	if issuerMetadata == nil {
		return nil, errors.New("no issuer metadata source specified")
	}

	var opts []goapicredentialschema.ResolveOpt

	credentialOpts, err := prepareCredentialsOpts(credentials)
	if err != nil {
		return nil, err
	}

	opts = append(opts, credentialOpts...)

	issuerMetadataOpts, err := prepareIssuerMetadataOpts(issuerMetadata, preferredLocale)
	if err != nil {
		return nil, err
	}

	opts = append(opts, issuerMetadataOpts...)

	return opts, nil
}

func prepareCredentialsOpts(credentials *Credentials) ([]goapicredentialschema.ResolveOpt, error) {
	var opts []goapicredentialschema.ResolveOpt

	if credentials.VCs != nil && credentials.VCs.Data != nil {
		opt, err := generateWithCredentialsOpt(credentials.VCs)
		if err != nil {
			return nil, err
		}

		opts = append(opts, opt)
	}

	if credentials.Reader != nil {
		opt, err := generateWithCredentialReaderOpt(credentials)
		if err != nil {
			return nil, err
		}

		opts = append(opts, opt)
	}

	return opts, nil
}

func generateWithCredentialsOpt(vcs *api.JSONArray) (goapicredentialschema.ResolveOpt, error) {
	var vcsRaw []interface{}

	err := json.Unmarshal(vcs.Data, &vcsRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal VCs into an array: %w", err)
	}

	var credentials []*verifiable.Credential

	for _, vcRaw := range vcsRaw {
		vcBytes, err := json.Marshal(vcRaw)
		if err != nil {
			return nil, err
		}

		credential, err := verifiable.ParseCredential(vcBytes,
			verifiable.WithJSONLDDocumentLoader(ld.NewDefaultDocumentLoader(common.DefaultHTTPClient())),
			verifiable.WithDisabledProofCheck())
		if err != nil {
			return nil, fmt.Errorf("failed to parse credential: %w", err)
		}

		credentials = append(credentials, credential)
	}

	return goapicredentialschema.WithCredentials(credentials), nil
}

func generateWithCredentialReaderOpt(credentials *Credentials) (goapicredentialschema.ResolveOpt, error) {
	var ids []string

	if credentials.IDs != nil {
		err := json.Unmarshal(credentials.IDs.Data, &ids)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal credential IDs into a []string: %w", err)
		}
	}

	opt := goapicredentialschema.WithCredentialReader(&gomobilewrappers.CredentialReaderWrapper{
		CredentialReader: credentials.Reader,
		DocumentLoader:   ld.NewDefaultDocumentLoader(common.DefaultHTTPClient()),
	}, ids)

	return opt, nil
}

func prepareIssuerMetadataOpts(issuerMetadata *IssuerMetadata,
	preferredLocale string,
) ([]goapicredentialschema.ResolveOpt, error) {
	var opts []goapicredentialschema.ResolveOpt

	if issuerMetadata.IssuerURI != "" {
		opt := goapicredentialschema.WithIssuerURI(issuerMetadata.IssuerURI)

		opts = append(opts, opt)
	}

	if issuerMetadata.Metadata != nil && issuerMetadata.Metadata.Data != nil {
		opt, err := generateWithIssuerMetadataOpt(issuerMetadata)
		if err != nil {
			return nil, err
		}

		opts = append(opts, opt)
	}

	if preferredLocale != "" {
		opt := goapicredentialschema.WithPreferredLocale(preferredLocale)

		opts = append(opts, opt)
	}

	return opts, nil
}

func generateWithIssuerMetadataOpt(issuerMetadata *IssuerMetadata) (goapicredentialschema.ResolveOpt, error) {
	var issuerMetadataParsed issuer.Metadata

	err := json.Unmarshal(issuerMetadata.Metadata.Data, &issuerMetadataParsed)
	if err != nil {
		return nil, err
	}

	opt := goapicredentialschema.WithIssuerMetadata(&issuerMetadataParsed)

	return opt, nil
}
