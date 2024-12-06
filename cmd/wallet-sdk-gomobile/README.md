# Wallet SDK Gomobile

This package contains the `gomobile`-compatible version of the SDK. It acts as a wrapper for the [Go SDK](../../README.md).

## Prerequisites

* [Go 1.23](https://go.dev/doc/install) or newer
* The gomobile tools:
  ```
  go install golang.org/x/mobile/cmd/gomobile@latest
  gomobile init
  ```
* `$GOPATH/bin` added to your path

### Android Bindings

* Android SDK (installable via the SDK Manager in [Android Studio](https://developer.android.com/studio/install))
* Android NDK (installable via the SDK Manager in [Android Studio](https://developer.android.com/studio/install))
* [JDK](https://www.oracle.com/java/technologies/downloads/)

### iOS Bindings

* A Mac
* Xcode Command Line Tools - these are included with [Xcode](https://developer.apple.com/xcode/), but can also be installed standalone with `xcode-select --install`
* Agreed to the Xcode licence agreement prompt via Terminal. Creating the iOS bindings will trigger the prompt if you haven't agreed to the licence yet

## Generating the Bindings

To generate both the Android and iOS bindings in one command:

```
make generate-all-bindings
```

To generate only the Android bindings:

```
make generate-android-bindings
```

To generate only the iOS bindings:

```
make generate-ios-bindings
```

The generated bindings can be found in the `bindings` folder.

## SDK Usage

For usage documentation + examples, see [here](docs/usage.md).

## Helpful Tips

* After importing `walletsdk.aar` into your Android project, to use the Wallet SDK packages in a Kotlin or
  Java file, you import them in a roughly similar way as you would the equivalent Go package:
  ```java
  // Follow the pattern below for other packages/types as well
  import dev.trustbloc.wallet.sdk.api.*;
  import dev.trustbloc.wallet.sdk.openid4ci.*;
  ```
* After importing the iOS bindings into Xcode, to use the Wallet SDK in a Swift file, add the following:
  ```swift
  // This imports everything
  import Walletsdk
  ```
  To use the Wallet SDK in an Objective-C file, do the following:
  ```objc
  #import <Walletsdk/Walletsdk.h>
  ```
  Alternatively, you can import individual files. For example:
  ```objc
  #import <Walletsdk/Api.objc.h>
  #import <Walletsdk/Storage.objc.h>
  ```
  Note that the various types, interfaces and methods will have the package names prefixed to them (e.g.
  `Signer` in Go becomes `CredentialsignerSigner` in Swift and Objective-C).
* If you've regenerated bindings on your machine, make sure you update your Android Studio or Xcode project accordingly.
