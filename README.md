[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://raw.githubusercontent.com/trustbloc/agent-sdk/main/LICENSE)
[![Release](https://img.shields.io/github/release/trustbloc/wallet-sdk.svg?style=flat-square)](https://github.com/trustbloc/wallet-sdk/releases/latest)
[![Godocs](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/trustbloc/wallet-sdk)

[![Build Status](https://github.com/trustbloc/wallet-sdk/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/trustbloc/wallet-sdk/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/trustbloc/wallet-sdk)](https://goreportcard.com/report/github.com/trustbloc/wallet-sdk)


# Wallet SDK

The TrustBloc Wallet SDK repo contains APIs to issue/present [W3C Verifiable Credential(VC)](https://www.w3.org/TR/vc-data-model/) signed/verified with [W3C Decentralized Identifier(DID)](https://www.w3.org/TR/did-core/). These APIs are useful for the holder role defined in W3C VC Specification.

This project can be used as a
- [Golang SDK](./pkg/)
- [GoMobile Bindings](./cmd/wallet-sdk-gomobile/) 
  - iOS (generates a .xcframework binary)
  - Android (generates a .aar binary)

The repo also has code to generate [Reference iOS/Android App](demo/app/) built using [Ionic](https://ionicframework.com/)/[Capacitor JavaScript](https://capacitorjs.com/) frameworks.


## Build/Run
- [GoMobile Bindings (iOS/Android)](./cmd/wallet-sdk-gomobile/README.md)
- [Demo/Reference App](demo/app/README.md)


## Contributing
Thank you for your interest in contributing. Please see our
[community contribution guidelines](https://github.com/trustbloc/community/blob/main/CONTRIBUTING.md) for more information.

## License
Apache License, Version 2.0 (Apache-2.0). See the [LICENSE](LICENSE) file.
