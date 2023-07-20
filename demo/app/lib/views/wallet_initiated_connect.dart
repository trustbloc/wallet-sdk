/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'dart:convert';
import 'dart:developer';
import 'package:app/models/connect_issuer_config.dart';
import 'package:app/models/connect_issuer_config_value.dart';
import 'package:app/services/config_service.dart';
import 'package:app/views/supported_credentials_list.dart';
import 'package:app/widgets/primary_button.dart';
import 'package:flutter/material.dart';
import 'package:app/widgets/common_title_appbar.dart';
import 'package:flutter/services.dart';
import 'package:app/wallet_sdk/wallet_sdk_mobile.dart';

class ConnectIssuerList extends StatefulWidget {
  const ConnectIssuerList({Key? key}) : super(key: key);

  @override
  State<ConnectIssuerList> createState() => ConnectIssuerListState();
}

class ConnectIssuerListState extends State<ConnectIssuerList> {
  List<ConnectIssuerConfig> connectIssuerConfigList = List.empty(growable: true);

  var WalletSDKPlugin = WalletSDK();
  final ConfigService _configService = ConfigService();


  void connect(String issuerURI) async {
    var connectResp = await WalletSDKPlugin.initializeWalletInitiatedFlow(issuerURI);
    //TODO: Implement the supported credential List dynamically
    log("$connectResp");
  }

  readConnectIssuerConfig() async {
    connectIssuerConfigList = await _configService.readConnectIssuerConfig();
    log("${connectIssuerConfigList.length}");
    setState(() {});
  }

  @override
  void initState() {
    super.initState();
    readConnectIssuerConfig();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: const CustomTitleAppBar(
        pageTitle: 'Wallet Supported Issuer',
        addCloseIcon: false,
        height: 50,
      ),
      body: Column(
        children: <Widget>[
          const SizedBox(height:20),
     Expanded(
       child: ListView.builder(
         padding: const EdgeInsets.fromLTRB(12, 12, 12, 12),
      itemCount: connectIssuerConfigList.length,
      itemBuilder: (context, index) {
        return Card(
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(12),
            ),
            color: Colors.grey.shade200,
            clipBehavior: Clip.antiAliasWithSaveLayer,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Container(
                  padding: const EdgeInsets.fromLTRB(12, 12, 12, 8),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      Text(
                        connectIssuerConfigList.elementAt(index).key,
                        style: const TextStyle(
                          fontSize: 14,
                          color: Color(0xff190C21),
                        ),
                      ),
                      // Add a space between the title and the text
                      Container(height: 10),
                      Text(
                        connectIssuerConfigList.elementAt(index).value.description,
                        style: TextStyle(
                          fontSize: 12,
                          color: Colors.grey[700],
                        ),
                      ),
                      Container(height: 10),
                      Row(
                        children: <Widget>[
                          const Spacer(),
                          PrimaryButton(
                            child: const Text(
                              "Connect",
                              style: TextStyle(color: Colors.white),
                            ),
                            onPressed: () {
                              connect(connectIssuerConfigList.elementAt(index).value.issuerURI);
                              Navigator.push(
                                context,
                                MaterialPageRoute(builder: (context) => SupportedCredentialsList(issuerName: connectIssuerConfigList.elementAt(index).key,)),
                              );
                            },
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
                Container(height: 5),
              ],
            ),
          );
       },
      ),
     ),
      ],
     ),
    );
  }
}
