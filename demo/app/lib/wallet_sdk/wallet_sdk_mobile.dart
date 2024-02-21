/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'dart:convert';
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

  Future<CreateDID> createDID(String didMethodType, String didKeyType) async {
    final createDIDResp = await methodChannel.invokeMethod<Map<Object?, Object?>?>(
        'createDID', <String, dynamic>{'didMethodType': didMethodType, 'didKeyType': didKeyType});
    return CreateDID.fromJson(jsonDecode(json.encode(createDIDResp)));
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

  Future<List<SupportedCredentials>> initializeWalletInitiatedFlow(
      String issuerURI, List<String> credentialTypes) async {
    try {
      List<dynamic> supportedCredentialResp = await methodChannel.invokeMethod('initializeWalletInitiatedFlow',
          <String, dynamic>{'issuerURI': issuerURI, 'credentialTypes': credentialTypes});
      return supportedCredentialResp.map((d) => SupportedCredentials.fromMap(d.cast<String, dynamic>())).toList();
    } on PlatformException catch (error) {
      debugPrint(error.toString());
      rethrow;
    }
  }

  Future<String> createAuthorizationURLWalletInitiatedFlow(List<String> scopes, List<String> credentialTypes,
      String credentialFormat, clientID, redirectURI, issuerURI) async {
    try {
      String authorizationURL =
          await methodChannel.invokeMethod('createAuthorizationURLWalletInitiatedFlow', <String, dynamic>{
        'scopes': scopes,
        'credentialTypes': credentialTypes,
        'credentialFormat': credentialFormat,
        'clientID': clientID,
        'redirectURI': redirectURI,
        'issuerURI': issuerURI
      });
      log('authorizationURL Wallet-Initiated-Flow -> $authorizationURL');
      return authorizationURL;
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

  Future<bool> requireAcknowledgment() async {
    var ackResp = await methodChannel.invokeMethod<bool>('requireAcknowledgment');
    return ackResp!;
  }

  void acknowledgeSuccess() async {
    try {
      await methodChannel.invokeMethod<bool>('acknowledgeSuccess');
    } on PlatformException catch (error) {
      debugPrint(error.toString());
      rethrow;
    }
  }

  void acknowledgeReject() async {
    try {
      await methodChannel.invokeMethod<bool>('acknowledgeReject');
    } on PlatformException catch (error) {
      debugPrint(error.toString());
      rethrow;
    }
  }

  Future<WalletSDKError> parseWalletSDKError({required String localizedErrorMessage}) async {
    var parsedWalletError = await methodChannel
        .invokeMethod('parseWalletSDKError', <String, dynamic>{'localizedErrorMessage': localizedErrorMessage});
    return WalletSDKError.fromJson(jsonDecode(json.encode(parsedWalletError)));
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

  Future<String> requestCredentialWithWalletInitiatedFlow(String redirectURIWithParams) async {
    try {
      var credentialResponse = await methodChannel.invokeMethod<String>('requestCredentialWithWalletInitiatedFlow',
          <String, dynamic>{'redirectURIWithParams': redirectURIWithParams});
      return credentialResponse!;
    } on PlatformException catch (error) {
      debugPrint(error.toString());
      rethrow;
    }
  }

  Future<bool> credentialStatusVerifier(String credential) async {
    try {
      var credentialStatusVerifier =
          await methodChannel.invokeMethod<bool>('credentialStatusVerifier', <String, dynamic>{
        'credentials': [credential]
      });
      return credentialStatusVerifier!;
    } on PlatformException catch (error) {
      if (error.toString().contains('status verification failed: revoked')) {
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
        'parseCredentialDisplay', <String, dynamic>{'resolvedCredentialDisplayData': resolvedCredentialDisplayData});
    return renderedCredDisplay.map((d) => CredentialDisplayData.fromMap(d.cast<String, dynamic>())).toList();
  }

  Future<List<String>> processAuthorizationRequest(
      {required String authorizationRequest, List<String>? storedCredentials}) async {
    return (await methodChannel.invokeMethod<List>('processAuthorizationRequest',
            <String, dynamic>{'authorizationRequest': authorizationRequest, 'storedCredentials': storedCredentials}))!
        .map((e) => e!.toString())
        .toList();
  }

  Future<List<Object?>> getCustomScope() async {
    return await methodChannel.invokeMethod('getCustomScope');
  }

  Future<List<SubmissionRequirement>> getSubmissionRequirements({required List<String>? storedCredentials}) async {
    return (await methodChannel.invokeMethod<List<dynamic>>(
            'getMatchedSubmissionRequirements', <String, dynamic>{'storedCredentials': storedCredentials}))!
        .map((obj) => SubmissionRequirement.fromMap(obj.cast<String, dynamic>()))
        .toList();
  }

  Future<Map<Object?, Object?>?> getVersionDetails() async {
    var versionDetailResp = await methodChannel.invokeMethod('getVersionDetails');
    log('getVersionDetails in the app, $versionDetailResp');
    return versionDetailResp;
  }

  Future<WellKnownDidConfig> wellKnownDidConfig(String issuerID) async {
    Map<String, dynamic> config =
        (await methodChannel.invokeMethod('wellKnownDidConfig', <String, dynamic>{'issuerID': issuerID}))
            .cast<String, dynamic>();
    return WellKnownDidConfig.fromMap(config);
  }

  Future<VerifierDisplayData> getVerifierDisplayData() async {
    var data = await methodChannel.invokeMethod('getVerifierDisplayData');
    return VerifierDisplayData(
        name: data['name'] as String,
        did: data['did'] as String,
        logoURI: data['logoUri'] as String,
        purpose: data['purpose'] as String);
  }

  Future<EvaluationResult?> evaluateIssuanceTrustInfo() async {
    var data = await methodChannel.invokeMethod(
        'evaluateIssuanceTrustInfo', {'evaluateIssuanceURL': 'https://krakend-k8s-dev3.dev.dts-dsa.com/trustregistry/wallet/interactions/issuance'});
    return EvaluationResult.fromMap(data.cast<String, dynamic>());
  }

  Future<EvaluationResult?> evaluatePresentationTrustInfo() async {
    var data = await methodChannel.invokeMethod('evaluatePresentationTrustInfo',
        {'evaluatePresentationURL': 'https://krakend-k8s-dev3.dev.dts-dsa.com/trustregistry/wallet/interactions/presentation'});
    return EvaluationResult.fromMap(data.cast<String, dynamic>());
  }

  Future<void> presentCredential(
      {required List<String> selectedCredentials, Map<String, dynamic>? customScopeList}) async {
    try {
      return await methodChannel.invokeMethod('presentCredential',
          <String, dynamic>{'selectedCredentials': selectedCredentials, 'customScopeList': customScopeList});
    } on PlatformException catch (error) {
      debugPrint(error.toString());
      rethrow;
    }
  }

  Future<String?> verifyIssuer() async {
    try {
      return await methodChannel.invokeMethod<String>('verifyIssuer');
    } on PlatformException catch (error) {
      debugPrint(error.toString());
      if (error.code == errorCode) {
        debugPrint(error.toString());
        return error.details.toString();
      }
    }
    return null;
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
      return await methodChannel.invokeMethod<String>('getIssuerID', <String, dynamic>{'credentials': credentials});
    } on PlatformException catch (error) {
      if (error.code == errorCode) {
        return error.details.toString();
      }
    }
    return null;
  }

  Future<CredentialOfferDisplayData> getCredentialOfferDisplayData() async {
    final offerDisplayData = await methodChannel.invokeMethod('getCredentialOfferDisplayData');
    return CredentialOfferDisplayData.fromMap(offerDisplayData.cast<String, dynamic>());
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
