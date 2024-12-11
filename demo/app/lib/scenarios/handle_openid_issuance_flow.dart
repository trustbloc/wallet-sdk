/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'dart:convert';
import 'dart:developer';
import 'package:app/views/issuance_preview.dart';
import 'package:flutter/material.dart';
import 'package:app/wallet_sdk/wallet_sdk.dart';
import 'package:flutter/services.dart';
import 'package:app/models/credential_offer.dart';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';

import 'package:app/views/custom_error.dart';

void handleOpenIDIssuanceFlow(BuildContext context, String qrCodeURL) async {
  var WalletSDKPlugin = WalletSDK();
  var authCodeArgs;
  if (qrCodeURL.contains('credential_offer_uri')) {
    authCodeArgs = await parseCredentialOfferUri(qrCodeURL);
    log('credential offer uri auth code  $authCodeArgs');
  } else if (qrCodeURL.contains('authorization_code')) {
    authCodeArgs = await readIssuerAuthFlowConfig(qrCodeURL);
    log('auth code arguments fetched from config file $authCodeArgs');
    // While fetching auth code args based on issuer key from file, if no key-value pair is found then set the
    // arguments to default scope and redirect url.
    authCodeArgs ??= {
      'scopes': ['openid', 'profile'],
      'redirectURI': 'trustbloc-wallet://openid4vci/authcodeflow/callback'
    };
  } else {
    // Fetching and persisting credential type from credential offer query
    await getCredentialType(qrCodeURL);
  }

  log('qr code url -  $qrCodeURL');

  Map<Object?, Object?>? flowTypeData;
  try {
    flowTypeData = await WalletSDKPlugin.initialize(qrCodeURL, authCodeArgs);
  } catch (error) {
    var errString = error.toString().replaceAll(r'\', '');
    Navigator.push(
        context,
        MaterialPageRoute(
            builder: (context) => CustomError(
                titleBar: 'QR Code Scanned',
                requestErrorTitleMsg: 'error while intializing the interaction',
                requestErrorSubTitleMsg: errString)));
    return;
  }
  var flowTypeDataEncoded = json.encode(flowTypeData);
  Map<String, dynamic> responseJson = json.decode(flowTypeDataEncoded);
  var authorizeResultPinRequired = responseJson['pinRequired'];
  log('pin required flow -  $authorizeResultPinRequired');
  if (authorizeResultPinRequired == true) {
    navigateToIssuancePreviewScreen(context, authorizeResultPinRequired);
    return;
  } else if (responseJson['authorizationURLLink'] != '') {
    // initiate authCode Flow
    log("initiating authCode Flow- ${responseJson["authorizationURLLink"]}");
    Uri uri = Uri.parse(responseJson['authorizationURLLink']);
    navigateToIssuancePreviewScreenAuthFlow(context, uri);
    return;
  } else {
    navigateToIssuancePreviewScreen(context, authorizeResultPinRequired);
    return;
  }
}

readIssuerAuthFlowConfig(String qrCodeURL) async {
  var decodedUri = Uri.decodeComponent(qrCodeURL);
  final uri = Uri.parse(decodedUri);
  var credentialsQuery = json.decode(uri.queryParameters['credential_offer']!);
  await persistCredentialType(credentialsQuery);
  final String response =
      await rootBundle.loadString('lib/assets/issuerAuthFlowConfig.json');
  final configData = await json.decode(response);
  return configData[credentialsQuery['credential_issuer']];
}

getCredentialType(String qrCodeURL) async {
  var decodedUri = Uri.decodeComponent(qrCodeURL);
  final uri = Uri.parse(decodedUri);
  var credentialsQuery = json.decode(uri.queryParameters['credential_offer']!);
  await persistCredentialType(credentialsQuery);
}

persistCredentialType(credentialsQuery) async {
  final SharedPreferences prefs = await SharedPreferences.getInstance();
  var types = credentialsQuery['credential_configuration_ids'];
  prefs.setStringList('credentialTypes', List<String>.from(types));
}

parseCredentialOfferUri(String qrCodeURL) async {
  var decodedUri = Uri.decodeComponent(qrCodeURL);
  final uri = Uri.parse(decodedUri);
  final response =
      await http.get(Uri.parse(uri.queryParameters['credential_offer_uri']!));
  if (response.statusCode == 200) {
    final String configResp =
        await rootBundle.loadString('lib/assets/issuerAuthFlowConfig.json');
    final configData = await json.decode(configResp);
    var resp = CredentialOfferObject.fromJson(jsonDecode(response.body));
    await persistCredentialType(json.decode(response.body.toString()));
    return configData[resp.credentialIssuer];
  } else {
    throw Exception('Failed to load credential offer uri');
  }
}

void navigateToIssuancePreviewScreen(
    BuildContext context, bool? authorizeResultPinRequired) async {
  Navigator.push(
      context,
      MaterialPageRoute(
          builder: (context) => IssuancePreview(
              authorizeResultPinRequired: authorizeResultPinRequired)));
}

void navigateToIssuancePreviewScreenAuthFlow(
    BuildContext context, Uri uri) async {
  Navigator.push(context,
      MaterialPageRoute(builder: (context) => IssuancePreview(uri: uri)));
}
