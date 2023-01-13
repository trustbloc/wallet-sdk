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
  String _requestCredentialErrorMsg = '';
  String? actionText = 'Save Credential';

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
        onTap: () {
      FocusManager.instance.primaryFocus?.unfocus();
    },
    child: Scaffold(
      appBar: CustomTitleAppBar(pageTitle: 'Add Credential', addCloseIcon: true),
      backgroundColor: const Color(0xffFBF8FC),
      body: Center(
        child: ListView(
          padding: const EdgeInsets.all(24),
          children: [
            ListTile(
              //todo Issue-174 read the meta data from the backend on page load
              leading: Image.asset('lib/assets/images/credLogo.png'),
              title: const Text('Utopian Issuer', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
              subtitle: const Text('utopia.example.com', style: TextStyle(fontSize: 12, fontWeight: FontWeight.normal)),
              trailing: Image.asset('lib/assets/images/verified.png', width: 82, height: 26),
            ),
            Container(
              padding: const EdgeInsets.fromLTRB(14, 4, 14, 16),
           child: ListTile(
              contentPadding: const EdgeInsets.all(12),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
              tileColor: Colors.white,
              title: const Text(
                'Credential',
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Color(0xff190C21),
                ),
                textAlign: TextAlign.start,
              ),
              //TODO need to add fallback and network image url
              leading: const Image(
                image: AssetImage('lib/assets/images/genericCredential.png'),
                width: 47,
                height: 47,
                fit: BoxFit.cover,
              ),
            ),
            ),
      Container(
        padding: const EdgeInsets.fromLTRB(14, 4, 14, 16),
        child: PrimaryInputField(textController: otpController, maxLength: 6, labelText: 'Enter OTP Code', textInputFormatter: FilteringTextInputFormatter.digitsOnly),
      ),
            Column(
              children: <Widget>[
                Column(
                  children: <Widget>[
                    Container(
                      alignment: Alignment.bottomCenter,
                      padding: const EdgeInsets.only(top: 400),
                    ),
                    PrimaryButton(
                      onPressed: () async {
                        setState(() {
                          _otp = otpController.text;
                        });
                        var requestCredentialResp =  await WalletSDKPlugin.requestCredential(_otp!);
                        var resolvedCredentialDisplay =  await WalletSDKPlugin.resolveCredentialDisplay();
                        // Making sure request credential response doesn't contain error
                        if (!resolvedCredentialDisplay!.contains("invalid pin")) {
                          _navigateToCredPreviewScreen(requestCredentialResp!, resolvedCredentialDisplay);
                        } else {
                          setState(() {
                            _requestCredentialErrorMsg = requestCredentialResp.toString();
                            actionText = 'Re-enter';
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
            Text(_requestCredentialErrorMsg ?? '', style: const TextStyle(fontSize: 12, color: Colors.redAccent)),

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

