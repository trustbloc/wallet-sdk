// Copyright Avast Software. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile

go 1.19

require (
	github.com/hyperledger/aries-framework-go v0.1.9-0.20221219221222-fdd5069d178e
	github.com/piprate/json-gold v0.4.2
	github.com/stretchr/testify v1.8.1
	github.com/trustbloc/wallet-sdk v0.0.0-00010101000000-000000000000
	github.com/trustbloc/wallet-sdk/cmd/utilities v0.0.0-00010101000000-000000000000
)

require (
	github.com/PaesslerAG/gval v1.1.0 // indirect
	github.com/PaesslerAG/jsonpath v0.1.1 // indirect
	github.com/VictoriaMetrics/fastcache v1.5.7 // indirect
	github.com/btcsuite/btcd v0.22.0-beta // indirect
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/evanphx/json-patch v4.1.0+incompatible // indirect
	github.com/go-jose/go-jose/v3 v3.0.1-0.20221117193127-916db76e8214 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/tink/go v1.7.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hyperledger/aries-framework-go-ext/component/vdr/longform v0.0.0-20221209153644-5a3273a805c1 // indirect
	github.com/hyperledger/aries-framework-go-ext/component/vdr/sidetree v1.0.0-rc3.0.20221104150937-07bfbe450122 // indirect
	github.com/hyperledger/aries-framework-go/component/storageutil v0.0.0-20220610133818-119077b0ec85 // indirect
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20221025204933-b807371b6f1e // indirect
	github.com/hyperledger/ursa-wrapper-go v0.3.1 // indirect
	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a // indirect
	github.com/kawamuray/jsonpath v0.0.0-20201211160320-7483bafabd7e // indirect
	github.com/kilic/bls12-381 v0.1.1-0.20210503002446-7b7597926c69 // indirect
	github.com/minio/blake2b-simd v0.0.0-20160723061019-3f5f724cb5b1 // indirect
	github.com/minio/sha256-simd v0.1.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.1.0 // indirect
	github.com/multiformats/go-multibase v0.1.1 // indirect
	github.com/multiformats/go-multihash v0.0.14 // indirect
	github.com/multiformats/go-varint v0.0.6 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/pquerna/cachecontrol v0.1.0 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/square/go-jose/v3 v3.0.0-20200630053402-0a67ce9b0693 // indirect
	github.com/teserakt-io/golang-ed25519 v0.0.0-20210104091850-3888c087a4c8 // indirect
	github.com/tidwall/gjson v1.14.3 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tidwall/sjson v1.1.4 // indirect
	github.com/trustbloc/sidetree-core-go v1.0.0-rc3.0.20221028171319-8d44bd1cace7 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.17.0 // indirect
	golang.org/x/crypto v0.1.0 // indirect
	golang.org/x/sys v0.4.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/trustbloc/wallet-sdk => ../../

replace github.com/trustbloc/wallet-sdk/cmd/utilities => ../../cmd/utilities
