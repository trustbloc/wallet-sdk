/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'package:app/widgets/primary_button.dart';
import 'package:flutter/material.dart';

import '../widgets/common_title_appbar.dart';

class SupportedCredentialsList extends StatefulWidget {
  String issuerName;

  SupportedCredentialsList({required this.issuerName, Key? key}) : super(key: key);

  @override
  State<StatefulWidget> createState() {
    return SupportedCredentialsListState();
  }
}

class SupportedCredentialsListState extends State<SupportedCredentialsList> {
  //TODO: To be removed in the follow up pr
  final List<String> supportedCredentials = [
    'Permanent Resident Card',
    'Vaccination Certificate',
    'Verified Employee',
  ];
  String? selectedDIDType;
  String? selectedKeyType;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: CustomTitleAppBar(
        pageTitle: '${widget.issuerName} Credentials',
        addCloseIcon: true,
        height: 50,
      ),
        body: Column(
          children: <Widget>[
            const SizedBox(height:20),
            Expanded(
              child: ListView.builder(
                padding: const EdgeInsets.fromLTRB(12, 12, 12, 12),
                itemCount: supportedCredentials.length,
                itemBuilder: (context, index) {
                  return Card(
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                    color: Colors.grey.shade200,
                    clipBehavior: Clip.antiAliasWithSaveLayer,
                    child: ListTile(
                      title:   Text(
                        supportedCredentials[index],
                        style: const TextStyle(
                          fontSize: 14,
                          color: Color(0xff190C21),
                        ),
                      ),
                      leading: IconButton(
                        icon: const Icon(Icons.credit_card_sharp),
                        tooltip: 'Credential',
                        onPressed: (){},
                      ),
                      trailing:  PrimaryButton(
                        child: const Text(
                          "Request",
                          style: TextStyle(fontSize: 12, color: Colors.white),
                        ),
                        onPressed: () {
                          // Todo: Call request credential here
                        },
                      ),
                    ),
                  );
                },
              ),
            ),
          ],
        ),
    );
  }
}