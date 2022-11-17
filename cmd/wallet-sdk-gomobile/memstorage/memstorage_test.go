/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package memstorage_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/memstorage"
)

var (
	//go:embed testdata/credential_university_degree.jsonld
	universityDegreeVC []byte
	//go:embed testdata/credential_drivers_license.jsonld
	driversLicenseDegreeVC []byte
)

func TestProvider(t *testing.T) {
	provider := memstorage.NewProvider()

	const universityDegreeVCID = "http://example.edu/credentials/1872"

	// Store two VCs.
	err := provider.Add(&api.JSONObject{Data: universityDegreeVC})
	require.NoError(t, err)

	err = provider.Add(&api.JSONObject{Data: driversLicenseDegreeVC})
	require.NoError(t, err)

	// Get each VC individually.
	retrievedVC, err := provider.Get(universityDegreeVCID)
	require.NoError(t, err)
	require.NotNil(t, retrievedVC)

	retrievedVC, err = provider.Get("https://eu.com/claims/DriversLicense")
	require.NoError(t, err)
	require.NotNil(t, retrievedVC)

	// Retrieve both VCs in one call.
	retrievedVCsJSONArray, err := provider.GetAll()
	require.NoError(t, err)

	var retrievedVCs []interface{}

	err = json.Unmarshal(retrievedVCsJSONArray.Data, &retrievedVCs)
	require.NoError(t, err)

	require.Len(t, retrievedVCs, 2)
	require.NotNil(t, retrievedVCs[0])
	require.NotNil(t, retrievedVCs[1])

	// Remove one of the VCs and verify that it's deleted.
	err = provider.Remove(universityDegreeVCID)
	require.NoError(t, err)

	retrievedVC, err = provider.Get(universityDegreeVCID)
	require.EqualError(t, err, fmt.Sprintf("no credential with an id of %s was found", universityDegreeVCID))
	require.Nil(t, retrievedVC)
}

func TestProvider_Add_Failure_Empty_JSON(t *testing.T) {
	provider := memstorage.NewProvider()

	err := provider.Add(&api.JSONObject{})
	require.EqualError(t, err, "unmarshal new credential: unexpected end of JSON input")
}
