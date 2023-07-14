//go:build js && wasm

/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package types

import (
	"fmt"
	"syscall/js"

	diddoc "github.com/hyperledger/aries-framework-go/component/models/did"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-js/jsinterop/jssupport"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-js/walletsdk"
	"github.com/trustbloc/wallet-sdk/pkg/models"
)

const (
	openid4ciRequestCredentialWithPreAuth = "requestCredentialWithPreAuth"
	openid4ciPreAuthorizedCodeGrantParams = "preAuthorizedCodeGrantParams"
	openid4ciIssuerURI                    = "issuerURI"
)

func SerializeOpenID4CIIssuerInitiatedInteraction(agentMethodsRunner *jssupport.AsyncRunner,
	interaction *walletsdk.OpenID4CIIssuerInitiatedInteraction,
) map[string]interface{} {
	return map[string]interface{}{
		openid4ciIssuerURI: js.FuncOf(func(this js.Value, args []js.Value) any {
			return interaction.Interaction.IssuerURI()
		}),
		openid4ciPreAuthorizedCodeGrantParams: agentMethodsRunner.CreateAsyncFunc(func(this js.Value, args []js.Value) (any, error) {
			params, err := interaction.Interaction.PreAuthorizedCodeGrantParams()
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"userPINRequired": params.PINRequired(),
			}, nil
		}),
		openid4ciRequestCredentialWithPreAuth: agentMethodsRunner.CreateAsyncFunc(
			func(this js.Value, args []js.Value) (any, error) {
				pin, err := jssupport.EnsureString(jssupport.GetNamedArgument(args, "pin"))
				if err != nil {
					return nil, err
				}

				doc, err := jssupport.GetNamedArgument(args, "didDoc")
				if err != nil {
					return nil, err
				}

				docResolution, err := DeserializeDIDDoc(doc.Value)
				if err != nil {
					return nil, err
				}

				// look for assertion method
				verificationMethods := docResolution.DIDDocument.VerificationMethods(diddoc.AssertionMethod)

				if len(verificationMethods[diddoc.AssertionMethod]) == 0 {
					return nil, fmt.Errorf("DID provided has no assertion method to use as a default signing key")
				}
				vm := verificationMethods[diddoc.AssertionMethod][0].VerificationMethod

				creds, err := interaction.RequestCredentialWithPreAuth(models.VerificationMethodFromDoc(&vm), pin)
				if err != nil {
					return nil, err
				}

				marshaledCreds, err := SerializeCredentialArray(creds)
				if err != nil {
					return nil, err
				}

				return marshaledCreds, err
			}),
	}
}
