/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ldproof_test

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/stretchr/testify/require"
	diddoc "github.com/trustbloc/did-go/doc/did"
	"github.com/trustbloc/kms-go/doc/jose/jwk/jwksupport"
	"github.com/trustbloc/kms-go/doc/util/jwkkid"
	"github.com/trustbloc/kms-go/spi/kms"
	"github.com/trustbloc/vc-go/dataintegrity/suite/ecdsa2019"
	"github.com/trustbloc/vc-go/dataintegrity/suite/eddsa2022"
	"github.com/trustbloc/vc-go/presexch"
	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/internal/testutil"
	"github.com/trustbloc/wallet-sdk/pkg/ldproof"
)

func TestLDProof_Add(t *testing.T) {
	var (
		ldpType *presexch.LdpType
		vm      *diddoc.VerificationMethod
	)

	tests := []struct {
		name  string
		setup func()
		check func(t *testing.T, err error)
	}{
		{
			name: "success EcdsaSecp256k1Signature2019",
			setup: func() {
				ldpType = &presexch.LdpType{
					ProofType: []string{"EcdsaSecp256k1Signature2019"},
				}

				privateKey, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
				require.NoError(t, err)

				jwk, err := jwksupport.JWKFromKey(&privateKey.PublicKey)
				require.NoError(t, err)

				vm, err = diddoc.NewVerificationMethodFromJWK("vmID", "EcdsaSecp256k1Signature2019", "", jwk)
				require.NoError(t, err)
			},
			check: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "success Ed25519Signature2018",
			setup: func() {
				ldpType = &presexch.LdpType{
					ProofType: []string{"Ed25519Signature2018"},
				}

				verKey, _, err := ed25519.GenerateKey(rand.Reader)
				require.NoError(t, err)

				vm = diddoc.NewVerificationMethodFromBytes("vmID", "Ed25519VerificationKey2018", "", verKey)
			},
			check: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "success Ed25519Signature2020",
			setup: func() {
				ldpType = &presexch.LdpType{
					ProofType: []string{"Ed25519Signature2020"},
				}

				verKey, _, err := ed25519.GenerateKey(rand.Reader)
				require.NoError(t, err)

				vm = diddoc.NewVerificationMethodFromBytes("vmID", "Ed25519VerificationKey2020", "", verKey)
			},
			check: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "success JsonWebSignature2020",
			setup: func() {
				ldpType = &presexch.LdpType{
					ProofType: []string{"JsonWebSignature2020"},
				}

				privateKey, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
				require.NoError(t, err)

				jwk, err := jwksupport.JWKFromKey(&privateKey.PublicKey)
				require.NoError(t, err)

				vm, err = diddoc.NewVerificationMethodFromJWK("vmID", "JsonWebKey2020", "", jwk)
				require.NoError(t, err)
			},
			check: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "success ecdsa-rdfc-2019",
			setup: func() {
				ldpType = &presexch.LdpType{
					ProofType: []string{ecdsa2019.SuiteTypeNew},
				}

				privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
				require.NoError(t, err)

				jwk, err := jwksupport.JWKFromKey(&privateKey.PublicKey)
				require.NoError(t, err)

				vm, err = diddoc.NewVerificationMethodFromJWK("vmID", "", "", jwk)
				require.NoError(t, err)
			},
			check: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "success eddsa-rdfc-2022",
			setup: func() {
				ldpType = &presexch.LdpType{
					ProofType: []string{eddsa2022.SuiteType},
				}

				verKey, _, err := ed25519.GenerateKey(rand.Reader)
				require.NoError(t, err)

				jwk, err := jwkkid.BuildJWK(verKey, kms.ED25519Type)
				require.NoError(t, err)

				vm, err = diddoc.NewVerificationMethodFromJWK("vmID", "", "", jwk)
				require.NoError(t, err)
			},
			check: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "missing key value for verification method",
			setup: func() {
				ldpType = &presexch.LdpType{
					ProofType: []string{"Ed25519Signature2020"},
				}

				vm = diddoc.NewVerificationMethodFromBytes("vmID", "Ed25519VerificationKey2020", "", nil)
			},
			check: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "missing key value for vmID verification method")
			},
		},
		{
			name: "unsupported verification method type",
			setup: func() {
				ldpType = &presexch.LdpType{
					ProofType: []string{"Ed25519VerificationKey2020"},
				}

				vm = diddoc.NewVerificationMethodFromBytes("vmID", "unsupported", "", []byte("key value"))
			},
			check: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "unsupported verification method type")
			},
		},
		{
			name: "unsupported data integrity key type",
			setup: func() {
				ldpType = &presexch.LdpType{
					ProofType: []string{ecdsa2019.SuiteTypeNew},
				}

				privateKey, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
				require.NoError(t, err)

				jwk, err := jwksupport.JWKFromKey(&privateKey.PublicKey)
				require.NoError(t, err)

				vm, err = diddoc.NewVerificationMethodFromJWK("vmID", "", "", jwk)
				require.NoError(t, err)
			},
			check: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "no supported ldp types found")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			resolverMock := &didResolverMock{
				DocResolution: &diddoc.DocResolution{
					DIDDocument: &diddoc.Doc{
						AssertionMethod: []diddoc.Verification{
							{
								VerificationMethod: *vm,
							},
						},
						Authentication: []diddoc.Verification{
							{
								VerificationMethod: *vm,
							},
						},
						VerificationMethod: []diddoc.VerificationMethod{
							*vm,
						},
					},
				},
			}

			ldProof := ldproof.New(
				&cryptoMock{Signature: []byte("signature")},
				testutil.DocumentLoader(t),
				resolverMock,
			)

			err := ldProof.Add(&verifiable.Presentation{},
				ldproof.WithLdpType(ldpType),
				ldproof.WithVerificationMethod(vm),
				ldproof.WithChallenge("nonce"),
				ldproof.WithDomain("example.com"),
			)
			tt.check(t, err)
		})
	}
}

type cryptoMock struct {
	Signature []byte
	Err       error
}

func (m *cryptoMock) Sign([]byte, string) ([]byte, error) {
	return m.Signature, m.Err
}

type didResolverMock struct {
	DocResolution *diddoc.DocResolution
	Err           error
}

func (m *didResolverMock) Resolve(did string) (*diddoc.DocResolution, error) {
	return m.DocResolution, m.Err
}
