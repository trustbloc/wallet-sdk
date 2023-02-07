
class CredentialData {
  final String rawCredential;
  final String credentialDisplayData;
  final String issuerURL;


  CredentialData({ required this.rawCredential, required this.issuerURL, required this.credentialDisplayData});

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> data = <String, dynamic>{};
    data['rawCredential'] = rawCredential;
    data['credentialDisplayData'] = credentialDisplayData;
    data['issuerURL'] = issuerURL;
    return data;
  }

  factory CredentialData.fromJson( Map<String, dynamic> json) {
    return CredentialData(
      rawCredential: json['rawCredential'],
      credentialDisplayData: json['credentialDisplayData'],
      issuerURL: json['issuerURL'],
    );
  }
}

