
import 'dart:developer';
import 'package:app/demo_method_channel.dart';
import 'package:app/models/credential_data.dart';
import 'package:app/models/credential_data_object.dart';
import 'package:app/views/presentation_preview.dart';
import 'package:app/views/presentation_preview_multi_cred.dart';
import 'package:flutter/material.dart';
import 'package:app/views/otp.dart';
import 'package:app/services/storage_service.dart';
import 'package:app/models/store_credential_data.dart';
import 'package:app/views/credential_preview.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../models/activity_data_object.dart';

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

void _navigateToPresentMultiCred(
    BuildContext context, List<CredentialData> credentialData) async {
  Navigator.push(
      context,
      MaterialPageRoute(
          builder: (context) =>
              PresentationPreviewMultiCred(credentialData: credentialData)));
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
  final Future<SharedPreferences> prefs = SharedPreferences.getInstance();
  late List<CredentialDataObject> storedCredentials;
  late List<String> credentials;
  log('received qr code url - $qrCodeURL');
  if (!qrCodeURL.contains("openid-vc")) {
    var authorizeResultPinRequired = await WalletSDKPlugin.authorize(qrCodeURL);
    log("pin required flow -  $authorizeResultPinRequired");
    if (authorizeResultPinRequired == true) {
      _navigateToOTPScreen(context);
      return;
    } else {
      final SharedPreferences pref = await prefs;
      var didID = await WalletSDKPlugin.createDID("jwk");
      log("created didID :$didID");
      pref.setString('userDID',didID!);
      String? credentials =  await WalletSDKPlugin.requestCredential('');
      String? issuerURL = await WalletSDKPlugin.issuerURI();
      String? resolvedCredentialDisplay =  await WalletSDKPlugin.resolveCredentialDisplay([credentials],issuerURL!);
      var activities = await WalletSDKPlugin.storeActivityLogger();
      var credID = await WalletSDKPlugin.getCredID([credentials]);
      log("activities and credID handle open id  -$activities and $credID");
      storageService.addActivities(ActivityDataObj(credID!, activities));
      _navigateToCredPreviewScreen(context, credentials, issuerURL, resolvedCredentialDisplay!);
    }
  } else {
    // Check if the flow is for the verifiable presentation or for issuance.
    UserLoginDetails userLoginDetails = await getUser();
    var username = userLoginDetails.username!;
    storedCredentials = await storageService.retrieveCredentials(username!);
    credentials = storedCredentials.map((e) => e.value.rawCredential).toList();
    var matchedCred = await WalletSDKPlugin.processAuthorizationRequest(
        authorizationRequest: qrCodeURL, storedCredentials: credentials);
    // Get the matched VCIDs from the submission request.
    var getSubmissionRequest = await WalletSDKPlugin.getSubmissionRequirements(storedCredentials: credentials);
    var matchedVCIds = handleSubmissionRequirement(getSubmissionRequest);

    if (matchedVCIds.length > 1) {
      // multiple matched vc ids are found therefore, invoking multiple credential Presentation Preview.
      log("multi cred flow");
      var credentialDisplayData = storedCredentials
          .where((element) => matchedCred.contains(element.value.rawCredential))
          .map((e) =>
          CredentialData(rawCredential: e.value.rawCredential, issuerURL: e.value.issuerURL, credentialDisplayData: e.value.credentialDisplayData))
          .toList();
      _navigateToPresentMultiCred(context, credentialDisplayData);
      return;
    }
    for (var cred in credentials) {
      for (var matchCred in matchedCred){
        log("choose flow ${credentials.first != matchedCred.first}");
        if (cred != matchCred){
          log("selective disclosure flow - credentialDisplayData as credential doesnt match matched credentials");
          var issuerURI = storedCredentials.map((e) => e.value.issuerURL).toList();
          var credentialDisplayData =  await WalletSDKPlugin.resolveCredentialDisplay(matchedCred, issuerURI.first);
          _navigateToPresentationPreviewScreen(context, matchedCred.first,
              CredentialData(rawCredential: '', credentialDisplayData: credentialDisplayData!, issuerURL: ''));
          return;
        } else {
          log("single matched vc id flow");
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

  }
}


List<String> handleSubmissionRequirement(List<SubmissionRequirement> submissionRequirement){
     var submission =  submissionRequirement.first;
     return submission.inputDescriptors.first.matchedVCsID;
}