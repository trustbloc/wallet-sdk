import 'package:app/models/credential_data.dart';
import 'package:flutter/material.dart';
import 'package:app/widgets/common_title_appbar.dart';
import 'package:app/widgets/success_card.dart';

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
            appBar: const CustomTitleAppBar(pageTitle: 'Credential Added', addCloseIcon: true, height: 60,),
            backgroundColor: const Color(0xffF4F1F5),
            body: Center(
              child: SuccessCard(credentialData: widget.credentialData, verifierName: '', subTitle: 'Credential has been added to your wallet',),
            ));
  }
}

