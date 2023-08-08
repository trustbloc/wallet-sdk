# Demo/Reference App

The TrustBloc Demo app demonstrates the wallet-sdk API usage. At a high level, the app uses the following APIs.
- Create a Decentralized Identifier (DID) API
- OpenID4CI Issuance APIs
- OpenID4VP Presentation APIs 

The Demo app built using Flutter framework currently supports Android, iOS and Web App targets.

### Installation
For help getting started with Flutter development, view the
[online documentation](https://docs.flutter.dev/), which offers tutorials,
samples, guidance on mobile development, and a full API reference.

- Download Flutter : https://docs.flutter.dev/get-started/install/
- Add export PATH="$PATH:`pwd`/flutter/bin"
- Verify that the flutter command is available by running: `which flutter`
- After installing the flutter Run `flutter doctor`. This command checks your environment and displays a report to the terminal window.
  - If everything is installed as required, the sample output will be like this
  ```
    [✓] Flutter (Channel stable, 3.10.6, on macOS 13.5 22G74 darwin-arm64, locale
    en-CA)
    [✓] Android toolchain - develop for Android devices (Android SDK version 33.0.1)
    [✓] Xcode - develop for iOS and macOS (Xcode 14.2)
    [✓] Chrome - develop for the web
    [✓] Android Studio (version 2022.3)
    [✓] VS Code (version 1.80.2)
    [✓] Connected device (2 available)
    [✓] Network resources
    ```
- The Dart SDK is bundled with Flutter; it is not necessary to install Dart separately.
- iOS
  - For ios setup - install [XCode](https://docs.flutter.dev/get-started/install/macos#install-xcode)
  - TrustBloc app uses the plugin, therefore we use third-party CocoaPods dependency manager `sudo gem install cocoapods`[See More here][https://docs.flutter.dev/get-started/install/macos#deploy-to-ios-devices]
  - Apple Developer Account
- Android
  - For android setup - install [Android Studio](https://docs.flutter.dev/get-started/install/macos#android-setup)
- Web App
  - For android setup - install [Android Studio](https://docs.flutter.dev/get-started/install/macos#android-setup)

### Build and Run
- Build the SDK Bindings 
  - For Android/iOS, [refer here](https://github.com/trustbloc/wallet-sdk/blob/main/cmd/wallet-sdk-gomobile/README.md).
  - For Web App, [refer here](https://github.com/trustbloc/wallet-sdk/blob/main/cmd/wallet-sdk-js/README.md).
- Android
  - Open the app in Android Studio
  - Either run the Emulator for Android or attach the usb device, Android studio will detect the device to run the app on your phone/test device.
- iOS
  - run `flutter build ios`
  - run `pod install` in the `ios` folder
  - Open the app in XCode
  - Install app to Simulator or Device from the XCode Build tab
- Web App
  - run `flutter run -d chrome` from the terminal

### Build with WalletSDK from Maven
- Android
  - Create GitHub personal access token (classic).
  - Run `WALLET_SDK_USR=GIT_USR WALLET_SDK_TKN=TKN flutter run`. Alternatively, add `wallet-sdk-pkg.usr` and `wallet-sdk-pkg.tkn` variables to `android/local.properties` and then run `flutter run`.

Note: If you are switching between the Maven version and a local version, the Gradle cache may need to be cleared. To do this, run `rm -r $HOME/.gradle/`.

### Frequent Issues and Troubleshooting
- Deprecation error in Android Studio
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
- Issues with Android SDK and NDK installation
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
