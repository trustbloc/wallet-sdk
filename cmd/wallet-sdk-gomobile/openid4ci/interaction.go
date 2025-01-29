/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

// This file has functions that are used by both the IssuerInitiatedInteraction and WalletInitiatedInteraction types.

import (
	"errors"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	"github.com/trustbloc/wallet-sdk/pkg/memstorage/legacy"
	openid4cigoapi "github.com/trustbloc/wallet-sdk/pkg/openid4ci"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"

	afgoverifiable "github.com/trustbloc/vc-go/verifiable"
)

func createGoAPIClientConfig(didResolver api.DIDResolver, opts *InteractionOpts) (*openid4cigoapi.ClientConfig, error) {
	activityLogger := createGoAPIActivityLogger(opts.activityLogger)

	httpClient := wrapper.NewHTTPClient(opts.httpTimeout, opts.additionalHeaders, opts.disableHTTPClientTLSVerification)

	goAPIClientConfig := &openid4cigoapi.ClientConfig{
		DIDResolver:                      &wrapper.VDRResolverWrapper{DIDResolver: didResolver},
		ActivityLogger:                   activityLogger,
		MetricsLogger:                    &wrapper.MobileMetricsLoggerWrapper{MobileAPIMetricsLogger: opts.metricsLogger},
		DisableVCProofChecks:             opts.disableVCProofChecks,
		NetworkDocumentLoaderHTTPTimeout: opts.httpTimeout,
		HTTPClient:                       httpClient,
	}

	if opts.documentLoader != nil {
		documentLoaderWrapper := &wrapper.DocumentLoaderWrapper{
			DocumentLoader: opts.documentLoader,
		}

		goAPIClientConfig.DocumentLoader = documentLoaderWrapper
	} else {
		dlHTTPClient := wrapper.NewHTTPClient(opts.httpTimeout, api.Headers{}, opts.disableHTTPClientTLSVerification)

		var err error

		goAPIClientConfig.DocumentLoader, err = common.CreateJSONLDDocumentLoader(dlHTTPClient, legacy.NewProvider())
		if err != nil {
			return nil, err
		}
	}

	return goAPIClientConfig, nil
}

func createGoAPIActivityLogger(mobileAPIActivityLogger api.ActivityLogger) goapi.ActivityLogger {
	if mobileAPIActivityLogger == nil {
		return nil // Will result in activity logging being disabled in the OpenID4CI IssuerInitiatedInteraction object.
	}

	return &wrapper.MobileActivityLoggerWrapper{MobileAPIActivityLogger: mobileAPIActivityLogger}
}

func convertToGoAPICreateAuthURLOpts(opts *CreateAuthorizationURLOpts) []openid4cigoapi.CreateAuthorizationURLOpt {
	if opts == nil {
		opts = NewCreateAuthorizationURLOpts()
	}

	if opts.scopes == nil {
		opts.scopes = api.NewStringArray()
	}

	goAPIOpts := []openid4cigoapi.CreateAuthorizationURLOpt{openid4cigoapi.WithScopes(opts.scopes.Strings)}

	if opts.issuerState != nil {
		goAPIOpts = append(goAPIOpts, openid4cigoapi.WithIssuerState(*opts.issuerState))
	}

	if opts.useOAuthDiscoverableClientIDScheme {
		goAPIOpts = append(goAPIOpts, openid4cigoapi.WithOAuthDiscoverableClientIDScheme())
	}

	return goAPIOpts
}

func createSigner(vm *api.VerificationMethod, crypto api.Crypto) (*common.JWSSigner, error) {
	if vm == nil {
		return nil, walleterror.NewInvalidSDKUsageError(openid4cigoapi.ErrorModule,
			errors.New("verification method must be provided"))
	}

	signer, err := common.NewJWSSigner(vm.ToSDKVerificationMethod(), crypto)
	if err != nil {
		return nil, err
	}

	return signer, nil
}

func toGomobileCredentials(credentials []*afgoverifiable.Credential) *verifiable.CredentialsArray {
	gomobileCredentials := verifiable.NewCredentialsArray()

	for i := range credentials {
		gomobileCredentials.Add(verifiable.NewCredential(credentials[i]))
	}

	return gomobileCredentials
}

func toGomobileCredentialsV2(
	credentials []*afgoverifiable.Credential,
	configIDs []string,
) *verifiable.CredentialsArrayV2 {
	credentialArray := verifiable.NewCredentialsArrayV2()

	for i := range credentials {
		credentialArray.Add(verifiable.NewCredential(credentials[i]), configIDs[i])
	}

	return credentialArray
}
