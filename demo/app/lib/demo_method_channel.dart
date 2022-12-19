import 'dart:ffi';

import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';

import 'demo_platform_interface.dart';

class MethodChannelWallet extends WalletPlatform {
  @visibleForTesting
  final methodChannel = const MethodChannel('WalletSDKPlugin');
  final errorCode = 'Exception';

  Future<void> initSDK() async {
    await methodChannel.invokeMethod<bool>('initSDK');
  }

  Future<String?> createDID() async {
    final createDIDMsg = await methodChannel.invokeMethod<String>('createDID');
    return createDIDMsg;
  }

  Future<bool?> authorize(String qrCode) async {
    final authorizeResult =
        await methodChannel.invokeMethod<bool>('authorize', <String, dynamic>{'requestURI': qrCode});
    return authorizeResult;
  }

  Future<String?> requestCredential(String userPinEntered) async {
    try {
      final credentialResponse =
      await methodChannel.invokeMethod<String>('requestCredential', <String, dynamic>{'otp': userPinEntered});
      print("request credential");
      return credentialResponse;
    } on PlatformException catch (error) {
      if (error.code == errorCode) {
        return error.details.toString();
      }
    }
    return null;
  }

  Future<String?> resolveCredentialDisplay() async {
    try {
      final credentialResponse =
      await methodChannel.invokeMethod<String>('resolveCredentialDisplay');
      return credentialResponse;
    } on PlatformException catch (error) {
      if (error.code == errorCode) {
        return error.details.toString();
      }
    }
    return null;
  }

  // TODO: return credentials display information after it implemented in go sdk
  Future<void> processAuthorizationRequest(
      {required String authorizationRequest, required List<String> storedCredentials}) async {
    await methodChannel.invokeMethod<String>('processAuthorizationRequest',
        <String, dynamic>{'authorizationRequest': authorizationRequest, 'storedCredentials': storedCredentials});
  }

  Future<void> presentCredential({required String signingKeyId}) async {
    await methodChannel.invokeMethod<String>('presentCredential', <String, dynamic>{'signingKeyId': signingKeyId});
  }
}
