import 'dart:developer';
import 'package:app/demo_method_channel.dart';
import 'package:app/scenarios/handle_openid_issuance_flow.dart';
import 'package:app/scenarios/handle_openid_vp_flow.dart';
import 'package:flutter/material.dart';

void handleOpenIDUrl(BuildContext context, String qrCodeURL) async {
  log('received qr code url - $qrCodeURL');
  // Check if the flow is for the verifiable presentation or for issuance.
  if (!qrCodeURL.contains("openid-vc")) {
    handleOpenIDIssuanceFlow(context, qrCodeURL);
  } else {
    handleOpenIDVpFlow(context, qrCodeURL);
  }
}


List<String> handleSubmissionRequirement(List<SubmissionRequirement> submissionRequirement){
     var submission =  submissionRequirement.first;
     return submission.inputDescriptors.first.matchedVCsID;
}