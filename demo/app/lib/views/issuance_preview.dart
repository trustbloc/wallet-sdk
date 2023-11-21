/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'dart:developer';

import 'package:app/main.dart';
import 'package:app/views/dashboard.dart';
import 'package:app/services/storage_service.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:app/widgets/common_title_appbar.dart';
import 'package:app/widgets/primary_button.dart';

import 'package:app/models/activity_data_object.dart';
import 'package:app/models/credential_data.dart';
import 'package:app/wallet_sdk/wallet_sdk.dart';
import 'credential_preview.dart';
import 'handle_redirect_uri.dart';
import 'otp.dart';

class IssuancePreview extends StatefulWidget {
  bool? authorizeResultPinRequired;
  Uri? uri;

  IssuancePreview({this.authorizeResultPinRequired, this.uri, Key? key}) : super(key: key);

  @override
  State<IssuancePreview> createState() => IssuancePreviewState();
}

class IssuancePreviewState extends State<IssuancePreview> {
  String issuerDisplayName = '';
  String credentialIssuer = '';
  String credentialDisplayName = '';
  String backgroundColor = '';
  String issuerDisplayURL = '';
  String textColor = '';
  String logoURL = '';
  String issuerLogoURL = '';
  final Future<SharedPreferences> prefs = SharedPreferences.getInstance();
  List<String>? credentialTypes;

  @override
  void initState() {
    super.initState();
    prefs.then((value) {
      credentialTypes = value.getStringList('credentialTypes');
    }).whenComplete(() => WalletSDKPlugin.getIssuerMetaData(credentialTypes!).then((response) {
      setState(() {
        credentialIssuer = response.first.credentialIssuer;
        issuerDisplayName = response.first.localizedIssuerDisplays.first.name;
        issuerDisplayURL = response.first.localizedIssuerDisplays.first.url;
        final issuerLogo = response.first.localizedIssuerDisplays.first.logo;
        if (issuerLogo != null) {
          issuerLogoURL = issuerLogo;
        }
        credentialDisplayName = response.first.supportedCredentials.first.display.first.name;
        final logo = response.first.supportedCredentials.first.display.first.logo;
        if (logo != null) {
          logoURL = logo;
        }
        backgroundColor =
            '0xff${response.first.supportedCredentials.first.display.first.backgroundColor.toString().replaceAll('#', '')}';
        textColor =
            '0xff${response.first.supportedCredentials.first.display.first.textColor.toString().replaceAll('#', '')}';
      });
    }),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: const CustomTitleAppBar(
        pageTitle: 'Issuance Preview',
        addCloseIcon: true,
        height: 60,
      ),
      body: Container(
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
              imageUrl: issuerLogoURL,
              placeholder: (context, url) => const CircularProgressIndicator(),
              errorWidget: (context, url, error) => Image.asset('lib/assets/images/logoIcon.png', fit: BoxFit.cover),
              width: 60,
              height: 80,
              fit: BoxFit.fitWidth,
            ),
            SizedBox(
              height: 40,
              child: Text(
                  textAlign: TextAlign.center,
                  style: const TextStyle(fontSize: 18, color: Color(0xff190C21), fontWeight: FontWeight.bold),
                  issuerDisplayName),
            ),
            SizedBox(
              height: 30,
              child: Text(
                  textAlign: TextAlign.center,
                  style: const TextStyle(fontSize: 12, color: Color(0xff190C21), fontWeight: FontWeight.normal),
                  issuerDisplayURL),
            ),
            const SizedBox(height: 20),
            Container(
                height: 80,
                alignment: Alignment.center,
                decoration: BoxDecoration(
                    color: backgroundColor.isNotEmpty ? Color(int.parse(backgroundColor)) : Colors.white,
                    borderRadius: BorderRadius.circular(12),
                    boxShadow: [BoxShadow(offset: const Offset(3, 3), color: Colors.grey.shade300, blurRadius: 5)]),
                child: ListTile(
                  title: Text(
                    credentialDisplayName,
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                      color: textColor.isNotEmpty ? Color(int.parse(textColor)) : const Color(0xff190C21),
                    ),
                    textAlign: TextAlign.start,
                  ),
                  subtitle: Text(
                    credentialIssuer,
                    style: TextStyle(
                      fontSize: 6,
                      fontWeight: FontWeight.bold,
                      color: textColor.isNotEmpty ? Color(int.parse(textColor)) : const Color(0xff190C21),
                    ),
                    textAlign: TextAlign.start,
                  ),
                  leading: CachedNetworkImage(
                    imageUrl: logoURL,
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
                          child: const Text('Add to Wallet', style: TextStyle(fontSize: 16, color: Colors.white))),
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

    Navigator.push(context, MaterialPageRoute(builder: (context) => CredentialPreview(credentialData: credentialData)));
  }

  Future<CredentialData> fetchPreviewScreenDetails() async {
    final Future<SharedPreferences> prefs = SharedPreferences.getInstance();
    var WalletSDKPlugin = WalletSDK();
    final StorageService storageService = StorageService();
    final SharedPreferences pref = await prefs;

    var didType = pref.getString('didType');
    var keyType = pref.getString('keyType');
    // choosing default if no selection is made
    didType = didType ?? 'ion';
    keyType = keyType ?? 'ED25519';

    var didResolution = await WalletSDKPlugin.createDID(didType, keyType);
    var didID = didResolution.did;
    var didDoc = didResolution.didDoc;
    log('created didID :$didID');

    pref.setString('userDID', didID);
    pref.setString('userDIDDoc', didDoc);

    String? credentials = await WalletSDKPlugin.requestCredential('');
    String? issuerURL = await WalletSDKPlugin.issuerURI();
    String? resolvedCredentialDisplay = await WalletSDKPlugin.serializeDisplayData([credentials], issuerURL!);
    log('resolvedCredentialDisplay $resolvedCredentialDisplay');

    var activities = await WalletSDKPlugin.storeActivityLogger();

    var credID = await WalletSDKPlugin.getCredID([credentials]);

    log('activities and credID handle open id  -$activities and $credID');
    storageService.addActivities(ActivityDataObj(credID!, activities));

    return CredentialData(
        rawCredential: credentials,
        issuerURL: issuerURL,
        credentialDisplayData: resolvedCredentialDisplay!,
        credentialDID: didID,
        credID: credID);
  }
}
