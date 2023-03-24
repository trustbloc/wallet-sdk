/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credential

import "github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"

// VerifiablePresentation typed wrapper around go implementation of verifiable presentation.
type VerifiablePresentation struct {
	wrapped *verifiable.Presentation
}

// wrapVerifiablePresentation wraps go implementation of verifiable presentation into gomobile compatible struct.
func wrapVerifiablePresentation(vp *verifiable.Presentation) *VerifiablePresentation {
	return &VerifiablePresentation{wrapped: vp}
}

// Content return marshaled representation of verifiable presentation.
func (vp *VerifiablePresentation) Content() ([]byte, error) {
	return vp.wrapped.MarshalJSON()
}
