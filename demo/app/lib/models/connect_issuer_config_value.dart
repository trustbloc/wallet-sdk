/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

class ConnectIssuerConfigValue {
  final String issuerURI;
  final List<String> scopes;
  final String clientID;
  final String redirectURI;
  final bool showIssuer;
  final String description;
  final String logo;
  final String backgroundColor;
  final String textColor;

  ConnectIssuerConfigValue(
      {required this.issuerURI,
      required this.scopes,
      required this.clientID,
      required this.redirectURI,
      required this.showIssuer,
      required this.description,
      required this.logo,
      required this.backgroundColor,
      required this.textColor});

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> data = <String, dynamic>{};
    data['issuerURI'] = issuerURI;
    data['scopes'] = scopes;
    data['clientID'] = clientID;
    data['redirectURI'] = redirectURI;
    data['showIssuer'] = showIssuer;
    data['description'] = description;
    data['logo'] = logo;
    data['backgroundColor'] = backgroundColor;
    data['textColor'] = textColor;
    return data;
  }

  factory ConnectIssuerConfigValue.fromJson(Map<String, dynamic> json) {
    return ConnectIssuerConfigValue(
        issuerURI: json['issuerURI'],
        scopes: json['scopes'].cast<String>(),
        clientID: json['clientID'],
        redirectURI: json['redirectURI'],
        showIssuer: json['showIssuer'],
        description: json['description'],
        logo: json['logo'],
        backgroundColor: json['backgroundColor'],
        textColor: json['textColor']);
  }
}
