package oauth2

import (
	"net/http"

	"github.com/trustbloc/wallet-sdk/pkg/api"
)

type opts struct {
	initialAccessBearerToken string
	httpClient               *http.Client
}

// An Opt is a single option for a call to RegisterClient.
type Opt func(opts *opts)

// WithInitialAccessBearerToken is an option for a call to RegisterClient that allows a caller to specify an initial
// access bearer token to use for the client registration request, which may be required by the server.
func WithInitialAccessBearerToken(token string) Opt {
	return func(opts *opts) {
		opts.initialAccessBearerToken = token
	}
}

// WithHTTPClient is an option for a call to RegisterClient that allows a caller to specify their own HTTP client.
func WithHTTPClient(httpClient *http.Client) Opt {
	return func(opts *opts) {
		opts.httpClient = httpClient
	}
}

func processOpts(options []Opt) *opts {
	opts := mergeOpts(options)

	if opts.httpClient == nil {
		opts.httpClient = &http.Client{Timeout: api.DefaultHTTPTimeout}
	}

	return opts
}

func mergeOpts(options []Opt) *opts {
	resolveOpts := &opts{}

	for _, opt := range options {
		if opt != nil {
			opt(resolveOpts)
		}
	}

	return resolveOpts
}
