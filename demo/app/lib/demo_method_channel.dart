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

  Future<String?> createDID(String didMethodType) async {
    final createDIDMsg = await methodChannel.invokeMethod<String>('createDID', <String, dynamic>{'didMethodType': didMethodType});
    return createDIDMsg;
  }

  Future<String?> fetchStoredDID(String didID) async {
    final fetchDIDMsg = await methodChannel.invokeMethod<String>('fetchDID', <String, dynamic>{'didID': didID});
    return fetchDIDMsg;
  }


  Future<bool?> authorize(String qrCode) async {
    final authorizeResult =
        await methodChannel.invokeMethod<bool>('authorize', <String, dynamic>{'requestURI': qrCode});
    return authorizeResult;
  }

  Future<String> requestCredential(String userPinEntered) async {
    try {
      final credentialResponse =
          await methodChannel.invokeMethod<String>('requestCredential', <String, dynamic>{'otp': userPinEntered});
      return credentialResponse!;
    } on PlatformException catch (error) {
      debugPrint(error.toString());
      rethrow;
    }
  }

  Future<String?> issuerURI() async {
    final issuerURI = await methodChannel.invokeMethod<String>('issuerURI');
    return issuerURI;
  }

  Future<String?> resolveCredentialDisplay(List<String> credentials , String issuerURI) async {
    try {
      final credentialResponse =
      await methodChannel.invokeMethod<String>('resolveCredentialDisplay', <String, dynamic>{'vcCredentials':credentials, 'uri':issuerURI});
      return credentialResponse;
    } on PlatformException catch (error) {
      if (error.code == errorCode) {
        return error.details.toString();
      }
    }
    return null;
  }

  Future<List<String>> processAuthorizationRequest(
      {required String authorizationRequest, required List<String> storedCredentials}) async {
    return (await methodChannel.invokeMethod<List>('processAuthorizationRequest',
            <String, dynamic>{'authorizationRequest': authorizationRequest, 'storedCredentials': storedCredentials}))!
        .map((e) => e!.toString())
        .toList();
  }

  Future<void> presentCredential() async {
    await methodChannel.invokeMethod('presentCredential');
  }

  Future<List<Object?>> activityLogger() async {
   var activityObj =  await methodChannel.invokeMethod('activityLogger');
   return activityObj;
  }
}