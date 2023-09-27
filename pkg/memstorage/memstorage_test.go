/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package memstorage_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/pkg/memstorage"
)

func TestProvider(t *testing.T) {
	provider := memstorage.NewProvider()

	vcToStore1, err := verifiable.CreateCredential(verifiable.CredentialContents{
		ID: "VC1",
	}, nil)
	require.NoError(t, err)

	vcToStore2, err := verifiable.CreateCredential(verifiable.CredentialContents{
		ID: "VC2",
	}, nil)
	require.NoError(t, err)

	// Store two VCs.
	err = provider.Add(vcToStore1)
	require.NoError(t, err)

	err = provider.Add(vcToStore2)
	require.NoError(t, err)

	// Get each VC individually.
	retrievedVC, err := provider.Get(vcToStore1.Contents().ID)
	require.NoError(t, err)
	require.NotNil(t, retrievedVC)
	require.Equal(t, vcToStore1.Contents().ID, retrievedVC.Contents().ID)

	retrievedVC, err = provider.Get(vcToStore2.Contents().ID)
	require.NoError(t, err)
	require.NotNil(t, retrievedVC)
	require.Equal(t, vcToStore2.Contents().ID, retrievedVC.Contents().ID)

	// Retrieve both VCs in one call.
	retrievedVCs, err := provider.GetAll()
	require.NoError(t, err)
	require.Len(t, retrievedVCs, 2)

	gotExpectedVCsOrder1 := vcToStore1.Contents().ID == retrievedVCs[0].Contents().ID &&
		vcToStore2.Contents().ID == retrievedVCs[1].Contents().ID
	gotExpectedVCsOrder2 := vcToStore1.Contents().ID == retrievedVCs[1].Contents().ID &&
		vcToStore2.Contents().ID == retrievedVCs[0].Contents().ID

	require.True(t, gotExpectedVCsOrder1 || gotExpectedVCsOrder2)

	// Remove one of the VCs and verify that it's deleted.
	err = provider.Remove(vcToStore1.Contents().ID)
	require.NoError(t, err)

	retrievedVC, err = provider.Get(vcToStore1.Contents().ID)
	require.EqualError(t, err, fmt.Sprintf("no credential with an id of %s was found", vcToStore1.Contents().ID))
	require.Nil(t, retrievedVC)

	// Attempt to store a nil VC
	err = provider.Add(nil)
	require.EqualError(t, err, "VC cannot be nil")
}
