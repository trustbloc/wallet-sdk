/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package common implements common functionality like jwt sign and did public key resolve.
package common

import (
	diddoc "github.com/trustbloc/did-go/doc/did"
	vdrapi "github.com/trustbloc/did-go/vdr/api"
	"github.com/trustbloc/vc-go/vermethod"

	"github.com/trustbloc/wallet-sdk/pkg/api"
)

// NewVDRKeyResolver creates VDRKeyResolver.
func NewVDRKeyResolver(resolver api.DIDResolver) *vermethod.VDRResolver {
	return vermethod.NewVDRResolver(&didResolverWrapper{didResolver: resolver})
}

type didResolverWrapper struct {
	didResolver api.DIDResolver
}

func (d *didResolverWrapper) Resolve(did string, _ ...vdrapi.DIDMethodOption) (*diddoc.DocResolution, error) {
	return d.didResolver.Resolve(did)
}
