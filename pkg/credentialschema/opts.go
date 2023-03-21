/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialschema

import (
	"errors"
	"net/http"

	"github.com/trustbloc/wallet-sdk/pkg/metricslogger/noop"

	"github.com/trustbloc/wallet-sdk/pkg/api"

	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"

	metadatafetcher "github.com/trustbloc/wallet-sdk/pkg/internal/issuermetadata"
	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
)

// credentialSource represents the different ways that credentials can be passed in to the Resolve function.
// At most one out of vcs and reader can be used for a given call to Resolve.
// If reader is specified, then ids must also be specified. The corresponding credentials will be
// retrieved from the credentialReader.
type credentialSource struct {
	// vcs is a slice of Verifiable Credentials.
	vcs []*verifiable.Credential
	// reader allows for access to VCs stored via some storage mechanism.
	reader credentialReader
	// ids specifies which credentials should be retrieved from the reader.
	ids []string
}

// issuerMetadataSource represents the different ways that issuer metadata can be specified in the Resolve function.
// At most one out of issuerURI and metadata can be used for a given call to Resolve.
// Setting issuerURI will cause the Resolve function to fetch an issuer's metadata by doing a lookup on its
// OpenID configuration endpoint. issuerURI is expected to be the base URL for the issuer.
// Alternatively, if metadata is set, then it will be used directly.
type issuerMetadataSource struct {
	issuerURI string
	metadata  *issuer.Metadata
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type resolveOpts struct {
	credentialSource     credentialSource
	issuerMetadataSource issuerMetadataSource
	preferredLocal       string
	metricsLogger        api.MetricsLogger
	httpClient           httpClient
}

// ResolveOpt represents an option for the Resolve function.
type ResolveOpt func(opts *resolveOpts)

// WithCredentials is an option allowing a caller to directly pass in the VCs that they want to have resolved.
func WithCredentials(vcs []*verifiable.Credential) ResolveOpt {
	return func(opts *resolveOpts) {
		opts.credentialSource.vcs = vcs
	}
}

// WithMetricsLogger is an option for an OpenID4VP instance that allows a caller to specify their MetricsLogger.
// If used, then performance metrics events will be pushed to the given MetricsLogger implementation.
// If this option is not used, then metrics logging will be disabled.
func WithMetricsLogger(metricsLogger api.MetricsLogger) ResolveOpt {
	return func(opts *resolveOpts) {
		opts.metricsLogger = metricsLogger
	}
}

// A credentialReader is capable of reading VCs from some underlying storage mechanism.
type credentialReader interface {
	// Get retrieves a VC.
	Get(id string) (*verifiable.Credential, error)
}

// WithCredentialReader is an option allowing a caller to specify the VCs they want to have resolved by providing their
// IDs along with a CredentialReader.
func WithCredentialReader(credentialReader credentialReader, ids []string) ResolveOpt {
	return func(opts *resolveOpts) {
		opts.credentialSource.reader = credentialReader
		opts.credentialSource.ids = ids
	}
}

// WithIssuerURI is an option allowing a caller to specify an issuer URI that will be used to fetch metadata. Using this
// option will cause the Resolve function to fetch an issuer's metadata by doing a lookup on its OpenID configuration
// endpoint. The issuer URI is expected to be the base URL for the issuer.
func WithIssuerURI(uri string) ResolveOpt {
	return func(opts *resolveOpts) {
		opts.issuerMetadataSource.issuerURI = uri
	}
}

// WithIssuerMetadata is an option allowing a caller to directly pass in the issuer's metadata to use for resolving VCs.
func WithIssuerMetadata(metadata *issuer.Metadata) ResolveOpt {
	return func(opts *resolveOpts) {
		opts.issuerMetadataSource.metadata = metadata
	}
}

// WithPreferredLocale is an option specifying the caller's preferred locale to look for while resolving VC display
// data. If the preferred locale is not available (or this option is no used), then the first locale specified by the
// issuer's metadata will be used during resolution. The actual locales used for various pieces of display information
// are available in the ResolvedDisplayData object.
func WithPreferredLocale(locale string) ResolveOpt {
	return func(opts *resolveOpts) {
		opts.preferredLocal = locale
	}
}

// WithHTTPClient is an option allowing a caller to specify their own HTTP client implementation.
func WithHTTPClient(httpClient httpClient) ResolveOpt {
	return func(opts *resolveOpts) {
		opts.httpClient = httpClient
	}
}

func processOpts(opts []ResolveOpt) ([]*verifiable.Credential, *issuer.Metadata, string, error) {
	mergedOpts := mergeOpts(opts)

	err := validateOpts(mergedOpts)
	if err != nil {
		return nil, nil, "", err
	}

	return processValidatedOpts(mergedOpts)
}

func mergeOpts(opts []ResolveOpt) *resolveOpts {
	resolveOpts := &resolveOpts{}
	for _, opt := range opts {
		opt(resolveOpts)
	}

	return resolveOpts
}

func validateOpts(opts *resolveOpts) error {
	err := validateVCOpts(&opts.credentialSource)
	if err != nil {
		return err
	}

	return validateIssuerMetadataOpts(&opts.issuerMetadataSource)
}

func validateVCOpts(credentialSource *credentialSource) error {
	if credentialSource.vcs == nil && credentialSource.reader == nil {
		return errors.New("no credentials specified")
	}

	if credentialSource.vcs != nil && credentialSource.reader != nil {
		return errors.New("cannot have multiple credential sources specified - must use either " +
			"WithCredentials or WithCredentialReader, but not both")
	}

	if credentialSource.reader != nil && len(credentialSource.ids) == 0 {
		return errors.New("credential IDs must be provided when using a credential reader")
	}

	return nil
}

func validateIssuerMetadataOpts(issuerMetadataSource *issuerMetadataSource) error {
	if issuerMetadataSource.issuerURI == "" && issuerMetadataSource.metadata == nil {
		return errors.New("no issuer metadata source specified")
	}

	return nil
}

func processValidatedOpts(opts *resolveOpts) ([]*verifiable.Credential, *issuer.Metadata, string, error) {
	vcs, err := processVCOpts(&opts.credentialSource)
	if err != nil {
		return nil, nil, "", err
	}

	var metricsLogger api.MetricsLogger

	if opts.metricsLogger == nil {
		metricsLogger = noop.NewMetricsLogger()
	} else {
		metricsLogger = opts.metricsLogger
	}

	if opts.httpClient == nil {
		opts.httpClient = http.DefaultClient
	}

	issuerMetadata, err := processIssuerMetadataOpts(&opts.issuerMetadataSource, opts.httpClient, metricsLogger)
	if err != nil {
		return nil, nil, "", err
	}

	return vcs, issuerMetadata, opts.preferredLocal, nil
}

func processVCOpts(credentialSource *credentialSource) ([]*verifiable.Credential, error) {
	var vcs []*verifiable.Credential

	if credentialSource.vcs != nil {
		vcs = credentialSource.vcs
	} else {
		for _, id := range credentialSource.ids {
			vc, err := credentialSource.reader.Get(id)
			if err != nil {
				return nil, err
			}

			vcs = append(vcs, vc)
		}
	}

	return vcs, nil
}

func processIssuerMetadataOpts(issuerMetadataSource *issuerMetadataSource, httpClient httpClient,
	metricsLogger api.MetricsLogger,
) (*issuer.Metadata, error) {
	if issuerMetadataSource.metadata != nil {
		return issuerMetadataSource.metadata, nil
	}

	metadata, err := metadatafetcher.Get(issuerMetadataSource.issuerURI,
		httpClient, metricsLogger, "Resolve display")
	if err != nil {
		return nil, err
	}

	return metadata, nil
}
