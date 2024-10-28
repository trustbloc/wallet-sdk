/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ldproof_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/piprate/json-gold/ld"
	"github.com/stretchr/testify/require"
	diddoc "github.com/trustbloc/did-go/doc/did"
	"github.com/trustbloc/kms-go/doc/util/jwkkid"
	"github.com/trustbloc/kms-go/spi/kms"
	"github.com/trustbloc/vc-go/presexch"
	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/internal/testutil"
	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/ldproof"
)

func TestNew(t *testing.T) {
	var (
		crypto         api.Crypto
		documentLoader ld.DocumentLoader
		ldpVPFormat    *presexch.LdpType
		keyType        kms.KeyType
	)

	tests := []struct {
		name  string
		setup func()
		check func(t *testing.T, ldProof *ldproof.LDProof, err error)
	}{
		{
			name: "success",
			setup: func() {
				crypto = &cryptoMock{}
				documentLoader = testutil.DocumentLoader(t)
				ldpVPFormat = &presexch.LdpType{
					ProofType: []string{"EcdsaSecp256k1Signature2019", "Ed25519Signature2018", "Ed25519Signature2020",
						"JsonWebSignature2020",
					},
				}
				keyType = kms.ECDSAP384IEEEP1363
			},
			check: func(t *testing.T, ldProof *ldproof.LDProof, err error) {
				require.NotNil(t, ldProof)
				require.NoError(t, err)
			},
		},
		{
			name: "no supported linked data proof found",
			setup: func() {
				documentLoader = testutil.DocumentLoader(t)
				ldpVPFormat = &presexch.LdpType{
					ProofType: []string{"Ed25519Signature2018", "Ed25519Signature2020"},
				}
				keyType = kms.ECDSAP384IEEEP1363
			},
			check: func(t *testing.T, ldProof *ldproof.LDProof, err error) {
				require.Nil(t, ldProof)
				require.ErrorContains(t, err, "no supported linked data proof found")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			ldProof, err := ldproof.New(crypto, documentLoader, ldpVPFormat, keyType)
			tt.check(t, ldProof, err)
		})
	}
}

func TestLDProof_Add(t *testing.T) {
	ldProof, err := ldproof.New(
		&cryptoMock{Signature: []byte("signature")},
		testutil.DocumentLoader(t),
		&presexch.LdpType{
			ProofType: []string{"EcdsaSecp256k1Signature2019", "Ed25519Signature2018", "Ed25519Signature2020",
				"JsonWebSignature2020",
			},
		},
		kms.ED25519Type,
	)
	require.NoError(t, err)

	verKey, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	jwk, err := jwkkid.BuildJWK(verKey, kms.ED25519Type)
	require.NoError(t, err)

	var (
		vp   *verifiable.Presentation
		opts []ldproof.Opt
	)

	tests := []struct {
		name  string
		setup func()
		check func(t *testing.T, err error)
	}{
		{
			name: "success",
			setup: func() {
				vp = &verifiable.Presentation{}

				var signingVM *diddoc.VerificationMethod

				signingVM, err = diddoc.NewVerificationMethodFromJWK("vmID", "Ed25519VerificationKey2018", "", jwk)
				require.NoError(t, err)

				opts = []ldproof.Opt{
					ldproof.WithSigningVM(signingVM),
					ldproof.WithNonce("nonce"),
					ldproof.WithDomain("example.com"),
				}
			},
			check: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "missing signing verification method",
			setup: func() {
				vp = &verifiable.Presentation{}

				opts = []ldproof.Opt{
					ldproof.WithNonce("nonce"),
					ldproof.WithDomain("example.com"),
				}
			},
			check: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "missing signing verification method")
			},
		},
		{
			name: "missing jwk for verification method",
			setup: func() {
				vp = &verifiable.Presentation{}

				signingVM := &diddoc.VerificationMethod{
					ID:   "vmID",
					Type: "Ed25519VerificationKey2018",
				}

				opts = []ldproof.Opt{
					ldproof.WithSigningVM(signingVM),
					ldproof.WithNonce("nonce"),
					ldproof.WithDomain("example.com"),
				}
			},
			check: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "missing jwk for vmID verification method")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			err = ldProof.Add(vp, opts...)
			tt.check(t, err)
		})
	}
}

type cryptoMock struct {
	Signature []byte
	Err       error
}

func (c *cryptoMock) Sign([]byte, string) ([]byte, error) {
	return c.Signature, c.Err
}
