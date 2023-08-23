//go:build js && wasm

/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package types

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-js/jsinterop/jssupport"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-js/walletsdk"
)

const (
	openid4VPGetQuery            = "getQuery"
	openid4VPPresentCredential   = "presentCredential"
	openid4VPVerifierDisplayData = "verifierDisplayData"
)

func SerializeOpenID4VPInteraction(agentMethodsRunner *jssupport.AsyncRunner,
	interaction *walletsdk.OpenID4VPInteraction,
) map[string]interface{} {
	return map[string]interface{}{
		openid4VPGetQuery: agentMethodsRunner.CreateAsyncFunc(func(this js.Value, args []js.Value) (any, error) {
			presentationDefinition := interaction.Interaction.GetQuery()

			pdBytes, err := json.Marshal(presentationDefinition)
			if err != nil {
				return nil, fmt.Errorf("presentation definition marshal: %w", err)
			}

			return string(pdBytes), nil
		}),
		openid4VPPresentCredential: agentMethodsRunner.CreateAsyncFunc(func(this js.Value, args []js.Value) (any, error) {
			credentials, err := jssupport.EnsureStringArray(jssupport.GetNamedArgument(args, "credentials"))
			if err != nil {
				return nil, err
			}

			var parsedCreds []*verifiable.Credential

			for _, cred := range credentials {
				verifiableCredential, err := verifiable.ParseCredential(
					[]byte(cred),
					verifiable.WithJSONLDDocumentLoader(interaction.DocLoader),
					verifiable.WithDisabledProofCheck())
				if err != nil {
					return nil, fmt.Errorf("parse creds: %w", err)
				}

				parsedCreds = append(parsedCreds, verifiableCredential)
			}

			err = interaction.Interaction.PresentCredential(parsedCreds)
			if err != nil {
				return nil, fmt.Errorf("present credentials: %w", err)
			}

			return nil, nil
		}),

		openid4VPVerifierDisplayData: agentMethodsRunner.CreateAsyncFunc(func(this js.Value, args []js.Value) (any, error) {
			data := interaction.Interaction.VerifierDisplayData()
			return map[string]interface{}{
				"name":    data.Name,
				"did":     data.DID,
				"purpose": data.Purpose,
				"logoURI": data.LogoURI,
			}, nil
		}),
	}
}
