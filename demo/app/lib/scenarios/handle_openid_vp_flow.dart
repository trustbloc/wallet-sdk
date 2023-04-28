import 'dart:developer';
import 'package:app/demo_method_channel.dart';
import 'package:app/views/custom_error.dart';
import 'package:flutter/material.dart';
import 'package:app/models/credential_data.dart';
import 'package:app/models/credential_data_object.dart';
import 'package:app/models/store_credential_data.dart';
import 'package:app/services/storage_service.dart';
import 'package:app/views/presentation_preview.dart';
import 'package:app/views/presentation_preview_multi_cred.dart';
import 'package:app/views/presentation_preview_multi_cred_radio.dart';

void handleOpenIDVpFlow(BuildContext context, String qrCodeURL) async {
  var WalletSDKPlugin = MethodChannelWallet();
  final StorageService storageService = StorageService();
  late List<CredentialDataObject> storedCredentials;
  late List<String> credentials;

  // Check if the flow is for the verifiable presentation or for issuance.
  UserLoginDetails userLoginDetails = await getUser();
  var username = userLoginDetails.username!;
  storedCredentials = await storageService.retrieveCredentials(username!);
  log("stored credentials -> $storedCredentials");

  credentials = storedCredentials.map((e) => e.value.rawCredential).toList();

  if (credentials.isEmpty) {
      log("credentials is empty now $credentials");
      Navigator.push(
          context,
          MaterialPageRoute(
              builder: (context) =>
                  CustomError(requestErrorTitleMsg: "No Credentials found",requestErrorSubTitleMsg: "Error found in the presentation flow")));
      return;
  }

  List<String> matchedCred = [];
  try {
    matchedCred =  await WalletSDKPlugin.processAuthorizationRequest(
        authorizationRequest: qrCodeURL, storedCredentials: credentials);
  } catch (error) {
    Navigator.push(
        context,
        MaterialPageRoute(
            builder: (context) =>
                CustomError(requestErrorTitleMsg:"No matching credential found" ,requestErrorSubTitleMsg: error.toString())));
  }

  // Get the matched VCIDs from the submission request.
  var getSubmissionRequest = await WalletSDKPlugin.getSubmissionRequirements(storedCredentials: credentials);
  var submission =  getSubmissionRequest.first;

  if (submission.count > 1) {
    // multiple matched vc ids are found therefore, invoking multiple credential Presentation Preview.
    log("multi cred flow $submission");
    var credentialDisplayData = storedCredentials
        .where((element) => credentials.contains(element.value.rawCredential))
        .map((e) =>
        CredentialData(rawCredential: e.value.rawCredential, issuerURL: e.value.issuerURL, credentialDisplayData: e.value.credentialDisplayData))
        .toList();
    navigateToPresentMultiCred(context, credentialDisplayData, "Choose ${submission.count} credentials to present");
    return;
  } else if (submission.count==1){
    var matchedVCsID = submission.inputDescriptors.first.matchedVCsID;
    if (matchedVCsID.length > 1) {
      log("matched length , more than matched vc ids are found ${matchedVCsID.length}");
      var credentialDisplayData = storedCredentials
          .where((element) => credentials.contains(element.value.rawCredential))
          .map((e) =>
          CredentialData(rawCredential: e.value.rawCredential, issuerURL: e.value.issuerURL, credentialDisplayData: e.value.credentialDisplayData))
          .toList();
      navigateToPresentMultiCredChooseOne(context, credentialDisplayData);
      return;
    } else {
      for (var cred in credentials) {
        log("single matched vc id flow");
        var credentialDisplayData = storedCredentials
            .where((element) => credentials.contains(element.value.rawCredential))
            .map((e) => e.value.credentialDisplayData);
        navigateToPresentationPreviewScreen(context,
            CredentialData(rawCredential: credentials.first, credentialDisplayData: credentialDisplayData.first, issuerURL: ''));
        return;
       }
      }
    }

}

void navigateToPresentMultiCred(
    BuildContext context, List<CredentialData> credentialData, String infoData) async {
  Navigator.push(
      context,
      MaterialPageRoute(
          builder: (context) =>
              PresentationPreviewMultiCredCheck(credentialData: credentialData, infoData: infoData )));
}

void navigateToPresentMultiCredChooseOne(
    BuildContext context, List<CredentialData> credentialData) async {
  Navigator.push(
      context,
      MaterialPageRoute(
          builder: (context) =>
              PresentationPreviewMultiCred(credentialData: credentialData)));
}

void navigateToPresentationPreviewScreen(
    BuildContext context, CredentialData credentialData) async {
  Navigator.push(
      context,
      MaterialPageRoute(
          builder: (context) =>
              PresentationPreview(credentialData: credentialData)));
}
