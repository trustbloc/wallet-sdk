import 'dart:convert';
import 'dart:developer';

import 'package:app/demo_method_channel.dart';
import 'package:app/models/credential_data.dart';
import 'package:app/models/credential_data_object.dart';
import 'package:app/views/presentation_preview.dart';
import 'package:flutter/material.dart';
import 'package:app/views/otp.dart';
import 'package:app/services/storage_service.dart';
import 'package:app/models/store_credential_data.dart';

import 'package:app/views/credential_preview.dart';

void _navigateToOTPScreen(BuildContext context) async {
  Navigator.push(context, MaterialPageRoute(builder: (context) => const OTP()));
}

void _navigateToPresentationPreviewScreen(
    BuildContext context, String matchedCredential, CredentialData credentialData) async {
  Navigator.push(
      context,
      MaterialPageRoute(
          builder: (context) =>
              PresentationPreview(matchedCredential: matchedCredential, credentialData: credentialData)));
}

_navigateToCredPreviewScreen(
    BuildContext context, String credentialResp, String issuerURL, String resolvedCredentialDisplay) async {
  Navigator.push(
      context,
      MaterialPageRoute(
          builder: (context) =>
              CredentialPreview(credentialData: CredentialData(rawCredential: credentialResp, issuerURL: issuerURL, credentialDisplayData: resolvedCredentialDisplay)),));
}

void handleOpenIDUrl(BuildContext context, String qrCodeURL) async {
  var WalletSDKPlugin = MethodChannelWallet();

  final StorageService storageService = StorageService();
  late List<CredentialDataObject> storedCredentials;
  log('received qr code url - $qrCodeURL');
  if (!qrCodeURL.contains("openid-vc")) {
    var authorizeResultPinRequired = await WalletSDKPlugin.authorize(qrCodeURL);
    log("pin required flow -  $authorizeResultPinRequired");
    if (authorizeResultPinRequired == true) {
      _navigateToOTPScreen(context);
      return;
    } else {
      String? requestCredentialResp =  await WalletSDKPlugin.requestCredential('');
      String? issuerURL = await WalletSDKPlugin.issuerURI();
      String? resolvedCredentialDisplay =  await WalletSDKPlugin.resolveCredentialDisplay([requestCredentialResp],issuerURL!);
      _navigateToCredPreviewScreen(context, requestCredentialResp, issuerURL, resolvedCredentialDisplay!);
    }
  } else {
    // Check if the flow is for the verifiable presentation or for issuance.
    UserLoginDetails userLoginDetails = await getUser();
    var username = userLoginDetails.username!;
    storedCredentials = await storageService.retrieveCredentials(username!);
    var credentials = storedCredentials.map((e) => e.value.rawCredential).toList();
    var matchedCred = await WalletSDKPlugin.processAuthorizationRequest(
        authorizationRequest: qrCodeURL, storedCredentials: credentials);
    if (credentials != matchedCred){
      var issuerURI = storedCredentials.map((e) => e.value.issuerURL).toList();
     var credentialDisplayData =  await WalletSDKPlugin.resolveCredentialDisplay(matchedCred, issuerURI.first);
      log("selective disclosure flow - credentialDisplayData");
      _navigateToPresentationPreviewScreen(context, matchedCred.first,
          CredentialData(rawCredential: '', credentialDisplayData: credentialDisplayData!, issuerURL: ''));
      return;
    } else {
      log("regular flow");
      var credentialDisplayData = storedCredentials
          .where((element) => matchedCred.contains(element.value.rawCredential))
          .map((e) => e.value.credentialDisplayData);
      log(matchedCred.length.toString());
      if (matchedCred.isNotEmpty) {
        //TODO: in future we can show all the credential
        _navigateToPresentationPreviewScreen(context, matchedCred.first,
            CredentialData(rawCredential: '', credentialDisplayData: credentialDisplayData.first, issuerURL: ''));
        return;
      }
    }

  }
}
