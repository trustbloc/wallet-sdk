/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credential_test

import (
	_ "embed"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/credential"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
)

var (
	//go:embed test_data/credential_university_degree.jsonld
	universityDegreeVC string
	//go:embed test_data/credential_drivers_license.jsonld
	driversLicenseDegreeVC string
)

func TestProvider(t *testing.T) {
	provider := credential.NewInMemoryDB()

	const universityDegreeVCID = "http://example.edu/credentials/1872"

	opts := verifiable.NewOpts()
	opts.DisableProofCheck()

	universityDegreeVerifiableCredential, err := verifiable.ParseCredential(universityDegreeVC, opts)
	require.NoError(t, err)

	// Store two VCs.
	err = provider.Add(universityDegreeVerifiableCredential)
	require.NoError(t, err)

	driversLicenseVerifiableCredential, err := verifiable.ParseCredential(driversLicenseDegreeVC, opts)
	require.NoError(t, err)

	err = provider.Add(driversLicenseVerifiableCredential)
	require.NoError(t, err)

	// Get each VC individually.
	retrievedVC, err := provider.Get(universityDegreeVCID)
	require.NoError(t, err)
	require.NotNil(t, retrievedVC)

	retrievedVC, err = provider.Get("https://eu.com/claims/DriversLicense")
	require.NoError(t, err)
	require.NotNil(t, retrievedVC)

	// Retrieve both VCs in one call.
	retrievedVCs, err := provider.GetAll()
	require.NoError(t, err)

	require.Equal(t, 2, retrievedVCs.Length())
	require.NotNil(t, retrievedVCs.AtIndex(0))
	require.NotNil(t, retrievedVCs.AtIndex(1))

	// Remove one of the VCs and verify that it's deleted.
	err = provider.Remove(universityDegreeVCID)
	require.NoError(t, err)

	retrievedVC, err = provider.Get(universityDegreeVCID)
	require.EqualError(t, err, fmt.Sprintf("no credential with an id of %s was found", universityDegreeVCID))
	require.Nil(t, retrievedVC)
}

func TestProvider_Add_Failure_Nil_VC(t *testing.T) {
	provider := credential.NewInMemoryDB()

	err := provider.Add(verifiable.NewCredential(nil))
	require.EqualError(t, err, "VC cannot be nil")
}
