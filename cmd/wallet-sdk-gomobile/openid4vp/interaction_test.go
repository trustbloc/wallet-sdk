/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp //nolint: testpackage

import (
	"crypto/ed25519"
	"crypto/rand"
	_ "embed" //nolint:gci // required for go:embed
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hyperledger/aries-framework-go/component/models/did"
	"github.com/hyperledger/aries-framework-go/component/models/presexch"
	afgoverifiable "github.com/hyperledger/aries-framework-go/component/models/verifiable"
	"github.com/piprate/json-gold/ld"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
	"github.com/trustbloc/wallet-sdk/internal/testutil"
	"github.com/trustbloc/wallet-sdk/pkg/models"
	"github.com/trustbloc/wallet-sdk/pkg/openid4vp"
)

var (
	//go:embed test_data/request_object.jwt
	requestObjectJWT string

	//go:embed test_data/credentials.jsonld
	credentialsJSONLD []byte

	//go:embed test_data/valid_doc_resolution.jsonld
	sampleDIDDocResolution []byte
)

func TestNewInteraction(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("OpenTelemetry disabled, custom headers used instead", func(t *testing.T) {
			requiredArgs := NewArgs(
				requestObjectJWT,
				&mockCrypto{},
				&mockDIDResolver{ResolveDocBytes: sampleDIDDocResolution},
			)

			// Note: in-depth testing of opts functionality is done in the integration tests.
			opts := NewOpts()

			additionalHeaders := api.NewHeaders()

			additionalHeaders.Add(api.NewHeader("header-name-1", "header-value-1"))
			additionalHeaders.Add(api.NewHeader("header-name-2", "header-value-2"))

			opts.AddHeaders(additionalHeaders)

			opts.DisableOpenTelemetry()

			instance, err := NewInteraction(requiredArgs, opts)
			require.NoError(t, err)
			require.NotNil(t, instance)
		})
		t.Run("All other options invoked", func(t *testing.T) {
			requiredArgs := NewArgs(
				requestObjectJWT,
				&mockCrypto{},
				&mockDIDResolver{ResolveDocBytes: sampleDIDDocResolution},
			)

			// Note: in-depth testing of opts functionality is done in the integration tests.
			opts := NewOpts()
			opts.SetDocumentLoader(&documentLoaderWrapper{goAPIDocumentLoader: testutil.DocumentLoader(t)})
			opts.SetActivityLogger(nil)
			opts.SetMetricsLogger(nil)
			opts.DisableHTTPClientTLSVerify()
			opts.SetHTTPTimeoutNanoseconds(0)

			instance, err := NewInteraction(requiredArgs, opts)
			require.NoError(t, err)
			require.NotNil(t, instance)
		})
	})
	t.Run("Failure - invalid authorization request", func(t *testing.T) {
		requiredArgs := NewArgs(
			requestObjectJWT,
			&mockCrypto{},
			&mockDIDResolver{},
		)

		instance, err := NewInteraction(requiredArgs, nil)
		testutil.RequireErrorContains(t, err, "INVALID_AUTHORIZATION_REQUEST")
		testutil.RequireErrorContains(t, err, "verify request object: parse JWT: "+
			"parse JWT from compact JWS: resolve DID did:ion:EiDYWcDuP-EDjVyFWGFdpgPncar9A7OGFykdeX71ZTU-wg")
		require.Nil(t, instance)
	})
}

func TestOpenID4VP_PresentCredential(t *testing.T) {
	mockKey, _, e := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, e)

	credentials := verifiable.NewCredentialsArray()

	var credentialData []json.RawMessage

	e = json.Unmarshal(credentialsJSONLD, &credentialData)
	require.NoError(t, e)

	for _, credBytes := range credentialData {
		cred, err := afgoverifiable.ParseCredential(credBytes,
			afgoverifiable.WithDisabledProofCheck(), afgoverifiable.WithCredDisableValidation())
		require.NoError(t, err)

		credentials.Add(verifiable.NewCredential(cred))
	}

	singleCredential := credentials.AtIndex(0)

	makeInteraction := func() *Interaction {
		return &Interaction{
			crypto:           &mockCrypto{},
			ldDocumentLoader: &documentLoaderWrapper{goAPIDocumentLoader: testutil.DocumentLoader(t)},
			goAPIOpenID4VP: &mockGoAPIInteraction{
				PresentCredentialErr: nil,
			},
			didResolver: &mockDIDResolver{ResolveDocBytes: mockResolution(t, &api.VerificationMethod{
				ID:   "did:example:12345#testId",
				Type: "Ed25519VerificationKey2018",
				Key:  models.VerificationKey{Raw: mockKey},
			})},
		}
	}

	t.Run("Success", func(t *testing.T) {
		instance := makeInteraction()

		err := instance.PresentCredential(credentials)
		require.NoError(t, err)
	})

	t.Run("Success Unsafe", func(t *testing.T) {
		instance := makeInteraction()

		err := instance.PresentCredentialUnsafe(singleCredential)
		require.NoError(t, err)
	})

	t.Run("Present credentials failed", func(t *testing.T) {
		instance := makeInteraction()

		instance.goAPIOpenID4VP = &mockGoAPIInteraction{
			PresentCredentialErr: errors.New("present credentials failed"),
		}

		err := instance.PresentCredential(credentials)
		require.Contains(t, err.Error(), "present credentials failed")
	})

	t.Run("Present credentials unsafe failed", func(t *testing.T) {
		instance := makeInteraction()

		instance.goAPIOpenID4VP = &mockGoAPIInteraction{
			PresentCredentialUnsafeErr: errors.New("present credentials failed"),
		}

		err := instance.PresentCredentialUnsafe(singleCredential)
		require.Contains(t, err.Error(), "present credentials failed")
	})
}

func TestInteraction_VerifierDisplayData(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		instance := &Interaction{
			goAPIOpenID4VP: &mockGoAPIInteraction{
				VerifierDisplayDataRes: &openid4vp.VerifierDisplayData{
					DID:     "DID",
					Name:    "testName",
					Purpose: "purpose",
					LogoURI: "logoURI",
				},
			},
		}

		data := instance.VerifierDisplayData()
		require.NotNil(t, data)
		require.Equal(t, "DID", data.DID())
		require.Equal(t, "testName", data.Name())
		require.Equal(t, "purpose", data.Purpose())
		require.Equal(t, "logoURI", data.LogoURI())
	})
}

type documentLoaderWrapper struct {
	goAPIDocumentLoader ld.DocumentLoader
}

func (dl *documentLoaderWrapper) LoadDocument(u string) (*api.LDDocument, error) {
	ldDoc, err := dl.goAPIDocumentLoader.LoadDocument(u)
	if err != nil {
		return nil, err
	}

	docBytes, err := json.Marshal(ldDoc.Document)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ld document: %w", err)
	}

	return &api.LDDocument{
		DocumentURL: ldDoc.DocumentURL,
		Document:    string(docBytes),
		ContextURL:  ldDoc.ContextURL,
	}, nil
}

type mockCrypto struct {
	SignResult []byte
	SignErr    error
	VerifyErr  error
}

func (c *mockCrypto) Sign(_ []byte, _ string) ([]byte, error) {
	return c.SignResult, c.SignErr
}

func (c *mockCrypto) Verify(signature, msg []byte, keyID string) error {
	return c.VerifyErr
}

type mockGoAPIInteraction struct {
	GetQueryResult             *presexch.PresentationDefinition
	PresentCredentialErr       error
	PresentCredentialUnsafeErr error
	VerifierDisplayDataRes     *openid4vp.VerifierDisplayData
}

func (o *mockGoAPIInteraction) GetQuery() *presexch.PresentationDefinition {
	return o.GetQueryResult
}

func (o *mockGoAPIInteraction) PresentCredential([]*afgoverifiable.Credential) error {
	return o.PresentCredentialErr
}

func (o *mockGoAPIInteraction) PresentCredentialUnsafe(*afgoverifiable.Credential) error {
	return o.PresentCredentialUnsafeErr
}

func (o *mockGoAPIInteraction) VerifierDisplayData() *openid4vp.VerifierDisplayData {
	return o.VerifierDisplayDataRes
}

type mockDIDResolver struct {
	ResolveDocBytes []byte
	ResolveErr      error
}

func (m *mockDIDResolver) Resolve(string) ([]byte, error) {
	return m.ResolveDocBytes, m.ResolveErr
}

func mockResolution(t *testing.T, vm *api.VerificationMethod) []byte {
	t.Helper()

	var (
		mockVM   *did.VerificationMethod
		err      error
		mockDID  string
		mockVMID string
	)

	idSplit := strings.Split(vm.ID, "#")
	switch len(idSplit) {
	case 1:
		mockVMID = idSplit[0]
	case 2:
		mockDID = idSplit[0]
		mockVMID = idSplit[1]
	default:
		t.Fail()
	}

	if vm.Key.JSONWebKey != nil {
		mockVM, err = did.NewVerificationMethodFromJWK(mockVMID, vm.Type, mockDID, vm.Key.JSONWebKey)
		require.NoError(t, err)
	} else {
		mockVM = did.NewVerificationMethodFromBytes(mockVMID, vm.Type, mockDID, vm.Key.Raw)
	}

	docRes := &did.DocResolution{
		DIDDocument: &did.Doc{
			ID:      mockDID,
			Context: []string{did.ContextV1},
			VerificationMethod: []did.VerificationMethod{
				*mockVM,
			},
			AssertionMethod: []did.Verification{
				{
					VerificationMethod: *mockVM,
				},
			},
		},
	}

	docBytes, err := docRes.JSONBytes()
	require.NoError(t, err)

	return docBytes
}
