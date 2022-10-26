import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';

import 'demo_platform_interface.dart';

/// An implementation of [HelloPlatform] that uses method channels.
class MethodChannelHello extends HelloPlatform {
  /// The method channel used to interact with the native platform.
  @visibleForTesting
  final methodChannel = const MethodChannel('HelloPlugin');

  Future<String?> sayHello(String name) async {
    final helloMsg = await methodChannel.invokeMethod<String>('sayHello', {"name":name});
    return helloMsg;
  }
  Future<String?> storeCredentials() async {
    final storeMsg = await methodChannel.invokeMethod<String>('storeCredentials');
    return storeMsg;
  }
}
