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
import 'package:app/views/handle_redirect_uri.dart';
import 'package:flutter/services.dart';
import 'dart:async';
import 'package:app/models/credential_offer.dart';
import 'package:http/http.dart' as http;

void handleOpenIDIssuanceFlow(BuildContext context, String qrCodeURL) async {
  var WalletSDKPlugin = MethodChannelWallet();
  final StorageService storageService = StorageService();
  final Future<SharedPreferences> prefs = SharedPreferences.getInstance();
  var authCodeArgs;
  if (qrCodeURL.contains("credential_offer_uri")){
    authCodeArgs = await parseCredentialOfferUri(qrCodeURL);
    log("credential offer uri auth code  ${authCodeArgs}");
  } else {
    if (qrCodeURL.contains("authorization_code")){
      authCodeArgs = await readIssuerAuthFlowConfig(qrCodeURL);
      log("auth code arguments fetched from config file ${authCodeArgs}");
    }
  }

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

 readIssuerAuthFlowConfig(String qrCodeURL) async {
  var decodedUri = Uri.decodeComponent(qrCodeURL);
  final uri = Uri.parse(decodedUri);
  var credentialIssuerKey = json.decode(uri.queryParameters["credential_offer"]!);
  final String response = await rootBundle.loadString('lib/assets/issuerAuthFlowConfig.json');
  final configData = await json.decode(response);
  return configData[credentialIssuerKey["credential_issuer"]];
}

parseCredentialOfferUri(String qrCodeURL) async {
  var decodedUri = Uri.decodeComponent(qrCodeURL);
  final uri = Uri.parse(decodedUri);
  final response = await http
      .get(Uri.parse(uri.queryParameters['credential_offer_uri']!));
  if (response.statusCode == 200) {
    final String configResp = await rootBundle.loadString('lib/assets/issuerAuthFlowConfig.json');
    final configData = await json.decode(configResp);
    var resp = CredentialOfferObject.fromJson(jsonDecode(response.body));
    return configData[resp.credentialIssuer];
  } else {
    throw Exception('Failed to load credential offer uri');
  }
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

  navigateToCredPreviewScreen(context, credentials, issuerURL, resolvedCredentialDisplay!, didID, credID);
}

void navigateToOTPScreen(BuildContext context) async {
  Navigator.push(context, MaterialPageRoute(builder: (context) => const OTP()));
}

navigateToCredPreviewScreen(
    BuildContext context, String credentialResp, String issuerURL, String resolvedCredentialDisplay, String didID, String credID) async {
  Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) =>
            CredentialPreview(credentialData: CredentialData(rawCredential: credentialResp, issuerURL: issuerURL, credentialDisplayData: resolvedCredentialDisplay, credentialDID: didID, credID: credID)),));
}


void navigateToAuthFlow(BuildContext context, Uri uri) async {
  Navigator.of(context).push(
      MaterialPageRoute(
          builder: (context) => HandleRedirectUri(uri)
      ));
}