/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialschema

import (
	"fmt"

	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
)

func validateCredentialsSupported(credentialsSupported []issuer.SupportedCredential) error {
	allSupportedCredentialObjectIDs := make(map[string]struct{})

	for i := range credentialsSupported {
		_, duplicateIDFound := allSupportedCredentialObjectIDs[credentialsSupported[i].ID]
		if duplicateIDFound {
			return fmt.Errorf("the ID %s appears in multiple supported credential objects", credentialsSupported[i].ID)
		}

		allSupportedCredentialObjectIDs[credentialsSupported[i].ID] = struct{}{}
	}

	return nil
}
