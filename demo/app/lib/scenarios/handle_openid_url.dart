/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'dart:developer';
import 'package:app/scenarios/handle_openid_issuance_flow.dart';
import 'package:app/scenarios/handle_openid_vp_flow.dart';
import 'package:app/wallet_sdk/wallet_sdk_model.dart';
import 'package:flutter/material.dart';

void handleOpenIDUrl(BuildContext context, String qrCodeURL) async {
  log('received qr code url - $qrCodeURL');
  // Check if the flow is for the verifiable presentation or for issuance.
  if (qrCodeURL.contains('openid-vc://') || qrCodeURL.contains('openid4vp://')) {
    handleOpenIDVpFlow(context, qrCodeURL);
  } else {
    handleOpenIDIssuanceFlow(context, qrCodeURL);
  }
}

List<String> handleSubmissionRequirement(List<SubmissionRequirement> submissionRequirement) {
  var submission = submissionRequirement.first;
  return submission.inputDescriptors.first.matchedVCsID;
}
