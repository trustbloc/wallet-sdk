
class CredentialData {
  final String rawCredential;
  final String credentialDisplayData;
  final String issuerURL;
  final String? credentialDID;
  final String? credID;


  CredentialData({ required this.rawCredential, required this.issuerURL, required this.credentialDisplayData, this.credentialDID, this.credID});

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> data = <String, dynamic>{};
    data['rawCredential'] = rawCredential;
    data['credentialDisplayData'] = credentialDisplayData;
    data['issuerURL'] = issuerURL;
    data['credentialDID'] = credentialDID;
    data['credID'] = credID;
    return data;
  }

  factory CredentialData.fromJson( Map<String, dynamic> json) {
    return CredentialData(
      rawCredential: json['rawCredential'],
      credentialDisplayData: json['credentialDisplayData'],
      issuerURL: json['issuerURL'],
      credentialDID: json['credentialDID'],
      credID: json['credID'],
    );
  }
}

