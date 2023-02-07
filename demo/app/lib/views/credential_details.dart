import 'dart:convert';
import 'package:app/widgets/Credential_card_outline.dart';
import 'package:app/widgets/credential_metadata_card.dart';
import 'package:app/widgets/credential_verified_information_view.dart';
import 'package:flutter/material.dart';

import 'package:app/models/credential_data.dart';
import 'package:intl/intl.dart';

class CredentialDetails extends StatelessWidget {
  CredentialData credentialData;
  bool isDashboardWidget = true;
  String credentialName;

  CredentialDetails({required this.credentialData, required this.isDashboardWidget, required this.credentialName, Key? key}) : super(key: key);

  final ScrollController credDataController = ScrollController();
  final ScrollController rawDataController = ScrollController();

  getCurrentDate() {
    final now = DateTime.now();
    // Todo instead of today's date always it will have to persist the date in the storage in the shared preference.
    String formatter = DateFormat('yMMMMd').format(now);// 28/03/2020
    return  formatter;
  }

  prettifyRawJson(){
    final parsedJson = json.decode(credentialData.credentialDisplayData!);
    final prettyString = const JsonEncoder.withIndent('  ').convert(parsedJson);
    return Text(prettyString);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Credential Details', textAlign: TextAlign.center, style: TextStyle(fontSize: 18, fontStyle: FontStyle.normal, fontWeight: FontWeight.w700, fontFamily: 'SF Pro')),
      backgroundColor: const Color(0xff261131),
      ),
      body: SingleChildScrollView(
        child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
          const SizedBox(height: 24.0),
          DefaultTabController(
              length: 2, // length of tabs
              initialIndex: 0,
              child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                const TabBar(
                  labelColor: Color(0xff190C21),
                  labelStyle: TextStyle(fontWeight: FontWeight.w700, fontSize: 14),
                  unselectedLabelColor: Color(0xff6C6D7C),
                  indicatorColor: Color(0xff8A35B7),
                  padding: EdgeInsets.fromLTRB(24, 0, 24, 0),
                  tabs: [
                    Tab(
                      child: SizedBox(
                        width: 100,
                        child: Text(
                          "Details",
                          textAlign: TextAlign.start,
                          style: TextStyle(
                               fontSize: 14,
                              fontWeight: FontWeight.w700,
                          ),
                        ),
                      ),
                    ),
                    Tab(  child: SizedBox(
                      width: 100,
                      child: Text(
                        "Raw Json",
                        textAlign: TextAlign.start,
                        style: TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                    ),
                    ),
                  ],
                ),
                Container(
                    height: MediaQuery.of(context).size.height*0.8,
                    padding: const EdgeInsets.all(24),
                    alignment: Alignment.center,
                    child:  TabBarView(children: <Widget>[
                    Column(
                    children: [
                      Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        Container(
                        height: 90,
                        padding: const EdgeInsets.all(16),
                        alignment: Alignment.topCenter,
                        decoration: BoxDecoration(
                            color: Colors.white,
                            borderRadius: BorderRadius.circular(12),
                            border: Border.all(
                              color: const Color(0xffDBD7DC),
                              width: 0.5,
                            ),
                           ),
                        child:
                        ListTile(
                          title: Text(
                            credentialName,
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
                        )

                    ),
                        isDashboardWidget?
                        const CredentialMetaDataCard(): Container(),
                        CredentialVerifiedInformation(credentialData: credentialData, height: MediaQuery.of(context).size.height*0.42,)
                      ],
                    )
             ]),
                      SizedBox(
                          height: 450,
                          child: Scrollbar(
                            thumbVisibility: true,
                            controller: rawDataController,
                            child: ListView.builder(
                                controller: rawDataController,
                                itemCount: 1,
                                itemBuilder: (BuildContext context, int index) {
                                  return Padding(
                                    padding: const EdgeInsets.all(8.0),
                                    child: prettifyRawJson(),
                                  );
                                }),
                          ))
                    ])
                )]),
          ),
        ])
      ),
    );
  }
}