# Demo/Reference App

The TrustBloc Demo app demonstrates the wallet-sdk API usage. At a high level, the app uses the following APIs.
- Create a Decentralized Identifier (DID) API
- OpenID4CI Issuance APIs
- OpenID4VP Presentation APIs 

## Getting Started

This project is a starting point for a Flutter application.

A few resources to get you started if this is your first Flutter project:

Download Flutter : https://docs.flutter.dev/get-started/install/

Add export PATH="$PATH:`pwd`/flutter/bin"
Verify that the flutter command is available by running: `which flutter`

Things you need to install before running the app in your local: 
- After installing the flutter Run `flutter doctor`. This command checks your environment and displays a report to the terminal window.
  - If everything is installed as required. Sample output will be like this 
  ```
      [✓] Flutter (Channel stable, 3.3.5, on macOS 13.0 22A380 darwin-arm, locale en-CA)
      [✓] Android toolchain - develop for Android devices (Android SDK version 33.0.0)
      [✓] Xcode - develop for iOS and macOS (Xcode 14.2)
      [✓] Chrome - develop for the web
      [✓] Android Studio (version 2021.3)
      [✓] Connected device (3 available)
      [✓] HTTP Host Availability
    ```
- The Dart SDK is bundled with Flutter; it is not necessary to install Dart separately.
- For ios setup - install [XCode](https://docs.flutter.dev/get-started/install/macos#install-xcode)
- Trustbloc app uses the plugin, therefore we uses third-party CocoaPods dependency manager `sudo gem install cocoapods`[See More here][https://docs.flutter.dev/get-started/install/macos#deploy-to-ios-devices]
  - Once we clone the project, run `pod install` to get all the latest dependencies.
- For android setup - install [Android Studio](https://docs.flutter.dev/get-started/install/macos#android-setup)
- Build the sdk bindings by following this [document](https://github.com/trustbloc/wallet-sdk/blob/main/cmd/wallet-sdk-gomobile/README.md)
- Either run the emulator for android or attach the usb device, android studio will detect the device to run the app on your phone/test device.
- Frequent Issue and solutions:  

```
Problem:
Launching lib/main.dart on sdk gphone64 arm64 in debug mode...
Note: /Users/user/.pub-cache/hosted/pub.dev/uni_links-0.5.1/android/src/main/java/name/avioli/unilinks/UniLinksPlugin.java uses or overrides a deprecated API.
Note: Recompile with -Xlint:deprecation for details.
```

```
Solution:
In your android studio: 
Preferences > Build, Execution, Deployment > Build Tools > Gradle > check "Generate .IML files for modules imported from Gradle 
```

```
Problem: 
Having issues with android sdk and ndk installation
```
```
Solution: 
Install the SDK Tools by following: 
Preferences | Appearance & Behavior | System Settings | Android SDK | SDK Tools | Android Tools 
Preferences | Appearance & Behavior | System Settings | Android SDK | SDK Tools | NDK
```

For help getting started with Flutter development, view the
[online documentation](https://docs.flutter.dev/), which offers tutorials,
samples, guidance on mobile development, and a full API reference.
