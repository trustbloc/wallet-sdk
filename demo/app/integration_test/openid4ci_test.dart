import 'package:app/demo_method_channel.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';

void main() async {
  final walletSDKPlugin = MethodChannelWallet();
  print("Init SDK");

  await walletSDKPlugin.initSDK();

  testWidgets('openid4ci-vp', (tester) async {
    const didMethodTypes = String.fromEnvironment("WALLET_DID_METHODS");
    print("didMethodTypes $didMethodTypes");

    var didMethodTypesList = didMethodTypes.split(' ');

    const issuanceURLs = String.fromEnvironment("INITIATE_ISSUANCE_URLS");
    print("issuanceURLs : $issuanceURLs");

    var issuanceURLsList = issuanceURLs.split(' ');

    const verificationURLs = String.fromEnvironment("INITIATE_VERIFICATION_URLS");
    print("verificationURLs $verificationURLs");

    var verificationURLsList = verificationURLs.split(' ');

    for (int i = 0; i < issuanceURLsList.length; i++) {
      String didMethodType = didMethodTypesList[i];
      print("didMethodType : $didMethodType");
      final didContent = await walletSDKPlugin.createDID(didMethodTypesList[i]);
      print("didContent : $didContent");

      String issuanceURL = issuanceURLsList[i];
      print("issuanceURL : $issuanceURL");
      bool? requirePIN = await walletSDKPlugin.authorize(issuanceURLsList[i]);

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

      final matchedCreds = await walletSDKPlugin
          .processAuthorizationRequest(authorizationRequest: verificationURLsList[i], storedCredentials: [credential]);

      expect(matchedCreds, hasLength(equals(1)));
      expect(matchedCreds[0], equals(credential));

      await walletSDKPlugin.presentCredential();
    }
  });
}
