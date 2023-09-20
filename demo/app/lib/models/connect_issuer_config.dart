/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'package:app/models/connect_issuer_config_value.dart';

class ConnectIssuerConfig {
  ConnectIssuerConfig(this.key, this.value);

  final String key;
  final ConnectIssuerConfigValue value;
}
