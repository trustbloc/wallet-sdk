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
  external String get did;

  external String get didDoc;
}

@JS()
@staticInterop
class JSResolvedDisplayData {}

extension JSResolvedDisplayDataExt on JSResolvedDisplayData {
  external String get issuerName;

  external List<dynamic> get credentialDisplays;
}

@JS()
@staticInterop

class JSVerifierDisplayData {}

extension JSVerifierDisplayDataExt on JSVerifierDisplayData {
  external String get name;
  external String get did;
  external String get logoURI;
  external String get purpose;
}

@JS()
@staticInterop

class JSWellKnownDIDConfig {}

extension JSWellKnownDIDConfigExt on JSWellKnownDIDConfig {
  external bool get isValid;
  external String get serviceURL;
}

@JS()
@staticInterop
class JSCredentialDisplayData {}

extension JSCredentialDisplayDataExt on JSCredentialDisplayData {
  external String get name;

  external String get logo;

  external String get backgroundColor;

  external String get textColor;

  external List<dynamic> get claims;
}

@JS()
@staticInterop
class JSClaimsDisplayData {}

extension JSClaimsDisplayDataExt on JSClaimsDisplayData {
  external String get rawValue;

  external String get valueType;

  external String get label;

  external String? get value;

  external int? get order;
}

@JS()
@staticInterop
class JSSubmissionRequirement {}

extension JSSubmissionRequirementExt on JSSubmissionRequirement {
  external String get name;

  external String get purpose;

  external String get rule;

  external int get count;

  external int get min;

  external int get max;

  external List<dynamic> descriptors;
  external List<dynamic> nested;
}

@JS()
@staticInterop
class JSInputDescriptor {}

extension JSInputDescriptorExt on JSInputDescriptor {
  external String get id;
  external String get name;
  external String get purpose;
  external List<String> get matchedVCsID;
  external List<dynamic> get matchedVCs;
}

@JS()
@staticInterop
class JSSupportedCredentials {}

extension JSSupportedCredentialsExt on JSSupportedCredentials {

  external String get format;

  external List<dynamic> get types;

  external List<dynamic> get display;
}

@JS()
@staticInterop
class JSIssuerMetadataLogo {}

extension JSIssuerMetadataLogoExt on JSIssuerMetadataLogo {
  external String get alt_text;
  external String get url;
}

@JS()
@staticInterop
class JSIssuerDisplayData {}

extension JSIssuerDisplayDataExt on JSIssuerDisplayData {
  external String get name;
  external String get locale;
  external String get url;
  external JSIssuerMetadataLogo? get logo;
  external String? get text_color;
  external String? get background_color;
}

@JS()
@staticInterop
class JSSupportedCredentialDisplayData {}

extension JSSupportedCredentialDisplayDataExt on JSSupportedCredentialDisplayData {
  external String get name;
  external String get locale;
  external JSIssuerMetadataLogo? get logo;
  external String? get text_color;
  external String? get background_color;
}

@JS()
@staticInterop
class JSIssuerMetadata {}

extension JSIssuerMetadataExt on JSIssuerMetadata {
  external String get credentialIssuer;
  external List<dynamic> get supportedCredentials;
  external List<dynamic> get localizedIssuerDisplays;
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

@JS()
external dynamic jsParseResolvedDisplayData(String resolvedCredentialDisplayData);

@JS()
external dynamic jsCreateOpenID4VPInteraction(authorizationRequest);

@JS()
external dynamic jsGetSubmissionRequirements(credentials);

@JS()
external dynamic jsPresentCredential(credentials);

@JS()
external dynamic jsVerifierDisplayData();

@JS()
external dynamic jsVerifyCredentialsStatus(String credential);

@JS()
external dynamic jsWellKnownDidConfig(String issuerID);

@JS()
external dynamic jsGetIssuerMetadata();

class WalletSDK extends WalletPlatform {
  @visibleForTesting
  final methodChannel = const MethodChannel('WalletSDKPlugin');
  final errorCode = 'Exception';

  Future<void> initSDK(String didResolverURI) async {
    await promiseToFuture(jsInitSDK(didResolverURI));
  }

  Future<CreateDID> createDID(String didMethodType, String didKeyType) async {
    DidDocResolution result = await promiseToFuture(jsCreateDID(didMethodType, didKeyType));
    return CreateDID(did: result.did, didDoc: result.didDoc);
  }

  Future<String?> fetchStoredDID(String didID) async {
    final fetchDIDMsg = await methodChannel.invokeMethod<String>('fetchDID', <String, dynamic>{'didID': didID});
    return fetchDIDMsg;
  }

  Future<Map<String, dynamic>> initialize(String initiateIssuanceURI, Map<String, dynamic>? authCodeArgs) async {
    try {
      final result =
          await promiseToFuture(jsCreateOpenID4CIInteraction(initiateIssuanceURI)) as CreateOpenID4CIInteractionResult;

      return {'pinRequired': result.userPINRequired};
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

  Future<WalletSDKError> parseWalletSDKError({required String localizedErrorMessage}) async {
    return throw Exception('Method not implemented');
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

  Future<bool> credentialStatusVerifier(String credential) async {
    try {
      await promiseToFuture(jsVerifyCredentialsStatus(credential));
      return true;
    } catch (error) {
      if (error.toString().contains('status verification failed: revoked')) {
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

  Future<List<CredentialDisplayData>> parseCredentialDisplayData(String resolvedCredentialDisplayData) async {
    final JSResolvedDisplayData data = await promiseToFuture(jsParseResolvedDisplayData(resolvedCredentialDisplayData));

    return data.credentialDisplays
        .map((e) => e as JSCredentialDisplayData)
        .map((cred) => CredentialDisplayData(
            issuerName: data.issuerName,
            overviewName: cred.name,
            logo: cred.logo,
            backgroundColor: cred.backgroundColor,
            textColor: cred.textColor,
            claims: cred.claims
                .map((e) => e as JSClaimsDisplayData)
                .map((claim) => CredentialDisplayClaim(
                    rawValue: claim.rawValue,
                    valueType: claim.valueType,
                    label: claim.label,
                    value: claim.value,
                    order: claim.order))
                .toList()))
        .toList();
  }

  Future<void> processAuthorizationRequest(
      {required String authorizationRequest, List<String>? storedCredentials}) async {
    await promiseToFuture(jsCreateOpenID4VPInteraction(authorizationRequest));
    return;
  }

  Future<List<SubmissionRequirement>> getSubmissionRequirements({required List<String> storedCredentials}) async {
    final List<dynamic> reqs = await promiseToFuture(jsGetSubmissionRequirements(storedCredentials));

    return Future.wait(reqs.map((s) => mapSubmissionRequirement(s)));
  }

  Future<InputDescriptor> mapInputDescriptor(JSInputDescriptor d) async {
    final List<dynamic> jsMatchedVCs = d.matchedVCs.map((e) => e).toList();

    final matchedVCs = jsMatchedVCs.map<String>((vc) => vc).toList();
    final matchedVCsIds = await Future.wait<String>(jsMatchedVCs.map((vc) => promiseToFuture(jsGetCredentialID(vc))));

    return InputDescriptor(
        name: d.name, purpose: d.purpose, id: d.id, matchedVCs: matchedVCs, matchedVCsID: matchedVCsIds
    );
  }

  Future<SubmissionRequirement> mapSubmissionRequirement(JSSubmissionRequirement jsReq) async {
    final descriptors = await Future.wait(jsReq.descriptors.map((e) => mapInputDescriptor(e)));
    final nested = await Future.wait(jsReq.nested.map((n) => mapSubmissionRequirement(n)));

    return SubmissionRequirement(
      name: jsReq.name,
      rule: jsReq.rule,
      count: jsReq.count,
      min: jsReq.min,
      max: jsReq.max,
      inputDescriptors: descriptors,
      nested: nested,
    );
  }

  Future<Map<Object?, Object?>?> getVersionDetails() async {
    var versionDetailResp = await methodChannel.invokeMethod('getVersionDetails');
    log('getVersionDetails in the app, $versionDetailResp');
    return versionDetailResp;
  }

  Future<WellKnownDidConfig> wellKnownDidConfig(String issuerID) async {
    final JSWellKnownDIDConfig jsConfig = await promiseToFuture(jsWellKnownDidConfig(issuerID));

    return WellKnownDidConfig(isValid: jsConfig.isValid, serviceURL: jsConfig.serviceURL);
  }

  Future<List<IssuerMetaData>> getIssuerMetaData() async {
    final JSIssuerMetadata data = await promiseToFuture(jsGetIssuerMetadata());

    final supportedCredentials = data.supportedCredentials
        .map((e) => e as JSSupportedCredentials)
        .map((supportedCredential) => SupportedCredentials(
            format: supportedCredential.format,
            types: supportedCredential.types.map((e) => e as String).toList(),
            display: supportedCredential.display
                .map((e) => e as JSSupportedCredentialDisplayData)
                .map((supportedCredentialDisplayData) => SupportedCredentialDisplayData(
                  name: supportedCredentialDisplayData.name,
                  locale: supportedCredentialDisplayData.locale,
                  logo: supportedCredentialDisplayData.logo?.url,
                  textColor: supportedCredentialDisplayData.text_color,
                  backgroundColor: supportedCredentialDisplayData.background_color
                  )
          ).toList(),
        )).toList();


    final localizedIssuerDisplays = data.localizedIssuerDisplays
        .map((e) => e as JSIssuerDisplayData)
        .map((localizedIssuerDisplay) => IssuerDisplayData(
          name: localizedIssuerDisplay.name,
          locale: localizedIssuerDisplay.locale,
          url: localizedIssuerDisplay.url,
          logo: localizedIssuerDisplay.logo?.url,
          textColor: localizedIssuerDisplay.text_color,
          backgroundColor: localizedIssuerDisplay.background_color
    )).toList();

    final List<IssuerMetaData> issuerMetadata = [IssuerMetaData(
      credentialIssuer: data.credentialIssuer,
      supportedCredentials: supportedCredentials,
      localizedIssuerDisplays: localizedIssuerDisplays,
    )];

    debugPrint('Issuer Metadata: $issuerMetadata');

    return issuerMetadata;
  }

  Future<VerifierDisplayData> getVerifierDisplayData() async {
    final JSVerifierDisplayData data = await promiseToFuture(jsVerifierDisplayData());
    return VerifierDisplayData(name: data.name, did: data.did, logoURI: data.logoURI, purpose: data.purpose);
  }

  Future<void> presentCredential({required List<String> selectedCredentials}) async {
    await promiseToFuture(jsPresentCredential(selectedCredentials));
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
      log('get issuerID - , $issuerID');
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
