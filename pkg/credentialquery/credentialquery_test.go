/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialquery_test

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/doc/presexch"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/internal/testutil"
	"github.com/trustbloc/wallet-sdk/pkg/credentialquery"
)

var (
	//go:embed test_data/presentation_definition.json
	presentationDefinition []byte

	//go:embed test_data/university_degree.jwt
	universityDegreeVC []byte

	//go:embed test_data/permanent_resident_card.jwt
	permanentResidentCardVC []byte
)

func TestInstance_Query(t *testing.T) {
	docLoader := testutil.DocumentLoader(t)
	pdQuery := &presexch.PresentationDefinition{}
	err := json.Unmarshal(presentationDefinition, pdQuery)
	require.NoError(t, err)

	contents := [][]byte{
		universityDegreeVC,
		permanentResidentCardVC,
	}

	var credentials []*verifiable.Credential

	for _, credContent := range contents {
		cred, credErr := verifiable.ParseCredential(credContent, verifiable.WithDisabledProofCheck(),
			verifiable.WithJSONLDDocumentLoader(docLoader))
		require.NoError(t, credErr)

		credentials = append(credentials, cred)
	}

	instance := credentialquery.NewInstance(docLoader)
	presentation, err := instance.Query(pdQuery, credentials)

	require.NoError(t, err)
	require.NotNil(t, presentation)

	require.Len(t, presentation.Credentials(), 1)
}
