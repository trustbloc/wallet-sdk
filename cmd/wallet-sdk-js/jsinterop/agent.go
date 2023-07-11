//go:build js && wasm

/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package jsinterop implements interop between Wallet-SDK and JS.
package jsinterop

import (
	"fmt"
	"syscall/js"

	"github.com/hyperledger/aries-framework-go/component/kmscrypto/kms"
	"github.com/hyperledger/aries-framework-go/component/storage/indexeddb"
	arieskms "github.com/hyperledger/aries-framework-go/spi/kms"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-js/jsinterop/errors"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-js/jsinterop/jssupport"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-js/jsinterop/types"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-js/walletsdk"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

const (
	dbNamespace = "wallet-sdk"
)

var agentInstance *walletsdk.Agent
var agentMethodsRunner jssupport.AsyncRunner

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

	indexedDBKMSProvider, err := indexeddb.NewProvider(dbNamespace)
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

func CreateOpenID4CIInteraction(_ js.Value, args []js.Value) (any, error) {
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

	interaction, err := agentInstance.CreateOpenID4CIInteraction(initiateIssuanceURI)
	if err != nil {
		return nil, err
	}

	return types.SerializeOpenID4CIInteraction(&agentMethodsRunner, interaction), nil
}

func ExportAgentFunctions() map[string]js.Func {
	return map[string]js.Func{
		"initAgent":                  agentMethodsRunner.CreateAsyncFunc(InitAgent),
		"createDID":                  agentMethodsRunner.CreateAsyncFunc(CreateDID),
		"createOpenID4CIInteraction": agentMethodsRunner.CreateAsyncFunc(CreateOpenID4CIInteraction),
	}

}
