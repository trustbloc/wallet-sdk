import 'dart:convert';
import 'dart:developer';
import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:app/demo_method_channel.dart';
import 'package:app/models/activity_data_object.dart';
import 'package:app/models/credential_data.dart';
import 'package:app/services/storage_service.dart';
import 'package:app/views/credential_preview.dart';
import 'package:app/views/otp.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:app/views/handle_redirect_uri.dart';

void handleOpenIDIssuanceFlow(BuildContext context, String qrCodeURL) async {
  var WalletSDKPlugin = MethodChannelWallet();
  final StorageService storageService = StorageService();
  final Future<SharedPreferences> prefs = SharedPreferences.getInstance();
  var authCodeArgs = readArgs();
  log("auth code supplied arguments  ${authCodeArgs}");
  var flowTypeData = await WalletSDKPlugin.initialize(qrCodeURL, authCodeArgs);
  var flowTypeDataEncoded = json.encode(flowTypeData);
  Map<String, dynamic> responseJson = json.decode(flowTypeDataEncoded);
  var authorizeResultPinRequired = responseJson["pinRequired"];
  log("pin required flow -  $authorizeResultPinRequired");
  if (authorizeResultPinRequired == true) {
    navigateToOTPScreen(context);
    return;
  } else if (responseJson["authorizationURLLink"] != '') {
    // initiate authCode Flow
    log("initiating authCode Flow- ${responseJson["authorizationURLLink"]}");
    Uri uri = Uri.parse(responseJson["authorizationURLLink"]);
    navigateToAuthFlow(context, uri);
    return;
  } else {
    navigateToWithoutPinFlow(context);
    return;
  }
}

Map<String, String> readArgs() {
  const scope1 = String.fromEnvironment("scope1");
  const scope2 = String.fromEnvironment("scope2");
  const clientID = String.fromEnvironment("clientID");
  const redirectURI = String.fromEnvironment("redirectURI");
  Map<String, String> authCodeArgsMap = {
    'scope1': scope1,
    'scope2': scope2,
    'clientID': clientID,
    'redirectURI': redirectURI
  };

  return authCodeArgsMap;
}

void navigateToWithoutPinFlow(BuildContext context) async{
  final Future<SharedPreferences> prefs = SharedPreferences.getInstance();
  var WalletSDKPlugin = MethodChannelWallet();
  final StorageService storageService = StorageService();
  final SharedPreferences pref = await prefs;

  var didType = pref.getString('didType');
  var keyType = pref.getString('keyType');
  // choosing default if no selection is made
  didType = didType ?? "ion";
  keyType = keyType ?? "ED25519";

  var didResolution = await WalletSDKPlugin.createDID(didType, keyType);
  var didDocEncoded = json.encode(didResolution);
  Map<String, dynamic> responseJson = json.decode(didDocEncoded);

  var didID = responseJson["did"];
  var didDoc = responseJson["didDoc"];
  log("created didID :$didID");
  pref.setString('userDID',didID);
  pref.setString('userDIDDoc',didDoc);

  String? credentials =  await WalletSDKPlugin.requestCredential('');
  String? issuerURL = await WalletSDKPlugin.issuerURI();
  String? resolvedCredentialDisplay =  await WalletSDKPlugin.serializeDisplayData([credentials],issuerURL!);

  var activities = await WalletSDKPlugin.storeActivityLogger();

  var credID = await WalletSDKPlugin.getCredID([credentials]);

  log("activities and credID handle open id  -$activities and $credID");
  storageService.addActivities(ActivityDataObj(credID!, activities));

  navigateToCredPreviewScreen(context, credentials, issuerURL, resolvedCredentialDisplay!, didID);
}

void navigateToOTPScreen(BuildContext context) async {
  Navigator.push(context, MaterialPageRoute(builder: (context) => const OTP()));
}

navigateToCredPreviewScreen(
    BuildContext context, String credentialResp, String issuerURL, String resolvedCredentialDisplay, String didID) async {
  Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) =>
            CredentialPreview(credentialData: CredentialData(rawCredential: credentialResp, issuerURL: issuerURL, credentialDisplayData: resolvedCredentialDisplay, credentialDID: didID)),));
}


void navigateToAuthFlow(BuildContext context, Uri uri) async {
  Navigator.of(context).push(
      MaterialPageRoute(
          builder: (context) => HandleRedirectUri(uri)
      ));
}