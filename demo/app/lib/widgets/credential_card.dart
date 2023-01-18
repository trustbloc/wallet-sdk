import 'dart:convert';

import 'package:app/models/credential_data.dart';
import 'package:app/views/credential_details.dart';
import 'package:flutter/material.dart';
import 'package:app/models/credential_preview.dart';

class CredentialCard extends StatefulWidget {
  CredentialData item;

  CredentialCard({required this.item, Key? key}) : super(key: key);
  @override
  State<CredentialCard> createState() => _CredentialCardState();
}

class _CredentialCardState extends State<CredentialCard> {

  Widget getCredentialDetails() {
    List<CredentialPreviewData> list;
    var data = json.decode(widget.item.credentialDisplayData);
    var credentialClaimsData = data['credential_displays'][0]['claims'] as List;
    list = credentialClaimsData.map<CredentialPreviewData>((json) => CredentialPreviewData.fromJson(json)).toList();
    return listViewWidget(list);
  }

  Widget listViewWidget(List<CredentialPreviewData> credPrev) {
    return ListView.builder(
        itemCount: credPrev.length,
        scrollDirection: Axis.vertical,
        itemBuilder: (context, position) {
          //TODO Ignoring the photo value for now due to extremely long text need to render in a separate issue-150
          return (credPrev[position].label != "Photo") ? Card(
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
    Map<String, dynamic> issuer = jsonDecode(widget.item.credentialDisplayData);
    final credentialDisplayName = issuer['credential_displays'][0]['overview']['name'];
    return Padding(
      padding: const EdgeInsets.fromLTRB(24, 60, 24, 24),
      child: Container(
          height: 70,
          alignment: Alignment.center,
          decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(12),
              boxShadow: [
                BoxShadow(
                    offset: const Offset(3, 3),
                    color: Colors.grey.shade300,
                    blurRadius: 5)
              ]),
          child: ListTile(
            title: Text(
              credentialDisplayName,
              style: const TextStyle(
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
            trailing: IconButton(
              icon: const Icon(Icons.arrow_circle_right, size: 32, color: Color(0xffB6B7C7)),
              onPressed: () async {
                Navigator.push(
                  context,
                  MaterialPageRoute(builder: (context) => CredentialDetails(item: widget.item,credentialName: credentialDisplayName,)),
                );
              },
            ),
          )),
    );
  }
}