// Copyright Avast Software. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/wallet-sdk/test/integration

go 1.19

require (
	github.com/google/uuid v1.3.0
	github.com/hyperledger/aries-framework-go v0.1.9-0.20221202141134-083803ecf0a3
	github.com/stretchr/testify v1.8.1
	github.com/trustbloc/cmdutil-go v0.0.0-20221125151303-09d42adcc811
	github.com/trustbloc/logutil-go v0.0.0-20221124174025-c46110e3ea42
	github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile v0.0.0-20221125140656-aa1c5329d343
	go.uber.org/zap v1.23.0
	golang.org/x/oauth2 v0.1.0
)

require (
	github.com/VictoriaMetrics/fastcache v1.5.7 // indirect
	github.com/btcsuite/btcd v0.22.0-beta // indirect
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/tink/go v1.7.0 // indirect
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20221025204933-b807371b6f1e // indirect
	github.com/hyperledger/ursa-wrapper-go v0.3.1 // indirect
	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a // indirect
	github.com/kilic/bls12-381 v0.1.1-0.20210503002446-7b7597926c69 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/minio/blake2b-simd v0.0.0-20160723061019-3f5f724cb5b1 // indirect
	github.com/minio/sha256-simd v0.1.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.1.0 // indirect
	github.com/multiformats/go-multibase v0.1.1 // indirect
	github.com/multiformats/go-multihash v0.0.13 // indirect
	github.com/multiformats/go-varint v0.0.5 // indirect
	github.com/piprate/json-gold v0.4.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/pquerna/cachecontrol v0.1.0 // indirect
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/square/go-jose/v3 v3.0.0-20200630053402-0a67ce9b0693 // indirect
	github.com/teserakt-io/golang-ed25519 v0.0.0-20210104091850-3888c087a4c8 // indirect
	github.com/trustbloc/wallet-sdk v0.0.0-00010101000000-000000000000 // indirect
	github.com/trustbloc/wallet-sdk/cmd/utilities v0.0.0-00010101000000-000000000000 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.1.0 // indirect
	golang.org/x/net v0.1.0 // indirect
	golang.org/x/sys v0.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/trustbloc/wallet-sdk => ../../

replace github.com/trustbloc/wallet-sdk/cmd/utilities => ../../cmd/utilities

replace github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile => ../../cmd/wallet-sdk-gomobile
