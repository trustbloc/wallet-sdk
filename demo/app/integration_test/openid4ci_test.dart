import 'package:app/demo_method_channel.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';

void main() async {
  final walletSDKPlugin = MethodChannelWallet();
  print("Init SDK");

  await walletSDKPlugin.initSDK();

  testWidgets('openid4ci-vp', (tester) async {
    final didContent = await walletSDKPlugin.createDID();
    print("didContent : $didContent");

    const issuanceURL = String.fromEnvironment("INITIATE_ISSUANCE_URL");
    print("issuanceURL $issuanceURL");

    bool? requirePIN = await walletSDKPlugin.authorize(issuanceURL);

    print("requirePIN: $requirePIN");

    final credential = await walletSDKPlugin.requestCredential("");
    debugPrint("content: $credential");
    for (final p in credential.split('.')) {
      print("----");
      print(p);
    }

    expect(credential, hasLength(greaterThan(0)));

    const verificationURL = String.fromEnvironment("INITIATE_VERIFICATION_URL");
    print("verificationURL $issuanceURL");

    final matchedCreds = await walletSDKPlugin
        .processAuthorizationRequest(authorizationRequest: verificationURL, storedCredentials: [credential]);

    expect(matchedCreds, hasLength(equals(1)));
    expect(matchedCreds[0], equals(credential));

    await walletSDKPlugin.presentCredential();
  });
}
