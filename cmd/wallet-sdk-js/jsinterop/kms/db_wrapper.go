//go:build js && wasm

/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package kms

import (
	"encoding/base64"
	"errors"
	"fmt"
	"syscall/js"
	"time"

	"github.com/trustbloc/kms-crypto-go/kms"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-js/jsinterop/jssupport"
)

type DBWrapper struct {
	dbObject js.Value
}

func WrapJSDB(dbObject js.Value) (*DBWrapper, error) {
	err := jssupport.EnsureMemberFunction(dbObject, "put")
	if err != nil {
		return nil, err
	}

	err = jssupport.EnsureMemberFunction(dbObject, "get")
	if err != nil {
		return nil, err
	}

	err = jssupport.EnsureMemberFunction(dbObject, "delete")
	if err != nil {
		return nil, err
	}

	return &DBWrapper{
		dbObject: dbObject,
	}, nil
}

func (w *DBWrapper) Put(keysetID string, keyData []byte) error {
	_, err := getResult(w.dbObject.Call("put", keysetID, base64.StdEncoding.EncodeToString(keyData)))
	if err != nil {
		return fmt.Errorf("put failed: %w", err)
	}

	return nil
}
func (w *DBWrapper) Get(keysetID string) ([]byte, error) {
	data, err := getResult(w.dbObject.Call("get", keysetID))
	if err != nil {
		return nil, fmt.Errorf("get failed: %w", err)
	}

	if !data.Truthy() {
		return nil, kms.ErrKeyNotFound
	}

	decoded, err := base64.StdEncoding.DecodeString(data.String())
	if err != nil {
		return nil, fmt.Errorf("data decoding failed: %w", err)
	}
	return decoded, err
}
func (w *DBWrapper) Delete(keysetID string) error {
	_, err := getResult(w.dbObject.Call("delete", keysetID))
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	return nil
}

func getResult(promise js.Value) (*js.Value, error) {
	onsuccess := make(chan js.Value)
	onerror := make(chan js.Value)

	const timeout = 10

	promise.Call("then", js.FuncOf(func(this js.Value, inputs []js.Value) interface{} {
		onsuccess <- inputs[0]
		return nil
	})).Set("catch", js.FuncOf(func(this js.Value, inputs []js.Value) interface{} {
		onerror <- inputs[0]
		return nil
	}))
	select {
	case value := <-onsuccess:
		return &value, nil
	case value := <-onerror:
		return nil, fmt.Errorf("%s %s", value.Get("name").String(),
			value.Get("message").String())
	case <-time.After(timeout * time.Second):
		return nil, errors.New("timeout waiting for eve")
	}
}
