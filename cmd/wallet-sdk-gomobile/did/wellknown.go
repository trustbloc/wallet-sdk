/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package did

import (
	"errors"
	"net/http"

	"github.com/hyperledger/aries-framework-go/pkg/client/didconfig"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	goapiwellknown "github.com/trustbloc/wallet-sdk/pkg/did/wellknown"
)

// A ValidationResult is the type returned from the ValidateLinkedDomains method.
// IsValid indicates if the given DID passed the service validation check, and ServiceURL indicates the URL of
// that service.
type ValidationResult struct {
	IsValid    bool
	ServiceURL string
}

// ValidateLinkedDomains validates the given DID's Linked Domains service against its well-known DID configuration.
// It returns a ValidationResult.
// The DID document must specify only a single service. If there are multiple URLs for a given service, only the
// first will be checked.
func ValidateLinkedDomains(did string, resolver api.DIDResolver) (*ValidationResult, error) {
	return validateLinkedDomains(did, resolver, http.DefaultClient)
}

func validateLinkedDomains(did string, resolver api.DIDResolver,
	client didconfig.HTTPClient,
) (*ValidationResult, error) {
	if resolver == nil {
		return nil, errors.New("no resolver provided")
	}

	vdrWrapper := &wrapper.VDRResolverWrapper{DIDResolver: resolver}

	valid, uri, err := goapiwellknown.ValidateLinkedDomains(did, vdrWrapper, client)
	if err != nil {
		return nil, wrapper.ToMobileError(err)
	}

	return &ValidationResult{IsValid: valid, ServiceURL: uri}, nil
}
