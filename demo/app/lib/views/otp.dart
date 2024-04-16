/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'dart:developer';

import 'package:app/models/activity_data_object.dart';
import 'package:app/models/credential_data.dart';
import 'package:flutter/material.dart';
import 'package:app/widgets/common_title_appbar.dart';
import 'package:flutter/services.dart';
import 'package:app/services/storage_service.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:app/widgets/primary_button.dart';
import 'package:app/widgets/primary_input_field.dart';
import 'package:app/wallet_sdk/wallet_sdk.dart';
import '../services/attestation.dart';
import '../widgets/loading_overlay.dart';
import 'credential_preview.dart';
import 'package:app/views/dashboard.dart';

class OTP extends StatefulWidget {
  const OTP({Key? key}) : super(key: key);

  @override
  State<OTP> createState() => _OTPPage();
}

class _OTPPage extends State<OTP> {
  final StorageService _storageService = StorageService();
  final Future<SharedPreferences> prefs = SharedPreferences.getInstance();
  final TextEditingController otpController = TextEditingController();
  var userDIDId = '';
  var userDIDDoc = '';

  var WalletSDKPlugin = WalletSDK();

  Future<String?> _createDid() async {
    final SharedPreferences pref = await prefs;
    var didType = pref.getString('didType');
    didType = didType ?? 'jwk';
    var keyType = pref.getString('keyType');
    keyType = keyType ?? 'ECDSAP384IEEEP1363';
    var didResolution = await WalletSDKPlugin.createDID(didType, keyType);
    var didID = didResolution.did;
    setState(() {
      userDIDId = didID;
      userDIDDoc = didResolution.didDoc;
    });
    return didID;
  }

  // This is the entered code
  // It will be displayed in a Text widget
  String? _otp;
  String _requestErrorSubTitleMsg = '';
  String _requestErrorTitleMsg = '';
  String _requestErrorDetailMsg = '';
  String? actionText = 'Submit';
  late double topPadding;
  bool show = false;
  bool showDetail = false;
  bool _isLoading = false;

  void detailToggle() {
    setState(() {
      showDetail = !showDetail;
    });
  }

  @override
  Widget build(BuildContext context) {
    final height = MediaQuery.of(context).size.height;
    return GestureDetector(
        onTap: () {
          FocusManager.instance.primaryFocus?.unfocus();
        },
        child: Scaffold(
          appBar: const CustomTitleAppBar(
            pageTitle: 'Enter OTP',
            addCloseIcon: true,
            height: 60,
          ),
          backgroundColor: const Color(0xffF4F1F5),
          body: Stack(
            children: [
              Center(
                child: Column(
                  children: [
                    Container(
                      padding: const EdgeInsets.fromLTRB(14, 4, 14, 16),
                    ),
                    Container(
                      padding: const EdgeInsets.fromLTRB(14, 4, 14, 16),
                      child: PrimaryInputField(
                          textController: otpController,
                          maxLength: 6,
                          labelText: 'Enter OTP Code',
                          textInputFormatter: FilteringTextInputFormatter.digitsOnly),
                    ),
                    Column(
                      children: <Widget>[
                        Column(
                          children: <Widget>[
                            Visibility(
                              visible: show,
                              child: Container(
                                padding: const EdgeInsets.all(12),
                                alignment: Alignment.center,
                                child: ListTile(
                                  shape: RoundedRectangleBorder(
                                    borderRadius: BorderRadius.circular(12),
                                  ),
                                  tileColor: const Color(0xffFBF8FC),
                                  title: SelectableText(
                                    _requestErrorTitleMsg ?? '',
                                    style: const TextStyle(
                                      fontSize: 16,
                                      fontWeight: FontWeight.bold,
                                      color: Color(0xff190C21),
                                    ),
                                    textAlign: TextAlign.start,
                                  ),
                                  subtitle: SelectableText(
                                    _requestErrorSubTitleMsg ?? '',
                                    style: const TextStyle(
                                      fontSize: 12,
                                      fontWeight: FontWeight.bold,
                                      color: Color(0xff6C6D7C),
                                    ),
                                    textAlign: TextAlign.start,
                                  ),
                                  trailing: IconButton(
                                    icon: const Icon(
                                      Icons.more_vert,
                                      size: 20.0,
                                    ),
                                    onPressed: () {
                                      detailToggle();
                                    },
                                  ),
                                  //TODO need to add fallback and network image url
                                  leading: const SizedBox(
                                      height: 24,
                                      width: 24,
                                      child: Image(
                                        image: AssetImage('lib/assets/images/errorVector.png'),
                                        width: 24,
                                        height: 24,
                                        fit: BoxFit.cover,
                                      )),
                                ),
                              ),
                            ),
                            Visibility(
                                visible: showDetail,
                                child: Padding(
                                  padding: const EdgeInsets.all(16),
                                  child: SelectableText(
                                    _requestErrorDetailMsg ?? '',
                                    style: const TextStyle(
                                      fontSize: 12,
                                      fontWeight: FontWeight.bold,
                                      color: Color(0xff6C6D7C),
                                    ),
                                    textAlign: TextAlign.start,
                                  ),
                                )),
                          ],
                        ),
                      ],
                    ),
                    const Spacer(),
                    Padding(
                      padding: const EdgeInsets.only(left: 16.0, top: 0, right: 16.0, bottom: 32.0),
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.end,
                        children: [
                          PrimaryButton(
                              onPressed: () async {
                                setState(() {
                                  _otp = otpController.text;
                                  _isLoading = true;
                                });

                                try {
                                  final SharedPreferences pref = await prefs;
                                  await _createDid();
                                  pref.setString('userDID', userDIDId);
                                  pref.setString('userDIDDoc', userDIDDoc);
                                  final credentials = await WalletSDKPlugin.requestCredential(
                                    _otp!,
                                    attestationVC: await AttestationService.returnAttestationVCIfEnabled(),
                                  );

                                  String? issuerURI = await WalletSDKPlugin.issuerURI();
                                  final resolvedCredentialsDisplay = await WalletSDKPlugin.resolveDisplayData(
                                      credentials.map((e) => e.content).toList(), issuerURI!);
                                  log('serializeDisplayData otp-> $resolvedCredentialsDisplay');
                                  var activities = await WalletSDKPlugin.storeActivityLogger();

                                  final result = <CredentialData>[];
                                  for (int i = 0; i < credentials.length; ++i) {
                                    var credential = credentials[i].content;
                                    var credID = credentials[i].id;

                                    log('activities and credID -$activities and $credID');
                                    _storageService.addActivities(ActivityDataObj(credID!, activities));
                                    pref.setString('credID', credID);

                                    result.add(CredentialData(
                                        rawCredential: credential,
                                        issuerURL: issuerURI,
                                        credentialDisplayData: resolvedCredentialsDisplay.credentialsDisplay[i],
                                        issuerDisplayData: resolvedCredentialsDisplay.issuerDisplay,
                                        credentialDID: userDIDId,
                                        credID: credID));
                                  }
                                  setState(() {
                                    _isLoading = false;
                                  });

                                  _navigateToCredPreviewScreen(result);
                                } catch (err) {
                                  String errorMessage = err.toString();
                                  log('errorMessage-> $errorMessage');
                                  if (err is PlatformException && err.message != null && err.message!.isNotEmpty) {
                                    log('err.details-> ${err.details}');
                                    var resp =
                                        await WalletSDKPlugin.parseWalletSDKError(localizedErrorMessage: err.details);
                                    log('resp-> $resp');
                                    (resp.category == 'INVALID_GRANT')
                                        ? {
                                            errorMessage = 'Try re-entering the PIN or scan a new QR code',
                                            _requestErrorDetailMsg = resp.details
                                          }
                                        : (resp.category == 'INVALID_TOKEN')
                                            ? {
                                                errorMessage = 'Try scanning a new QR code',
                                                _requestErrorDetailMsg = resp.details
                                              }
                                            : {errorMessage = resp.details, _requestErrorDetailMsg = resp.traceID};
                                  }
                                  setState(() {
                                    _requestErrorSubTitleMsg = errorMessage;
                                    _requestErrorTitleMsg = 'Oops! Something went wrong!';
                                    actionText = 'Re-enter';
                                    show = true;
                                    topPadding = height * 0.20;
                                    _clearOTPInput();
                                    _isLoading = false;
                                  });
                                }
                              },
                              width: MediaQuery.of(context).size.width,
                              child: Text(actionText!, style: const TextStyle(fontSize: 16, color: Colors.white))),
                          const Padding(
                            padding: EdgeInsets.only(top: 8),
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
                    ),
                  ],
                ),
              ),
              // Loading overlay
              if (_isLoading) const LoadingOverlay(),
            ],
          ),
        ));
  }

  _navigateToCredPreviewScreen(List<CredentialData> credentialsData) async {
    Navigator.push(
        context,
        MaterialPageRoute(
            builder: (context) => CredentialPreview(
                  credentialsData: credentialsData,
                )));
  }

  _navigateToDashboard() async {
    Navigator.push(context, MaterialPageRoute(builder: (context) => const Dashboard()));
  }

  _clearOTPInput() {
    otpController.clear();
  }
}
