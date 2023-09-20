/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'package:app/models/connect_issuer_config_value.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';

import 'package:app/wallet_sdk/wallet_sdk_mobile.dart';
import 'package:app/wallet_sdk/wallet_sdk_model.dart';
import 'package:app/widgets/common_title_appbar.dart';
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

  String convertToFlutterColor(String color) {
    return '0xff${color.toString().replaceAll('#', '')}';
  }

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
          const Text(
            'Please select a credential type from below ',
            style: TextStyle(fontSize: 14, color: Color(0xff190C21), fontFamily: 'SF Pro', fontWeight: FontWeight.bold),
          ),
          Expanded(
            child: ListView.builder(
              padding: const EdgeInsets.fromLTRB(24, 12, 24, 12),
              itemCount: widget.supportedCredentialList.length,
              itemBuilder: (context, index) {
                return Container(
                  height: 80,
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                      color: (widget.supportedCredentialList.elementAt(index).display.first.backgroundColor != null)
                          ? Color(int.parse(convertToFlutterColor(
                              widget.supportedCredentialList.elementAt(index).display.first.backgroundColor!)))
                          : Colors.white,
                      borderRadius: BorderRadius.circular(12),
                      boxShadow: [BoxShadow(offset: const Offset(3, 3), color: Colors.grey.shade300, blurRadius: 5)]),
                  child: ListTile(
                    title: Text(
                      widget.supportedCredentialList.elementAt(index).display.first.name,
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.bold,
                        color: (widget.supportedCredentialList.elementAt(index).display.first.textColor != null)
                            ? Color(int.parse(convertToFlutterColor(
                                widget.supportedCredentialList.elementAt(index).display.first.textColor!)))
                            : const Color(0xff190C21),
                      ),
                      textAlign: TextAlign.start,
                    ),
                    leading: widget.supportedCredentialList.elementAt(index).display.first.logo == null
                        ? const SizedBox.shrink()
                        : CachedNetworkImage(
                            imageUrl: widget.supportedCredentialList.elementAt(index).display.first.logo!,
                            placeholder: (context, url) => const CircularProgressIndicator(),
                            errorWidget: (context, url, error) =>
                                Image.asset('lib/assets/images/genericCredential.png', fit: BoxFit.contain),
                            width: 60,
                            height: 60,
                            fit: BoxFit.contain,
                          ),
                    trailing: ElevatedButton(
                      style: ElevatedButton.styleFrom(
                        backgroundColor:
                            (widget.supportedCredentialList.elementAt(index).display.first.textColor != null)
                                ? Color(int.parse(convertToFlutterColor(
                                    widget.supportedCredentialList.elementAt(index).display.first.textColor!)))
                                : Colors.white,
                        shadowColor: Colors.transparent,
                        shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(8), side: const BorderSide(color: Color(0xffC7C3C8))),
                      ),
                      onPressed: () async {
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
                      child: const Text(
                        'Request',
                        style: TextStyle(fontSize: 12, color: Color(0xff190C21)),
                      ),
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
        .push(MaterialPageRoute(builder: (context) => HandleRedirectUri(uri, 'wallet-initiated-flow', issuerURI)));
  }
}
