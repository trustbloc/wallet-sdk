import 'dart:convert';

import 'package:app/main.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'credential_preview.dart';

class OTP extends StatefulWidget {
  const OTP({Key? key}) : super(key: key);

  @override
  State<OTP> createState() => _OTPPage();
}
class _OTPPage extends State<OTP> {
  // 4 text editing controllers that associate with the 4 input fields
  final TextEditingController _fieldOne = TextEditingController();
  final TextEditingController _fieldTwo = TextEditingController();
  final TextEditingController _fieldThree = TextEditingController();
  final TextEditingController _fieldFour = TextEditingController();
  final TextEditingController _fieldFive = TextEditingController();
  final TextEditingController _fieldSix = TextEditingController();

  // This is the entered code
  // It will be displayed in a Text widget
  String? _otp;
  String _requestCredentialErrorMsg = '';
  String? actionText = 'Submit';

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Credential Preview'),
      ),
      body: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Text('Please enter OTP', style: TextStyle(fontSize: 30)),
          const SizedBox(
            height: 30,
          ),
          // Implement 4 input fields
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceEvenly,
            children: [
              OtpInput(_fieldOne, true), // auto focus
              OtpInput(_fieldTwo, false),
              OtpInput(_fieldThree, false),
              OtpInput(_fieldFour, false),
              OtpInput(_fieldFive, false),
              OtpInput(_fieldSix, false),
            ],
          ),
          const SizedBox(
            height: 30,
          ),
          ElevatedButton(
              onPressed: () async {
                setState(() {
                  _otp = _fieldOne.text +
                      _fieldTwo.text +
                      _fieldThree.text +
                      _fieldFour.text+
                      _fieldFive.text +
                      _fieldSix.text;
                });
                  var requestCredentialResp =  await WalletSDKPlugin.requestCredential(_otp!);
                  var resolvedCredentialDisplay =  await WalletSDKPlugin.resolveCredentialDisplay();
                  // Making sure request credential response doesn't contain error
                  if (!resolvedCredentialDisplay!.contains("generic-error")) {
                    _navigateToCredPreviewScreen(requestCredentialResp!, resolvedCredentialDisplay);
                  } else {
                    setState(() {
                      _requestCredentialErrorMsg = requestCredentialResp.toString();
                      actionText = 'Re-enter';
                      _clearOTPInput();
                    });
                  }
              },

              child: Text(actionText!, style: const TextStyle(fontSize: 20))
          ),
          const SizedBox(
            height: 30,
          ),
          // Display the entered OTP code
          Text(
            _otp ?? '',
            style: const TextStyle(fontSize: 30),
          ),
          Text(_requestCredentialErrorMsg ?? '', style: const TextStyle(fontSize: 12, color: Colors.redAccent)),
        ],
      ),
    );
  }
  _navigateToCredPreviewScreen(String credentialResp, String credentialResolveDisplay) async {
    Navigator.push(context, MaterialPageRoute(builder: (context) => CredentialPreview(rawCredential: credentialResp,credentialDisplay: credentialResolveDisplay)));
  }

  _clearOTPInput(){
    _fieldOne.clear();
    _fieldTwo.clear();
    _fieldThree.clear();
    _fieldFour.clear();
    _fieldFive.clear();
    _fieldSix.clear();
  }
}

// Create an input widget that takes only one digit
class OtpInput extends StatelessWidget {
  final TextEditingController controller;
  final bool autoFocus;
  const OtpInput(this.controller, this.autoFocus, {Key? key}) : super(key: key);
  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 60,
      width: 50,
      child: TextField(
        autofocus: autoFocus,
        textAlign: TextAlign.center,
        keyboardType: TextInputType.number,
        controller: controller,
        maxLength: 1,
        cursorColor: Theme.of(context).primaryColor,
        decoration: const InputDecoration(
            border: OutlineInputBorder(),
            counterText: '',
            hintStyle: TextStyle(color: Colors.black, fontSize: 20.0)),
        onChanged: (value) {
          if (value.length == 1) {
            FocusScope.of(context).nextFocus();
          }
        },
      ),
    );
  }
}

