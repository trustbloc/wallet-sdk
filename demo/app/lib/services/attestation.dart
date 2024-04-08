import '../wallet_sdk/wallet_sdk.dart';
import 'config_service.dart';

class AttestationService {
  static String get attestationURL => ConfigService.config.attestationURL;

  static bool get attestationEnabled => attestationURL.isNotEmpty;

  static Future<String> getAttestationVC() async {
    _attestationVC ??=
        await WalletSDK().getAttestationVC(attestationURL: attestationURL, attestationPayload: ConfigService.config.attestationPayload);

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
