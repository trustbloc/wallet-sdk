import 'dart:io';

import 'package:app/demo_method_channel.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';

void main() async {
  final walletSDKPlugin = MethodChannelWallet();
  print("Init SDK");

  await walletSDKPlugin.initSDK();

  testWidgets('Testing openid4vc with a single credential', (tester) async {
    const didMethodTypes = String.fromEnvironment("WALLET_DID_METHODS");
    var didMethodTypesList = didMethodTypes.split(' ');

    const issuanceURLs = String.fromEnvironment("INITIATE_ISSUANCE_URLS");
    var issuanceURLsList = issuanceURLs.split(' ');

    const verificationURLs = String.fromEnvironment("INITIATE_VERIFICATION_URLS");
    var verificationURLsList = verificationURLs.split(' ');

    for (int i = 0; i < issuanceURLsList.length; i++) {
      String didMethodType = didMethodTypesList[i];
      print("wallet DID Type : $didMethodType");
      final didContent = await walletSDKPlugin.createDID(didMethodTypesList[i]);
      print("wallet DID : $didContent");

      String issuanceURL = issuanceURLsList[i];
      print("issuanceURL : $issuanceURL");

      bool? requirePIN = await walletSDKPlugin.authorize(issuanceURL);
      print("requirePIN: $requirePIN");

      final credential = await walletSDKPlugin.requestCredential("");
      debugPrint("content: $credential");
      for (final p in credential.split('.')) {
        print("----");
        print(p);
      }

      expect(credential, hasLength(greaterThan(0)));

      String verificationURL = verificationURLsList[i];
      print("verificationURL : $verificationURL");

      await walletSDKPlugin
          .processAuthorizationRequest(authorizationRequest: verificationURL);

      print("getSubmissionRequirements");

      final requirements = await walletSDKPlugin.getSubmissionRequirements( storedCredentials: [credential]);

      print("getSubmissionRequirements finished");

      expect(requirements, hasLength(equals(1)));
      expect(requirements[0].inputDescriptors, hasLength(equals(1)));
      expect(requirements[0].inputDescriptors[0].matchedVCsID, hasLength(equals(1)));

      await walletSDKPlugin.presentCredential(selectedCredentials: [credential]);
    }
  });

  testWidgets('Testing openid4vc with multiple credentials', (tester) async {
    const didMethodTypes = String.fromEnvironment("WALLET_DID_METHODS");
    var didMethodTypesList = didMethodTypes.split(' ');
    String didMethodType = didMethodTypesList[0];
    print("wallet DID type : $didMethodType");
    final didContent = await walletSDKPlugin.createDID(didMethodTypesList[0]);
    print("wallet DID : $didContent");

    const issuanceURLs = String.fromEnvironment("INITIATE_ISSUANCE_URLS_MULTIPLE_CREDS");
    var issuanceURLsList = issuanceURLs.split(' ');

    List<String> credentials = [];

    // Issue multiple credentials
    for (int i = 0; i < issuanceURLsList.length; i++) {
      String issuanceURL = issuanceURLsList[i];
      print("issuanceURL : $issuanceURL");
      bool? requirePIN = await walletSDKPlugin.authorize(issuanceURL);

      print("requirePIN: $requirePIN");

      final credential = await walletSDKPlugin.requestCredential("");

      expect(credential, hasLength(greaterThan(0)));

      credentials.add(credential);
    }
    print("issued credentials: $credentials");

    const verificationURLs = String.fromEnvironment("INITIATE_VERIFICATION_URLS_MULTIPLE_CREDS");
    var verificationURLsList = verificationURLs.split(' ');
    String verificationURL = verificationURLsList[0];
    print("verificationURL : $verificationURL");

    final matchedCreds = await walletSDKPlugin
        .processAuthorizationRequest(authorizationRequest: verificationURL, storedCredentials: credentials);

    print("matchedCreds : $matchedCreds");

    expect(matchedCreds, hasLength(equals(3)));
    expect(matchedCreds[0], equals(credentials[0]));
    expect(matchedCreds[1], equals(credentials[1]));
    expect(matchedCreds[2], equals(credentials[2]));

    await walletSDKPlugin.presentCredential();
  });
}
