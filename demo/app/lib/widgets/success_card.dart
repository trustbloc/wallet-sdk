import 'package:app/models/credential_data.dart';
import 'package:flutter/material.dart';
import 'package:app/widgets/credential_card.dart';

class SuccessCard extends StatelessWidget {
  String? verifierName;
  String? subTitle;
  List<CredentialData> credentialDatas;

  SuccessCard({required this.credentialDatas, this.verifierName, this.subTitle, Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Scaffold(
        backgroundColor: const Color(0xffF4F1F5),
        body: Center(
          child: ListView(
            padding: const EdgeInsets.fromLTRB(24, 24, 24, 0),
            children: [
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
                title: const Text('Success', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                subtitle: Text(subTitle! + verifierName!,
                    style: const TextStyle(fontSize: 12, fontWeight: FontWeight.normal)),
              ),
              for (final credData in credentialDatas)
                CredentialCard(
                  credentialData: credData,
                  isDashboardWidget: false,
                  isDetailArrowRequired: false,
                ),
            ],
          ),
        ));
  }
}
