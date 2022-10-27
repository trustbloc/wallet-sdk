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

The repo also has code to generate [Reference iOS/Android App](demo/app/) built using [Flutter](https://flutter.dev/) framework.


## Build/Run
- [GoMobile Bindings (iOS/Android)](./cmd/wallet-sdk-gomobile/README.md)
- [Demo/Reference App](demo/app/README.md)

## Library/Package

### Android
The Wallet SDK Android package is available on GitHub Maven Repository. Please check [this page](https://github.com/trustbloc-cicd/snapshot/packages/1690705) for the latest SNAPSHOT versions.

```
<dependency>
 <groupId>dev.trustbloc</groupId>
 <artifactId>vc-wallet-sdk</artifactId>
 <version>0.1.0-SNAPSHOT-7e3a6ed</version>
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
SNAPSHOT_REPO_URL=https://maven.pkg.github.com/trustbloc-cicd/snapshot
RELEASE_REPO_URL=https://maven.pkg.github.com/trustbloc/wallet-sdk


## Contributing
Thank you for your interest in contributing. Please see our
[community contribution guidelines](https://github.com/trustbloc/community/blob/main/CONTRIBUTING.md) for more information.

## License
Apache License, Version 2.0 (Apache-2.0). See the [LICENSE](LICENSE) file.
