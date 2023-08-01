/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'dart:developer';

import 'package:app/models/connect_issuer_config.dart';
import 'package:app/models/connect_issuer_config_value.dart';
import 'package:app/widgets/primary_button.dart';
import 'package:flutter/material.dart';

import 'package:app/wallet_sdk/wallet_sdk_mobile.dart';
import 'package:app/wallet_sdk/wallet_sdk_model.dart';
import 'package:app/widgets/common_title_appbar.dart';

import 'package:app/services/config_service.dart';

import 'handle_redirect_uri.dart';

class SupportedCredentialsList extends StatefulWidget {
  String issuerName;
  String? issuerURI;
  final List<SupportedCredentials> supportedCredentialList;
  final ConnectIssuerConfigValue connectIssuerConfigValue;

  SupportedCredentialsList(
      {required this.issuerName,
      required this.supportedCredentialList,
      required this.connectIssuerConfigValue,
      this.issuerURI,
      Key? key})
      : super(key: key);

  @override
  State<StatefulWidget> createState() {
    return SupportedCredentialsListState();
  }
}

class SupportedCredentialsListState extends State<SupportedCredentialsList> {
  var walletSDKPlugin = WalletSDK();

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: CustomTitleAppBar(
        pageTitle: '${widget.issuerName} Credentials',
        addCloseIcon: true,
        height: 50,
      ),
      body: Column(
        children: <Widget>[
          const SizedBox(height: 20),
          Expanded(
            child: ListView.builder(
              padding: const EdgeInsets.fromLTRB(12, 12, 12, 12),
              itemCount: widget.supportedCredentialList.length,
              itemBuilder: (context, index) {
                return Card(
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                  color: Colors.grey.shade200,
                  clipBehavior: Clip.antiAliasWithSaveLayer,
                  child: ListTile(
                    title: widget.supportedCredentialList.elementAt(index).types[index] == "VerifiableCredential"
                        ? Text(
                            widget.supportedCredentialList.elementAt(index).types[index + 1],
                            style: const TextStyle(
                              fontSize: 14,
                              color: Color(0xff190C21),
                            ),
                          )
                        : null,
                    leading: IconButton(
                      icon: const Icon(Icons.credit_card_sharp),
                      tooltip: 'Credential',
                      onPressed: () {},
                    ),
                    trailing: PrimaryButton(
                      child: const Text(
                        "Request",
                        style: TextStyle(fontSize: 12, color: Colors.white),
                      ),
                      onPressed: () async {
                        log("scopes ${widget.connectIssuerConfigValue.scopes}");
                        log("clientID ${widget.connectIssuerConfigValue.clientID}");
                        log("redirectURI ${widget.connectIssuerConfigValue.redirectURI}");
                        var authorizationURL = await walletSDKPlugin.createAuthorizationURLWalletInitiatedFlow(
                            widget.connectIssuerConfigValue.scopes,
                            widget.supportedCredentialList.elementAt(index).types,
                            widget.supportedCredentialList.elementAt(index).format,
                            widget.connectIssuerConfigValue.clientID,
                            widget.connectIssuerConfigValue.redirectURI,
                            widget.issuerURI);
                        Uri uri = Uri.parse(authorizationURL);
                        navigateToAuthFlow(context, uri, widget.issuerURI);
                      },
                    ),
                  ),
                );
              },
            ),
          ),
        ],
      ),
    );
  }

  void navigateToAuthFlow(BuildContext context, Uri uri, String? issuerURI) async {
    Navigator.of(context)
        .push(MaterialPageRoute(builder: (context) => HandleRedirectUri(uri, "wallet-initiated-flow", issuerURI)));
  }
}
