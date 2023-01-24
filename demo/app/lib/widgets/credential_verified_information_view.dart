import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:app/models/credential_data.dart';
import 'package:app/models/credential_preview.dart';

class CredentialVerifiedInformation extends StatelessWidget {

  CredentialData credentialData;
  double? height;
  CredentialVerifiedInformation({required this.credentialData, this.height,   Key? key}) : super(key: key);

  final ScrollController credDataController = ScrollController();

  Widget getCredentialDetails() {
    List<CredentialPreviewData> list;
    var data = json.decode(credentialData.credentialDisplayData);
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
    return Container(
        height: height,
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
                  height: MediaQuery.of(context).size.height*0.3,
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
        ));
  }
}