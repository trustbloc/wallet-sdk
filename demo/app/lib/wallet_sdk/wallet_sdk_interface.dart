/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'package:plugin_platform_interface/plugin_platform_interface.dart';

import 'wallet_sdk.dart';

abstract class WalletPlatform extends PlatformInterface {
  /// Constructs a HelloPlatform.
  WalletPlatform() : super(token: _token);

  static final Object _token = Object();

  static WalletPlatform _instance = WalletSDK();
  static WalletPlatform get instance => _instance;

  /// Platform-specific implementations should set this with their own
  /// platform-specific class that extends [WalletPlatform] when
  /// they register themselves.
  static set instance(WalletPlatform instance) {
    PlatformInterface.verifyToken(instance, _token);
    _instance = instance;
  }

  Future<String?> getPlatformVersion() {
    throw UnimplementedError('platformVersion() has not been implemented.');
  }
}
