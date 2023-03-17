import 'dart:convert';
import 'dart:developer';
import 'package:flutter/material.dart';
import 'package:app/models/credential_data.dart';
import 'package:app/models/credential_preview.dart';

import 'package:app/main.dart';

class CredentialVerifiedInformation extends StatefulWidget {

  CredentialData credentialData;
  double? height;

  CredentialVerifiedInformation({required this.credentialData, this.height, Key? key}) : super(key: key);

  @override
  State<CredentialVerifiedInformation> createState() => CredentialVerifiedState();
}
  class CredentialVerifiedState extends State< CredentialVerifiedInformation> {
    bool showMaskedValue = true;
    Color maskIconColor = Colors.blueGrey;
    final ScrollController credDataController = ScrollController();
    dynamic credentialClaimsData;
    bool isLoading = false;
    @override
    void initState() {
      setState(() {
        isLoading = true;
      });
      WalletSDKPlugin.resolveCredDisplayRendering(widget.credentialData.credentialDisplayData).then(
              (response) {
            setState(() {
              var credentialDisplayEncodeData = json.encode(response);
              List<dynamic> responseJson = json.decode(credentialDisplayEncodeData);
              credentialClaimsData = responseJson.first['claims'];
              isLoading = false;
            });
          });
      super.initState();
    }

  Future<Widget> getCredentialDetails() async {
    var credentialClaimsDataList = credentialClaimsData as List;
    List<CredentialPreviewData> list = credentialClaimsDataList.map<CredentialPreviewData>((json) => CredentialPreviewData.fromJson(json)).toList();
      list.sort((a, b) {
        int compare = a.order.compareTo(b.order);
        return compare;
      });
    return listViewWidget(list);
  }


  Widget listViewWidget(List<CredentialPreviewData> credPrev) {
    return ListView.builder(
        itemCount: credPrev.length,
        scrollDirection: Axis.vertical,
        controller: credDataController,
        shrinkWrap: true,
        itemBuilder: (context, position) {
          if (credPrev[position].valueType != "image"){
              return Row(
                children: [
                  const Divider(
                    thickness: 2,
                    color: Color(0xffDBD7DC),
                  ),
                  Expanded(
                    child: ListTile(
                      title: Text(
                          credPrev[position].label,
                          style: const TextStyle(
                              fontSize: 14, fontFamily: 'SF Pro', fontWeight: FontWeight.w400, color: Color(0xff6C6D7C))
                      ),
                      subtitle: showMaskedValue && credPrev[position].value.isNotEmpty? Text(
                        credPrev[position].value,
                        style: const TextStyle(
                            fontSize: 16,
                            color: Color(0xff190C21),
                            fontFamily: 'SF Pro',
                            fontWeight: FontWeight.normal),
                      ):  Text(
                        credPrev[position].rawValue,
                        style: const TextStyle(
                            fontSize: 16,
                            color: Color(0xff190C21),
                            fontFamily: 'SF Pro',
                            fontWeight: FontWeight.normal),
                      ),
                      trailing: credPrev[position].value.isNotEmpty?  Column(
                        children: [
                          IconButton(
                            icon: Icon(Icons.remove_red_eye_rounded, size: 32,
                                color: maskIconColor),
                            onPressed: () async {
                              setState(() {
                                showMaskedValue = !showMaskedValue;
                                if(maskIconColor == Colors.blueGrey){
                                    maskIconColor = Colors.green;
                                    }else{
                                    maskIconColor = Colors.blueGrey;
                              }});
                            },
                          )
                        ],
                      ) :  Column()
                    ),
                  ),
                ],
              );
          } else {
          return Row(
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
                      Image.memory(const Base64Decoder().convert(credPrev[position].rawValue.split(',').last), width: 80, height: 80,),
                    ],
                  ),
                ),
              ),
            ],
          );
        }});
  }

  @override
  Widget build(BuildContext context) {
    return isLoading ? const Center(child: LinearProgressIndicator()) : Container(
        height: widget.height,
        padding: const EdgeInsets.fromLTRB(0, 24, 0, 0),
        child: SingleChildScrollView(
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
                          return FutureBuilder<Widget>(
                              future:  getCredentialDetails(),
                              builder: (context, AsyncSnapshot<Widget> snapshot) {
                                return Padding(padding: const EdgeInsets.all(8.0), child: snapshot.data,);
                              }
                          );
                        }),
                  )),
            ]
        )));
  }
}