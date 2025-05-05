// Copyright Avast Software. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/wallet-sdk/test/integration/helper

go 1.23.0

toolchain go1.23.4

require github.com/trustbloc/wallet-sdk/test/integration v0.0.0-20221207181956-419a3951143f

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/trustbloc/cmdutil-go v1.0.0 // indirect
	github.com/trustbloc/logutil-go v1.0.0 // indirect
	go.opentelemetry.io/otel v1.32.0 // indirect
	go.opentelemetry.io/otel/trace v1.32.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/oauth2 v0.24.0 // indirect
)

replace github.com/trustbloc/wallet-sdk/test/integration => ../

replace github.com/trustbloc/wallet-sdk => ../../../

replace github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile => ../../../cmd/wallet-sdk-gomobile
