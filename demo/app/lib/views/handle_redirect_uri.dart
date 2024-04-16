/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'dart:async';

import 'package:flutter/material.dart';
import 'dart:developer';
import 'package:app/wallet_sdk/wallet_sdk_mobile.dart';
import 'package:app/services/storage_service.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:app/models/activity_data_object.dart';
import 'package:app/models/credential_data.dart';
import 'credential_preview.dart';
import 'package:flutter/foundation.dart';
import 'package:uni_links/uni_links.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:app/views/custom_error.dart';

class HandleRedirectUri extends StatefulWidget {
  Uri uri;
  String flowType;
  String? issuerURI;

  HandleRedirectUri(this.uri, this.flowType, this.issuerURI, {super.key});

  @override
  State<HandleRedirectUri> createState() => HandleRedirectUriState();
}

class HandleRedirectUriState extends State<HandleRedirectUri> {
  Uri? _redirectUri;
  Object? _err;

  var WalletSDKPlugin = WalletSDK();
  final StorageService _storageService = StorageService();
  final Future<SharedPreferences> prefs = SharedPreferences.getInstance();
  var userDIDId = '';
  var userDIDDoc = '';

  StreamSubscription? _sub;
  Future<Widget>? credentialPreview;

  @override
  void initState() {
    super.initState();
    _launchUrl(widget.uri);
  }

  @override
  void dispose() {
    _sub?.cancel();
    super.dispose();
  }

  Future<String?> _createDid() async {
    final SharedPreferences pref = await prefs;
    var didType = pref.getString('didType');
    didType = didType ?? 'jwk';
    var keyType = pref.getString('keyType');
    keyType = keyType ?? 'ECDSAP384IEEEP1363';
    var didResolution = await WalletSDKPlugin.createDID(didType, keyType);
    var didID = didResolution.did;
    setState(() {
      userDIDId = didID;
      userDIDDoc = didResolution.didDoc;
    });
    return didID;
  }

  Future<Widget>? launchCredPreview() async {
    final SharedPreferences pref = await prefs;
    await _createDid();
    pref.setString('userDID', userDIDId);
    pref.setString('userDIDDoc', userDIDDoc);
    if (_redirectUri != null) {
      try {
        if (widget.flowType == 'issuer-initiated-flow') {
          var issuerURI = await WalletSDKPlugin.issuerURI();
          var credentials = await WalletSDKPlugin.requestCredentialWithAuth(_redirectUri.toString());
          var serializedDisplayData = await WalletSDKPlugin.resolveDisplayData([credentials], issuerURI!);
          log('serializedDisplayData -> $serializedDisplayData');
          var activities = await WalletSDKPlugin.storeActivityLogger();
          var credID = await WalletSDKPlugin.getCredID([credentials]);
          await _storageService.addActivities(ActivityDataObj(credID!, activities));
          pref.setString('credID', credID);
          setState(() {});
          _navigateToCredPreviewScreen([
            CredentialData(
                rawCredential: credentials,
                issuerURL: issuerURI,
                issuerDisplayData: serializedDisplayData.issuerDisplay,
                credentialDisplayData: serializedDisplayData.credentialsDisplay[0],
                credentialDID: userDIDId,
                credID: credID),
          ]);
        } else {
          log('_redirectUri.toString() ${_redirectUri.toString()}');
          var credentials = await WalletSDKPlugin.requestCredentialWithWalletInitiatedFlow(_redirectUri.toString());
          var issuerURI = widget.issuerURI;
          var serializedDisplayData = await WalletSDKPlugin.resolveDisplayData([credentials], issuerURI!);
          log('serializedDisplayData -> $serializedDisplayData');
          // TODO: Issue-518 Add activity logger support for wallet-initiated-flow
          _navigateToCredPreviewScreen([
            CredentialData(
                rawCredential: credentials,
                issuerURL: issuerURI,
                issuerDisplayData: serializedDisplayData.issuerDisplay,
                credentialDisplayData: serializedDisplayData.credentialsDisplay[0],
                credentialDID: userDIDId,
                credID: ''),
          ]);
        }
      } catch (error) {
        Navigator.push(
            context,
            MaterialPageRoute(
                builder: (context) => CustomError(
                    titleBar: 'Handle redirect URI',
                    requestErrorTitleMsg: 'Redirect uri error',
                    requestErrorSubTitleMsg: error.toString())));
      }
    }
    return Container();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
        body: FutureBuilder<Widget>(
      future: credentialPreview,
      builder: (BuildContext context, AsyncSnapshot snapshot) {
        return SizedBox(
          height: MediaQuery.of(context).size.height / 1.3,
          child: const Center(
            child: CircularProgressIndicator(),
          ),
        );
      },
    ));
  }

  _launchUrl(Uri uri) async {
    if (!await launchUrl(uri, mode: LaunchMode.externalApplication)) {
      throw 'Failed to launch $uri';
    } else {
      _handleIncomingLinks();
    }
  }

  _handleIncomingLinks() async {
    if (!kIsWeb) {
      _sub = uriLinkStream.listen(
        (Uri? uri) {
          if (!mounted) return;
          credentialPreview = launchCredPreview();
          _sub?.cancel();
          setState(() {
            _redirectUri = uri;
            _err = null;
          });
        },
        onError: (Object err) {
          if (!mounted) return;
          setState(() {
            _redirectUri = null;
            if (err is FormatException) {
              _err = err;
            } else {
              _err = null;
            }
          });
        },
      );
    }
  }

  _navigateToCredPreviewScreen(List<CredentialData> credentialsData) async {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      Navigator.push(
          context, MaterialPageRoute(builder: (context) => CredentialPreview(credentialsData: credentialsData)));
    });
  }
}
