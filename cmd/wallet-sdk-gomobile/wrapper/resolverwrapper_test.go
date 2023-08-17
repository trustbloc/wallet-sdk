/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package wrapper //nolint:testpackage

import (
	_ "embed" //nolint:gci // required for go:embed
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed test_data/valid_doc_resolution.jsonld
var validDocResolution []byte

func TestGomobileVDRKeyResolverAdapter_Resolve(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		resolver := &VDRResolverWrapper{DIDResolver: &mocksDIDResolver{
			ResolveDocBytes: validDocResolution,
		}}

		doc, err := resolver.Resolve("did:example:123456")
		require.NoError(t, err)
		require.NotNil(t, doc)
	})

	t.Run("Resolve failed", func(t *testing.T) {
		resolver := &VDRResolverWrapper{DIDResolver: &mocksDIDResolver{
			ResolveErr: errors.New("resolve failed"),
		}}

		doc, err := resolver.Resolve("did:example:123456")
		require.Contains(t, err.Error(), "resolve failed")
		require.Nil(t, doc)
	})

	t.Run("Parse resolution failed", func(t *testing.T) {
		resolver := &VDRResolverWrapper{DIDResolver: &mocksDIDResolver{
			ResolveDocBytes: []byte("random string"),
		}}

		doc, err := resolver.Resolve("did:example:123456")
		require.Contains(t, err.Error(), "document resolution parsing failed")
		require.Nil(t, doc)
	})
}

type mocksDIDResolver struct {
	ResolveDocBytes []byte
	ResolveErr      error
}

func (m *mocksDIDResolver) Resolve(string) ([]byte, error) {
	return m.ResolveDocBytes, m.ResolveErr
}
