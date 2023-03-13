import 'dart:convert';
import 'dart:developer';
import 'package:app/main.dart';
import 'package:app/models/credential_data.dart';
import 'package:app/views/credential_added.dart';
import 'package:app/views/dashboard.dart';
import 'package:app/services/storage_service.dart';
import 'package:flutter/material.dart';
import 'package:uuid/uuid.dart';
import 'package:app/models/credential_data_object.dart';
import 'package:app/models/store_credential_data.dart';
import 'package:app/widgets/common_title_appbar.dart';
import 'package:app/widgets/credential_card.dart';
import 'package:app/widgets/primary_button.dart';

class CredentialPreview extends StatefulWidget {
  final CredentialData credentialData;
  const CredentialPreview({super.key, required this.credentialData});

  @override
  State<CredentialPreview> createState() => CredentialPreviewState();
}

class CredentialPreviewState extends State<CredentialPreview> {
  final StorageService _storageService = StorageService();
  var uuid = const Uuid();
  late final String userLoggedIn;
  late String issuerDisplayData;

  @override
  void initState() {
    super.initState();
    WalletSDKPlugin.resolveCredDisplayRendering(widget.credentialData.credentialDisplayData!).then(
            (response) {
              setState(() {
                var credentialDisplayEncodeData = json.encode(response);
                List<dynamic> responseJson = json.decode(credentialDisplayEncodeData);
                issuerDisplayData = responseJson.first["issuerName"];
                log("issuerDisplayData state $issuerDisplayData");
              });
        });
    WidgetsBinding.instance?.addPostFrameCallback((timeStamp) async {
      UserLoginDetails userLoginDetails =  await getUser();
      userLoggedIn = userLoginDetails.username!;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar:  const CustomTitleAppBar(pageTitle: 'Credential Preview', addCloseIcon: true, height: 60,),
      body: Container(
        padding: const EdgeInsets.fromLTRB(24, 40, 16, 24),
        child: Column(
        mainAxisAlignment: MainAxisAlignment.start,
        children: [
          const SizedBox(height: 50),
          SizedBox(
            height: 50,
            child:  Text(
                textAlign: TextAlign.center,
                style: const TextStyle(fontSize: 22, color: Color(0xff190C21), fontWeight: FontWeight.bold),
                issuerDisplayData),
          ),
          const SizedBox(
            child:  Text(
                textAlign: TextAlign.center,
                style: TextStyle(fontSize: 18, color: Colors.black),
                "wants to issue the credential"),
          ),
          CredentialCard(credentialData: widget.credentialData, isDashboardWidget: false, isDetailArrowRequired: false),
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
                            _storageService.addCredential(CredentialDataObject("$userLoggedIn-${uuid.v1()}", widget.credentialData!));
                            _navigateToCredentialAdded();
                        },
                        width: double.infinity,
                        child: const Text('Save Credential', style: TextStyle(fontSize: 16, color: Colors.white))
                    ),
                    const Padding(
                      padding: EdgeInsets.fromLTRB(24, 0, 24, 8),
                    ),
                    PrimaryButton(
                      onPressed: (){
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

  _navigateToCredentialAdded() async {
    Navigator.push(context, MaterialPageRoute(builder: (context) => CredentialAdded(credentialData: widget.credentialData)));
  }
  _navigateToDashboard() async {
    Navigator.push(context, MaterialPageRoute(builder: (context) => const Dashboard()));
  }
}