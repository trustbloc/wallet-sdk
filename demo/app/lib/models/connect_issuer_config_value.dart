class ConnectIssuerConfigValue {
  final String issuerURI;
  final List<String> scopes;
  final String clientID;
  final String redirectURI;
  final bool showIssuer;
  final String description;

  ConnectIssuerConfigValue(
      {
        required this.issuerURI,
        required this.scopes,
        required this.clientID,
        required this.redirectURI,
        required this.showIssuer,
        required this.description,
      });

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> data = <String, dynamic>{};
    data['issuerURI'] = issuerURI;
    data['scopes'] = scopes;
    data['clientID'] = clientID;
    data['redirectURI'] = redirectURI;
    data['showIssuer'] = showIssuer;
    data['description'] = description;
    return data;
  }

  factory ConnectIssuerConfigValue.fromJson(Map<String, dynamic> json) {
    return ConnectIssuerConfigValue(
        issuerURI: json['issuerURI'],
        scopes: json['scopes'].cast<String>(),
        clientID: json['clientID'],
        redirectURI: json['redirectURI'],
        showIssuer: json['showIssuer'],
        description: json['description']
    );
  }
}