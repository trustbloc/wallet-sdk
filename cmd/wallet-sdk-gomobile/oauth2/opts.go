/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package oauth2

import (
	"time"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

// RegisterClientOpts contains all optional arguments that can be passed into the RegisterClient and
// RegisterClientWithInitialAccessToken functions.
type RegisterClientOpts struct {
	initialAccessBearerToken         string
	additionalHeaders                api.Headers
	disableHTTPClientTLSVerification bool
	httpTimeout                      *time.Duration
}

// NewRegisterClientOpts returns a new RegisterClientOpts object.
func NewRegisterClientOpts() *RegisterClientOpts {
	return &RegisterClientOpts{}
}

// SetInitialAccessBearerToken is an option for the RegisterClient function that allows a caller to set an
// access bearer token to use for the client registration request, which may be required by the server.
func (o *RegisterClientOpts) SetInitialAccessBearerToken(token string) *RegisterClientOpts {
	o.initialAccessBearerToken = token

	return o
}

// SetHTTPTimeoutNanoseconds sets the timeout (in nanoseconds) for HTTP calls.
// Passing in 0 will disable timeouts.
func (o *RegisterClientOpts) SetHTTPTimeoutNanoseconds(timeout int64) *RegisterClientOpts {
	timeoutDuration := time.Duration(timeout)
	o.httpTimeout = &timeoutDuration

	return o
}

// AddHeaders adds the given HTTP headers to all REST calls made to the issuer during the OpenID4CI flow.
func (o *RegisterClientOpts) AddHeaders(headers *api.Headers) *RegisterClientOpts {
	headersAsArray := headers.GetAll()

	for i := range headersAsArray {
		o.additionalHeaders.Add(&headersAsArray[i])
	}

	return o
}

// AddHeader adds the given HTTP header to all REST calls made to the issuer during the OpenID4CI flow.
func (o *RegisterClientOpts) AddHeader(header *api.Header) *RegisterClientOpts {
	o.additionalHeaders.Add(header)

	return o
}

// DisableHTTPClientTLSVerify disables tls verification, should be used only for test purposes.
func (o *RegisterClientOpts) DisableHTTPClientTLSVerify() *RegisterClientOpts {
	o.disableHTTPClientTLSVerification = true

	return o
}
