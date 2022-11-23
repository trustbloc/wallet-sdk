/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialquery_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/piprate/json-gold/ld"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/credentialquery"
	"github.com/trustbloc/wallet-sdk/internal/testutil"
)

var (
	//go:embed test_data/presentation_definition.json
	presentationDefinition []byte

	//go:embed test_data/university_degree.jwt
	universityDegreeVC string

	//go:embed test_data/permanent_resident_card.jwt
	permanentResidentCardVC string
)

func TestInstance_Query(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		query := credentialquery.NewQuery(&documentLoaderReverseWrapper{
			DocumentLoader: testutil.DocumentLoader(t),
		})

		presentation, err := query.Query(presentationDefinition,
			createCredJSONArray(t, []string{universityDegreeVC, permanentResidentCardVC}),
		)
		require.NoError(t, err)
		require.NotNil(t, presentation)
	})

	t.Run("No matched credential", func(t *testing.T) {
		query := credentialquery.NewQuery(&documentLoaderReverseWrapper{
			DocumentLoader: testutil.DocumentLoader(t),
		})

		_, err := query.Query(presentationDefinition,
			createCredJSONArray(t, []string{permanentResidentCardVC}),
		)
		require.Contains(t, err.Error(), "credentials do not satisfy requirements")
	})

	t.Run("PD parse failed", func(t *testing.T) {
		query := credentialquery.NewQuery(&documentLoaderReverseWrapper{
			DocumentLoader: testutil.DocumentLoader(t),
		})

		_, err := query.Query(nil,
			createCredJSONArray(t, []string{universityDegreeVC, permanentResidentCardVC}),
		)

		require.Contains(t, err.Error(), "unmarshal of presentation definition failed:")
	})

	t.Run("PD validation failed", func(t *testing.T) {
		query := credentialquery.NewQuery(&documentLoaderReverseWrapper{
			DocumentLoader: testutil.DocumentLoader(t),
		})

		_, err := query.Query([]byte("{}"),
			createCredJSONArray(t, []string{universityDegreeVC, permanentResidentCardVC}),
		)

		require.Contains(t, err.Error(), "validation of presentation definition failed:")
	})

	t.Run("PD parse failed", func(t *testing.T) {
		query := credentialquery.NewQuery(&documentLoaderReverseWrapper{
			DocumentLoader: testutil.DocumentLoader(t),
		})

		_, err := query.Query(presentationDefinition,
			&credentialquery.Credentials{VCs: &api.JSONArray{}},
		)

		require.Contains(t, err.Error(), "unmarshal of credentials array failed, should be json array of jwt strings")
	})

	t.Run("VC parse failed", func(t *testing.T) {
		query := credentialquery.NewQuery(&documentLoaderReverseWrapper{
			DocumentLoader: testutil.DocumentLoader(t),
		})

		_, err := query.Query(presentationDefinition,
			createCredJSONArray(t, []string{"{}"}),
		)

		require.Contains(t, err.Error(), "verifiable credential parse failed")
	})

	t.Run("Nil credentials and nil reader", func(t *testing.T) {
		query := credentialquery.NewQuery(&documentLoaderReverseWrapper{
			DocumentLoader: testutil.DocumentLoader(t),
		})

		_, err := query.Query(presentationDefinition, &credentialquery.Credentials{})

		require.Contains(t, err.Error(), "either credential reader or vc array should be set")
	})
}

func createCredJSONArray(t *testing.T, creds []string) *credentialquery.Credentials {
	t.Helper()

	arr, err := json.Marshal(creds)
	require.NoError(t, err)

	return &credentialquery.Credentials{
		VCs: &api.JSONArray{Data: arr},
	}
}

type documentLoaderReverseWrapper struct {
	DocumentLoader ld.DocumentLoader
}

func (l *documentLoaderReverseWrapper) LoadDocument(u string) (*api.LDDocument, error) {
	doc, err := l.DocumentLoader.LoadDocument(u)
	if err != nil {
		return nil, err
	}

	wrappedDoc := &api.LDDocument{
		DocumentURL: doc.DocumentURL,
		ContextURL:  doc.ContextURL,
	}

	wrappedDoc.Document, err = json.Marshal(doc.Document)
	if err != nil {
		return nil, fmt.Errorf("fail to unmarshal ld document bytes: %w", err)
	}

	return wrappedDoc, nil
}
