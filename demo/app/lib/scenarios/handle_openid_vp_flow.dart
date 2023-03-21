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

  var matchedVCIds = handleSubmissionRequirement(getSubmissionRequest);

  if (matchedVCIds.length > 1) {
    // multiple matched vc ids are found therefore, invoking multiple credential Presentation Preview.
    log("multi cred flow");
    var credentialDisplayData = storedCredentials
        .where((element) => matchedCred.contains(element.value.rawCredential))
        .map((e) =>
        CredentialData(rawCredential: e.value.rawCredential, issuerURL: e.value.issuerURL, credentialDisplayData: e.value.credentialDisplayData))
        .toList();
    navigateToPresentMultiCred(context, credentialDisplayData);
    return;
  }

  for (var cred in credentials) {
    for (var matchCred in matchedCred){
      log("choose flow ${credentials.first != matchedCred.first}");
      if (cred != matchCred){
        log("selective disclosure flow - credentialDisplayData as credential doesnt match matched credentials");
        var issuerURI = storedCredentials.map((e) => e.value.issuerURL).toList();

        var credentialDisplayData =  await WalletSDKPlugin.serializeDisplayData(matchedCred, issuerURI.first);
        navigateToPresentationPreviewScreen(context, matchedCred.first,
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
          navigateToPresentationPreviewScreen(context, matchedCred.first,
              CredentialData(rawCredential: '', credentialDisplayData: credentialDisplayData.first, issuerURL: ''));
          return;
        }
      }
    }

  }

}

List<String> handleSubmissionRequirement(List<SubmissionRequirement> submissionRequirement){
  var submission =  submissionRequirement.first;
  return submission.inputDescriptors.first.matchedVCsID;
}

void navigateToPresentMultiCred(
    BuildContext context, List<CredentialData> credentialData) async {
  Navigator.push(
      context,
      MaterialPageRoute(
          builder: (context) =>
              PresentationPreviewMultiCred(credentialData: credentialData)));
}

void navigateToPresentationPreviewScreen(
    BuildContext context, String matchedCredential, CredentialData credentialData) async {
  Navigator.push(
      context,
      MaterialPageRoute(
          builder: (context) =>
              PresentationPreview(matchedCredential: matchedCredential, credentialData: credentialData)));
}
