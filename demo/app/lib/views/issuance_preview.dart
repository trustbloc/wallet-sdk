/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'dart:developer';

import 'package:app/main.dart';
import 'package:app/views/dashboard.dart';
import 'package:app/services/storage_service.dart';
import 'package:app/wallet_sdk/wallet_sdk_model.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:app/widgets/common_title_appbar.dart';
import 'package:app/widgets/primary_button.dart';

import 'package:app/models/activity_data_object.dart';
import 'package:app/models/credential_data.dart';
import 'package:app/widgets/domain_verification_component.dart';
import '../services/attestation.dart';
import 'credential_preview.dart';
import 'handle_redirect_uri.dart';
import 'otp.dart';

class IssuancePreview extends StatefulWidget {
  final bool? authorizeResultPinRequired;
  final Uri? uri;

  const IssuancePreview({this.authorizeResultPinRequired, this.uri, Key? key}) : super(key: key);

  @override
  State<IssuancePreview> createState() => IssuancePreviewState();
}

Color _convertColor(String color) {
  return Color(int.parse('0xff${color.replaceAll('#', '')}'));
}

class IssuancePreviewState extends State<IssuancePreview> {
  String? issuerServiceURL;
  EvaluationResult? trustInfoEvaluationResult;
  CredentialOfferDisplayData? offerDisplayData;

  @override
  void initState() {
    super.initState();
    _init();
  }

  void _init() async {
    final trustInfoEvaluationResult = await WalletSDKPlugin.evaluateIssuanceTrustInfo();
    setState(() {
      this.trustInfoEvaluationResult = trustInfoEvaluationResult;
    });

    final requestedAttestations = trustInfoEvaluationResult?.requestedAttestations ?? [];
    if (requestedAttestations.where((attestation) => attestation == 'urn:attestation:compliance:fcra').isNotEmpty) {
      await AttestationService.issueAttestationVC();
    }

    final offerDisplayData = await WalletSDKPlugin.getCredentialOfferDisplayData();
    setState(() {
      this.offerDisplayData = offerDisplayData;
    });

    final serviceURL = await WalletSDKPlugin.verifyIssuer();
    setState(() {
      issuerServiceURL = serviceURL;
    });
  }

  @override
  Widget build(BuildContext context) {
    final offerDisplayData = this.offerDisplayData;

    return Scaffold(
      appBar: const CustomTitleAppBar(
        pageTitle: 'Issuance Preview',
        addCloseIcon: true,
        height: 60,
      ),
      body: offerDisplayData == null
          ? const Center(child: CircularProgressIndicator())
          : Container(
              padding: const EdgeInsets.fromLTRB(24, 40, 16, 24),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.start,
                children: [
                  const SizedBox(height: 50),
                  const SizedBox(
                    child: Text(
                        textAlign: TextAlign.center,
                        style: TextStyle(fontSize: 18, color: Colors.black),
                        'Add this credential ?'),
                  ),
                  const SizedBox(height: 30),
                  CachedNetworkImage(
                    imageUrl: offerDisplayData.issuer.logo ?? '',
                    placeholder: (context, url) => const CircularProgressIndicator(),
                    errorWidget: (context, url, error) =>
                        Image.asset('lib/assets/images/logoIcon.png', fit: BoxFit.cover),
                    width: 60,
                    height: 80,
                    fit: BoxFit.fitWidth,
                  ),
                  SizedBox(
                    height: 30,
                    child: Text(
                        textAlign: TextAlign.center,
                        style: const TextStyle(fontSize: 18, color: Color(0xff190C21), fontWeight: FontWeight.bold),
                        offerDisplayData.issuer.name),
                  ),
                  FittedBox(
                    child: issuerServiceURL != null
                        ? const DomainVerificationComponent(
                            status: 'Verified',
                            imagePath: 'lib/assets/images/tick-checked.svg',
                          )
                        : const DomainVerificationComponent(
                            status: 'Unverified',
                            imagePath: 'lib/assets/images/error_icon.svg',
                          ),
                  ),
                  const SizedBox(height: 4),
                  if (trustInfoEvaluationResult != null)
                    FittedBox(
                      child: trustInfoEvaluationResult!.allowed
                          ? const DomainVerificationComponent(
                              status: 'Trust info verified',
                              imagePath: 'lib/assets/images/tick-checked.svg',
                            )
                          : Column(
                              children: [
                                const DomainVerificationComponent(
                                  status: 'Trust info not verified',
                                  imagePath: 'lib/assets/images/error_icon.svg',
                                ),
                                Text(trustInfoEvaluationResult!.errorCode)
                              ],
                            ),
                    ),
                  SizedBox(
                    height: 30,
                    child: Text(
                        textAlign: TextAlign.center,
                        style: const TextStyle(fontSize: 12, color: Color(0xff190C21), fontWeight: FontWeight.normal),
                        offerDisplayData.issuer.url),
                  ),
                  const SizedBox(height: 20),
                  for (final cred in offerDisplayData.offeredCredentials)
                    Container(
                        height: 80,
                        alignment: Alignment.center,
                        decoration: BoxDecoration(
                            color: cred.backgroundColor.isNotEmpty ? _convertColor(cred.backgroundColor) : Colors.white,
                            borderRadius: BorderRadius.circular(12),
                            boxShadow: [
                              BoxShadow(offset: const Offset(3, 3), color: Colors.grey.shade300, blurRadius: 5)
                            ]),
                        child: ListTile(
                          title: Text(
                            cred.overviewName,
                            style: TextStyle(
                              fontSize: 14,
                              fontWeight: FontWeight.bold,
                              color:
                                  cred.textColor.isNotEmpty ? _convertColor(cred.textColor) : const Color(0xff190C21),
                            ),
                            textAlign: TextAlign.start,
                          ),
                          subtitle: Text(
                            offerDisplayData.issuer.name,
                            style: TextStyle(
                              fontSize: 6,
                              fontWeight: FontWeight.bold,
                              color:
                                  cred.textColor.isNotEmpty ? _convertColor(cred.textColor) : const Color(0xff190C21),
                            ),
                            textAlign: TextAlign.start,
                          ),
                          leading: CachedNetworkImage(
                            imageUrl: cred.logo ?? '',
                            placeholder: (context, url) => const CircularProgressIndicator(),
                            errorWidget: (context, url, error) =>
                                Image.asset('lib/assets/images/genericCredential.png', fit: BoxFit.contain),
                            width: 60,
                            height: 60,
                            fit: BoxFit.contain,
                          ),
                        )),
                  Expanded(
                    child: Align(
                      alignment: Alignment.bottomCenter,
                      child: Container(
                        height: 150,
                        padding: const EdgeInsets.all(16),
                        alignment: Alignment.topCenter,
                        decoration: BoxDecoration(
                          color: Colors.white,
                          borderRadius: BorderRadius.circular(12),
                          border: Border.all(
                            color: const Color(0xffDBD7DC),
                            width: 0.5,
                          ),
                        ),
                        child: Column(
                          children: [
                            const Padding(
                              padding: EdgeInsets.fromLTRB(24, 0, 24, 0),
                            ),
                            PrimaryButton(
                                onPressed: () async {
                                  if (widget.authorizeResultPinRequired == true) {
                                    navigateToOTPScreen(context);
                                  } else if (widget.uri != null) {
                                    navigateToAuthFlow(context, widget.uri!);
                                  } else {
                                    navigateToWithoutPinFlow(context);
                                  }
                                },
                                width: double.infinity,
                                child:
                                    const Text('Add to Wallet', style: TextStyle(fontSize: 16, color: Colors.white))),
                            const Padding(
                              padding: EdgeInsets.fromLTRB(24, 0, 24, 8),
                            ),
                            PrimaryButton(
                              onPressed: () {
                                _navigateToDashboard();
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
                  ),
                ],
              ),
            ),
    );
  }

  void navigateToOTPScreen(BuildContext context) async {
    Navigator.push(context, MaterialPageRoute(builder: (context) => const OTP()));
  }

  _navigateToDashboard() async {
    Navigator.push(context, MaterialPageRoute(builder: (context) => const Dashboard()));
  }

  void navigateToAuthFlow(BuildContext context, Uri uri) async {
    Navigator.of(context)
        .push(MaterialPageRoute(builder: (context) => HandleRedirectUri(uri, 'issuer-initiated-flow', '')));
  }

  void navigateToWithoutPinFlow(BuildContext context) async {
    var credentialData = await fetchPreviewScreenDetails();

    Navigator.push(
        context, MaterialPageRoute(builder: (context) => CredentialPreview(credentialsData: credentialData)));
  }

  Future<List<CredentialData>> fetchPreviewScreenDetails() async {
    final Future<SharedPreferences> prefs = SharedPreferences.getInstance();
    final StorageService storageService = StorageService();
    final SharedPreferences pref = await prefs;

    var didType = pref.getString('didType');
    var keyType = pref.getString('keyType');
    // choosing default if no selection is made
    didType = didType ?? 'jwk';
    keyType = keyType ?? 'ECDSAP384IEEEP1363';

    var didResolution = await WalletSDKPlugin.createDID(didType, keyType);
    var didID = didResolution.did;
    var didDoc = didResolution.didDoc;
    log('created didID :$didID');

    pref.setString('userDID', didID);
    pref.setString('userDIDDoc', didDoc);

    final credentials = await WalletSDKPlugin.requestCredential(
      '',
      attestationVC: await AttestationService.returnAttestationVCIfEnabled(),
    );
    final issuerURL = await WalletSDKPlugin.issuerURI();
    final resolvedCredentialsDisplay =
        await WalletSDKPlugin.resolveDisplayData(credentials.map((e) => e.content).toList(), issuerURL!);

    var activities = await WalletSDKPlugin.storeActivityLogger();

    final result = <CredentialData>[];

    for (int i = 0; i < credentials.length; ++i) {
      var credential = credentials[i].content;
      var credID = credentials[i].id;

      log('activities and credID handle open id  -$activities and $credID');
      storageService.addActivities(ActivityDataObj(credID, activities));

      result.add(CredentialData(
          rawCredential: credential,
          issuerURL: issuerURL,
          credentialDisplayData: resolvedCredentialsDisplay.credentialsDisplay[i],
          issuerDisplayData: resolvedCredentialsDisplay.issuerDisplay,
          credentialDID: didID,
          credID: credID));
    }

    return result;
  }
}
