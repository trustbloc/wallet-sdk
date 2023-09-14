import 'dart:convert';
import 'dart:developer';

import 'package:app/models/connect_issuer_config.dart';
import 'package:app/models/connect_issuer_config_value.dart';
import 'package:flutter/services.dart';

class ConfigService {
  Future<List<ConnectIssuerConfig>> readConnectIssuerConfig() async {
    List<ConnectIssuerConfig> connectIssuerConfigList = List.empty(growable: true);
    final String configResponse = await rootBundle.loadString('lib/assets/walletInitiatedIssuerConfig.json');
    final configResponseDecoded = jsonDecode(configResponse);

    configResponseDecoded.forEach((key, value) {
      ConnectIssuerConfig connectIssuerConfig = ConnectIssuerConfig(key, ConnectIssuerConfigValue.fromJson(value));
      connectIssuerConfigList.add(connectIssuerConfig);
    });

    log('decodedResponse $configResponseDecoded');
    return connectIssuerConfigList;
  }
}
