/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

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

  Future<Map<String, dynamic>> readCustomScopeConfig(List<Object?> customScopeList) async {
    Map<String, dynamic> customScopeConfigList = {};
    final String scopeConfigResponse = await rootBundle.loadString('lib/assets/customScopes.json');
    final configDataResp = await json.decode(scopeConfigResponse);

    List<String> customScopeStr = customScopeList.cast<String>();
    for (var customScope in customScopeStr) {
      var claimJSON = configDataResp[customScope];
      customScopeConfigList.addAll(
          {customScope.toString(): jsonEncode(claimJSON)}
      );
    }
    log('customScope config list $customScopeConfigList');
    return customScopeConfigList;
  }
}
