import 'dart:convert';

import 'package:flutter/material.dart';

import '../models/credential_data.dart';
import '../models/credential_preview.dart';
import 'package:intl/intl.dart';

class CredentialDetails extends StatelessWidget {
  CredentialData item;
  String credentialName;

  CredentialDetails({required this.item, required this.credentialName, Key? key}) : super(key: key);

  final ScrollController credDataController = ScrollController();
  final ScrollController rawDataController = ScrollController();

  getCurrentDate() {
    final now = DateTime.now();
    String formatter = DateFormat('yMMMMd').format(now);// 28/03/2020
    return  formatter;
  }

  prettifyRawJson(){
    final parsedJson = json.decode(item.credentialDisplayData);
    final prettyString = const JsonEncoder.withIndent('  ').convert(parsedJson);
    return Text(prettyString);
  }

  Widget getCredentialDetails() {
    List<CredentialPreviewData> list;
    var data = json.decode(item.credentialDisplayData);
    var credentialClaimsData = data['credential_displays'][0]['claims'] as List;
    list = credentialClaimsData.map<CredentialPreviewData>((json) => CredentialPreviewData.fromJson(json)).toList();
    return listViewWidget(list);
  }

  Widget listViewWidget(List<CredentialPreviewData> credPrev) {
    return ListView.builder(
        itemCount: credPrev.length,
        scrollDirection: Axis.vertical,
        controller: credDataController,
        shrinkWrap: true,
        itemBuilder: (context, position) {
          return (credPrev[position].label != "Photo") ? Row (
            children: [
              const Divider(
                thickness: 2,
                color: Color(0xffDBD7DC),
              ),
              Expanded(
                child: ListTile(
                title: Text(
                    credPrev[position].label,
                    style: const TextStyle(fontSize: 14, fontFamily: 'SF Pro', fontWeight: FontWeight.w400, color: Color(0xff6C6D7C))
                ),
                subtitle: Text(
                  credPrev[position].value,
                  style: const TextStyle(
                      fontSize: 16,
                      color: Color(0xff190C21),
                      fontFamily: 'SF Pro',
                      fontWeight: FontWeight.normal),
                ),
              ),
              ),
            ],
          ):
          Row(
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [
                Expanded(
                  child: Padding(
                    padding: const EdgeInsets.all(24.0),
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.spaceAround,
                      crossAxisAlignment: CrossAxisAlignment.start,
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Text(credPrev[position].label, style: const TextStyle(fontSize: 14, fontFamily: 'SF Pro', fontWeight: FontWeight.w400, color: Color(0xff6C6D7C)),),
                      ],
                    ),
                  ),
                ),
                Flexible(
                  fit: FlexFit.tight,
                  child: Padding(
                    padding: const EdgeInsets.all(24.0),
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.spaceAround,
                      crossAxisAlignment: CrossAxisAlignment.end,
                      mainAxisSize: MainAxisSize.min,
                      children: <Widget>[
                        Image.memory(const Base64Decoder().convert(credPrev[position].value.split(',').last), width: 80, height: 80,),
                      ],
                    ),
                  ),
                ),
              ],
          );
        });
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
                    height: MediaQuery.of(context).size.height*1.2,
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
                        Container(
                          decoration: BoxDecoration(
                              color: Colors.white,
                              borderRadius: const BorderRadius.only(
                                bottomLeft: Radius.circular(12),
                                bottomRight: Radius.circular(12),
                              ),
                              boxShadow: [
                                BoxShadow(
                                    color: Colors.grey.shade300,
                                    blurRadius: 4,
                                    offset: const Offset(4, 4),
                                )
                              ]),
                            padding: const EdgeInsets.fromLTRB(0, 0, 0, 16),
                            child:Column(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                Flexible(
                                    child: SizedBox(
                                      height: 60,
                                      child:  ListTile(
                                          title: const Text(
                                            'Added on',
                                            style: TextStyle(
                                              fontSize: 14,
                                              fontWeight: FontWeight.bold,
                                              color: Color(0xff190C21),
                                            ),
                                            textAlign: TextAlign.start,
                                          ),
                                          //TODO need to add fallback and network image url
                                          subtitle: Text(
                                            getCurrentDate(),
                                            style: const TextStyle(
                                              fontSize: 14,
                                              color: Color(0xff6C6D7C),
                                            ),
                                            textAlign: TextAlign.start,
                                          )
                                      ),
                                    )
                                ),
                                const Flexible(
                                    child: SizedBox(
                                        height: 60,
                                      child:    ListTile(
                                          title: Text(
                                            'Expires on',
                                            style: TextStyle(
                                              fontSize: 14,
                                              fontWeight: FontWeight.bold,
                                              color: Color(0xff190C21),
                                            ),
                                            textAlign: TextAlign.start,
                                          ),
                                          //TODO need to add fallback and network image url
                                          subtitle: Text(
                                            'Never',
                                            style: TextStyle(
                                              fontSize: 14,
                                              color: Color(0xff6C6D7C),
                                            ),
                                            textAlign: TextAlign.start,
                                          )
                                      )
                                    )
                                ),
                                const Flexible(
                                    child: SizedBox(
                                        height: 60,
                                      child: ListTile(
                                            title: Text(
                                              'Last used',
                                              style: TextStyle(
                                                fontSize: 14,
                                                fontWeight: FontWeight.bold,
                                                color: Color(0xff190C21),
                                              ),
                                              textAlign: TextAlign.start,
                                            ),
                                            //TODO need to add fallback and network image url
                                            subtitle: Text(
                                              'Never',
                                              style: TextStyle(
                                                fontSize: 14,
                                                color: Color(0xff6C6D7C),
                                              ),
                                              textAlign: TextAlign.start,
                                            )
                                        )
                                    )
                                )
                              ],
                          )

                        ),
                        Container(
                            height: 370,
                            padding: const EdgeInsets.fromLTRB(0, 24, 0, 0),
                            child: Column(
                                children: <Widget>[
                              Row(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: const <Widget>[
                              Text( "Verified information", style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700))
                                ]
                              ),
                              const SizedBox(height: 10),
                              const Divider(
                                thickness: 1,
                                color: Color(0xffC7C3C8),
                              ),
                                  SizedBox(
                                      height: 280,
                                      // When using the PrimaryScrollController and a Scrollbar
                                      // together, only one ScrollPosition can be attached to the
                                      // PrimaryScrollController at a time. Providing a
                                      // unique scroll controller to this scroll view prevents it
                                      // from attaching to the PrimaryScrollController.
                                      child: Scrollbar(
                                        thumbVisibility: true,
                                        controller: credDataController,
                                        child: ListView.builder(
                                            itemCount: 1,
                                            itemBuilder: (BuildContext context, int index) {
                                              return Padding(
                                                padding: const EdgeInsets.all(8.0),
                                                child: getCredentialDetails(),
                                              );
                                            }),
                                      )),
                          ]
                        ))
                      ],
                    )
             ]),
                      SizedBox(
                          height: 450,
                          // When using the PrimaryScrollController and a Scrollbar
                          // together, only one ScrollPosition can be attached to the
                          // PrimaryScrollController at a time. Providing a
                          // unique scroll controller to this scroll view prevents it
                          // from attaching to the PrimaryScrollController.
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