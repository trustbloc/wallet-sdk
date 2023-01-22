import 'package:app/models/credential_data.dart';
import 'package:flutter/material.dart';
import 'package:app/widgets/common_title_appbar.dart';
import 'package:app/widgets/credential_card.dart';

class CredentialAdded extends StatefulWidget {
  CredentialData credentialData;

  CredentialAdded({required this.credentialData, Key? key}) : super(key: key);

  @override
  State<CredentialAdded> createState() => CredentialAddedPage();
}
class CredentialAddedPage extends State<CredentialAdded> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
            appBar: CustomTitleAppBar(pageTitle: 'Credential Added', addCloseIcon: true),
            backgroundColor: const Color(0xffF4F1F5),
            body: Center(
              child: ListView(
                padding: const EdgeInsets.fromLTRB(24,24, 24, 0),
                children: [
                  ListTile(
                    horizontalTitleGap: 0.5,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                    tileColor: const Color(0xffFBF8FC),
                    //todo Issue-174 read the meta data from the backend on page load
                    leading: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      crossAxisAlignment: CrossAxisAlignment.center,
                      children: [
                        Image.asset('lib/assets/images/success.png')
                      ],
                    ),
                    title: const Text('Success', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                    subtitle: const Text('Credential has been added to your wallet', style: TextStyle(fontSize: 12, fontWeight: FontWeight.normal)),
                  ),
                  CredentialCard(item: widget.credentialData, isDashboardWidget: false),
                ],
              ),
            ));
  }
}

