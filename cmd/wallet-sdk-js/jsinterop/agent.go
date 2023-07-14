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

func InitAgent(_ js.Value, args []js.Value) (any, error) {
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

func CreateDID(_ js.Value, args []js.Value) (any, error) {
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

func CreateOpenID4CIIssuerInitiatedInteraction(_ js.Value, args []js.Value) (any, error) {
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

func ResolveDisplayData(_ js.Value, args []js.Value) (any, error) {
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

func GetCredentialID(_ js.Value, args []js.Value) (any, error) {
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
		"initAgent": agentMethodsRunner.CreateAsyncFunc(InitAgent),
		"createDID": agentMethodsRunner.CreateAsyncFunc(CreateDID),
		"createOpenID4CIIssuerInitiatedInteraction": agentMethodsRunner.CreateAsyncFunc(CreateOpenID4CIIssuerInitiatedInteraction),
		"resolveDisplayData":                        agentMethodsRunner.CreateAsyncFunc(ResolveDisplayData),
		"getCredentialID":                           agentMethodsRunner.CreateAsyncFunc(GetCredentialID),
	}
}
