import 'dart:convert';
import 'dart:developer';

import 'package:app/main.dart';
import 'package:flutter/material.dart';
import 'package:app/widgets/common_title_appbar.dart';
import 'package:flutter/services.dart';
import '../widgets/primary_button.dart';
import '../widgets/primary_input_field.dart';
import 'credential_preview.dart';

class OTP extends StatefulWidget {
  const OTP({Key? key}) : super(key: key);

  @override
  State<OTP> createState() => _OTPPage();
}
class _OTPPage extends State<OTP> {
  // 4 text editing controllers that associate with the 4 input fields
  final TextEditingController otpController = TextEditingController();

  // This is the entered code
  // It will be displayed in a Text widget
  String? _otp;
  String _requestErrorSubTitleMsg = '';
  String _requestErrorTitleMsg = '';
  String? actionText = 'Submit';
  double topPadding = 520;
  bool show = false;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
        onTap: () {
      FocusManager.instance.primaryFocus?.unfocus();
    },
    child: Scaffold(
      appBar: CustomTitleAppBar(pageTitle: 'Enter OTP', addCloseIcon: true),
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
                          title: Text(
                            _requestErrorTitleMsg ?? '',
                            style: const TextStyle(
                              fontSize: 16,
                              fontWeight: FontWeight.bold,
                              color: Color(0xff190C21),
                            ),
                            textAlign: TextAlign.start,
                          ),
                          subtitle:  Text(
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
                      padding: EdgeInsets.only(top: topPadding),
                    ),
                    PrimaryButton(
                      onPressed: () async {
                        setState(() {
                          _otp = otpController.text;
                        });
                        String? requestCredentialResp;
                        String? resolvedCredentialDisplay;
                        try {
                          log('inside requestCredential');
                          requestCredentialResp =  await WalletSDKPlugin.requestCredential(_otp!);
                          resolvedCredentialDisplay =  await WalletSDKPlugin.resolveCredentialDisplay();
                          _navigateToCredPreviewScreen(requestCredentialResp!, resolvedCredentialDisplay!);
                        } catch (err) {
                          String errorMessage = err.toString();
                          if (err is PlatformException &&
                              err.message != null &&
                              err.message!.isNotEmpty) {
                            errorMessage = err.message!;
                          }
                            setState(() {
                              _requestErrorSubTitleMsg = errorMessage;
                              _requestErrorTitleMsg = 'Error';
                              actionText = 'Re-enter';
                               show = true;
                              topPadding = 440;
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
                )
              ],
            ),
          ],
    ),
 )));
  }
  _navigateToCredPreviewScreen(String credentialResp, String credentialResolveDisplay) async {
    Navigator.push(context, MaterialPageRoute(builder: (context) => CredentialPreview(rawCredential: credentialResp,credentialDisplay: credentialResolveDisplay)));
  }

  _clearOTPInput(){
    otpController.clear();
  }
}

