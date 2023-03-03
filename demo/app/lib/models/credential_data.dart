
class CredentialData {
  final String rawCredential;
  final String credentialDisplayData;
  final String issuerURL;
  final String? credentialDID;


  CredentialData({ required this.rawCredential, required this.issuerURL, required this.credentialDisplayData, this.credentialDID});

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> data = <String, dynamic>{};
    data['rawCredential'] = rawCredential;
    data['credentialDisplayData'] = credentialDisplayData;
    data['issuerURL'] = issuerURL;
    data['credentialDID'] = credentialDID;
    return data;
  }

  factory CredentialData.fromJson( Map<String, dynamic> json) {
    return CredentialData(
      rawCredential: json['rawCredential'],
      credentialDisplayData: json['credentialDisplayData'],
      issuerURL: json['issuerURL'],
      credentialDID: json['credentialDID'],
    );
  }
}

