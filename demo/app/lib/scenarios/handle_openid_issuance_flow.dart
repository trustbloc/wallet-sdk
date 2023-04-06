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

void handleOpenIDIssuanceFlow(BuildContext context, String qrCodeURL) async {
  var WalletSDKPlugin = MethodChannelWallet();
  final StorageService storageService = StorageService();
  final Future<SharedPreferences> prefs = SharedPreferences.getInstance();
  var authorizeResultPinRequired = await WalletSDKPlugin.authorize(qrCodeURL);
  log("pin required flow -  $authorizeResultPinRequired");
  if (authorizeResultPinRequired == true) {
    navigateToOTPScreen(context);
    return;
  } else {
    final SharedPreferences pref = await prefs;
    var didType = pref.getString('didType');
    var keyType = pref.getString('keyType');
    // choosing default if no selection is made
    didType = didType ?? "jwk";
    keyType = keyType ?? "ECDSAP384IEEEP1363";
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