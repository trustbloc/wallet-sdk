/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'dart:convert';
import 'dart:io';

import 'package:app/wallet_sdk/wallet_sdk.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_test/flutter_test.dart';

void main() async {
  print('Init Wallet SDK Plugin');
  final walletSDKPlugin = WalletSDK();
  print('Init SDK');
  const didResolverURI = String.fromEnvironment('DID_RESOLVER_URI');
  await walletSDKPlugin.initSDK(didResolverURI);
  var didKeyType = 'ED25519';
  testWidgets('Testing openid4vc with a single credential', (tester) async {
    const didMethodTypes = String.fromEnvironment('WALLET_DID_METHODS');
    var didMethodTypesList = didMethodTypes.split(' ');

    const issuanceURLs = String.fromEnvironment('INITIATE_ISSUANCE_URLS');
    var issuanceURLsList = issuanceURLs.split(' ');

    const verificationURLs = String.fromEnvironment('INITIATE_VERIFICATION_URLS');
    var verificationURLsList = verificationURLs.split(' ');

    for (int i = 0; i < issuanceURLsList.length; i++) {
      String didMethodType = didMethodTypesList[i];
      print('wallet DID Type : $didMethodType');
      print('wallet DID Key Type : $didKeyType');
      var didDocData = await walletSDKPlugin.createDID(didMethodTypesList[i], didKeyType);
      final didContent = didDocData.did;
      print('wallet DID : $didContent');

      String issuanceURL = issuanceURLsList[i];
      print('issuanceURL : $issuanceURL');

      var initializeResp = await walletSDKPlugin.initialize(issuanceURL, null);
      var initializeRespEncoded = json.encode(initializeResp!);
      Map<String, dynamic> initializeRespJson = json.decode(initializeRespEncoded);
      var requirePIN = initializeRespJson['pinRequired'];
      print('requirePIN: $requirePIN');

      final attestationVC = await walletSDKPlugin.getAttestationVC(
          attestationURL: 'https://localhost:8097/profiles/profileID/profileVersion/wallet/attestation/',
          disableTLSVerify: true,
          attestationToken: 'token',
          attestationPayload: '''{
							"type": "urn:attestation:application:midy",
							"application": {
								"type":    "MidyWallet",
								"name":    "Midy Wallet",
								"version": "2.0"
							},
							"compliance": {
								"type": "fcra"				
							}
						}''',);

      List<String> credentials;
      await tester.runAsync(() async {
        credentials = (await walletSDKPlugin.requestCredential('', attestationVC: attestationVC))
            .map((e) => e.content)
            .toList();
        print('credentials');
        expect(credentials, hasLength(greaterThan(0)));

        String verificationURL = verificationURLsList[i];
        print('verificationURL : $verificationURL');

        final l = await walletSDKPlugin.processAuthorizationRequest(authorizationRequest: verificationURL);
        print('processAuthorizationRequest --> : $l');

        // Add another delay if needed
        await Future.delayed(Duration(seconds: 2));

        final requirements = await walletSDKPlugin.getSubmissionRequirements(storedCredentials: credentials);
        print('getSubmissionRequirements finished');

        expect(requirements, hasLength(equals(1)));
        expect(requirements[0].inputDescriptors, hasLength(equals(1)));
        expect(requirements[0].inputDescriptors[0].matchedVCsID, hasLength(equals(1)));

        var customScopesList = {
          'registration': jsonEncode({'email': 'test@example.com'}),
        };
        await walletSDKPlugin.presentCredential(
            selectedCredentials: credentials, customScopeList: customScopesList, attestationVC: attestationVC);
        print('credential presented');
      });
    }
  },
      timeout: Timeout.none );

  testWidgets('Testing openid4vc with multiple credentials', (tester) async {
    const didMethodTypes = String.fromEnvironment('WALLET_DID_METHODS');
    var didMethodTypesList = didMethodTypes.split(' ');
    String didMethodType = didMethodTypesList[0];
    print('wallet DID type : $didMethodType');
    print('wallet DID Key type : $didKeyType');

    await Future.delayed(Duration(seconds: 2));
    var didDocData = await walletSDKPlugin.createDID(didMethodTypesList[0], didKeyType);
    var didContent = didDocData.did;
    print('wallet DID : $didContent');

    const issuanceURLs = String.fromEnvironment('INITIATE_ISSUANCE_URLS_MULTIPLE_CREDS');
    var issuanceURLsList = issuanceURLs.split(' ');

    List<String> credentials = [];

    // Issue multiple credentials
    for (int i = 0; i < issuanceURLsList.length; i++) {
      String issuanceURL = issuanceURLsList[i];
      print('issuanceURL : $issuanceURL');

      var initializeResp = await walletSDKPlugin.initialize(issuanceURL, null);
      var initializeRespEncoded = json.encode(initializeResp!);
      Map<String, dynamic> initializeRespJson = json.decode(initializeRespEncoded);
      var requirePIN = initializeRespJson['pinRequired'];
      print('requirePIN: $requirePIN');

      final requestdCreds = await walletSDKPlugin.requestCredential('');

      expect(requestdCreds, hasLength(greaterThan(0)));

      credentials.addAll(requestdCreds.map((e) => e.content).toList());
    }
    print('issued credentials');

    const verificationURLs = String.fromEnvironment('INITIATE_VERIFICATION_URLS_MULTIPLE_CREDS');
    var verificationURLsList = verificationURLs.split(' ');
    String verificationURL = verificationURLsList[0];
    print('verificationURL : $verificationURL');

    final matchedCreds = await walletSDKPlugin.processAuthorizationRequest(
        authorizationRequest: verificationURL, storedCredentials: credentials);

    print('matchedCreds : $matchedCreds');

    expect(matchedCreds, hasLength(equals(3)));

    expect(matchedCreds[0], equals(credentials[0]));
    expect(matchedCreds[1], equals(credentials[1]));
    expect(matchedCreds[2], equals(credentials[2]));
    var customScopesList = {
      'registration': jsonEncode({'email': 'test@example.com'}),
      'testscope': jsonEncode({'data': 'testdata'}),
    };

    // Add a delay before presenting credentials
    await Future.delayed(Duration(seconds: 2));
    await walletSDKPlugin.presentCredential(selectedCredentials: matchedCreds, customScopeList: customScopesList);
  },       timeout: Timeout.none );

  testWidgets('Testing openid4vc with the auth code flow', (tester) async {
    const didMethodTypes = String.fromEnvironment('WALLET_DID_METHODS');
    var didMethodTypesList = didMethodTypes.split(' ');
    String didMethodType = didMethodTypesList[0];
    print('wallet DID type : $didMethodType');
    print('wallet DID Key type : $didKeyType');
    var didDocData = await walletSDKPlugin.createDID(didMethodTypesList[0], didKeyType);
    print('wallet didDocData : $didDocData');
    var didContent = didDocData.did;
    print('wallet DID : $didContent');

    const issuanceURL = String.fromEnvironment('INITIATE_ISSUANCE_URLS_AUTH_CODE_FLOW');
    debugPrint('issuanceURLs Auth Code Flow: $issuanceURL');

    var authCodeArgs = {
      'scopes': ['openid', 'profile'],
      'clientID': 'oidc4vc_client',
      'redirectURI': 'http://127.0.0.1/callback'
    };

    var initializeResp = await walletSDKPlugin.initialize(issuanceURL, authCodeArgs);
    var initializeRespEncoded = json.encode(initializeResp!);
    Map<String, dynamic> initializeRespJson = json.decode(initializeRespEncoded);
    var authorizationURLLink = initializeRespJson['authorizationURLLink'];
    debugPrint('authorizationURLLink: $authorizationURLLink');
    // fetching redirect uri
    String redirectURI = '';
    final client = HttpClient();
    var redirectUrl = Uri.parse(authorizationURLLink);
    var request = await client.getUrl(redirectUrl);
    request.followRedirects = false;
    var response = await request.close();
    while (response.isRedirect) {
      response.drain();
      final location = response.headers.value(HttpHeaders.locationHeader);
      if (location != null) {
        redirectUrl = redirectUrl.resolve(location);
        if (location.contains('http://127.0.0.1/callback')) {
          redirectURI = location;
          break;
        }
        if (redirectUrl.host.contains('cognito-mock.trustbloc.local')) {
          redirectUrl = Uri.parse(redirectUrl.toString().replaceAll('cognito-mock.trustbloc.local', 'localhost'));
          print('uri updated $redirectUrl');
        }
        request = await client.getUrl(redirectUrl);
        request.followRedirects = false;
        response = await request.close();
      }
    }

    debugPrint('redirectURI $redirectURI');

    final credential = await walletSDKPlugin.requestCredentialWithAuth(redirectURI);
    debugPrint('content: $credential');
    for (final p in credential.split('.')) {
      print('----');
      print(p);
    }

    expect(credential, hasLength(greaterThan(0)));
  },       timeout: Timeout.none );
}
