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

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-js/walletsdk"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

const (
	dbNamespace = "wallet-sdk"
)

var agentInstance *walletsdk.Agent
var agentMethodsRunner AsyncRunner

func InitAgent(_ js.Value, args []js.Value) (any, error) {
	didResolverURI, err := ensureString(getOptionalNamedArgument(args, "didResolverURI"))
	if err != nil {
		return nil, err
	}

	if agentInstance != nil {
		return nil, walleterror.NewExecutionError(
			Module,
			InitializationFailedCode,
			InitializationFailedError,
			fmt.Errorf("agent instance already initialized"))
	}

	indexedDBKMSProvider, err := indexeddb.NewProvider(dbNamespace)
	if err != nil {
		return nil, walleterror.NewExecutionError(
			Module,
			InitializationFailedCode,
			InitializationFailedError,
			fmt.Errorf("failed to create IndexedDB provider: %w", err))
	}

	kmsStore, err := kms.NewAriesProviderWrapper(indexedDBKMSProvider)
	if err != nil {
		return nil, walleterror.NewExecutionError(
			Module,
			InitializationFailedCode,
			InitializationFailedError,
			fmt.Errorf("failed to create Aries KMS store wrapper %w", err))
	}

	agentInstance, err = walletsdk.NewAgent(didResolverURI, kmsStore)
	if err != nil {
		return nil, walleterror.NewExecutionError(
			Module,
			InitializationFailedCode,
			InitializationFailedError,
			fmt.Errorf("failed to create agant %w", err))
	}

	return nil, nil
}

func CreateDID(_ js.Value, args []js.Value) (any, error) {
	if agentInstance == nil {
		return nil, walleterror.NewExecutionError(
			Module,
			InitializationFailedCode,
			InitializationFailedError,
			fmt.Errorf("agent instance is not initialized"))
	}

	didMethod, err := ensureString(getNamedArgument(args, "didMethod"))
	if err != nil {
		return nil, err
	}

	keyType, err := ensureString(getOptionalNamedArgument(args, "keyType"))
	if err != nil {
		return nil, err
	}

	verificationType, err := ensureString(getOptionalNamedArgument(args, "verificationType"))
	if err != nil {
		return nil, err
	}

	didDoc, err := agentInstance.CreateDID(didMethod, arieskms.KeyType(keyType), verificationType)
	if err != nil {
		return nil, err
	}

	content, err := didDoc.JSONBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize did:%w", err)
	}

	return map[string]interface{}{
		"id":      didDoc.DIDDocument.ID,
		"content": content,
	}, nil
}

func ExportAgentFunctions() map[string]js.Func {
	return map[string]js.Func{
		"initAgent": agentMethodsRunner.CreateAsyncFunc(InitAgent),
		"createDID": agentMethodsRunner.CreateAsyncFunc(CreateDID),
	}

}
