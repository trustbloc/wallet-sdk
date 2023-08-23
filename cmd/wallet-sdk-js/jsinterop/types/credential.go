//go:build js && wasm

/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package types

import "github.com/trustbloc/vc-go/verifiable"

func SerializeCredential(cred *verifiable.Credential) (any, error) {
	marshalledVC, err := cred.MarshalJSON()
	if err != nil {
		return "", err
	}

	return string(marshalledVC), err
}

func SerializeCredentialArray(creds []*verifiable.Credential) ([]any, error) {
	var result []any
	for _, cred := range creds {
		marshalledVC, err := SerializeCredential(cred)
		if err != nil {
			return nil, err
		}

		result = append(result, marshalledVC)
	}

	return result, nil
}
