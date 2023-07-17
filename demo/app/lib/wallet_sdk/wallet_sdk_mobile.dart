/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'dart:developer';

import 'package:app/wallet_sdk/wallet_sdk_model.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';

import 'wallet_sdk_interface.dart';

class WalletSDK extends WalletPlatform {
  @visibleForTesting
  final methodChannel = const MethodChannel('WalletSDKPlugin');
  final errorCode = 'Exception';

  Future<void> initSDK(String didResolverURI) async {
    await methodChannel.invokeMethod<bool>('initSDK', <String, dynamic>{'didResolverURI': didResolverURI});
  }

  Future<Map<Object?, Object?>?> createDID(String didMethodType, String didKeyType) async {
    final createDIDMsg = await methodChannel.invokeMethod<Map<Object?, Object?>?>(
        'createDID', <String, dynamic>{'didMethodType': didMethodType, 'didKeyType': didKeyType});
    return createDIDMsg;
  }

  Future<String?> fetchStoredDID(String didID) async {
    final fetchDIDMsg = await methodChannel.invokeMethod<String>('fetchDID', <String, dynamic>{'didID': didID});
    return fetchDIDMsg;
  }

  Future<Map<Object?, Object?>?> initialize(String qrCode, Map<String, dynamic>? authCodeArgs) async {
    try {
      final flowTypeData = await methodChannel
          .invokeMethod('initialize', <String, dynamic>{'requestURI': qrCode, 'authCodeArgs': authCodeArgs});
      return flowTypeData;
    } on PlatformException catch (error) {
      debugPrint(error.toString());
      rethrow;
    }
  }

  Future<String> requestCredential(String userPinEntered) async {
    try {
      var credentialResponse =
          await methodChannel.invokeMethod<String>('requestCredential', <String, dynamic>{'otp': userPinEntered});
      return credentialResponse!;
    } on PlatformException catch (error) {
      debugPrint(error.toString());
      rethrow;
    }
  }

  Future<String> requestCredentialWithAuth(String redirectURIWithParams) async {
    try {
      var credentialResponse = await methodChannel.invokeMethod<String>(
          'requestCredentialWithAuth', <String, dynamic>{'redirectURIWithParams': redirectURIWithParams});
      return credentialResponse!;
    } on PlatformException catch (error) {
      debugPrint(error.toString());
      rethrow;
    }
  }

  Future<bool?> credentialStatusVerifier(List<String> credentials) async {
    try {
      var credentialStatusVerifier = await methodChannel
          .invokeMethod<bool>('credentialStatusVerifier', <String, dynamic>{'credentials': credentials});
      return credentialStatusVerifier!;
    } on PlatformException catch (error) {
      if (error.toString().contains("status verification failed: revoked")) {
        return false;
      } else {
        debugPrint(error.toString());
      }
      rethrow;
    }
  }

  Future<String?> issuerURI() async {
    final issuerURI = await methodChannel.invokeMethod<String>('issuerURI');
    return issuerURI;
  }

  Future<String?> serializeDisplayData(List<String> credentials, String issuerURI) async {
    final credentialResponse = await methodChannel.invokeMethod<String>(
        'serializeDisplayData', <String, dynamic>{'vcCredentials': credentials, 'uri': issuerURI});
    return credentialResponse;
  }

  Future<List<CredentialDisplayData>> parseCredentialDisplayData(String resolvedCredentialDisplayData) async {
    List<dynamic> renderedCredDisplay = await methodChannel.invokeMethod(
        'resolveCredentialDisplay', <String, dynamic>{'resolvedCredentialDisplayData': resolvedCredentialDisplayData});
    return renderedCredDisplay.map((d) => CredentialDisplayData.fromMap(d.cast<String, dynamic>())).toList();
  }

  Future<List<String>> processAuthorizationRequest(
      {required String authorizationRequest, List<String>? storedCredentials}) async {
    return (await methodChannel.invokeMethod<List>('processAuthorizationRequest',
            <String, dynamic>{'authorizationRequest': authorizationRequest, 'storedCredentials': storedCredentials}))!
        .map((e) => e!.toString())
        .toList();
  }

  Future<List<SubmissionRequirement>> getSubmissionRequirements({required List<String>? storedCredentials}) async {
    return (await methodChannel.invokeMethod<List<dynamic>>(
            'getMatchedSubmissionRequirements', <String, dynamic>{'storedCredentials': storedCredentials}))!
        .map((obj) => SubmissionRequirement.fromMap(obj.cast<String, dynamic>()))
        .toList();
  }

  Future<Map<Object?, Object?>?> getVersionDetails() async {
    var versionDetailResp = await methodChannel.invokeMethod('getVersionDetails');
    log("getVersionDetails in the app, $versionDetailResp");
    return versionDetailResp;
  }

  Future<Map<Object?, Object?>?> wellKnownDidConfig(String issuerID) async {
    var didLinkedResp = await methodChannel.invokeMethod('wellKnownDidConfig', <String, dynamic>{'issuerID': issuerID});
    log("well known config, $didLinkedResp");
    return didLinkedResp;
  }

  Future<Map<Object?, Object?>?> getVerifierDisplayData() async {
    var verifierDisplayData = await methodChannel.invokeMethod('getVerifierDisplayData');
    return verifierDisplayData;
  }

  Future<void> presentCredential({List<String>? selectedCredentials}) async {
    await methodChannel
        .invokeMethod('presentCredential', <String, dynamic>{'selectedCredentials': selectedCredentials});
  }

  Future<List<Object?>> storeActivityLogger() async {
    var activityObj = await methodChannel.invokeMethod('activityLogger');
    return activityObj;
  }

  Future<List<Object?>> parseActivities(List<dynamic> activities) async {
    var activityObj = await methodChannel.invokeMethod('parseActivities', <String, dynamic>{'activities': activities});
    return activityObj;
  }

  Future<String?> getIssuerID(List<String> credentials) async {
    try {
      final issuerID =
          await methodChannel.invokeMethod<String>('getIssuerID', <String, dynamic>{'vcCredentials': credentials});
      log("get issuerID - , $issuerID");
      return issuerID;
    } on PlatformException catch (error) {
      if (error.code == errorCode) {
        return error.details.toString();
      }
    }
    return null;
  }

  Future<String?> getCredID(List<String> credentials) async {
    try {
      final credentialID =
          await methodChannel.invokeMethod<String>('getCredID', <String, dynamic>{'vcCredentials': credentials});
      return credentialID;
    } on PlatformException catch (error) {
      if (error.code == errorCode) {
        return error.details.toString();
      }
    }
    return null;
  }
}
