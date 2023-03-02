/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/metricslogger/stderr"

	diddoc "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/doc/util"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdr"
	"github.com/piprate/json-gold/ld"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/credential"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/did"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
	sdkapi "github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	"github.com/trustbloc/wallet-sdk/pkg/did/resolver"
)

func TestCredentialAPI(t *testing.T) {
	kms, e := localkms.NewKMS(localkms.NewMemKMSStore())
	require.NoError(t, e)

	crypto := kms.GetCrypto()

	credStore := credential.NewInMemoryDB()

	ldLoader := ld.NewDefaultDocumentLoader(common.DefaultHTTPClient())

	ldLoaderWrapper := &documentLoaderReverseWrapper{DocumentLoader: ldLoader}

	didResolver, e := did.NewResolver("")
	require.NoError(t, e)

	signer, e := credential.NewSigner(credStore, didResolver, crypto, ldLoaderWrapper)
	require.NoError(t, e)

	c, e := did.NewCreatorWithKeyWriter(kms)
	require.NoError(t, e)

	sdkResolver, e := resolver.NewDIDResolver("")
	require.NoError(t, e)

	verifier := jwtvcVerifier{
		ldLoader:         ldLoader,
		publicKeyFetcher: verifiable.NewVDRKeyResolver(&didResolverWrapper{didResolver: sdkResolver}).PublicKeyFetcher(),
	}

	testCases := []struct {
		name          string
		didMethod     string
		getCredByName bool
	}{
		{
			name:          "did:ion signing DID",
			didMethod:     "ion",
			getCredByName: false,
		},
		{
			name:          "did:ion signing DID, with stored credential",
			didMethod:     "ion",
			getCredByName: true,
		},
		{
			name:          "did:key signing DID",
			didMethod:     "key",
			getCredByName: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			didDoc, err := c.Create(tc.didMethod, &api.CreateDIDOpts{MetricsLogger: stderr.NewMetricsLogger()})
			require.NoError(t, err)

			docID, err := didDoc.ID()
			require.NoError(t, err)

			templateCredential := &verifiable.Credential{
				ID:      "cred-ID",
				Types:   []string{verifiable.VCType},
				Context: []string{verifiable.ContextURI},
				Subject: verifiable.Subject{
					ID: "foo",
				},
				Issuer: verifiable.Issuer{
					ID: docID,
				},
				Issued: util.NewTime(time.Now()),
			}

			err = credStore.Add(api.NewVerifiableCredential(templateCredential))
			require.NoError(t, err)

			var cred *api.VerifiableCredential
			var credID string

			if tc.getCredByName {
				credID = templateCredential.ID
			} else {
				cred = api.NewVerifiableCredential(templateCredential)
			}

			issuedCred, err := signer.Issue(cred, credID, docID)
			require.NoError(t, err)

			require.NoError(t, verifier.verify(issuedCred))
		})
	}
}

type jwtvcVerifier struct {
	ldLoader         ld.DocumentLoader
	publicKeyFetcher verifiable.PublicKeyFetcher
}

func (j *jwtvcVerifier) verify(cred []byte) error {
	_, err := verifiable.ParseCredential(
		cred,
		verifiable.WithJSONLDDocumentLoader(j.ldLoader),
		verifiable.WithPublicKeyFetcher(j.publicKeyFetcher),
	)

	return err
}

type didResolverWrapper struct {
	didResolver sdkapi.DIDResolver
}

func (d *didResolverWrapper) Resolve(did string, _ ...vdr.DIDMethodOption) (*diddoc.DocResolution, error) {
	return d.didResolver.Resolve(did)
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
