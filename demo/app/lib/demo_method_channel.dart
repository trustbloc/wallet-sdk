import 'dart:ffi';

import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';

import 'demo_platform_interface.dart';

class MethodChannelWallet extends WalletPlatform {
  @visibleForTesting
  final methodChannel = const MethodChannel('WalletSDKPlugin');

  Future<String?> createDID() async {
    final createDIDMsg = await methodChannel.invokeMethod<String>('createDID');
    return createDIDMsg;
  }

  Future<bool?> authorize(String? qrCode) async {
    final authorizeResult = await methodChannel.invokeMethod<bool>('authorize', <String, dynamic>{'requestURI': qrCode});
    return authorizeResult;
  }

  Future<String?> requestCredential(String? userPinEntered) async {
    final credentialResponse = await methodChannel.invokeMethod<String>('requestCredential', <String, dynamic>{'otp': userPinEntered});
    return credentialResponse;
  }
}
