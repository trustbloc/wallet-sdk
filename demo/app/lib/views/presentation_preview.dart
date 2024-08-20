/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'dart:developer';

import 'package:app/views/credential_shared.dart';
import 'package:app/views/dashboard.dart';
import 'package:app/wallet_sdk/wallet_sdk_model.dart';
import 'package:app/widgets/credential_verified_information_view.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:uuid/uuid.dart';
import 'package:app/widgets/common_title_appbar.dart';
import 'package:app/models/credential_data.dart';
import 'package:app/widgets/primary_button.dart';
import 'package:app/widgets/credential_metadata_card.dart';
import 'package:app/main.dart';
import 'package:app/services/storage_service.dart';
import 'package:app/models/activity_data_object.dart';
import 'package:app/widgets/credential_card.dart';
import 'package:app/views/custom_error.dart';

import 'package:app/services/config_service.dart';

import '../services/attestation.dart';

class PresentationPreview extends StatefulWidget {
  final CredentialData credentialData;

  const PresentationPreview({super.key, required this.credentialData});

  @override
  State<PresentationPreview> createState() => PresentationPreviewState();
}

class PresentationPreviewState extends State<PresentationPreview> {
  final StorageService _storageService = StorageService();
  final Future<SharedPreferences> prefs = SharedPreferences.getInstance();
  var uuid = const Uuid();

  late String verifierName = '';
  late String verifierLogoURL = '';
  late String verifierPurpose = '';
  late String serviceURL = '';
  bool verifiedDomain = true;
  EvaluationResult? trustInfoEvaluationResult;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((timeStamp) async {
      final verifiedDisplayData = await WalletSDKPlugin.getVerifierDisplayData();
      log('verifiedDisplayData ${verifiedDisplayData.logoURI}');
      var resp = await WalletSDKPlugin.wellKnownDidConfig(verifiedDisplayData.did);
      setState(() {
        verifierName = verifiedDisplayData.name;
        verifierLogoURL = verifiedDisplayData.logoURI;
        verifierPurpose = verifiedDisplayData.purpose;
        serviceURL = resp.serviceURL;
        verifiedDomain = resp.isValid;
      });

      trustInfoEvaluationResult = await WalletSDKPlugin.evaluatePresentationTrustInfo();
      setState(() {});
      final requestedAttestations = trustInfoEvaluationResult?.requestedAttestations ?? [];
      if (requestedAttestations.where((attestation) => attestation == 'wallet_authentication').isNotEmpty) {
        await AttestationService.issueAttestationVC();
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: const CustomTitleAppBar(
        pageTitle: 'Share Credential',
        addCloseIcon: true,
        height: 60,
      ),
      body: SingleChildScrollView(
        child: Container(
          padding: const EdgeInsets.fromLTRB(24, 12, 24, 0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.start,
            children: [
              ListTile(
                leading: verifierLogoURL == ''
                    ? const SizedBox.shrink()
                    : CachedNetworkImage(
                        imageUrl: verifierLogoURL,
                        placeholder: (context, url) => const CircularProgressIndicator(),
                        errorWidget: (context, url, error) =>
                            Image.asset('lib/assets/images/credLogo.png', fit: BoxFit.contain),
                        width: 60,
                        height: 60,
                        fit: BoxFit.contain,
                      ),
                title: Text(verifierName, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                subtitle: Text(serviceURL != '' ? serviceURL : 'verifier.com',
                    style: const TextStyle(fontSize: 12, fontWeight: FontWeight.normal)),
                trailing: FittedBox(
                    child: verifiedDomain
                        ? const Row(children: [
                            Text.rich(
                              textAlign: TextAlign.center,
                              TextSpan(
                                children: [
                                  WidgetSpan(
                                      child: Icon(
                                    Icons.verified_user_outlined,
                                    color: Colors.lightGreen,
                                    size: 18,
                                  )),
                                  TextSpan(
                                    text: 'Verified',
                                    style: TextStyle(
                                      fontSize: 14,
                                      fontWeight: FontWeight.bold,
                                      color: Colors.lightGreen,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ])
                        : const Row(
                            children: [
                              Text.rich(
                                textAlign: TextAlign.center,
                                TextSpan(
                                  children: [
                                    WidgetSpan(
                                        child: Icon(
                                      Icons.dangerous_outlined,
                                      color: Colors.redAccent,
                                      size: 18,
                                    )),
                                    TextSpan(
                                      text: 'Unverified',
                                      style: TextStyle(
                                        fontSize: 14,
                                        fontWeight: FontWeight.bold,
                                        color: Colors.redAccent,
                                      ),
                                    ),
                                  ],
                                ),
                              )
                            ],
                          )),
              ),
              if (trustInfoEvaluationResult != null && trustInfoEvaluationResult!.allowed)
                const Text('Trust info verified',
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                      color: Colors.lightGreen,
                    )),
              if (trustInfoEvaluationResult != null && !trustInfoEvaluationResult!.allowed)
                Text('Trust info not verified:${trustInfoEvaluationResult!.errorCode}',
                    style: const TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                      color: Colors.red,
                    )),
              if (trustInfoEvaluationResult != null && trustInfoEvaluationResult!.denyReason.isNotEmpty)
                Text('Warning: ${trustInfoEvaluationResult!.denyReason}',
                    style: const TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.normal,
                      color: Colors.yellow,
                    )),

              Text(verifierPurpose),
              CredentialCard(
                  credentialData: widget.credentialData, isDashboardWidget: false, isDetailArrowRequired: false),
              CredentialMetaDataCard(credentialData: widget.credentialData),
              CredentialVerifiedInformation(
                credentialData: widget.credentialData,
                height: MediaQuery.of(context).size.height * 0.38,
              ),
              Align(
                alignment: Alignment.bottomCenter,
                child: SingleChildScrollView(
                  padding: const EdgeInsets.all(8),
                  child: Column(
                    children: [
                      const Padding(
                        padding: EdgeInsets.fromLTRB(24, 0, 24, 0),
                      ),
                      PrimaryButton(
                          onPressed: _presentCredential,
                          width: double.infinity,
                          child: const Text('Share Credential', style: TextStyle(fontSize: 16, color: Colors.white))),
                      const Padding(
                        padding: EdgeInsets.fromLTRB(12, 0, 12, 8),
                      ),
                      PrimaryButton(
                        onPressed: () {
                          _callNoConsentAcknowledgment();
                        },
                        width: double.infinity,
                        gradient: const LinearGradient(
                            begin: Alignment.topCenter,
                            end: Alignment.bottomCenter,
                            colors: [Color(0xffFFFFFF), Color(0xffFFFFFF)]),
                        child: const Text('Cancel', style: TextStyle(fontSize: 16, color: Color(0xff6C6D7C))),
                      ),
                    ],
                  ),
                ), //last one
              ),
            ],
          ),
        ),
      ),
    );
  }

  void  _callNoConsentAcknowledgment() async {
    var ackResp = await  WalletSDKPlugin.noConsentAcknowledgement();
    if (ackResp != null) {
      _navigateToDashboard();
    }
  }

  void _presentCredential() async {
    final SharedPreferences pref = await prefs;
    Map<String, dynamic> customScopeConfigList = {};
    final ConfigService configService = ConfigService();
    final attestationVC = await AttestationService.returnAttestationVCIfEnabled();
    WalletSDKPlugin.getCustomScope()
        .then((customScopeList) async {
      customScopeConfigList = await configService.readCustomScopeConfig(customScopeList);
    })
        .whenComplete(() => WalletSDKPlugin.presentCredential(
        selectedCredentials: [widget.credentialData.rawCredential],
        attestationVC: attestationVC,
        customScopeList: customScopeConfigList))
        .onError((error, stackTrace) {
      var errString = error.toString().replaceAll(r'\', '');
      _navigateToCustomError(errString);
    });
    var activities = await WalletSDKPlugin.storeActivityLogger();
    var credID = pref.getString('credID');
    _storageService.addActivities(ActivityDataObj(credID!, activities));
    _navigateToCredentialShareSuccess(verifierName);
  }

  _navigateToDashboard() async {
    Navigator.push(context, MaterialPageRoute(builder: (context) => const Dashboard()));
  }

  _navigateToCustomError(String errString) async {
    Navigator.push(
        context,
        MaterialPageRoute(
            builder: (context) => CustomError(
                titleBar: 'Presentation Preview',
                requestErrorTitleMsg: 'error while presenting credential',
                requestErrorSubTitleMsg: errString)));
  }

  _navigateToCredentialShareSuccess(String verifierName) async {
    Navigator.push(
        context,
        MaterialPageRoute(
            builder: (context) => CredentialShared(
                  verifierName: verifierName,
                  credentialData: [widget.credentialData],
                )));
  }
}
