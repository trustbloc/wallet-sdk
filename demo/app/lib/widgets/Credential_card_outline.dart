

import 'dart:convert';
import 'package:app/models/credential_data.dart';
import 'package:flutter/material.dart';

class CredentialCardOutline extends StatelessWidget {
  CredentialData item;

  CredentialCardOutline({required this.item,  Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    Map<String, dynamic> issuer = jsonDecode(item.credentialDisplayData!);
    final credentialDisplayName = issuer['credential_displays'][0]['overview']['name'];
    return Padding(
      padding: const EdgeInsets.fromLTRB(0, 0, 0, 2),
      child: Container(
          height: 80,
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
          )),
    );
  }
}
