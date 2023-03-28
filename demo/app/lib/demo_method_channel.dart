import 'dart:convert';
import 'dart:developer';
import 'dart:ffi';

import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';

import 'demo_platform_interface.dart';

class SubmissionRequirement {
  final String rule;
  final String name;
  final int min;
  final int max;
  final int count;
  final List<SubmissionRequirement> nested;
  final List<InputDescriptor> inputDescriptors;

//<editor-fold desc="Data Methods">
  const SubmissionRequirement({
    required this.rule,
    required this.name,
    required this.min,
    required this.max,
    required this.count,
    required this.nested,
    required this.inputDescriptors,
  });

  @override
  String toString() {
    return 'SubmissionRequirement{ rule: $rule, name: $name, min: $min, max: $max, count: $count, nested: $nested, inputDescriptors: $inputDescriptors,}';
  }

  factory SubmissionRequirement.fromMap(Map<String, dynamic> map) {
    return SubmissionRequirement(
      rule: map['rule'] as String,
      name: map['name'] as String,
      min: map['min'] as int,
      max: map['max'] as int,
      count: map['count'] as int,
      nested: (map['nested'] as List<dynamic>)
          .map((obj) => SubmissionRequirement.fromMap(obj.cast<String, dynamic>()))
          .toList(),
      inputDescriptors: (map['inputDescriptors'] as List<dynamic>)
          .map((obj) => InputDescriptor.fromMap(obj.cast<String, dynamic>()))
          .toList(),
    );
  }

//</editor-fold>
}

class InputDescriptor {
  final String id;
  final String name;
  final String purpose;
  final List<String> matchedVCsID;

  const InputDescriptor({
    required this.id,
    required this.name,
    required this.purpose,
    required this.matchedVCsID,
  });

  @override
  String toString() {
    return 'InputDescriptor{ id: $id, name: $name, purpose: $purpose, matchedVCsID: $matchedVCsID,}';
  }

  factory InputDescriptor.fromMap(Map<String, dynamic> map) {
    return InputDescriptor(
      id: map['id'] as String,
      name: map['name'] as String,
      purpose: map['purpose'] as String,
      matchedVCsID: map['matchedVCsID'].cast<String>(),
    );
  }
}

class MethodChannelWallet extends WalletPlatform {
  @visibleForTesting
  final methodChannel = const MethodChannel('WalletSDKPlugin');
  final errorCode = 'Exception';

  Future<void> initSDK() async {
    await methodChannel.invokeMethod<bool>('initSDK');
  }

  Future<Map<Object?, Object?>?> createDID(String didMethodType) async {
    final createDIDMsg =
        await methodChannel.invokeMethod<Map<Object?, Object?>?>('createDID', <String, dynamic>{'didMethodType': didMethodType});
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
      var credentialResponse =
          await methodChannel.invokeMethod<String>('requestCredential', <String, dynamic>{'otp': userPinEntered});
      return credentialResponse!;
    } on PlatformException catch (error) {
      debugPrint(error.toString());
      rethrow;
    }
  }

  Future<bool?> credentialStatusVerifier(List<String> credentials) async {
    log("inside credentialStatusVerifier");
    try {
      var credentialStatusVerifier =
      await methodChannel.invokeMethod<bool>('credentialStatusVerifier', <String, dynamic>{'credentials': credentials});
      return credentialStatusVerifier!;
    } on PlatformException catch (error) {
      if (error.toString().contains("status verification failed: revoked")){
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
    try {
      final credentialResponse = await methodChannel.invokeMethod<String>(
          'serializeDisplayData', <String, dynamic>{'vcCredentials': credentials, 'uri': issuerURI});
      return credentialResponse;
    } on PlatformException catch (error) {
      if (error.code == errorCode) {
        return error.details.toString();
      }
    }
    return null;
  }

  Future<List<Object?>> resolveCredentialDisplay(String resolvedCredentialDisplayData) async {
    var renderedCredDisplay = await methodChannel.invokeMethod('resolveCredentialDisplay', <String, dynamic>{'resolvedCredentialDisplayData': resolvedCredentialDisplayData});
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
   var versionDetailResp = await methodChannel
        .invokeMethod('getVersionDetails');
   log("getVersionDetails in the app, $versionDetailResp");
   return versionDetailResp;
  }

  Future<void> presentCredential({List<String>? selectedCredentials}) async {
    await methodChannel
        .invokeMethod('presentCredential', <String, dynamic>{'selectedCredentials': selectedCredentials});
  }

  Future<List<Object?>> storeActivityLogger() async {
    var activityObj = await methodChannel.invokeMethod('activityLogger');
    return activityObj;
  }

  Future<List<Object?>> parseActivities(List activities) async {
    var activityObj = await methodChannel.invokeMethod('parseActivities', <String, dynamic>{'activities': activities});
    return activityObj;
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
