/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'package:app/views/credential_shared.dart';
import 'package:app/views/dashboard.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:uuid/uuid.dart';
import 'package:app/widgets/common_title_appbar.dart';
import 'package:app/models/credential_data.dart';
import 'package:app/widgets/primary_button.dart';
import 'package:app/main.dart';
import 'package:app/services/storage_service.dart';
import 'package:app/models/activity_data_object.dart';
import 'package:app/widgets/credential_card.dart';
import 'package:app/views/custom_error.dart';
import 'dart:developer';

import '../services/attestation.dart';

class PresentationPreviewMultiCredCheck extends StatefulWidget {
  final List<CredentialData> credentialData;
  final String? infoData;

  const PresentationPreviewMultiCredCheck({super.key, required this.credentialData, this.infoData});

  @override
  State<PresentationPreviewMultiCredCheck> createState() => PresentationPreviewMultiCredCheckState();
}

class PresentationPreviewMultiCredCheckState extends State<PresentationPreviewMultiCredCheck> {
  final StorageService _storageService = StorageService();
  final Future<SharedPreferences> prefs = SharedPreferences.getInstance();
  var uuid = const Uuid();
  bool checked = false;
  List multipleSelected = [];
  late var checkListItems = widget.credentialData;
  List<CredentialData> selectedCredentialData = [];
  var selectedIndexes = [];

  late String verifierName = '';
  late String serviceURL = '';
  late String verifierPurpose = '';
  late String verifierLogoURL = '';
  bool verifiedDomain = true;
  bool rememberMe = false;
  bool showErrorMessage = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((timeStamp) async {
      final verifiedDisplayData = await WalletSDKPlugin.getVerifierDisplayData();

      var resp = await WalletSDKPlugin.wellKnownDidConfig(verifiedDisplayData.did);
      setState(() {
        verifierName = verifiedDisplayData.name;
        verifierLogoURL = verifiedDisplayData.logoURI;
        verifierPurpose = verifiedDisplayData.purpose;
        serviceURL = resp.serviceURL;
        verifiedDomain = resp.isValid;
      });
    });
  }

  @override
  Widget build(BuildContext context) {
    final width = MediaQuery.of(context).size.width;
    return Scaffold(
      appBar: const CustomTitleAppBar(
        pageTitle: 'Share Multi Credential',
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
              Text(verifierPurpose),
              Text.rich(
                textAlign: TextAlign.center,
                TextSpan(
                  children: [
                    TextSpan(
                      text: widget.infoData,
                      style: const TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.normal,
                        color: Colors.black,
                      ),
                    ),
                  ],
                ),
              ),
              Column(
                children: List.generate(
                  checkListItems.length,
                  (index) => CheckboxListTile(
                    controlAffinity: ListTileControlAffinity.leading,
                    contentPadding: EdgeInsets.zero,
                    dense: true,
                    title: CredentialCard(
                        credentialData: checkListItems[index], isDashboardWidget: false, isDetailArrowRequired: true),
                    value: selectedIndexes.contains(index),
                    onChanged: (value) {
                      setState(() {
                        log('selected item ${selectedIndexes.contains(index)}');
                        if (selectedIndexes.contains(index)) {
                          selectedIndexes.remove(index);
                          multipleSelected.remove(checkListItems[index].rawCredential);
                          log('multiple selected removing $multipleSelected');
                          selectedCredentialData.removeAt(index);
                          setState(() => rememberMe = value!);
                        } else {
                          selectedIndexes.add(index);
                          multipleSelected.add(checkListItems[index].rawCredential);
                          log('multiple selected adding $multipleSelected');
                          selectedCredentialData.add(CredentialData(
                              rawCredential: checkListItems[index].rawCredential,
                              credentialDisplayData: checkListItems[index].credentialDisplayData,
                              issuerDisplayData: checkListItems[index].issuerDisplayData,
                              issuerURL: '',
                              credentialDID: checkListItems[index].credentialDID,
                              credID: checkListItems[index].credID));
                          setState(() => rememberMe = value!);
                          setState(() => showErrorMessage = false);
                        }
                      });
                    },
                  ),
                ),
              ),
              Padding(
                padding: EdgeInsets.only(top: width * 0.7),
              ),
              showErrorMessage
                  ? Container(
                      decoration: BoxDecoration(color: Colors.redAccent, borderRadius: BorderRadius.circular(8.0)),
                      child: Padding(
                          padding: const EdgeInsets.all(12),
                          child: SelectableText(
                            widget.infoData!,
                            style: TextStyle(fontWeight: FontWeight.bold, color: Colors.white.withOpacity(0.8)),
                          )))
                  : Container(),
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
                          onPressed: () async {
                            if (rememberMe != true) {
                              setState(() => showErrorMessage = true);
                            } else {
                              setState(() => showErrorMessage = false);
                            }
                            final SharedPreferences pref = await prefs;
                            try {
                              await WalletSDKPlugin.presentCredential(
                                  attestationVC :await AttestationService.returnAttestationVCIfEnabled(),
                                  selectedCredentials: multipleSelected.cast<String>());
                            } catch (error) {
                              log(error.toString());
                              if (!error.toString().contains('OVP1-0002')) {
                                var errString = error.toString().replaceAll(r'\', '');
                                Navigator.push(
                                    context,
                                    MaterialPageRoute(
                                        builder: (context) => CustomError(
                                            titleBar: 'Multi Presentation Preview',
                                            requestErrorTitleMsg: 'error while presenting credential',
                                            requestErrorSubTitleMsg: errString)));
                                return;
                              } else {
                                await WalletSDKPlugin.presentCredential(
                                    attestationVC :await AttestationService.returnAttestationVCIfEnabled(),
                                    selectedCredentials: multipleSelected.cast<String>());
                              }
                            }
                            var activities = await WalletSDKPlugin.storeActivityLogger();
                            var credID = pref.getString('credID');
                            _storageService.addActivities(ActivityDataObj(credID!, activities));
                            _navigateToCredentialShareSuccess(verifierName);
                          },
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
  _navigateToDashboard() async {
    Navigator.push(context, MaterialPageRoute(builder: (context) => const Dashboard()));
  }

  _navigateToCredentialShareSuccess(String verifierName) async {
    Navigator.push(
        context,
        MaterialPageRoute(
            builder: (context) =>
                CredentialShared(verifierName: verifierName, credentialData: selectedCredentialData)));
  }
}
