import 'dart:convert';
import 'package:app/models/credential_preview.dart';
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
    /// Await your Future here (This function only called once after the layout is Complete)
    WidgetsBinding.instance?.addPostFrameCallback((timeStamp) async {
      userLoggedIn = (await _storageService.retrieve("username"))!;
    });
  }

  Future<List<CredentialPreviewData>> getData() async {
    List<CredentialPreviewData> list;
      var data = json.decode(widget.credentialResponse);
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
          return (credPrev[position].label != "photo") ? Card(
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
    Map<String, dynamic> issuer = jsonDecode(widget.credentialResponse);
    final issuerDisplayData = issuer['issuer_display']['name'];
    return Scaffold(
      appBar: AppBar(
        title: const Text('Credential Preview'),
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
                        "$issuerDisplayData"),
                  ),
                  const SizedBox(
                    child:  Text(
                        textAlign: TextAlign.center,
                        style: TextStyle(fontSize: 18, color: Colors.black),
                        "wants to issue the credential"),
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
                    final StorageItem? newItem = await showDialog<StorageItem>(
                        context: context, builder: (_) => AddDataDialog());
                    if (newItem != null) {
                      _storageService.add(StorageItem("$userLoggedIn-credential-${uuid.v1()}", widget.credentialResponse));
                      _navigateToDashboard(userLoggedIn);
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
  _navigateToDashboard(String userLoggedIn) async {
    Navigator.push(context, MaterialPageRoute(builder: (context) => Dashboard(user: userLoggedIn)));
  }
}