import 'dart:convert';
import 'dart:developer';
import 'dart:typed_data';
import 'package:app/models/credential_data.dart';
import 'package:app/models/credential_preview.dart';
import 'package:app/views/dashboard.dart';
import 'package:app/services/storage_service.dart';
import 'package:app/widgets/add_credential_dialog.dart';
import 'package:flutter/material.dart';
import 'package:uuid/uuid.dart';
import '../models/credential_data_object.dart';
import '../models/store_credential_data.dart';

class CredentialPreview extends StatefulWidget {
  final String rawCredential;
  final String credentialDisplay;
  const CredentialPreview({super.key, required this.rawCredential, required this.credentialDisplay});

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
      UserLoginDetails userLoginDetails =  await getUser();
      userLoggedIn = userLoginDetails.username!;
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
          return (credPrev[position].label != "Photo") ? Card(
            child: ListTile(
              title: Text(
                credPrev[position].label,
                style: const TextStyle(fontSize: 14, fontFamily: 'SF Pro', fontWeight: FontWeight.w400, color: Color(0xff6C6D7C))
              ),
              subtitle: Text(
                credPrev[position].value,
                style: const TextStyle(
                    fontSize: 16,
                    color: Color(0xff190C21),
                    fontFamily: 'SF Pro',
                    fontWeight: FontWeight.normal),
              ),
            ),
          ):
          Card(
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [
                Expanded(
                  child: Padding(
                    padding: const EdgeInsets.all(24.0),
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.spaceAround,
                      crossAxisAlignment: CrossAxisAlignment.start,
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Text(credPrev[position].label, style: const TextStyle(fontSize: 14, fontFamily: 'SF Pro', fontWeight: FontWeight.w400, color: Color(0xff6C6D7C)),),
                      ],
                    ),
                  ),
                ),
                Flexible(
                  fit: FlexFit.tight,
                  child: Padding(
                    padding: const EdgeInsets.all(24.0),
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.spaceAround,
                      crossAxisAlignment: CrossAxisAlignment.end,
                      mainAxisSize: MainAxisSize.min,
                      children: <Widget>[
                    Image.memory(const Base64Decoder().convert(credPrev[position].value.split(',').last), width: 80, height: 80,),
                     ],
                    ),
                  ),
                ),
              ],
            ),
          );
        });
  }
  @override
  Widget build(BuildContext context) {
    Map<String, dynamic> issuer = jsonDecode(widget.credentialDisplay);
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
                       _storageService.addCredential(CredentialDataObject("$userLoggedIn-${uuid.v1()}",CredentialData(rawCredential: widget.rawCredential, credentialDisplayData: widget.credentialDisplay)));
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
    Navigator.push(context, MaterialPageRoute(builder: (context) => const Dashboard()));
  }
}