/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'dart:convert';
import 'dart:developer';
import 'dart:math' as math;

import 'package:app/wallet_sdk/wallet_sdk_model.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';

import '../services/config_service.dart';
import 'wallet_sdk_interface.dart';

class WalletSDK extends WalletPlatform {
  @visibleForTesting
  final methodChannel = const MethodChannel('WalletSDKPlugin');
  final errorCode = 'Exception';

  Future<void> initSDK(String didResolverURI) async {
    await methodChannel.invokeMethod<bool>(
        'initSDK', <String, dynamic>{'didResolverURI': didResolverURI});
  }

  Future<CreateDID> createDID(String didMethodType, String didKeyType) async {
    final createDIDResp = await methodChannel
        .invokeMethod<Map<Object?, Object?>?>('createDID', <String, dynamic>{
      'didMethodType': didMethodType,
      'didKeyType': didKeyType
    });
    return CreateDID.fromJson(jsonDecode(json.encode(createDIDResp)));
  }

  Future<String?> fetchStoredDID(String didID) async {
    final fetchDIDMsg = await methodChannel
        .invokeMethod<String>('fetchDID', <String, dynamic>{'didID': didID});
    return fetchDIDMsg;
  }

  Future<Map<Object?, Object?>?> initialize(
      String qrCode, Map<String, dynamic>? authCodeArgs) async {
    try {
      var correlationID = generateRandomString(8);

      debugPrint('initialize->correlationID: $correlationID');

      final flowTypeData =
          await methodChannel.invokeMethod('initialize', <String, dynamic>{
        'requestURI': qrCode,
        'authCodeArgs': authCodeArgs,
        'correlationID': correlationID
      });
      return flowTypeData;
    } on PlatformException catch (error) {
      debugPrint(error.toString());
      rethrow;
    }
  }

  Future<List<SupportedCredentials>> initializeWalletInitiatedFlow(
      String issuerURI, List<String> credentialTypes) async {
    try {
      List<dynamic> supportedCredentialResp = await methodChannel.invokeMethod(
          'initializeWalletInitiatedFlow', <String, dynamic>{
        'issuerURI': issuerURI,
        'credentialTypes': credentialTypes
      });
      return supportedCredentialResp
          .map((d) => SupportedCredentials.fromMap(d.cast<String, dynamic>()))
          .toList();
    } on PlatformException catch (error) {
      debugPrint(error.toString());
      rethrow;
    }
  }

  Future<String> createAuthorizationURLWalletInitiatedFlow(
      List<String> scopes,
      List<String> credentialTypes,
      String credentialFormat,
      clientID,
      redirectURI,
      issuerURI) async {
    try {
      String authorizationURL = await methodChannel.invokeMethod(
          'createAuthorizationURLWalletInitiatedFlow', <String, dynamic>{
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

  Future<List<CredentialWithId>> requestCredential(String userPinEntered,
      {String? attestationVC}) async {
    try {
      List<dynamic> credentialResponse = await methodChannel
          .invokeMethod('requestCredentials', <String, dynamic>{
        'otp': userPinEntered,
        'attestationVC': attestationVC,
      });

      return credentialResponse
          .map((e) => CredentialWithId.fromMap(e.cast<String, dynamic>()))
          .toList();
    } on PlatformException catch (error) {
      debugPrint(error.toString());
      rethrow;
    }
  }

  Future<bool> requireAcknowledgment() async {
    var ackResp =
        await methodChannel.invokeMethod<bool>('requireAcknowledgment');
    return ackResp!;
  }

  Future<String> noConsentAcknowledgement() async {
    var data = await methodChannel.invokeMethod('noConsentAcknowledgement');
    return data;
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

  Future<WalletSDKError> parseWalletSDKError(
      {required String localizedErrorMessage}) async {
    var parsedWalletError = await methodChannel.invokeMethod(
        'parseWalletSDKError',
        <String, dynamic>{'localizedErrorMessage': localizedErrorMessage});
    return WalletSDKError.fromJson(jsonDecode(json.encode(parsedWalletError)));
  }

  Future<String> requestCredentialWithAuth(String redirectURIWithParams) async {
    try {
      var credentialResponse = await methodChannel.invokeMethod<String>(
          'requestCredentialWithAuth',
          <String, dynamic>{'redirectURIWithParams': redirectURIWithParams});
      return credentialResponse!;
    } on PlatformException catch (error) {
      debugPrint(error.toString());
      rethrow;
    }
  }

  Future<String> requestCredentialWithWalletInitiatedFlow(
      String redirectURIWithParams) async {
    try {
      var credentialResponse = await methodChannel.invokeMethod<String>(
          'requestCredentialWithWalletInitiatedFlow',
          <String, dynamic>{'redirectURIWithParams': redirectURIWithParams});
      return credentialResponse!;
    } on PlatformException catch (error) {
      debugPrint(error.toString());
      rethrow;
    }
  }

  Future<bool> credentialStatusVerifier(String credential) async {
    try {
      var credentialStatusVerifier = await methodChannel
          .invokeMethod<bool>('credentialStatusVerifier', <String, dynamic>{
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

  Future<CredentialsDisplayData> resolveDisplayData(
      List<String> credentials, String issuerURI) async {
    final credentialResponse = await methodChannel.invokeMethod(
        'resolveDisplayData',
        <String, dynamic>{'vcCredentials': credentials, 'uri': issuerURI});

    return CredentialsDisplayData.fromMap(
        credentialResponse.cast<String, dynamic>());
  }

  Future<CredentialDisplayData> parseCredentialDisplayData(
      String resolvedCredentialDisplayData) async {
    final renderedCredDisplay = await methodChannel.invokeMethod(
        'parseCredentialDisplay', <String, dynamic>{
      'resolvedCredentialDisplayData': resolvedCredentialDisplayData
    });
    return CredentialDisplayData.fromMap(
        renderedCredDisplay.cast<String, dynamic>());
  }

  Future<IssuerDisplayData> parseIssuerDisplay(String issuerDisplayData) async {
    var issuerDisplay = await methodChannel.invokeMethod('parseIssuerDisplay',
        <String, dynamic>{'issuerDisplayData': issuerDisplayData});
    return IssuerDisplayData.fromMap(issuerDisplay.cast<String, dynamic>());
  }

  Future<List<String>> processAuthorizationRequest(
      {required String authorizationRequest,
      List<String>? storedCredentials}) async {
    var correlationID = generateRandomString(8);

    debugPrint('processAuthorizationRequest->correlationID: $correlationID');

    return (await methodChannel.invokeMethod<List>(
            'processAuthorizationRequest', <String, dynamic>{
      'authorizationRequest': authorizationRequest,
      'storedCredentials': storedCredentials,
      'correlationID': correlationID
    }))!
        .map((e) => e!.toString())
        .toList();
  }

  Future<List<Object?>> getCustomScope() async {
    return await methodChannel.invokeMethod('getCustomScope');
  }

  Future<List<SubmissionRequirement>> getSubmissionRequirements(
      {required List<String>? storedCredentials}) async {
    return (await methodChannel.invokeMethod<List<dynamic>>(
            'getMatchedSubmissionRequirements',
            <String, dynamic>{'storedCredentials': storedCredentials}))!
        .map(
            (obj) => SubmissionRequirement.fromMap(obj.cast<String, dynamic>()))
        .toList();
  }

  Future<String> getAttestationVC(
      {required String attestationURL,
      bool disableTLSVerify = false,
      required String attestationPayload,
      String? attestationToken}) async {
    var attestationVC = await methodChannel
        .invokeMethod<String>('getAttestationVC', <String, dynamic>{
      'attestationURL': attestationURL,
      'disableTLSVerify': disableTLSVerify,
      'attestationPayload': attestationPayload,
      'attestationToken': attestationToken,
    });
    return attestationVC!;
  }

  Future<Map<Object?, Object?>?> getVersionDetails() async {
    var versionDetailResp =
        await methodChannel.invokeMethod('getVersionDetails');
    log('getVersionDetails in the app, $versionDetailResp');
    return versionDetailResp;
  }

  Future<WellKnownDidConfig> wellKnownDidConfig(String issuerID) async {
    Map<String, dynamic> config = (await methodChannel.invokeMethod(
            'wellKnownDidConfig', <String, dynamic>{'issuerID': issuerID}))
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
    var data = await methodChannel.invokeMethod('evaluateIssuanceTrustInfo',
        {'evaluateIssuanceURL': ConfigService.config.evaluateIssuanceURL});
    return EvaluationResult.fromMap(data.cast<String, dynamic>());
  }

  Future<EvaluationResult?> evaluatePresentationTrustInfo() async {
    var data = await methodChannel.invokeMethod(
        'evaluatePresentationTrustInfo', {
      'evaluatePresentationURL': ConfigService.config.evaluatePresentationURL
    });
    return EvaluationResult.fromMap(data.cast<String, dynamic>());
  }

  Future<void> presentCredential(
      {required List<String> selectedCredentials,
      Map<String, dynamic>? customScopeList,
      String? attestationVC}) async {
    try {
      return await methodChannel
          .invokeMethod('presentCredential', <String, dynamic>{
        'selectedCredentials': selectedCredentials,
        'customScopeList': customScopeList,
        'attestationVC': attestationVC,
      });
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
    var activityObj = await methodChannel.invokeMethod(
        'parseActivities', <String, dynamic>{'activities': activities});
    return activityObj;
  }

  Future<String?> getIssuerID(List<String> credentials) async {
    try {
      var correlationID = generateRandomString(8);

      debugPrint('getIssuerID->correlationID: $correlationID');

      return await methodChannel.invokeMethod<String>(
          'getIssuerID', <String, dynamic>{
        'credentials': credentials,
        'correlationID': correlationID
      });
    } on PlatformException catch (error) {
      if (error.code == errorCode) {
        return error.details.toString();
      }
    }
    return null;
  }

  Future<CredentialOfferDisplayData> getCredentialOfferDisplayData() async {
    final offerDisplayData =
        await methodChannel.invokeMethod('getCredentialOfferDisplayData');
    return CredentialOfferDisplayData.fromMap(
        offerDisplayData.cast<String, dynamic>());
  }

  Future<String?> getCredID(List<String> credentials) async {
    try {
      final credentialID = await methodChannel.invokeMethod<String>(
          'getCredID', <String, dynamic>{'vcCredentials': credentials});
      return credentialID;
    } on PlatformException catch (error) {
      if (error.code == errorCode) {
        return error.details.toString();
      }
    }
    return null;
  }
}

String generateRandomString(int length) {
  const characters =
      'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
  math.Random random = math.Random();
  StringBuffer randomString = StringBuffer();

  for (int i = 0; i < length; i++) {
    randomString.write(characters[random.nextInt(characters.length)]);
  }

  return randomString.toString();
}
