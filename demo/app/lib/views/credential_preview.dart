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

  @override
  Widget build(BuildContext context) {
    widget.credentialResponse;
    return Scaffold(
      appBar: AppBar(
        title: const Text('Credential Preview'),
      ),
        body: Center(
            child:  Column (
              children: [
                const SizedBox(height: 50),
                const Text(
                  'Add a Verified ID',
                  textAlign: TextAlign.center,
                  style: TextStyle(fontSize: 22, color: Colors.black),
                ),
                const SizedBox(height: 30),
                MaterialButton(
                  onPressed: () { },
                  padding: const EdgeInsets.all(0.0),
                  child: Ink(
                    decoration: const BoxDecoration(
                      gradient: LinearGradient(
                        colors: <Color>[Colors.lightBlueAccent, Colors.blueAccent],
                      ),
                    ),
                    child: Container(
                      constraints: const BoxConstraints(maxWidth: 350, minHeight: 120.0), // min sizes for Material buttons
                      alignment: Alignment.center,
                      child: const Text(
                        'Permanent Resident Card',
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
                      _storageService.add(StorageItem("credential_prefix_${uuid.v1()})", widget.credentialResponse));
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