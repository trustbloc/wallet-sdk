import 'dart:convert';

import 'package:app/views/credential_shared.dart';
import 'package:app/views/dashboard.dart';
import 'package:flutter/material.dart';
import 'package:uuid/uuid.dart';
import '../models/credential_preview.dart';
import '../services/storage_service.dart';

class PresentationPreview extends StatefulWidget {
  final String matchedCredential;
  final String credentialDisplay;
  const PresentationPreview({super.key, required this.matchedCredential, required this.credentialDisplay});

  @override
  State<PresentationPreview> createState() => PresentationPreviewState();
}

class PresentationPreviewState extends State<PresentationPreview> {
  final StorageService _storageService = StorageService();
  var uuid = const Uuid();
  late final String userLoggedIn;
  // Todo fetch the name of the  verifier name from the presentation
  late String verifierName = 'Utopian Background Check';

  @override
  void initState() {
    super.initState();
    /// Await your Future here (This function only called once after the layout is Complete)
    WidgetsBinding.instance?.addPostFrameCallback((timeStamp) async {
      userLoggedIn = (await _storageService.retrieve("username"))!;
    });
  }

  Future<List<CredentialPreviewData>> getData() async {
    List<CredentialPreviewData> list;
    var data = json.decode(widget.credentialDisplay);
    var credentialClaimsData = data['credential_displays'][0]['claims'] as List;
    list = credentialClaimsData.map<CredentialPreviewData>((json) => CredentialPreviewData.fromJson(json)).toList();
    return list;
  }
  Widget listViewWidget(List<CredentialPreviewData> credPrev) {
    return ListView.builder(
        itemCount: credPrev.length,
        scrollDirection: Axis.vertical,
        shrinkWrap: true,
        itemBuilder: (context, position) {
          //TODO Ignoring the photo value for now due to extremely long text need to render in a separate issue-881
          return (credPrev[position].label != "photo" && credPrev[position].label != "ID") ? Card(
            child: ListTile(
              title: Text(
                credPrev[position].label,
                style: const TextStyle(
                    fontSize: 16,
                    color: Colors.black,
                    fontWeight: FontWeight.normal),
              ),
              subtitle: Text(
                credPrev[position].value,
                style: const TextStyle(
                    fontSize: 18.0,
                    color: Colors.green,
                    fontWeight: FontWeight.normal),
              ),
            ),
          ):
          Container();
        });
  }
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Share Credential'),
      ),
      body: FutureBuilder(
          future: getData(),
          builder: (context, snapshot) {
            if (!snapshot.hasData) {
              return const Center(child: CircularProgressIndicator());
            } else {
              return Column(
                children: [
                  const SizedBox(height: 50),
                  SizedBox(
                    height: 50,
                    child:  Text(
                        textAlign: TextAlign.center,
                        style: const TextStyle(fontSize: 22, color: Colors.green, fontWeight: FontWeight.bold),
                        verifierName),
                  ),
                  const SizedBox(
                    child:  Text(
                        textAlign: TextAlign.center,
                        style: TextStyle(fontSize: 18, color: Colors.black),
                        "requesting the following credential"),
                  ),
                  Expanded( // wrap in Expanded
                    child: listViewWidget(snapshot.data!),
                  ),
                ],
              );
            }
          }
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
                    _navigateToDashboard(userLoggedIn);
                  },
                  child: const Text("Cancel"),
                ),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: ElevatedButton(
                  onPressed: () async {
                    _navigateToCredentialShareSuccess(verifierName);
                  },
                  child: const Text("Share Credential"),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
  _navigateToDashboard(String userLoggedIn) async {
    Navigator.push(context, MaterialPageRoute(builder: (context) => Dashboard(user: userLoggedIn)));
  }
  _navigateToCredentialShareSuccess(String verifierName) async {
    Navigator.push(context, MaterialPageRoute(builder: (context) => CredentialShared(verifierName: verifierName)));
  }
}