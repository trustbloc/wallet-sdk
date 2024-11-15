/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package wellknown contains a function for validating a DID's service against its well-known DID configuration.
package wellknown

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	diddoc "github.com/trustbloc/did-go/doc/did"
	vdrapi "github.com/trustbloc/did-go/vdr/api"
	didconfig "github.com/trustbloc/vc-go/didconfig/client"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	diderrors "github.com/trustbloc/wallet-sdk/pkg/did"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

const linkedDomainsServiceType = "LinkedDomains"

// HTTPClient represents an HTTP client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// ValidateLinkedDomains validates the given DID's Linked Domains service against its well-known DID configuration.
// It returns a bool indicating whether it's valid and the service's URL.
// The DID document associated with the given DID must specify only a single service.
// If there are multiple URLs for a given service, only the first will be checked.
// The HTTP client parameter is optional. If not provided, then a default client will be used.
func ValidateLinkedDomains(did string, resolver api.DIDResolver,
	httpClient HTTPClient,
) (bool, string, error) {
	if resolver == nil {
		return false, "",
			walleterror.NewExecutionError(
				diderrors.Module,
				diderrors.WellknownInitializationCode,
				diderrors.WellknownInitializationFailed,
				errors.New("no resolver provided"))
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: api.DefaultHTTPTimeout}
	}

	didDocResolution, err := resolver.Resolve(did)
	if err != nil {
		return false, "", walleterror.NewExecutionError(
			diderrors.Module,
			diderrors.WellknownInitializationCode,
			diderrors.WellknownInitializationFailed,
			fmt.Errorf("failed to resolve DID: %w", err))
	}

	linkedDomainsService, err := getLinkedDomainsService(didDocResolution.DIDDocument)
	if err != nil {
		return false, "", err
	}

	if linkedDomainsService == nil {
		return false, "", nil
	}

	client := didconfig.New(didconfig.WithHTTPClient(httpClient),
		didconfig.WithVDRegistry(&didResolverWrapper{didResolver: resolver}))

	// Note that in the case of multiple origins, this method will only return the first one.
	uri, err := linkedDomainsService.ServiceEndpoint.URI()
	if err != nil {
		return false, "", walleterror.NewExecutionError(
			diderrors.Module,
			diderrors.WellknownInitializationCode,
			diderrors.WellknownInitializationFailed, err)
	}

	didBelongsToDomain := true
	uri = strings.TrimSuffix(uri, "/")

	verErr := client.VerifyDIDAndDomain(did, uri)
	if verErr != nil {
		didBelongsToDomain = false
	}

	return didBelongsToDomain, uri, nil
}

func getLinkedDomainsService(didDoc *diddoc.Doc) (*diddoc.Service, error) {
	var linkedDomainsService *diddoc.Service

	for i := range didDoc.Service {
		serviceType, isString := didDoc.Service[i].Type.(string)
		if !isString {
			return nil, fmt.Errorf("resolved DID document is not supported since it contains a service "+
				"type at index %d that is not a simple string", i)
		}

		if strings.EqualFold(serviceType, linkedDomainsServiceType) {
			linkedDomainsService = &didDoc.Service[i]

			break
		}
	}

	return linkedDomainsService, nil
}

type didResolverWrapper struct {
	didResolver api.DIDResolver
}

func (d *didResolverWrapper) Resolve(did string, _ ...vdrapi.DIDMethodOption) (*diddoc.DocResolution, error) {
	return d.didResolver.Resolve(did)
}
