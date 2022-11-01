import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';

import 'demo_platform_interface.dart';

class MethodChannelWallet extends WalletPlatform {
  @visibleForTesting
  final methodChannel = const MethodChannel('WalletSDKPlugin');

  Future<String?> createDID() async {
    final createDIDMsg = await methodChannel.invokeMethod<String>('createDID');
    return createDIDMsg;
  }
}
