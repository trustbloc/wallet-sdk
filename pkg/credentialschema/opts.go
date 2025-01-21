/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialschema

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/trustbloc/vc-go/jwt"
	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	metadatafetcher "github.com/trustbloc/wallet-sdk/pkg/internal/issuermetadata"
	"github.com/trustbloc/wallet-sdk/pkg/metricslogger/noop"
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
	// credentialConfigIDs holds the config IDs for credentials, with each ID corresponding to the credential
	// at the same index in the vcs slice.
	credentialConfigIDs []string
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

// credentialConfigMapping represents a mapping of Credential to its corresponding CredentialConfigurationSupported.
type credentialConfigMapping struct {
	credential *verifiable.Credential
	config     map[string]*issuer.CredentialConfigurationSupported // config ID -> CredentialConfigurationSupported
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
	maskingString        *string
	signatureVerifier    jwt.ProofChecker
	skipNonClaimData     bool
}

// ResolveOpt represents an option for the Resolve function.
type ResolveOpt func(opts *resolveOpts)

// WithCredentials is an option allowing a caller to directly pass in the VCs that they want to have resolved.
func WithCredentials(vcs []*verifiable.Credential, configID ...string) ResolveOpt {
	return func(opts *resolveOpts) {
		opts.credentialSource.vcs = vcs
		opts.credentialSource.credentialConfigIDs = configID
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

// WithMaskingString is an option allowing a caller to specify a string to be used when creating masked values for
// display. The substitution is done on a character-by-character basis, whereby each individual character to be masked
// will be replaced by the entire string. See the examples below to better understand exactly how the
// substitution works.
//
// (Note that any quote characters in the examples below are only there for readability reasons - they're not actually
// part of the values.)
//
// Scenario: The unmasked display value is 12345, and the issuer's metadata specifies that the first 3 characters are
// to be masked. The most common use-case is to substitute every masked character with a single character. This is
// achieved by specifying just a single character in the maskingString. Here's what the masked value would look like
// with different maskingString choices:
//
// maskingString: "•"    -->    •••45
// maskingString: "*"    -->    ***45
//
// It's also possible to specify multiple characters in the maskingString, or even an empty string if so desired.
// Here's what the masked value would like in such cases:
//
// maskingString: "???"  -->    ?????????45
// maskingString: ""     -->    45
//
// If this option isn't used, then by default "•" characters (without the quotes) will be used for masking.
func WithMaskingString(maskingString string) ResolveOpt {
	return func(opts *resolveOpts) {
		opts.maskingString = &maskingString
	}
}

// WithJWTSignatureVerifier is an option that allows a caller to pass in a signature verifier. If the issuer metadata is
// retrieved from the issuer via an issuerURI, and it's signed, then a signature verifier must be provided so that
// the issuer metadata's signature can be verified.
func WithJWTSignatureVerifier(signatureVerifier jwt.ProofChecker) ResolveOpt {
	return func(opts *resolveOpts) {
		opts.signatureVerifier = signatureVerifier
	}
}

// WithSkipNonClaimData skips the non-claims related data like issue and expiry date.
func WithSkipNonClaimData() ResolveOpt {
	return func(opts *resolveOpts) {
		opts.skipNonClaimData = true
	}
}

func processOpts(opts []ResolveOpt) ([]*credentialConfigMapping, *issuer.Metadata, string, *string, error) {
	mergedOpts := mergeOpts(opts)

	err := validateOpts(mergedOpts)
	if err != nil {
		return nil, nil, "", nil, err
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

	if credentialSource.vcs != nil {
		if credentialSource.reader != nil {
			return errors.New("cannot have multiple credential sources specified - must use either " +
				"WithCredentials or WithCredentialReader, but not both")
		}

		numConfigIDs := len(credentialSource.credentialConfigIDs)
		numVCs := len(credentialSource.vcs)

		if numConfigIDs > 0 && numConfigIDs != numVCs {
			return fmt.Errorf("mismatch between the number of credentials (%d) and the number of config IDs (%d)",
				numVCs, numConfigIDs)
		}
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

//nolint:gocyclo
func processValidatedOpts(opts *resolveOpts) ([]*credentialConfigMapping, *issuer.Metadata, string, *string, error) {
	credentialConfigMappings, err := processVCOpts(&opts.credentialSource)
	if err != nil {
		return nil, nil, "", nil, err
	}

	var metricsLogger api.MetricsLogger

	if opts.metricsLogger == nil {
		metricsLogger = noop.NewMetricsLogger()
	} else {
		metricsLogger = opts.metricsLogger
	}

	if opts.httpClient == nil {
		opts.httpClient = &http.Client{Timeout: api.DefaultHTTPTimeout}
	}

	issuerMetadata, err := processIssuerMetadataOpts(&opts.issuerMetadataSource, opts.httpClient, metricsLogger,
		opts.signatureVerifier)
	if err != nil {
		return nil, nil, "", nil, err
	}

	for _, m := range credentialConfigMappings {
		vc := m.credential

		if len(m.config) > 0 {
			for configID := range m.config {
				config, ok := issuerMetadata.CredentialConfigurationsSupported[configID]
				if !ok {
					return nil, nil, "", nil, fmt.Errorf("credential configuration with ID %s not found", configID)
				}

				m.config[configID] = config
			}

			continue
		}

		for configID, config := range issuerMetadata.CredentialConfigurationsSupported {
			if !haveMatchingTypes(config, vc.Contents().Types) {
				continue
			}

			m.config[configID] = config

			break
		}
	}

	return credentialConfigMappings, issuerMetadata, opts.preferredLocal, opts.maskingString, nil
}

func processVCOpts(credentialSource *credentialSource) ([]*credentialConfigMapping, error) {
	var credentialConfigMappings []*credentialConfigMapping

	if credentialSource.vcs != nil {
		numVCs := len(credentialSource.vcs)
		numConfigIDs := len(credentialSource.credentialConfigIDs)

		for i := range numVCs {
			m := &credentialConfigMapping{
				credential: credentialSource.vcs[i],
				config:     make(map[string]*issuer.CredentialConfigurationSupported),
			}

			if numConfigIDs > 0 && i < numConfigIDs {
				m.config[credentialSource.credentialConfigIDs[i]] = nil
			}

			credentialConfigMappings = append(credentialConfigMappings, m)
		}
	} else {
		for _, id := range credentialSource.ids {
			vc, err := credentialSource.reader.Get(id)
			if err != nil {
				return nil, err
			}

			credentialConfigMappings = append(credentialConfigMappings,
				&credentialConfigMapping{
					credential: vc,
					config:     make(map[string]*issuer.CredentialConfigurationSupported),
				},
			)
		}
	}

	return credentialConfigMappings, nil
}

func processIssuerMetadataOpts(issuerMetadataSource *issuerMetadataSource, httpClient httpClient,
	metricsLogger api.MetricsLogger, signatureVerifier jwt.ProofChecker,
) (*issuer.Metadata, error) {
	if issuerMetadataSource.metadata != nil {
		return issuerMetadataSource.metadata, nil
	}

	metadata, err := metadatafetcher.Get(issuerMetadataSource.issuerURI,
		httpClient, metricsLogger, "Resolve display", signatureVerifier)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}
