/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'package:app/models/connect_issuer_config.dart';
import 'package:app/models/connect_issuer_config_value.dart';
import 'package:app/services/config_service.dart';
import 'package:app/views/supported_credentials_list.dart';
import 'package:app/wallet_sdk/wallet_sdk_model.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:app/widgets/common_title_appbar.dart';
import 'package:app/wallet_sdk/wallet_sdk_mobile.dart';
import 'package:flutter/services.dart';
import 'package:shared_preferences/shared_preferences.dart';

class ConnectIssuerList extends StatefulWidget {
  const ConnectIssuerList({Key? key}) : super(key: key);

  @override
  State<ConnectIssuerList> createState() => ConnectIssuerListState();
}

class ConnectIssuerListState extends State<ConnectIssuerList> {
  List<ConnectIssuerConfig> connectIssuerConfigList = List.empty(growable: true);
  final Future<SharedPreferences> prefs = SharedPreferences.getInstance();

  var walletSDKPlugin = WalletSDK();
  final ConfigService _configService = ConfigService();
  List<String>? credentialTypes;

  Future<List<SupportedCredentials>> connect(String issuerURI) async {
    final SharedPreferences pref = await prefs;
    credentialTypes = pref.getStringList('credentialTypes');
    return await walletSDKPlugin.initializeWalletInitiatedFlow(issuerURI, credentialTypes!);
  }

  readConnectIssuerConfig() async {
    connectIssuerConfigList = await _configService.readConnectIssuerConfig();
    setState(() {});
  }

  String _requestErrorSubTitleMsg = '';
  String _requestErrorTitleMsg = '';
  bool show = false;

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
          const SizedBox(height: 20),
          Expanded(
            child: ListView.builder(
              padding: const EdgeInsets.fromLTRB(24, 12, 24, 12),
              itemCount: connectIssuerConfigList.length,
              itemBuilder: (context, index) {
                return Column(children: [
                  Container(
                    height: 80,
                    alignment: Alignment.center,
                    decoration: BoxDecoration(
                        color: connectIssuerConfigList.elementAt(index).value.backgroundColor.isNotEmpty
                            ? Color(int.parse(
                                '0xff${connectIssuerConfigList.elementAt(index).value.backgroundColor.replaceAll('#', '')}'))
                            : Colors.white,
                        borderRadius: BorderRadius.circular(12),
                        boxShadow: [BoxShadow(offset: const Offset(3, 3), color: Colors.grey.shade300, blurRadius: 5)]),
                    /*     color: connectIssuerConfigList.elementAt(index).value.backgroundColor.isNotEmpty
                      ? Color(int.parse(
                          '0xff${connectIssuerConfigList.elementAt(index).value.backgroundColor.replaceAll('#', '')}'))
                      : Colors.grey.shade200,*/
                    clipBehavior: Clip.antiAliasWithSaveLayer,
                    child: ListTile(
                        title: Text(
                          connectIssuerConfigList.elementAt(index).key,
                          style: TextStyle(
                            fontWeight: FontWeight.bold,
                            fontSize: 16,
                            color: connectIssuerConfigList.elementAt(index).value.textColor.isNotEmpty
                                ? Color(int.parse(
                                    '0xff${connectIssuerConfigList.first.value.textColor.replaceAll('#', '')}'))
                                : const Color(0xff190C21),
                          ),
                        ),
                        leading: connectIssuerConfigList.elementAt(index).value.logo == null
                            ? const SizedBox.shrink()
                            : CachedNetworkImage(
                                imageUrl: connectIssuerConfigList.elementAt(index).value.logo,
                                placeholder: (context, url) => const CircularProgressIndicator(),
                                errorWidget: (context, url, error) =>
                                    Image.asset('lib/assets/images/logoIcon.png', fit: BoxFit.contain),
                                width: 50,
                                height: 50,
                                fit: BoxFit.contain,
                              ),
                        trailing: IconButton(
                          icon: const Icon(Icons.arrow_circle_right_outlined, size: 24, color: Color(0xffB6B7C7)),
                          onPressed: () async {
                            try {
                              var supportedCredentials =
                                  await connect(connectIssuerConfigList.elementAt(index).value.issuerURI);
                              var connectIssuerConfigValue = ConnectIssuerConfigValue(
                                  issuerURI: '',
                                  scopes: connectIssuerConfigList.elementAt(index).value.scopes,
                                  clientID: connectIssuerConfigList.elementAt(index).value.clientID,
                                  redirectURI: connectIssuerConfigList.elementAt(index).value.redirectURI,
                                  showIssuer: true,
                                  description: '',
                                  backgroundColor: '',
                                  textColor: '',
                                  logo: '');
                              _navigateToSupportedCredentialScreen(
                                  connectIssuerConfigList.elementAt(index).key,
                                  connectIssuerConfigList.elementAt(index).value.issuerURI,
                                  supportedCredentials,
                                  connectIssuerConfigValue);
                            } catch (err) {
                              if (err is PlatformException && err.message != null && err.message!.isNotEmpty) {
                                var resp = await walletSDKPlugin.parseWalletSDKError(
                                    localizedErrorMessage: err.details.toString());
                                setState(() {
                                  _requestErrorSubTitleMsg = resp.details;
                                  _requestErrorTitleMsg = 'Oops! Something went wrong!';
                                  show = true;
                                });
                              }
                            }
                          },
                        )),
                  ),
                  Column(
                    children: <Widget>[
                      Column(
                        children: <Widget>[
                          Visibility(
                            visible: show,
                            child: Container(
                              padding: const EdgeInsets.all(12),
                              alignment: Alignment.center,
                              child: ListTile(
                                shape: RoundedRectangleBorder(
                                  borderRadius: BorderRadius.circular(12),
                                ),
                                tileColor: const Color(0xffFBF8FC),
                                title: SelectableText(
                                  _requestErrorTitleMsg ?? '',
                                  style: const TextStyle(
                                    fontSize: 16,
                                    fontWeight: FontWeight.bold,
                                    color: Color(0xff190C21),
                                  ),
                                  textAlign: TextAlign.start,
                                ),
                                subtitle: SelectableText(
                                  _requestErrorSubTitleMsg ?? '',
                                  style: const TextStyle(
                                    fontSize: 12,
                                    fontWeight: FontWeight.bold,
                                    color: Color(0xff6C6D7C),
                                  ),
                                  textAlign: TextAlign.start,
                                ),
                                leading: const SizedBox(
                                    height: 24,
                                    width: 24,
                                    child: Image(
                                      image: AssetImage('lib/assets/images/errorVector.png'),
                                      width: 24,
                                      height: 24,
                                      fit: BoxFit.cover,
                                    )),
                              ),
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ]);
              },
            ),
          ),
        ],
      ),
    );
  }

  _navigateToSupportedCredentialScreen(String issuerName, issuerURI, List<SupportedCredentials> supportedCredentials,
      ConnectIssuerConfigValue connectIssuerConfigValue) async {
    Navigator.push(
      context,
      MaterialPageRoute(
          builder: (context) => SupportedCredentialsList(
                issuerName: issuerName,
                issuerURI: issuerURI,
                supportedCredentialList: supportedCredentials,
                connectIssuerConfigValue: connectIssuerConfigValue,
              )),
    );
  }
}
