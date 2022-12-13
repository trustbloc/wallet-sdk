import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:app/models/store_credential_data.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:app/models/credential_preview.dart';

class CredentialCard extends StatefulWidget {
  StorageItem item;

  CredentialCard({required this.item, Key? key}) : super(key: key);

  @override
  State<CredentialCard> createState() => _CredentialCardState();
}

class _CredentialCardState extends State<CredentialCard> {
  bool _visibility = false;

  Widget getCredentialDetails() {
    List<CredentialPreviewData> list;
    var data = json.decode(widget.item.value);
    var credentialClaimsData = data['credential_displays'][0]['claims'] as List;
    list = credentialClaimsData.map<CredentialPreviewData>((json) => CredentialPreviewData.fromJson(json)).toList();
    return listViewWidget(list);
  }

  Widget listViewWidget(List<CredentialPreviewData> credPrev) {
    return ListView.builder(
        itemCount: credPrev.length,
        scrollDirection: Axis.vertical,
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
    Map<String, dynamic> issuer = jsonDecode(widget.item.value);
    final credentialDisplayName = issuer['credential_displays'][0]['overview']['name'];
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Container(
          decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(16),
              boxShadow: [
                BoxShadow(
                    offset: const Offset(3, 3),
                    color: Colors.grey.shade300,
                    blurRadius: 5)
              ]),
          child: ListTile(
            onLongPress: () {
              setState(() {
                _visibility = !_visibility;
              });
            },
            title: Text(
              credentialDisplayName,
              style: const TextStyle(
                fontSize: 18,
              ),
            ),
            subtitle: Visibility(
              visible: _visibility,
              child: SizedBox(
                height: 500,
                child: Column(
                  children: [
                    Flexible( // wrap in Expanded
                      child: getCredentialDetails(),
                    ),
                  ],
                ),
              ),
            ),
            leading: const Icon(Icons.card_membership_outlined, size: 30),
            trailing: IconButton(
              icon: const Icon(Icons.arrow_downward),
              onPressed: () async {

              },
            ),
          )),
    );
  }
}
