import 'package:app/demo_method_channel.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';

void main() {
  testWidgets('openid4ci', (tester) async {
    final walletSDKPlugin = MethodChannelWallet();

    await walletSDKPlugin.initSDK();

    final didContent = await walletSDKPlugin.createDID();
    print("didContent : $didContent");

    const issuanceURL = String.fromEnvironment("INITIATE_ISSUANCE_URL");
    print("issuanceURL $issuanceURL");

    bool? requirePIN = await walletSDKPlugin.authorize(issuanceURL);

    print("requirePIN: $requirePIN");

    final credential = await walletSDKPlugin.requestCredential("");
    print("credential content: $credential");

    expect(credential, hasLength(greaterThan(0)));
  });
}
