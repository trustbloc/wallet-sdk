//go:build js && wasm

/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package types

import (
	"syscall/js"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/vc-go/presexch"
	"github.com/trustbloc/vc-go/verifiable"
)

func TestSerializeMatchedSubmissionRequirement(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		req := &presexch.MatchedSubmissionRequirement{
			Name:    "",
			Purpose: "",
			Rule:    "",
			Count:   0,
			Min:     0,
			Max:     0,
			Descriptors: []*presexch.MatchedInputDescriptor{
				{
					ID:      "",
					Name:    "",
					Purpose: "",
					MatchedVCs: []*verifiable.Credential{
						{},
					},
				},
			},
			Nested: nil,
		}
		serialized, err := SerializeMatchedSubmissionRequirement(req)

		require.NoError(t, err)
		require.NotNil(t, serialized)

		js.ValueOf(serialized)
	})
}
