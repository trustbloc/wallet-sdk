/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credential_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/piprate/json-gold/ld"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/credential"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
	"github.com/trustbloc/wallet-sdk/internal/testutil"
)

var (
	//go:embed test_data/multi_inputs_pd.json
	multiInputPD []byte

	//go:embed test_data/nested_submission_requirements_pd.json
	nestedRequirementsPD []byte

	//go:embed test_data/university_degree.jwt
	universityDegreeVCJWT []byte

	//go:embed test_data/permanent_resident_card.jwt
	permanentResidentCardVC []byte

	//go:embed test_data/drivers_license.jwt
	driverLicenseVC []byte

	//go:embed test_data/verified_employee.jwt
	verifiedEmployeeVC []byte

	//go:embed test_data/citizenship_pd.json
	citizenshipPD []byte

	//go:embed test_data/citizenship_vc.json
	citizenshipVC []byte
)

func TestNewInquirer(t *testing.T) {
	t.Run("Using the default network-based document loader", func(t *testing.T) {
		opts := credential.NewInquirerOpts().SetHTTPTimeoutNanoseconds(0)

		inquirer, err := credential.NewInquirer(opts)
		require.NoError(t, err)
		require.NotNil(t, inquirer)
	})

	t.Run("Default opts", func(t *testing.T) {
		inquirer, err := credential.NewInquirer(nil)
		require.NoError(t, err)
		require.NotNil(t, inquirer)
	})
}

func TestInstance_GetSubmissionRequirements(t *testing.T) {
	contents := [][]byte{
		universityDegreeVCJWT,
		permanentResidentCardVC,
		driverLicenseVC,
		verifiedEmployeeVC,
	}

	opts := credential.NewInquirerOpts()

	documentLoader := &documentLoaderReverseWrapper{
		DocumentLoader: testutil.DocumentLoader(t),
	}

	opts.SetDocumentLoader(documentLoader)
	opts.SetDIDResolver(&mocksDIDResolver{})

	t.Run("Success", func(t *testing.T) {
		query, err := credential.NewInquirer(opts)
		require.NoError(t, err)

		requirements, err := query.GetSubmissionRequirements(multiInputPD, createCredJSONArray(t, contents))

		require.NoError(t, err)
		require.Equal(t, requirements.Len(), 1)
		req1 := requirements.AtIndex(0)
		require.Equal(t, req1.DescriptorLen(), 3)
		require.Equal(t, req1.Name(), "Information")
		require.Equal(t, req1.Purpose(), "test purpose")
		require.Equal(t, req1.Rule(), "pick")

		require.Equal(t, req1.Count(), 1)
		require.Equal(t, req1.Min(), 0)
		require.Equal(t, req1.Max(), 0)
		require.Equal(t, req1.NestedRequirementLength(), 0)

		desc1 := req1.DescriptorAtIndex(0)

		require.Equal(t, desc1.ID, "VerifiedEmployee")
		require.Equal(t, desc1.Name, "Verified Employee")
		require.Equal(t, desc1.Purpose, "test purpose")
		require.Equal(t, desc1.MatchedVCs.Length(), 1)
	})

	t.Run("Success nested requirements", func(t *testing.T) {
		query, err := credential.NewInquirer(opts)
		require.NoError(t, err)

		requirements, err := query.GetSubmissionRequirements(nestedRequirementsPD, createCredJSONArray(t, contents))

		require.NoError(t, err)
		require.Equal(t, requirements.Len(), 1)
		req1 := requirements.AtIndex(0)
		require.Equal(t, req1.DescriptorLen(), 0)
		require.Equal(t, req1.Name(), "Nested requirements")
		require.Equal(t, req1.Rule(), "all")

		require.Equal(t, req1.Count(), 2)
		require.Equal(t, req1.Min(), 0)
		require.Equal(t, req1.Max(), 0)
		require.Equal(t, req1.NestedRequirementLength(), 2)

		nestedReq1 := req1.NestedRequirementAtIndex(0)

		require.Equal(t, nestedReq1.DescriptorLen(), 2)

		desc1 := nestedReq1.DescriptorAtIndex(0)

		require.Equal(t, desc1.ID, "VerifiedEmployee")
		require.Equal(t, desc1.Name, "Verified Employee")
		require.Equal(t, desc1.MatchedVCs.Length(), 1)
	})

	t.Run("PD parse failed", func(t *testing.T) {
		query, err := credential.NewInquirer(opts)
		require.NoError(t, err)

		_, err = query.GetSubmissionRequirements(nil,
			createCredJSONArray(t, [][]byte{universityDegreeVCJWT, permanentResidentCardVC}),
		)

		require.Contains(t, err.Error(), "unmarshal of presentation definition failed:")
	})

	t.Run("PD validation failed", func(t *testing.T) {
		query, err := credential.NewInquirer(opts)
		require.NoError(t, err)

		_, err = query.GetSubmissionRequirements([]byte("{}"),
			createCredJSONArray(t, [][]byte{universityDegreeVCJWT, permanentResidentCardVC}),
		)

		require.Contains(t, err.Error(), "validation of presentation definition failed:")
	})

	t.Run("Nil credentials", func(t *testing.T) {
		query, err := credential.NewInquirer(opts)
		require.NoError(t, err)

		submissionRequirements, err := query.GetSubmissionRequirements(nil, nil)

		require.EqualError(t, err, "credentials must be provided")
		require.Nil(t, submissionRequirements)
	})
}

func TestInstance_GetSubmissionRequirementsCitizenship(t *testing.T) {
	contents := [][]byte{
		citizenshipVC,
	}

	opts := credential.NewInquirerOpts()

	documentLoader := &documentLoaderReverseWrapper{
		DocumentLoader: testutil.DocumentLoader(t),
	}

	opts.SetDocumentLoader(documentLoader)
	opts.SetDIDResolver(&mocksDIDResolver{})

	t.Run("Success", func(t *testing.T) {
		query, err := credential.NewInquirer(opts)
		require.NoError(t, err)

		requirements, err := query.GetSubmissionRequirements(citizenshipPD, createCredJSONArray(t, contents))

		require.NoError(t, err)
		require.Equal(t, requirements.Len(), 1)
		req1 := requirements.AtIndex(0)
		require.Equal(t, req1.DescriptorLen(), 1)

		require.Equal(t, req1.Count(), 1)
		require.Equal(t, req1.Min(), 0)
		require.Equal(t, req1.Max(), 0)
		require.Equal(t, req1.NestedRequirementLength(), 0)

		desc1 := req1.DescriptorAtIndex(0)

		require.Equal(t, desc1.MatchedVCs.Length(), 1)
	})
}

func createCredJSONArray(t *testing.T, creds [][]byte) *verifiable.CredentialsArray {
	t.Helper()

	credsArray := verifiable.NewCredentialsArray()

	for _, credContent := range creds {
		opts := verifiable.NewOpts()
		opts.DisableProofCheck()

		vc, err := verifiable.ParseCredential(string(credContent), opts)
		require.NoError(t, err)

		credsArray.Add(vc)
	}

	return credsArray
}

type documentLoaderReverseWrapper struct {
	DocumentLoader ld.DocumentLoader
}

func (l *documentLoaderReverseWrapper) LoadDocument(u string) (*api.LDDocument, error) {
	doc, err := l.DocumentLoader.LoadDocument(u)
	if err != nil {
		return nil, err
	}

	documentBytes, err := json.Marshal(doc.Document)
	if err != nil {
		return nil, fmt.Errorf("fail to unmarshal ld document bytes: %w", err)
	}

	wrappedDoc := &api.LDDocument{
		DocumentURL: doc.DocumentURL,
		Document:    string(documentBytes),
		ContextURL:  doc.ContextURL,
	}

	return wrappedDoc, nil
}

type mocksDIDResolver struct {
	ResolveDocBytes []byte
	ResolveErr      error
}

func (m *mocksDIDResolver) Resolve(string) ([]byte, error) {
	return m.ResolveDocBytes, m.ResolveErr
}
