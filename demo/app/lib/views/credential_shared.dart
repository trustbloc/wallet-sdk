/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'package:flutter/material.dart';
import 'package:app/models/credential_data.dart';
import 'package:app/widgets/credential_card.dart';
import 'package:app/views/dashboard.dart';

class CredentialShared extends StatefulWidget {
  final String? verifierName;
  final List<CredentialData> credentialData;
  const CredentialShared({
    super.key,
    this.verifierName,
    required this.credentialData,
  });

  @override
  State<CredentialShared> createState() => CredentialSharedState();
}

class CredentialSharedState extends State<CredentialShared> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Credential Shared'),
        actions: [
          IconButton(
            onPressed: () {
              _navigateToDashboard();
            },
            icon: const Icon(Icons.close),
            color: Colors.white,
          ),
        ],
        flexibleSpace: Container(
          decoration: const BoxDecoration(
              gradient: LinearGradient(begin: Alignment.topLeft, end: Alignment.bottomRight, stops: [
            0.0,
            1.0
          ], colors: <Color>[
            Color(0xff261131),
            Color(0xff100716),
          ])),
        ),
      ),
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: [
          const SizedBox(height: 24),
          ListTile(
            horizontalTitleGap: 0.5,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(12),
            ),
            tileColor: const Color(0xffFBF8FC),
            leading: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [Image.asset('lib/assets/images/success.png')],
            ),
            title: const Text('Success',
                textAlign: TextAlign.left, style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
            subtitle: Text('Credentials have been shared with ${widget.verifierName}',
                style: const TextStyle(fontSize: 14, fontWeight: FontWeight.normal)),
          ),
          Expanded(
            child: ListView.builder(
              padding: const EdgeInsets.only(left: 24, right: 24, top: 24, bottom: 8),
              itemCount: widget.credentialData.length,
              itemBuilder: (BuildContext context, int index) {
                return CredentialCard(
                  credentialData: widget.credentialData[index],
                  isDashboardWidget: false,
                  isDetailArrowRequired: false,
                );
              },
            ),
          )
        ],
      ),
    );
  }

  _navigateToDashboard() async {
    Navigator.push(context, MaterialPageRoute(builder: (context) => const Dashboard()));
  }
}
