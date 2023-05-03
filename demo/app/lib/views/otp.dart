import 'dart:convert';
import 'dart:developer';

import 'package:app/main.dart';
import 'package:app/models/activity_data_object.dart';
import 'package:app/models/credential_data.dart';
import 'package:flutter/material.dart';
import 'package:app/widgets/common_title_appbar.dart';
import 'package:flutter/services.dart';
import 'package:app/services/storage_service.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:app/widgets/primary_button.dart';
import 'package:app/widgets/primary_input_field.dart';
import 'credential_preview.dart';

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

  Future<String?> _createDid() async {
    final SharedPreferences pref = await prefs;
    var didType = pref.getString('didType');
    didType = didType ?? "ion";
    var keyType = pref.getString('keyType');
    keyType = keyType ?? "ED25519";
    var didResolution = await WalletSDKPlugin.createDID(didType, keyType);
    var didDocEncoded = json.encode(didResolution!);
    Map<String, dynamic> responseJson = json.decode(didDocEncoded);
    var didID = responseJson["did"];
    var didDoc = responseJson["didDoc"];
    setState(() {
      userDIDId = didID;
      userDIDDoc = didDoc;
    });
    return didID;
  }


  // This is the entered code
  // It will be displayed in a Text widget
  String? _otp;
  String _requestErrorSubTitleMsg = '';
  String _requestErrorTitleMsg = '';
  String? actionText = 'Submit';
  late double topPadding;
  bool show = false;

  @override
  Widget build(BuildContext context) {
    final height = MediaQuery.of(context).size.height;
    final width = MediaQuery.of(context).size.width;
    return GestureDetector(
        onTap: () {
      FocusManager.instance.primaryFocus?.unfocus();
    },
    child: Scaffold(
      appBar: const CustomTitleAppBar(pageTitle: 'Enter OTP', addCloseIcon: true, height: 60,),
      backgroundColor: const Color(0xffF4F1F5),
      body: Center(
        child: ListView(
          padding: const EdgeInsets.all(24),
          children: [
            Container(
              padding: const EdgeInsets.fromLTRB(14, 4, 14, 16),
            ),
      Container(
        padding: const EdgeInsets.fromLTRB(14, 4, 14, 16),
        child: PrimaryInputField(textController: otpController, maxLength: 6, labelText: 'Enter OTP Code', textInputFormatter: FilteringTextInputFormatter.digitsOnly),
      ),
            Column(
              children: <Widget>[
                Column(
                  children: <Widget>[
                    Visibility(
                      visible: show,
                      child: Container (
                        padding: const EdgeInsets.all(12),
                        alignment: Alignment.center,
                        child: ListTile(
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(12),
                          ),
                          tileColor:  const Color(0xffFBF8FC),
                          title: SelectableText(
                            _requestErrorTitleMsg ?? '',
                            style: const TextStyle(
                              fontSize: 16,
                              fontWeight: FontWeight.bold,
                              color: Color(0xff190C21),
                            ),
                            textAlign: TextAlign.start,
                          ),
                          subtitle:  SelectableText(
                            _requestErrorSubTitleMsg ?? '',
                            style: const TextStyle(
                              fontSize: 12,
                              fontWeight: FontWeight.bold,
                              color: Color(0xff6C6D7C),
                            ),
                            textAlign: TextAlign.start,
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
                              )
                          ),
                        ),
                      ),
                    ),
                   Padding(
                      padding: EdgeInsets.only(top: height-width),
                    ),
                    PrimaryButton(
                      onPressed: () async {
                        setState(() {
                          _otp = otpController.text;
                        });

                        String? credentials;
                        String? resolvedCredentialDisplay;
                        try {
                          final SharedPreferences pref = await prefs;
                             await _createDid();
                             pref.setString('userDID',userDIDId);
                             pref.setString('userDIDDoc',userDIDDoc);
                          credentials =  await WalletSDKPlugin.requestCredential(_otp!);
                          String? issuerURI = await WalletSDKPlugin.issuerURI();
                          resolvedCredentialDisplay = await WalletSDKPlugin.serializeDisplayData([credentials], issuerURI!);
                          log("resolvedCredentialDisplay -> $resolvedCredentialDisplay");
                          var renderedCredDisplay =  await WalletSDKPlugin.resolveCredentialDisplay(resolvedCredentialDisplay!);
                          log("rendered cred display otp -${renderedCredDisplay}");
                          var activities = await WalletSDKPlugin.storeActivityLogger();
                          var credID = await WalletSDKPlugin.getCredID([credentials]);
                          log("activities and credID -$activities and $credID");
                          _storageService.addActivities(ActivityDataObj(credID!, activities));
                           pref.setString("credID", credID);
                          _navigateToCredPreviewScreen(credentials, issuerURI, resolvedCredentialDisplay!, userDIDId);
                        } catch (err) {
                          String errorMessage = err.toString();
                          if (err is PlatformException &&
                              err.message != null &&
                              err.message!.isNotEmpty) {
                            errorMessage = err.details!;
                          }
                            setState(() {
                              _requestErrorSubTitleMsg = errorMessage;
                              _requestErrorTitleMsg = 'Error';
                              actionText = 'Re-enter';
                               show = true;
                              topPadding = height*0.20;
                              _clearOTPInput();
                            });
                        }
                      },
                      width: double.infinity,
                     child: Text(actionText!, style: const TextStyle(fontSize: 16, color: Colors.white))
                    ),
                    const Padding(
                      padding: EdgeInsets.only(top: 8),
                    ),
                    PrimaryButton(
                      onPressed: (){},
                      width: double.infinity,
                      gradient: const LinearGradient(
                          begin: Alignment.topCenter,
                          end: Alignment.bottomCenter,
                          colors: [Color(0xffFFFFFF), Color(0xffFFFFFF)]),
                      child: const Text('Cancel', style: TextStyle(fontSize: 16, color: Color(0xff6C6D7C))),
                    ),
                  ],
                ),
              ],
            ),
          ],
    ),
 )));
  }
  _navigateToCredPreviewScreen(String credentialResp, String issuerURI, String credentialDisplayData, String didID) async {
    Navigator.push(context, MaterialPageRoute(builder: (context) => CredentialPreview(credentialData: CredentialData(rawCredential: credentialResp, issuerURL: issuerURI, credentialDisplayData: credentialDisplayData, credentialDID: didID),)));
  }

  _clearOTPInput(){
    otpController.clear();
  }
}

