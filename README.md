[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://raw.githubusercontent.com/trustbloc/agent-sdk/main/LICENSE)
[![Release](https://img.shields.io/github/release/trustbloc/wallet-sdk.svg?style=flat-square)](https://github.com/trustbloc/wallet-sdk/releases/latest)
[![Godocs](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/trustbloc/wallet-sdk)

[![Build Status](https://github.com/trustbloc/wallet-sdk/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/trustbloc/wallet-sdk/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/trustbloc/wallet-sdk)](https://goreportcard.com/report/github.com/trustbloc/wallet-sdk)


# Wallet SDK

The TrustBloc Wallet SDK repo contains APIs to receive and present [W3C Verifiable Credentials(VCs)](https://www.w3.org/TR/vc-data-model/) 
signed using [W3C Decentralized Identifiers(DIDs)](https://www.w3.org/TR/did-core/). These APIs are useful for the holder 
role defined in the [W3C VC Specification](https://www.w3.org/TR/vc-data-model/#dfn-holders).

The project contains the following source code components and build scripts for SDKs.
- Go SDK
  - Build native Go applications using the [core pkg](pkg).
- Android and iOS SDK
  - [A gomobile-compatible Go SDK](cmd/wallet-sdk-gomobile)
    - For generating gomobile-compatible bindings (see below).
    - To jump straight to usage documentation, see [here](cmd/wallet-sdk-gomobile/docs/usage.md).
  - [Scripts to generate Android and iOS-compatible bindings](cmd/wallet-sdk-gomobile/README.md)
    - Allows the Go SDK to be used in an Android or iOS app.

The repo also has code to generate a [Reference iOS and Android](demo/app/) built using the [Flutter](https://flutter.dev/) framework.

## Specifications

The TrustBloc Wallet SDK implements following specifications.
- W3C [Verifiable Credential Data Model (VCDM)](https://www.w3.org/TR/vc-data-model/)
- W3C [Decentralized Identifier (DID)](https://www.w3.org/TR/did-core/)
- OIDF [OpenID for Verifiable Credential Issuance](https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0.html)
  - Pre Authorization Code flow
  - Authorization Code low
- OIDF [OpenID for Verifiable Presentation](https://openid.net/specs/openid-4-verifiable-presentations-1_0.html)
- DIF [Presentation Exchange](https://identity.foundation/presentation-exchange/)
- DIF [Well Known DID Configuration](https://identity.foundation/.well-known/resources/did-configuration/)

## Build/Run
- [GoMobile Bindings (iOS/Android)](cmd/wallet-sdk-gomobile/README.md)
- [Demo/Reference App](demo/app/README.md)

## Library/Package

### Android
The Wallet SDK Android package is available on GitHub Maven Repository. Please refer 
[wallet-sdk maven packages](https://github.com/trustbloc/wallet-sdk/packages/1769347) for the latest releases.

```
<dependency>
  <groupId>dev.trustbloc</groupId>
  <artifactId>vc-wallet-sdk</artifactId>
  <version>1.0.0</version>
</dependency>
```

Refer [here](./cmd/wallet-sdk-gomobile/README.md#helpful-tips) for using the android package in your project.

#### Gradle Config

Add the following GitHub maven repository to dependencyResolutionManagement.repositories
```
maven {
  url = $URL
  credentials {
    username = $GITHUB_USER
    password = $GITHUB_TOKEN
  }
 }
```

Use the following URL based on snapshot or release dependency:
RELEASE_REPO_URL=https://maven.pkg.github.com/trustbloc/wallet-sdk

### iOS
The Wallet SDK iOS xcframework packages are distributed through Swift Package Manager (SPM). Please refer
[wallet-sdk tags](https://github.com/trustbloc/wallet-sdk/tags) with the suffix `-swift-pm` (e.g., `1.0.0-swift-pm`) for the
latest releases.


## Project structure

The Go SDK is defined in [pkg](pkg). If you want to build a native Go application, then this is what you'd use.

The `gomobile`-compatible version of the aforementioned Go SDK is defined in [cmd/wallet-sdk-gomobile](cmd/wallet-sdk-gomobile). It's similar to the [Go SDK](pkg), except that the various functions, methods, and interfaces only use a subset of Go types that are compatible with the `gomobile` tool. The `gomobile`-compatible SDK generally acts as a wrapper for the [Go SDK](pkg). Internally, it converts between the `gomobile`-compatible types and the types used by the [Go SDK](pkg) as needed.

## Contributing
Thank you for your interest in contributing. Please see our
[community contribution guidelines](https://github.com/trustbloc/community/blob/main/CONTRIBUTING.md) for more information.

## License
Apache License, Version 2.0 (Apache-2.0). See the [LICENSE](LICENSE) file.
