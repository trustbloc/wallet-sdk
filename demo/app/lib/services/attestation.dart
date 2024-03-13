import '../wallet_sdk/wallet_sdk.dart';

class AttestationService {
  static String get attestationURL => const String.fromEnvironment('attestationURL');

  static bool get attestationEnabled => attestationURL.isNotEmpty;

  static Future<String> getAttestationVC() async {
    _attestationVC ??=
        await WalletSDK().getAttestationVC(attestationURL: attestationURL, authenticationMethod: 'system_biometry');

    return _attestationVC!;
  }

  static Future<String?> returnAttestationVCIfEnabled() async {
    if (attestationEnabled) {
      return await getAttestationVC();
    }

    return null;
  }

  static String? _attestationVC;
}
