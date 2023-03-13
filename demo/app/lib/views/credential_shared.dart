import 'package:app/widgets/success_card.dart';
import 'package:flutter/material.dart';
import 'package:app/models/credential_data.dart';
import 'dashboard.dart';

class CredentialShared extends StatefulWidget {
  final String? verifierName;
  CredentialData credentialData;
  CredentialShared({super.key, this.verifierName, required this.credentialData,});

  @override
  State<CredentialShared> createState() => CredentialSharedState ();
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
          ),
        ],
        flexibleSpace: Container(
          decoration: const BoxDecoration(
              gradient: LinearGradient(
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                  stops: [0.0, 1.0],
                  colors: <Color>[
                    Color(0xff261131),
                    Color(0xff100716),
                  ])
          ),
        ),
      ),
      body: Center(
        child: SuccessCard(credentialData: widget.credentialData, verifierName: widget.verifierName, subTitle: 'Credential has been shared with ',),
      ),
    );
  }
  _navigateToDashboard() async {
    Navigator.push(context, MaterialPageRoute(builder: (context) => const Dashboard()));
  }
}