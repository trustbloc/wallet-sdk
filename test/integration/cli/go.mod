// Copyright Avast Software. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/wallet-sdk/test/integration/helper

go 1.20

require github.com/trustbloc/wallet-sdk/test/integration v0.0.0-20221207181956-419a3951143f

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/testify v1.8.1 // indirect
	github.com/trustbloc/cmdutil-go v0.0.0-20221125151303-09d42adcc811 // indirect
	github.com/trustbloc/logutil-go v0.0.0-20221124174025-c46110e3ea42 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.23.0 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/oauth2 v0.7.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/trustbloc/wallet-sdk/test/integration => ../

replace github.com/trustbloc/wallet-sdk => ../../../

replace github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile => ../../../cmd/wallet-sdk-gomobile
