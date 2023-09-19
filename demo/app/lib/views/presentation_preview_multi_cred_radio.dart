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

class PresentationPreviewMultiCred extends StatefulWidget {
  final List<CredentialData> credentialData;

  const PresentationPreviewMultiCred({super.key, required this.credentialData});

  @override
  State<PresentationPreviewMultiCred> createState() => PresentationPreviewMultiCredState();
}

class PresentationPreviewMultiCredState extends State<PresentationPreviewMultiCred> {
  final StorageService _storageService = StorageService();
  final Future<SharedPreferences> prefs = SharedPreferences.getInstance();
  var uuid = const Uuid();
  bool checked = false;
  late CredentialData selectedCredentialData = widget.credentialData[0];
  int selectedRadio = 0;
  late String verifierLogoURL = '';
  late String verifierName = '';
  late String serviceURL = '';
  late String verifierPurpose = '';
  bool verifiedDomain = true;

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

  setSelectedRadio(int val) {
    setState(() {
      selectedRadio = val;
    });
  }

  @override
  Widget build(BuildContext context) {
    final width = MediaQuery.of(context).size.width;
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
              Text(verifierPurpose),
              for (var i = 0; i < widget.credentialData.length; i++)
                RadioListTile(
                  controlAffinity: ListTileControlAffinity.leading,
                  title: CredentialCard(
                      credentialData: widget.credentialData[i], isDashboardWidget: false, isDetailArrowRequired: true),
                  activeColor: Colors.deepPurple,
                  autofocus: false,
                  value: i,
                  groupValue: selectedRadio,
                  onChanged: (val) {
                    print('Radio $val');
                    selectedCredentialData = widget.credentialData[i];
                    setSelectedRadio(val!);
                  },
                ),
              Padding(
                padding: EdgeInsets.only(top: width * 0.8),
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
                          onPressed: () async {
                            final SharedPreferences pref = await prefs;
                            await WalletSDKPlugin.presentCredential(
                                selectedCredentials: [selectedCredentialData.rawCredential]);
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
            ],
          ),
        ),
      ),
    );
  }

  _navigateToDashboard() async {
    Navigator.push(context, MaterialPageRoute(builder: (context) => const Dashboard()));
  }

  _navigateToCredentialShareSuccess(String verifierName) async {
    Navigator.push(
        context,
        MaterialPageRoute(
            builder: (context) =>
                CredentialShared(verifierName: verifierName, credentialData: [selectedCredentialData])));
  }
}
