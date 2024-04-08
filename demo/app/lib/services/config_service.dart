/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'dart:convert';
import 'dart:developer';
import 'dart:io';

import 'package:app/models/connect_issuer_config.dart';
import 'package:app/models/connect_issuer_config_value.dart';
import 'package:flutter/services.dart';

import '../models/config.dart';

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
    final configDataResp = jsonDecode(scopeConfigResponse);

    List<String> customScopeStr = customScopeList.cast<String>();
    for (var customScope in customScopeStr) {
      var claimJSON = configDataResp[customScope];
      customScopeConfigList.addAll({customScope.toString(): jsonEncode(claimJSON)});
    }
    return customScopeConfigList;
  }

  Future<Config> readConfig() async {
    final String configContent = await rootBundle.loadString('lib/assets/config.json');
    final Map<String, Object> configValues = jsonDecode(configContent).cast<String, Object>();

    String? overrideContent;

    try {
      overrideContent = await rootBundle.loadString('lib/assets/config.override.json');
    } catch (_) {}

    final overrideValues = jsonDecode(overrideContent ?? '{}').cast<String, Object>();

    var didResolverURI = configValues['didResolverURI'] as String;
    if (overrideValues['didResolverURI'] != null) {
      didResolverURI = overrideValues['didResolverURI'] as String;
    }

    var evaluateIssuanceURL = configValues['evaluateIssuanceURL'] as String;
    if (overrideValues['evaluateIssuanceURL'] != null) {
      evaluateIssuanceURL = overrideValues['evaluateIssuanceURL'] as String;
    }

    var evaluatePresentationURL = configValues['evaluatePresentationURL'] as String;
    if (overrideValues['evaluatePresentationURL'] != null) {
      evaluatePresentationURL = overrideValues['evaluatePresentationURL'] as String;
    }

    var attestationURL = configValues['attestationURL'] as String;
    if (overrideValues['attestationURL'] != null) {
      attestationURL = overrideValues['attestationURL'] as String;
    }

    var attestationPayload = jsonEncode(configValues['attestationPayload']);
    if (overrideValues['attestationPayload'] != null) {
      attestationPayload = jsonEncode(overrideValues['attestationPayload']);
    }

    return Config(
      didResolverURI: didResolverURI,
      evaluateIssuanceURL: evaluateIssuanceURL,
      evaluatePresentationURL: evaluatePresentationURL,
      attestationURL: attestationURL,
      attestationPayload: attestationPayload,
    );
  }

  static Future<void> init() async {
    _config = await ConfigService().readConfig();
  }

  static Config get config => _config;

  static late final Config _config;
}
