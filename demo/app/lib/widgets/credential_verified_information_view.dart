import 'dart:convert';
import 'package:app/wallet_sdk/wallet_sdk_model.dart';
import 'package:flutter/material.dart';
import 'package:app/models/credential_data.dart';

import 'package:app/main.dart';

class CredentialVerifiedInformation extends StatefulWidget {
  CredentialData credentialData;
  double? height;

  CredentialVerifiedInformation({required this.credentialData, this.height, Key? key}) : super(key: key);

  @override
  State<CredentialVerifiedInformation> createState() => CredentialVerifiedState();
}

class CredentialVerifiedState extends State<CredentialVerifiedInformation> {
  bool showMaskedValue = true;
  Color maskIconColor = Colors.blueGrey;
  final ScrollController credDataController = ScrollController();
  List<CredentialDisplayClaim> credentialClaimsData = [];
  bool isLoading = false;
  @override
  void initState() {
    setState(() {
      isLoading = true;
    });
    WalletSDKPlugin.parseCredentialDisplayData(widget.credentialData.credentialDisplayData).then((response) {
      setState(() {
        if(response.claims.isNotEmpty){
        credentialClaimsData = response.claims;

        credentialClaimsData.sort((a, b) {
          final aOrder = a.order ?? -1;
          final bOrder = b.order ?? -1;

          return aOrder.compareTo(bOrder);
        });

        isLoading = false;
      }});
    });
    super.initState();
  }

  Widget listViewWidget(List<CredentialDisplayClaim> credPrev) {
    return ListView.builder(
        itemCount: credPrev.length,
        scrollDirection: Axis.vertical,
        controller: credDataController,
        shrinkWrap: true,
        itemBuilder: (context, position) {
          if (credPrev[position].valueType == 'string') {
            return Row(
              children: [
                const Divider(
                  thickness: 2,
                  color: Color(0xffDBD7DC),
                ),
                Expanded(
                  child: ListTile(
                      title: Text(credPrev[position].label,
                          style: const TextStyle(
                              fontSize: 14,
                              fontFamily: 'SF Pro',
                              fontWeight: FontWeight.w400,
                              color: Color(0xff6C6D7C))),
                      subtitle: showMaskedValue && credPrev[position].value != null
                          ? Text(
                              credPrev[position].value!,
                              style: const TextStyle(
                                  fontSize: 16,
                                  color: Color(0xff190C21),
                                  fontFamily: 'SF Pro',
                                  fontWeight: FontWeight.normal),
                            )
                          : Text(
                              credPrev[position].rawValue,
                              style: const TextStyle(
                                  fontSize: 16,
                                  color: Color(0xff190C21),
                                  fontFamily: 'SF Pro',
                                  fontWeight: FontWeight.normal),
                            ),
                      trailing: credPrev[position].value != null
                          ? Column(
                              children: [
                                IconButton(
                                  icon: Icon(Icons.remove_red_eye_rounded, size: 32, color: maskIconColor),
                                  onPressed: () async {
                                    setState(() {
                                      showMaskedValue = !showMaskedValue;
                                      if (maskIconColor == Colors.blueGrey) {
                                        maskIconColor = Colors.green;
                                      } else {
                                        maskIconColor = Colors.blueGrey;
                                      }
                                    });
                                  },
                                )
                              ],
                            )
                          : const Column()),
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
                        Text(
                          credPrev[position].label,
                          style: const TextStyle(
                              fontSize: 14,
                              fontFamily: 'SF Pro',
                              fontWeight: FontWeight.w400,
                              color: Color(0xff6C6D7C)),
                        ),
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
                       credPrev[position].uri == null ?
                       Image.memory(
                         const Base64Decoder().convert(credPrev[position].rawValue.split(',').last),
                         width: 80,
                         height: 80,
                       ):
                       Image.memory(
                         const Base64Decoder().convert(credPrev[position].uri!.split(',').last),
                         width: 80,
                         height: 80,
                       )
                      ],
                    ),
                  ),
                ),
              ],
            );
          }
        });
  }

  @override
  Widget build(BuildContext context) {
    return isLoading
        ? const Center(child: LinearProgressIndicator())
        : Container(
            height: widget.height,
            padding: const EdgeInsets.fromLTRB(0, 24, 0, 0),
            child: SingleChildScrollView(
                child: Column(children: <Widget>[
              const Row(crossAxisAlignment: CrossAxisAlignment.start, children: <Widget>[
                Text('Verified information', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700))
              ]),
              const SizedBox(height: 10),
              const Divider(
                thickness: 1,
                color: Color(0xffC7C3C8),
              ),
              SizedBox(
                  height: MediaQuery.of(context).size.height * 0.3,
                  child: Scrollbar(
                    thumbVisibility: true,
                    controller: credDataController,
                    child: ListView.builder(
                        itemCount: 1,
                        itemBuilder: (BuildContext context, int index) {
                          return Padding(
                              padding: const EdgeInsets.all(8.0), child: listViewWidget(credentialClaimsData));
                        }),
                  )),
            ])));
  }
}
