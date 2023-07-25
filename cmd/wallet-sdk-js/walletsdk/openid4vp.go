/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package walletsdk

import (
	jsonld "github.com/piprate/json-gold/ld"

	"github.com/trustbloc/wallet-sdk/pkg/openid4vp"
)

// OpenID4VPInteraction wraps openid4vp.Interaction and necessary dependencies.
type OpenID4VPInteraction struct {
	Interaction *openid4vp.Interaction
	DocLoader   jsonld.DocumentLoader
}
