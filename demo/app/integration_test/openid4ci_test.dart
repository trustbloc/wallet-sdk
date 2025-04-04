/*
Copyright Gen Digital Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

import 'dart:convert';
import 'dart:io';

import 'package:app/wallet_sdk/wallet_sdk.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  final walletSDKPlugin = WalletSDK();
  const didResolverURI = String.fromEnvironment('DID_RESOLVER_URI');
  const didKeyType = 'ED25519';

  // Shared setup for all tests
  setUpAll(() async {
    print('Initializing Wallet SDK');
    await walletSDKPlugin.initSDK(didResolverURI);
  });

  group('Single Credential Flow', () {
    const didMethodTypes = String.fromEnvironment('WALLET_DID_METHODS');
    final didMethodTypesList = didMethodTypes.split(' ');
    const issuanceURLs = String.fromEnvironment('INITIATE_ISSUANCE_URLS');
    final issuanceURLsList = issuanceURLs.split(' ');
    const verificationURLs = String.fromEnvironment('INITIATE_VERIFICATION_URLS');
    final verificationURLsList = verificationURLs.split(' ');

    testWidgets('Test with each DID method', (tester) async {
      for (int i = 0; i < issuanceURLsList.length; i++) {
        final didMethodType = didMethodTypesList[i];
        print('Testing with DID Type: $didMethodType');

        final didDocData = await walletSDKPlugin.createDID(didMethodType, didKeyType);
        print('Created DID: ${didDocData.did}');

        await _testSingleCredentialFlow(
          walletSDKPlugin,
          tester,
          issuanceURLsList[i],
          verificationURLsList[i],
        );
      }
    }, timeout: Timeout(Duration(minutes: 3)));
  });

  group('Multiple Credentials Flow', () {
    testWidgets('Test multiple credentials issuance and verification', (tester) async {
      const didMethodTypes = String.fromEnvironment('WALLET_DID_METHODS');
      final didMethodType = didMethodTypes.split(' ')[0];

      final didDocData = await walletSDKPlugin.createDID(didMethodType, didKeyType);
      print('Created DID: ${didDocData.did}');

      const issuanceURLs = String.fromEnvironment('INITIATE_ISSUANCE_URLS_MULTIPLE_CREDS');
      const verificationURLs = String.fromEnvironment('INITIATE_VERIFICATION_URLS_MULTIPLE_CREDS');

      await _testMultipleCredentialsFlow(
        walletSDKPlugin,
        tester,
        issuanceURLs.split(' '),
        verificationURLs.split(' ')[0],
      );
    }, timeout: Timeout(Duration(minutes: 5)));
  });

  group('Auth Code Flow', () {
    testWidgets('Test authorization code flow', (tester) async {
      const didMethodTypes = String.fromEnvironment('WALLET_DID_METHODS');
      final didMethodType = didMethodTypes.split(' ')[0];

      final didDocData = await walletSDKPlugin.createDID(didMethodType, didKeyType);
      print('Created DID: ${didDocData.did}');

      const issuanceURL = String.fromEnvironment('INITIATE_ISSUANCE_URLS_AUTH_CODE_FLOW');

      await _testAuthCodeFlow(walletSDKPlugin, tester, issuanceURL);
    }, timeout: Timeout(Duration(minutes: 5)));
  });
}

Future<void> _testSingleCredentialFlow(
    WalletSDK walletSDKPlugin,
    WidgetTester tester,
    String issuanceURL,
    String verificationURL,
    ) async {
  print('Testing issuance URL: $issuanceURL');

  final initializeResp = await walletSDKPlugin.initialize(issuanceURL, null);
  final requirePIN = json.decode(json.encode(initializeResp!))['pinRequired'];
  print('PIN required: $requirePIN');

  final attestationVC = await walletSDKPlugin.getAttestationVC(
    attestationURL: 'https://localhost:8097/profiles/profileID/profileVersion/wallet/attestation/',
    disableTLSVerify: true,
    attestationToken: 'token',
    attestationPayload: json.encode({
      "type": "urn:attestation:application:midy",
      "application": {
        "type": "MidyWallet",
        "name": "Midy Wallet",
        "version": "2.0"
      },
      "compliance": {
        "type": "fcra"
      }
    }),
  );

  await tester.runAsync(() async {
    final credentials = (await walletSDKPlugin.requestCredential('', attestationVC: attestationVC))
        .map((e) => e.content)
        .toList();

    expect(credentials, isNotEmpty);
    print('Verification URL: $verificationURL');

    await walletSDKPlugin.processAuthorizationRequest(
      authorizationRequest: verificationURL,
    );

    final requirements = await walletSDKPlugin.getSubmissionRequirements(
      storedCredentials: credentials,
    );

    expect(requirements, hasLength(1));
    expect(requirements[0].inputDescriptors, hasLength(1));
    expect(requirements[0].inputDescriptors[0].matchedVCsID, hasLength(1));

    await walletSDKPlugin.presentCredential(
      selectedCredentials: credentials,
      customScopeList: {
        'registration': jsonEncode({'email': 'test@example.com'}),
      },
      attestationVC: attestationVC,
    );
  });
}

Future<void> _testMultipleCredentialsFlow(
    WalletSDK walletSDKPlugin,
    WidgetTester tester,
    List<String> issuanceURLs,
    String verificationURL,
    ) async {
  final credentials = <String>[];

  for (final issuanceURL in issuanceURLs) {
    print('Issuing credential from: $issuanceURL');

    final initializeResp = await walletSDKPlugin.initialize(issuanceURL, null);
    final requirePIN = json.decode(json.encode(initializeResp!))['pinRequired'];
    print('PIN required: $requirePIN');

    final requestedCreds = await walletSDKPlugin.requestCredential('');
    expect(requestedCreds, isNotEmpty);
    credentials.addAll(requestedCreds.map((e) => e.content).toList());
  }

  print('Processing verification with URL: $verificationURL');
  final matchedCreds = await walletSDKPlugin.processAuthorizationRequest(
    authorizationRequest: verificationURL,
    storedCredentials: credentials,
  );

  expect(matchedCreds, hasLength(3));
  expect(matchedCreds, [credentials[0], credentials[1], credentials[2]]);

  await walletSDKPlugin.presentCredential(
    selectedCredentials: matchedCreds,
    customScopeList: {
      'registration': jsonEncode({'email': 'test@example.com'}),
      'testscope': jsonEncode({'data': 'testdata'}),
    },
  );
}

Future<void> _testAuthCodeFlow(
    WalletSDK walletSDKPlugin,
    WidgetTester tester,
    String issuanceURL,
    ) async {
  print('Testing auth code flow with URL: $issuanceURL');

  final authCodeArgs = {
    'scopes': ['openid', 'profile'],
    'clientID': 'oidc4vc_client',
    'redirectURI': 'http://127.0.0.1/callback'
  };

  final initializeResp = await walletSDKPlugin.initialize(issuanceURL, authCodeArgs);
  final authorizationURLLink = json.decode(json.encode(initializeResp!))['authorizationURLLink'];
  print('Authorization URL: $authorizationURLLink');

  final redirectURI = await _followRedirects(authorizationURLLink);
  print('Final redirect URI: $redirectURI');

  final credential = await walletSDKPlugin.requestCredentialWithAuth(redirectURI);
  expect(credential, isNotEmpty);
}

Future<String> _followRedirects(String initialUrl) async {
  var redirectUrl = Uri.parse(initialUrl);
  final client = HttpClient();
  String redirectURI = '';

  try {
    var request = await client.getUrl(redirectUrl);
    request.followRedirects = false;
    var response = await request.close();

    while (response.isRedirect) {
      response.drain();
      final location = response.headers.value(HttpHeaders.locationHeader);

      if (location == null) break;

      redirectUrl = redirectUrl.resolve(location);

      if (location.contains('http://127.0.0.1/callback')) {
        redirectURI = location;
        break;
      }

      if (redirectUrl.host.contains('cognito-mock.trustbloc.local')) {
        redirectUrl = Uri.parse(redirectUrl.toString().replaceAll('cognito-mock.trustbloc.local', 'localhost'));
        print('Updated URI: $redirectUrl');
      }

      request = await client.getUrl(redirectUrl);
      request.followRedirects = false;
      response = await request.close();
    }
  } finally {
    client.close();
  }

  return redirectURI;
}