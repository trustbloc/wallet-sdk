//go:build js && wasm

/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package jsinterop implements interop between Wallet-SDK and JS.
package jsinterop

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/hyperledger/aries-framework-go/component/kmscrypto/kms"
	arieskms "github.com/hyperledger/aries-framework-go/spi/kms"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-js/jsinterop/errors"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-js/jsinterop/indexeddb"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-js/jsinterop/jssupport"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-js/jsinterop/types"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-js/walletsdk"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

const (
	dbNamespace = "wallet-sdk"
)

var (
	agentInstance      *walletsdk.Agent
	agentMethodsRunner jssupport.AsyncRunner
)

func initAgent(_ js.Value, args []js.Value) (any, error) {
	didResolverURI, err := jssupport.EnsureString(jssupport.GetOptionalNamedArgument(args, "didResolverURI"))
	if err != nil {
		return nil, err
	}

	if agentInstance != nil {
		return nil, walleterror.NewExecutionError(
			errors.Module,
			errors.InitializationFailedCode,
			errors.InitializationFailedError,
			fmt.Errorf("agent instance already initialized"))
	}

	indexedDBKMSProvider, err := indexeddb.NewProvider(dbNamespace, []string{
		kms.AriesWrapperStoreName,
	})
	if err != nil {
		return nil, walleterror.NewExecutionError(
			errors.Module,
			errors.InitializationFailedCode,
			errors.InitializationFailedError,
			fmt.Errorf("failed to create IndexedDB provider: %w", err))
	}

	kmsStore, err := kms.NewAriesProviderWrapper(indexedDBKMSProvider)
	if err != nil {
		return nil, walleterror.NewExecutionError(
			errors.Module,
			errors.InitializationFailedCode,
			errors.InitializationFailedError,
			fmt.Errorf("failed to create Aries KMS store wrapper %w", err))
	}

	agentInstance, err = walletsdk.NewAgent(didResolverURI, kmsStore)
	if err != nil {
		return nil, walleterror.NewExecutionError(
			errors.Module,
			errors.InitializationFailedCode,
			errors.InitializationFailedError,
			fmt.Errorf("failed to create agant %w", err))
	}

	return nil, nil
}

func createDID(_ js.Value, args []js.Value) (any, error) {
	if agentInstance == nil {
		return nil, walleterror.NewExecutionError(
			errors.Module,
			errors.InitializationFailedCode,
			errors.InitializationFailedError,
			fmt.Errorf("agent instance is not initialized"))
	}

	didMethod, err := jssupport.EnsureString(jssupport.GetNamedArgument(args, "didMethod"))
	if err != nil {
		return nil, err
	}

	keyType, err := jssupport.EnsureString(jssupport.GetOptionalNamedArgument(args, "keyType"))
	if err != nil {
		return nil, err
	}

	verificationType, err := jssupport.EnsureString(jssupport.GetOptionalNamedArgument(args, "verificationType"))
	if err != nil {
		return nil, err
	}

	didDoc, err := agentInstance.CreateDID(didMethod, arieskms.KeyType(keyType), verificationType)
	if err != nil {
		return nil, err
	}

	return types.SerializeDIDDoc(didDoc)
}

func createOpenID4CIIssuerInitiatedInteraction(_ js.Value, args []js.Value) (any, error) {
	if agentInstance == nil {
		return nil, walleterror.NewExecutionError(
			errors.Module,
			errors.InitializationFailedCode,
			errors.InitializationFailedError,
			fmt.Errorf("agent instance is not initialized"))
	}

	initiateIssuanceURI, err := jssupport.EnsureString(jssupport.GetNamedArgument(args, "initiateIssuanceURI"))
	if err != nil {
		return nil, err
	}

	interaction, err := agentInstance.CreateOpenID4CIIssuerInitiatedInteraction(initiateIssuanceURI)
	if err != nil {
		return nil, err
	}

	return types.SerializeOpenID4CIIssuerInitiatedInteraction(&agentMethodsRunner, interaction), nil
}

func resolveDisplayData(_ js.Value, args []js.Value) (any, error) {
	if agentInstance == nil {
		return nil, walleterror.NewExecutionError(
			errors.Module,
			errors.InitializationFailedCode,
			errors.InitializationFailedError,
			fmt.Errorf("agent instance is not initialized"))
	}

	issuerURI, err := jssupport.EnsureString(jssupport.GetNamedArgument(args, "issuerURI"))
	if err != nil {
		return nil, err
	}

	credentials, err := jssupport.EnsureStringArray(jssupport.GetNamedArgument(args, "credentials"))
	if err != nil {
		return nil, err
	}

	dispData, err := agentInstance.ResolveDisplayData(issuerURI, credentials)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(dispData)
	if err != nil {
		return nil, err
	}

	return string(data), err
}

func parseResolvedDisplayData(_ js.Value, args []js.Value) (any, error) {
	if agentInstance == nil {
		return nil, walleterror.NewExecutionError(
			errors.Module,
			errors.InitializationFailedCode,
			errors.InitializationFailedError,
			fmt.Errorf("agent instance is not initialized"))
	}

	resolvedCredentialDisplayData, err :=
		jssupport.EnsureString(jssupport.GetNamedArgument(args, "resolvedCredentialDisplayData"))
	if err != nil {
		return nil, err
	}

	displayData, err := agentInstance.ParseResolvedDisplayData(resolvedCredentialDisplayData)
	if err != nil {
		walleterror.NewValidationError(
			errors.Module,
			errors.InvalidDisplayDataCode,
			errors.InvalidDisplayDataError,
			fmt.Errorf("display data parsing failed: %w", err))
	}

	var credentialDisplays []any

	for _, credDisp := range displayData.CredentialDisplays {
		var claims []any

		for _, claimDisp := range credDisp.Claims {
			claim := map[string]any{
				"rawValue":  claimDisp.RawValue,
				"valueType": claimDisp.ValueType,
				"label":     claimDisp.Label,
				"value":     claimDisp.Value,
			}

			if claimDisp.Order != nil {
				claim["order"] = *claimDisp.Order
			}

			claims = append(claims, claim)

		}

		credentialDisplays = append(credentialDisplays, map[string]any{
			"name":            credDisp.Overview.Name,
			"logo":            credDisp.Overview.Logo.URL,
			"backgroundColor": credDisp.Overview.BackgroundColor,
			"textColor":       credDisp.Overview.TextColor,
			"claims":          claims,
		})
	}

	return map[string]any{
		"issuerName":         displayData.IssuerDisplay.Name,
		"credentialDisplays": credentialDisplays,
	}, nil
}

func getCredentialID(_ js.Value, args []js.Value) (any, error) {
	credential, err := jssupport.EnsureString(jssupport.GetNamedArgument(args, "credential"))
	if err != nil {
		return nil, err
	}

	parsed, err := agentInstance.ParseCredential(credential)
	if err != nil {
		return nil, err
	}

	return parsed.ID, nil
}

func ExportAgentFunctions() map[string]any {
	return map[string]any{
		"initAgent": agentMethodsRunner.CreateAsyncFunc(initAgent),
		"createDID": agentMethodsRunner.CreateAsyncFunc(createDID),
		"createOpenID4CIIssuerInitiatedInteraction": agentMethodsRunner.CreateAsyncFunc(createOpenID4CIIssuerInitiatedInteraction),
		"resolveDisplayData":                        agentMethodsRunner.CreateAsyncFunc(resolveDisplayData),
		"getCredentialID":                           agentMethodsRunner.CreateAsyncFunc(getCredentialID),
		"parseResolvedDisplayData":                  agentMethodsRunner.CreateAsyncFunc(parseResolvedDisplayData),
	}
}
