/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package wellknown contains a function for validating a DID's service against its well-known DID configuration.
package wellknown

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hyperledger/aries-framework-go/pkg/client/didconfig"
	diddoc "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdr"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	diderrors "github.com/trustbloc/wallet-sdk/pkg/did"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

const linkedDomainsServiceType = "LinkedDomains"

// ValidateLinkedDomains validates the given DID's Linked Domains service against its well-known DID configuration.
// It returns a bool indicating whether it's valid and the service's URL.
// The DID document associated with the given DID must specify only a single service.
// If there are multiple URLs for a given service, only the first will be checked.
// The HTTP client parameter is optional. If not provided, then a default client will be used.
func ValidateLinkedDomains(did string, resolver api.DIDResolver,
	httpClient didconfig.HTTPClient,
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
		httpClient = common.DefaultHTTPClient()
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

	err = client.VerifyDIDAndDomain(did, uri)
	if err != nil {
		return false, "", walleterror.NewExecutionError(
			diderrors.Module,
			diderrors.DomainAndDidVerificationCode,
			diderrors.DomainAndDidVerificationFailed,
			fmt.Errorf("DID service validation failed: %w", err))
	}

	return true, uri, nil
}

func getLinkedDomainsService(didDoc *diddoc.Doc) (*diddoc.Service, error) {
	var linkedDomainsService *diddoc.Service

	for i := range didDoc.Service {
		serviceType, isString := didDoc.Service[0].Type.(string)
		if !isString {
			return nil, fmt.Errorf("resolved DID document is not supported since it contains a service "+
				"type at index %d that is not a simple string", i)
		}

		if strings.EqualFold(serviceType, linkedDomainsServiceType) {
			if linkedDomainsService != nil {
				return nil, errors.New("validating multiple Linked Domains services not supported")
			}

			linkedDomainsService = &didDoc.Service[i]
		}
	}

	if linkedDomainsService == nil {
		return nil, errors.New("resolved DID document has no Linked Domains services specified")
	}

	return linkedDomainsService, nil
}

type didResolverWrapper struct {
	didResolver api.DIDResolver
}

func (d *didResolverWrapper) Resolve(did string, _ ...vdr.DIDMethodOption) (*diddoc.DocResolution, error) {
	return d.didResolver.Resolve(did)
}
