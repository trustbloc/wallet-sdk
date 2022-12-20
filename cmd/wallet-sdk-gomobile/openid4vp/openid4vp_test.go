/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp //nolint: testpackage

import (
	_ "embed" //nolint:gci // required for go:embed
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/doc/presexch"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/piprate/json-gold/ld"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/internal/testutil"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
)

var (
	//go:embed test_data/request_object.jwt
	requestObjectJWT string

	//go:embed test_data/presentation.jsonld
	presentationJSONLD []byte
)

func TestOpenID4VP_GetQuery(t *testing.T) {
	t.Run("NewInteraction success", func(t *testing.T) {
		instance := NewInteraction(
			requestObjectJWT,
			&mockKeyHandleReader{},
			&mockCrypto{},
			&mocksDIDResolver{},
			&documentLoaderWrapper{goAPIDocumentLoader: testutil.DocumentLoader(t)})

		require.NotNil(t, instance)
		require.NotNil(t, instance.crypto)
		require.NotNil(t, instance.ldDocumentLoader)
		require.NotNil(t, instance.keyHandleReader)
		require.NotNil(t, instance.goAPIOpenID4VP)
	})

	t.Run("GetQuery success", func(t *testing.T) {
		instance := &Interaction{
			keyHandleReader: &mockKeyHandleReader{},
			goAPIOpenID4VP: &mocGoAPIInteraction{
				GetQueryResult: &presexch.PresentationDefinition{},
			},
		}

		query, err := instance.GetQuery()
		require.NoError(t, err)
		require.NotNil(t, query)
	})

	t.Run("GetQuery failed", func(t *testing.T) {
		instance := &Interaction{
			goAPIOpenID4VP: &mocGoAPIInteraction{
				GetQueryError: errors.New("get query failed"),
			},
		}

		query, err := instance.GetQuery()
		require.Contains(t, err.Error(), "get query failed")
		require.Nil(t, query)
	})
}

func TestOpenID4VP_PresentCredential(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		instance := &Interaction{
			keyHandleReader:  &mockKeyHandleReader{},
			crypto:           &mockCrypto{},
			ldDocumentLoader: &documentLoaderWrapper{goAPIDocumentLoader: testutil.DocumentLoader(t)},
			goAPIOpenID4VP: &mocGoAPIInteraction{
				PresentCredentialErr: nil,
			},
		}

		err := instance.PresentCredential(presentationJSONLD,
			api.NewVerificationMethod("did:example:12345#testId", "Ed25519VerificationKey2018"))
		require.NoError(t, err)
	})

	t.Run("Fail to get signature algorithm", func(t *testing.T) {
		instance := &Interaction{
			keyHandleReader:  &mockKeyHandleReader{},
			crypto:           &mockCrypto{},
			ldDocumentLoader: &documentLoaderWrapper{goAPIDocumentLoader: testutil.DocumentLoader(t)},
			goAPIOpenID4VP: &mocGoAPIInteraction{
				PresentCredentialErr: nil,
			},
		}

		err := instance.PresentCredential(presentationJSONLD,
			&api.VerificationMethod{ID: "did:example:12345#testId", Type: "Invalid"})
		require.Contains(t, err.Error(), "create signer failed")
	})

	t.Run("parse presentation failed", func(t *testing.T) {
		instance := &Interaction{
			keyHandleReader:  &mockKeyHandleReader{},
			crypto:           &mockCrypto{},
			ldDocumentLoader: &documentLoaderWrapper{goAPIDocumentLoader: testutil.DocumentLoader(t)},
			goAPIOpenID4VP: &mocGoAPIInteraction{
				PresentCredentialErr: nil,
			},
		}

		err := instance.PresentCredential([]byte("random value"),
			&api.VerificationMethod{ID: "did:example:12345#testId", Type: "Ed25519VerificationKey2018"})
		require.Contains(t, err.Error(), "parse presentation failed")
	})

	t.Run("Present credentials failed", func(t *testing.T) {
		instance := &Interaction{
			keyHandleReader:  &mockKeyHandleReader{},
			crypto:           &mockCrypto{},
			ldDocumentLoader: &documentLoaderWrapper{goAPIDocumentLoader: testutil.DocumentLoader(t)},
			goAPIOpenID4VP: &mocGoAPIInteraction{
				PresentCredentialErr: errors.New("present credentials failed"),
			},
		}

		err := instance.PresentCredential(presentationJSONLD,
			&api.VerificationMethod{ID: "did:example:12345#testId", Type: "Ed25519VerificationKey2018"})
		require.Contains(t, err.Error(), "present credentials failed")
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
		Document:    docBytes,
		ContextURL:  ldDoc.ContextURL,
	}, nil
}

type mockKeyHandleReader struct {
	exportPubKeyResult []byte
	exportPubKeyErr    error
}

func (m *mockKeyHandleReader) ExportPubKey(string) ([]byte, error) {
	return m.exportPubKeyResult, m.exportPubKeyErr
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

type mocGoAPIInteraction struct {
	GetQueryResult       *presexch.PresentationDefinition
	GetQueryError        error
	PresentCredentialErr error
}

func (o *mocGoAPIInteraction) GetQuery() (*presexch.PresentationDefinition, error) {
	return o.GetQueryResult, o.GetQueryError
}

func (o *mocGoAPIInteraction) PresentCredential(
	presentation *verifiable.Presentation,
	jwtSigner goapi.JWTSigner,
) error {
	return o.PresentCredentialErr
}

type mocksDIDResolver struct {
	ResolveDocBytes []byte
	ResolveErr      error
}

func (m *mocksDIDResolver) Resolve(did string) ([]byte, error) {
	return m.ResolveDocBytes, m.ResolveErr
}
