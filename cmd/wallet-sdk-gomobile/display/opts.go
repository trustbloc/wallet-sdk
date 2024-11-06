/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package display

import (
	"time"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

// Opts contains all optional arguments that can be passed into the Resolve function.
type Opts struct {
	preferredLocale                  string
	metricsLogger                    api.MetricsLogger
	additionalHeaders                api.Headers
	httpTimeout                      *time.Duration
	disableHTTPClientTLSVerification bool
	maskingString                    *string
	didResolver                      api.DIDResolver
	skipNonClaimData                 bool
	credentialConfigIDs              []string
}

// NewOpts returns a new Opts object.
func NewOpts() *Opts {
	return &Opts{}
}

// SetPreferredLocale sets the preferred locale to use while resolving VC display data.
// If the preferred locale is not available (or no preferred locale is specified), then the first locale specified by
// the issuer's metadata will be used during resolution.
// The actual locales used for various pieces of display information are available in the Data object.
func (o *Opts) SetPreferredLocale(preferredLocale string) *Opts {
	o.preferredLocale = preferredLocale

	return o
}

// SetMetricsLogger sets a metrics logger to use.
func (o *Opts) SetMetricsLogger(metricsLogger api.MetricsLogger) *Opts {
	o.metricsLogger = metricsLogger

	return o
}

// SetHTTPTimeoutNanoseconds sets the timeout (in nanoseconds) for HTTP calls.
// Passing in 0 will disable timeouts.
func (o *Opts) SetHTTPTimeoutNanoseconds(timeout int64) *Opts {
	timeoutDuration := time.Duration(timeout)
	o.httpTimeout = &timeoutDuration

	return o
}

// AddHeaders adds the given HTTP headers to all REST calls made to the issuer during display resolution.
func (o *Opts) AddHeaders(headers *api.Headers) *Opts {
	headersAsArray := headers.GetAll()

	for i := range headersAsArray {
		o.additionalHeaders.Add(&headersAsArray[i])
	}

	return o
}

// DisableHTTPClientTLSVerify disables TLS verification. Should be used for testing purposes only.
func (o *Opts) DisableHTTPClientTLSVerify() *Opts {
	o.disableHTTPClientTLSVerification = true

	return o
}

// SetMaskingString sets the string to be used when creating masked values for display.
// The substitution is done on a character-by-character basis, whereby each individual character to be masked
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
func (o *Opts) SetMaskingString(maskingString string) *Opts {
	o.maskingString = &maskingString

	return o
}

// SetDIDResolver sets a DID resolver to be used. If the issuer metadata is signed, then a DID resolver must be
// provided so that the issuer metadata's signature can be verified.
func (o *Opts) SetDIDResolver(didResolver api.DIDResolver) *Opts {
	o.didResolver = didResolver

	return o
}

// SkipNonClaimData skips the non-claims related data like issue and expiry date.
func (o *Opts) SkipNonClaimData() *Opts {
	o.skipNonClaimData = true

	return o
}
