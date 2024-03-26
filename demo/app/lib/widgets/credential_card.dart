import 'dart:developer';

import 'package:app/models/credential_data.dart';
import 'package:app/views/credential_details.dart';
import 'package:app/widgets/primary_button.dart';
import 'package:flutter/material.dart';
import 'package:app/main.dart';
import 'credential_metadata_card.dart';
import 'credential_verified_information_view.dart';
import 'package:cached_network_image/cached_network_image.dart';

class CredentialCard extends StatefulWidget {
  CredentialData credentialData;
  bool isDashboardWidget;
  List<Object?>? activityLogger;
  bool isDetailArrowRequired;
  Function? delete;

  CredentialCard(
      {required this.credentialData,
      required this.isDashboardWidget,
      this.activityLogger,
      this.delete,
      required this.isDetailArrowRequired,
      Key? key})
      : super(key: key);

  @override
  State<CredentialCard> createState() => _CredentialCardState();
}

class _CredentialCardState extends State<CredentialCard> {
  bool showWidget = false;
  String credentialDisplayName = '';
  String? logoURL;
  String backgroundColor = '';
  String textColor = '';

  @override
  void initState() {
    WalletSDKPlugin.parseCredentialDisplayData(widget.credentialData.credentialDisplayData).then((response) {
      setState(() {
          credentialDisplayName = response.overviewName;
          logoURL = response.logo;
          backgroundColor = '0xff${response.backgroundColor.toString().replaceAll('#', '')}';
          textColor = '0xff${response.textColor.toString().replaceAll('#', '')}';
      });
    });
    super.initState();
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(0, 16, 0, 16),
      child: Column(
        children: [
          Container(
              height: 80,
              alignment: Alignment.center,
              decoration: BoxDecoration(
                  color: backgroundColor.isNotEmpty ? Color(int.parse(backgroundColor)) : Colors.white,
                  borderRadius: BorderRadius.circular(12),
                  boxShadow: [BoxShadow(offset: const Offset(3, 3), color: Colors.grey.shade300, blurRadius: 5)]),
              child: ListTile(
                title: Text(
                  credentialDisplayName,
                  style: TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.bold,
                    color: textColor.isNotEmpty ? Color(int.parse(textColor)) : const Color(0xff190C21),
                  ),
                  textAlign: TextAlign.start,
                ),
                leading: logoURL == null
                    ? const SizedBox.shrink()
                    : CachedNetworkImage(
                        imageUrl: logoURL!,
                        placeholder: (context, url) => const CircularProgressIndicator(),
                        errorWidget: (context, url, error) =>
                            Image.asset('lib/assets/images/genericCredential.png', fit: BoxFit.contain),
                        width: 60,
                        height: 60,
                        fit: BoxFit.contain,
                      ),
                trailing: Wrap(
                  spacing: -16,
                  children: <Widget>[
                    widget.isDetailArrowRequired == false
                        ? IconButton(
                            icon: const Icon(Icons.arrow_circle_right, size: 24, color: Color(0xffB6B7C7)),
                            onPressed: () async {
                              Navigator.push(
                                context,
                                MaterialPageRoute(
                                    builder: (context) => CredentialDetails(
                                          credentialData: widget.credentialData,
                                          credentialName: credentialDisplayName,
                                          isDashboardWidget: widget.isDashboardWidget,
                                          activityLogger: widget.activityLogger,
                                        )),
                              );
                            },
                          )
                        : IconButton(
                            icon: const Icon(Icons.expand_circle_down_sharp, size: 28, color: Color(0xffB6B7C7)),
                            onPressed: () async {
                              setState(() {
                                showWidget = !showWidget;
                              });
                            },
                          ), // icon-1
                    widget.isDashboardWidget == true
                        ? IconButton(
                            icon: const Icon(Icons.delete, size: 24, color: Color(0xffB6B7C7)),
                            onPressed: () async {
                              await showDialog(
                                context: context,
                                builder: (context) {
                                  return StatefulBuilder(
                                    builder: (context, setState) {
                                      return AlertDialog(
                                        shape: const RoundedRectangleBorder(
                                            borderRadius: BorderRadius.all(Radius.circular(10.0))),
                                        content: const Spacer(),
                                        title: const Text('Are you sure you want to delete the credential?',
                                            textAlign: TextAlign.center,
                                            style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                                        actions: <Widget>[
                                          Row(
                                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                            children: <Widget>[
                                              Row(
                                                children: <Widget>[
                                                  PrimaryButton(
                                                    onPressed: () => Navigator.pop(context, false),
                                                    gradient: const LinearGradient(
                                                        begin: Alignment.topCenter,
                                                        end: Alignment.bottomCenter,
                                                        colors: [Color(0xffFFFFFF), Color(0xffFFFFFF)]),
                                                    child: const Text('Cancel',
                                                        style: TextStyle(fontSize: 16, color: Color(0xff6C6D7C))),
                                                  ),
                                                ],
                                              ),
                                              PrimaryButton(
                                                  onPressed: () async {
                                                    Navigator.pop(context, true);
                                                    widget.delete!();
                                                  },
                                                  child: const Text('Delete',
                                                      style: TextStyle(fontSize: 16, color: Colors.white))),
                                            ],
                                          ),
                                        ],
                                      );
                                    },
                                  );
                                },
                              );
                            },
                          )
                        : const Column(),
                  ],
                ),
              )),
          showWidget
              ? Column(
                  children: [
                    CredentialMetaDataCard(credentialData: widget.credentialData),
                    CredentialVerifiedInformation(
                      credentialData: widget.credentialData,
                      height: MediaQuery.of(context).size.height * 0.38,
                    ),
                  ],
                )
              : Container(),
        ],
      ),
    );
  }
}
