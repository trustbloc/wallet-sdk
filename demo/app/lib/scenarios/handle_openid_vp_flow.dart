/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'dart:developer';
import 'dart:convert';
import 'package:app/wallet_sdk/wallet_sdk.dart';
import 'package:app/views/custom_error.dart';
import 'package:flutter/material.dart';
import 'package:app/models/credential_data.dart';
import 'package:app/models/credential_data_object.dart';
import 'package:app/models/store_credential_data.dart';
import 'package:app/services/storage_service.dart';
import 'package:app/views/presentation_preview.dart';
import 'package:app/views/presentation_preview_multi_cred.dart';
import 'package:app/views/presentation_preview_multi_cred_radio.dart';
import 'package:flutter/services.dart';
import 'package:jwt_decode/jwt_decode.dart';

void handleOpenIDVpFlow(BuildContext context, String qrCodeURL) async {
  var walletSDKPlugin = WalletSDK();
  final StorageService storageService = StorageService();
  late List<CredentialDataObject> storedCredentials;
  late List<String> credentials;

  // Check if the flow is for the verifiable presentation or for issuance.
  UserLoginDetails userLoginDetails = await getUser();
  var username = userLoginDetails.username;
  storedCredentials = await storageService.retrieveCredentials(username!);

  credentials = storedCredentials.map((e) => e.value.rawCredential).toList();
  try {
    await walletSDKPlugin.processAuthorizationRequest(authorizationRequest: qrCodeURL, storedCredentials: credentials);
  } on PlatformException catch (error) {
    if (!context.mounted) return;
    Navigator.push(
        context,
        MaterialPageRoute(
            builder: (context) => CustomError(
                titleBar: 'Processing Presentation',
                requestErrorTitleMsg: error.message!,
                requestErrorSubTitleMsg: error.details)));
  }
  // Get the matched VCIDs from the submission request.
  var getSubmissionRequest = await walletSDKPlugin.getSubmissionRequirements(storedCredentials: credentials);
  var submission = getSubmissionRequest.first;
  if (submission.count > 1) {
    // multiple matched vc ids are found therefore, invoking multiple credential Presentation Preview.
    List<CredentialData> credentialDisplayDataList = [];
    for (var cred in credentials) {
      log('multi cred flow $submission and ${credentials.length}');
      for (var inputDescriptor in submission.inputDescriptors) {
        Map<String, dynamic>? payload;
        try {
          payload = Jwt.parseJwt(cred);
        } catch (jwtError) {
          try {
            payload = json.decode(cred) as Map<String, dynamic>;
          } catch (jsonError) {
            log('error while parsing cred as json: $jsonError');
            continue;
          }
        }
        var key = payload['jti'] ?? payload['id'];
        if (key == null) {
          log('no key found in cred payload');
          continue;
        }
        if (inputDescriptor.matchedVCsID.contains(key)) {
          log('matched vc with id $key added to list');
          var credentialDisplayData = storedCredentials
              .where((element) => cred.contains(element.value.rawCredential))
              .map((e) => e.value)
              .toList();
          credentialDisplayDataList.add(credentialDisplayData.first);
        }
      }
    }
    navigateToPresentMultiCred(context, credentialDisplayDataList, 'Choose ${submission.count} credentials to present');
    return;
  } else if (submission.count == 1) {
    var matchedVCsID = submission.inputDescriptors.first.matchedVCsID;
    if (matchedVCsID.length > 1) {
      log('matched length, more than matched vc ids are found ${matchedVCsID.length}');
      var credentialDisplayData = storedCredentials
          .where((element) => matchedVCsID.contains(element.value.credID))
          .map((e) => e.value)
          .toList();
      navigateToPresentMultiCredChooseOne(context, credentialDisplayData);
      return;
    } else {
      log('single matched vc id flow');

      for (var inputDes in submission.inputDescriptors) {
        for (var matchVC in inputDes.matchedVCs) {
          var credID = (await walletSDKPlugin.getCredID([matchVC]))!;
          var issuerURI = storedCredentials
              .where((element) => credID.contains(element.value.credID))
              .map((e) => e.value.issuerURL)
              .toList();
          var credentialDID =
              storedCredentials.firstWhere((element) => credID.contains(element.value.credID)).value.credentialDID;

          log('matched issuerURI found: $issuerURI');
          final credentialDisplayData = (await walletSDKPlugin.resolveDisplayData([matchVC], issuerURI.first))!;
          log('credentialDisplayData -> $credentialDisplayData');
          navigateToPresentationPreviewScreen(
              context,
              CredentialData(
                  rawCredential: matchVC,
                  issuerURL: issuerURI.first,
                  credentialDisplayData: credentialDisplayData.credentialsDisplay[0],
                  credentialDID: credentialDID,
                  issuerDisplayData: credentialDisplayData.issuerDisplay,
                  credID: credID));
          return;
        }
      }
    }
  }
}

void navigateToPresentMultiCred(BuildContext context, List<CredentialData> credentialData, String infoData) async {
  Navigator.push(
      context,
      MaterialPageRoute(
          builder: (context) => PresentationPreviewMultiCredCheck(credentialData: credentialData, infoData: infoData)));
}

void navigateToPresentMultiCredChooseOne(BuildContext context, List<CredentialData> credentialData) async {
  Navigator.push(
      context, MaterialPageRoute(builder: (context) => PresentationPreviewMultiCred(credentialData: credentialData)));
}

void navigateToPresentationPreviewScreen(BuildContext context, CredentialData credentialData) async {
  Navigator.push(context, MaterialPageRoute(builder: (context) => PresentationPreview(credentialData: credentialData)));
}
