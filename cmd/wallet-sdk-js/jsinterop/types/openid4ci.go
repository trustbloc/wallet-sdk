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

	diddoc "github.com/trustbloc/vc-go/did"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-js/jsinterop/jssupport"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-js/util"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-js/walletsdk"
	"github.com/trustbloc/wallet-sdk/pkg/models"
)

const (
	openid4ciRequestCredentialWithPreAuth = "requestCredentialWithPreAuth"
	openid4ciPreAuthorizedCodeGrantParams = "preAuthorizedCodeGrantParams"
	openid4ciIssuerURI                    = "issuerURI"
	openid4ciIssuerMetadata               = "issuerMetadata"
)

func SerializeOpenID4CIIssuerInitiatedInteraction(agentMethodsRunner *jssupport.AsyncRunner,
	interaction *walletsdk.OpenID4CIIssuerInitiatedInteraction,
) map[string]interface{} {
	return map[string]interface{}{
		openid4ciIssuerURI: js.FuncOf(func(this js.Value, args []js.Value) any {
			return interaction.Interaction.IssuerURI()
		}),
		openid4ciIssuerMetadata: agentMethodsRunner.CreateAsyncFunc(func(this js.Value, args []js.Value) (any, error) {
			issuerMetadata, err := interaction.Interaction.IssuerMetadata()
			if err != nil {
				return nil, err
			}
			supportedCredentialsBytes, err := json.Marshal(issuerMetadata.CredentialsSupported)
			if err != nil {
				return nil, err
			}
			localizedIssuerDisplaysBytes, err := json.Marshal(issuerMetadata.LocalizedIssuerDisplays)
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"credentialIssuer":        issuerMetadata.CredentialIssuer,
				"supportedCredentials":    string(supportedCredentialsBytes),
				"localizedIssuerDisplays": string(localizedIssuerDisplaysBytes),
			}, nil
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

				marshaledCreds, err := util.MapTo(creds, SerializeCredential)
				if err != nil {
					return nil, err
				}

				return marshaledCreds, err
			}),
	}
}
