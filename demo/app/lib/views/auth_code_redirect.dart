import 'package:flutter/material.dart';
import 'package:app/widgets/primary_button.dart';
import 'dart:developer';
import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:app/demo_method_channel.dart';
import 'package:app/widgets/common_title_appbar.dart';
import 'package:app/services/storage_service.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:app/models/activity_data_object.dart';
import 'package:app/models/credential_data.dart';
import 'credential_preview.dart';

class RedirectPathTextBox extends StatefulWidget {
  RedirectPathTextBox({Key? key});

  @override
  State<RedirectPathTextBox> createState() => RedirectPathTextBoxState();
}

class RedirectPathTextBoxState extends State<RedirectPathTextBox> {
  final textController = TextEditingController();
  var WalletSDKPlugin = MethodChannelWallet();
  final StorageService _storageService = StorageService();
  final Future<SharedPreferences> prefs = SharedPreferences.getInstance();
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

  @override
  Widget build(BuildContext context) {
    return new Scaffold(
      appBar: const CustomTitleAppBar(
        pageTitle: 'Redirect Path Simulator',
        addCloseIcon: true,
        height: 50,
      ),
        body: new Column(
      crossAxisAlignment: CrossAxisAlignment.center,
      children: <Widget>[
        Padding(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 32),
            child: GestureDetector(
              onDoubleTap:() {
                if(textController.text.isNotEmpty) {
                  textController.selection = TextSelection(baseOffset: 0, extentOffset:textController.text.length);
                }
              },
              child: TextField(
                controller: textController,
                keyboardType: TextInputType.multiline,
                minLines: 1,
                maxLines: 7,
                autofocus: true,
                decoration: const InputDecoration(
                  border: OutlineInputBorder(),
                  labelText: 'Paste the redirect url here',
                ),
              ),
            )),
        const Padding(
          padding: EdgeInsets.fromLTRB(24, 0, 24, 0),
        ),
        PrimaryButton(
            onPressed: () async {
              final SharedPreferences pref = await prefs;
              await _createDid();
              pref.setString('userDID',userDIDId);
              pref.setString('userDIDDoc',userDIDDoc);
              var credentials = await WalletSDKPlugin.requestCredentialWithAuth(textController.text.toString());
              String? issuerURI = await WalletSDKPlugin.issuerURI();
              var resolvedCredentialDisplay = await WalletSDKPlugin.serializeDisplayData([credentials], issuerURI!);
              log("resolvedCredentialDisplay -> $resolvedCredentialDisplay");
              var renderedCredDisplay =  await WalletSDKPlugin.resolveCredentialDisplay(resolvedCredentialDisplay!);
              var activities = await WalletSDKPlugin.storeActivityLogger();
              var credID = await WalletSDKPlugin.getCredID([credentials]);
              _storageService.addActivities(ActivityDataObj(credID!, activities));
              pref.setString("credID", credID);
              _navigateToCredPreviewScreen(credentials, issuerURI, resolvedCredentialDisplay!, userDIDId);
            },
            width: double.infinity,
            child: const Text('Submit', style: TextStyle(fontSize: 16, color: Colors.white))
        ),
      ],
    )
    );
  }

 _navigateToCredPreviewScreen(String credentialResp, String issuerURI, String credentialDisplayData, String didID) async {
  log("about to navigate here");
  WidgetsBinding.instance.addPostFrameCallback((_) {
    Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => CredentialPreview(credentialData: CredentialData(rawCredential: credentialResp, issuerURL: issuerURI, credentialDisplayData: credentialDisplayData, credentialDID: didID),)));
  });
}
}