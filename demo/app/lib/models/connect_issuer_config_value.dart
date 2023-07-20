import 'dart:ffi';

class ConnectIssuerConfigValue {
  final String issuerURI;
  final List<String> scopes;
  final String clientID;
  final bool showIssuer;
  final String description;
  final String redirectURI;

  ConnectIssuerConfigValue(
      {
        required this.issuerURI,
        required this.scopes,
        required this.clientID,
        required this.showIssuer,
        required this.description,
        required this.redirectURI
      });

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> data = <String, dynamic>{};
    data['issuerURI'] = issuerURI;
    data['scopes'] = scopes;
    data['clientID'] = clientID;
    data['showIssuer'] = showIssuer;
    data['description'] = description;
    return data;
  }

  factory ConnectIssuerConfigValue.fromJson(Map<String, dynamic> json) {
    return ConnectIssuerConfigValue(
        issuerURI: json['issuerURI'],
        scopes: json['scopes'].cast<String>(),
        clientID: json['clientID'],
        showIssuer: json['showIssuer'],
        description: json['description'],
        redirectURI: const String.fromEnvironment("redirectURI")
    );
  }
}