/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

@JS()
library script.js;

import 'package:js/js_util.dart';

import 'dart:developer';

import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';

import 'wallet_sdk_interface.dart';
import 'wallet_sdk_model.dart';

import 'package:js/js.dart';

@JS()
@staticInterop
class CreateOpenID4CIInteractionResult {}

extension CreateOpenID4CIInteractionExt on CreateOpenID4CIInteractionResult {
  external bool get userPINRequired;
}

@JS()
@staticInterop
class DidDocResolution {}

extension DidDocResolutionExt on DidDocResolution {
  external String get id;

  external String get content;
}

@JS()
external dynamic jsInitSDK(String didResolverURI);

@JS()
external dynamic jsCreateDID(String didMethod, String keyType);

@JS()
external dynamic jsCreateOpenID4CIInteraction(String initiateIssuanceURI);

@JS()
external dynamic jsRequestCredentialWithPreAuth(String userPinEntered);

@JS()
external dynamic jsIssuerURI();

@JS()
external dynamic jsResolveDisplayData(String issuerURI, List<String> credentials);

@JS()
external dynamic jsGetCredentialID(String credential);

class WalletSDK extends WalletPlatform {
  @visibleForTesting
  final methodChannel = const MethodChannel('WalletSDKPlugin');
  final errorCode = 'Exception';

  Future<void> initSDK(String didResolverURI) async {
    await promiseToFuture(jsInitSDK(didResolverURI));
  }

  Future<Map<String, dynamic>> createDID(String didMethodType, String didKeyType) async {
    DidDocResolution result = await promiseToFuture(jsCreateDID(didMethodType, didKeyType));
    return {"did": result.id, "didDoc": result.content};
  }

  Future<String?> fetchStoredDID(String didID) async {
    final fetchDIDMsg = await methodChannel.invokeMethod<String>('fetchDID', <String, dynamic>{'didID': didID});
    return fetchDIDMsg;
  }

  Future<Map<String, dynamic>> initialize(String initiateIssuanceURI, Map<String, dynamic>? authCodeArgs) async {
    try {
      final result =
          await promiseToFuture(jsCreateOpenID4CIInteraction(initiateIssuanceURI)) as CreateOpenID4CIInteractionResult;

      return {"pinRequired": result.userPINRequired};
    } on PlatformException catch (error) {
      debugPrint(error.toString());
      rethrow;
    }
  }

  Future<String> requestCredential(String userPinEntered) async {
    try {
      String credentialResponse = await promiseToFuture(jsRequestCredentialWithPreAuth(userPinEntered));
      return credentialResponse;
    } catch (error) {
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

  Future<String> issuerURI() async {
    final String issuerURI = jsIssuerURI();
    return issuerURI;
  }

  Future<String> serializeDisplayData(List<String> credentials, String issuerURI) async {
    return await promiseToFuture(jsResolveDisplayData(issuerURI, credentials));
  }

  Future<List<Object?>> resolveCredentialDisplay(String resolvedCredentialDisplayData) async {
    var renderedCredDisplay = await methodChannel.invokeMethod(
        'resolveCredentialDisplay', <String, dynamic>{'resolvedCredentialDisplayData': resolvedCredentialDisplayData});
    return renderedCredDisplay;
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
    return [];
  }

  Future<List<Object?>> parseActivities(List<dynamic> activities) async {
    return [];
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

  Future<String> getCredID(List<String> credentials) async {
    return promiseToFuture(jsGetCredentialID(credentials[0]));
  }
}
