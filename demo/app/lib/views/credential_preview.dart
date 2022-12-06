import 'dart:convert';
import 'package:app/views/dashboard.dart';
import 'package:app/services/storage_service.dart';
import 'package:app/widgets/add_credential_dialog.dart';
import 'package:flutter/material.dart';
import 'package:uuid/uuid.dart';
import '../models/store_credential_data.dart';

class CredentialPreview extends StatefulWidget {
  final String credentialResponse;
  const CredentialPreview({super.key, required this.credentialResponse});

  @override
  State<CredentialPreview> createState() => CredentialPreviewState();
}

class CredentialPreviewState extends State<CredentialPreview> {
  final StorageService _storageService = StorageService();
  var uuid = const Uuid();
  late final String userLoggedIn;

  @override
  void initState() {
    super.initState();
    userLoggedIn = _storageService.retrieve("username").toString();
  }


  @override
  Widget build(BuildContext context) {
    Map<String, dynamic> issuer = jsonDecode(widget.credentialResponse);
    final issuerDisplayData = issuer['issuer_display']['name'];
    final credentialName = issuer['credential_displays'][0]['overview']['name'];
    final credentialLogoURL = issuer['credential_displays'][0]['overview']['logo']['url'];
    return Scaffold(
      appBar: AppBar(
        title: const Text('Credential Preview'),
      ),
        body: Center(
            child:  Column (
              children: [
                const SizedBox(height: 50),
                Text(
                  issuerDisplayData,
                  textAlign: TextAlign.center,
                  style: TextStyle(fontSize: 22, color: Colors.black),
                ),
                const Text(
                  'wants to issue the credential',
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 30),
                MaterialButton(
                  onPressed: () { },
                  padding: const EdgeInsets.all(0.0),
                  child: Ink(
                    decoration: BoxDecoration(
                      image: DecorationImage(
                        image: AssetImage(credentialLogoURL.toString()),
                        fit: BoxFit.none,
                      ),
                      gradient: const LinearGradient(
                        colors: <Color>[Colors.lightBlueAccent, Colors.blueAccent],
                      ),
                    ),
                    child: Container(
                      constraints: const BoxConstraints(maxWidth: 350, minHeight: 120.0), // min sizes for Material buttons
                      alignment: Alignment.center,
                      child: Text(
                        credentialName,
                        textAlign: TextAlign.center,
                      ),
                    ),
                  ),
                ),
              ],
            )
        ),
      floatingActionButton: SizedBox(
        width: double.infinity,
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 8),
          child: Row(
            children: [
              Expanded(
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(primary: Colors.red),
                  onPressed: () async {
                    _navigateToDashboard();
                  },
                  child: const Text("Cancel"),
                ),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: ElevatedButton(
                  onPressed: () async {
                    final StorageItem? newItem = await showDialog<StorageItem>(
                        context: context, builder: (_) => AddDataDialog());
                    if (newItem != null) {
                      _storageService.add(StorageItem("$userLoggedIn-credential-${uuid.v1()}", widget.credentialResponse));
                      _navigateToDashboard();
                    }
                  },
                  child: const Text("Save Credential"),
                ),
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
}