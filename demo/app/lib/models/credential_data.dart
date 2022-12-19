
class CredentialData {
  final String rawCredential;
  final String credentialDisplayData;

  CredentialData({ required this.rawCredential, required this.credentialDisplayData});

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> data = <String, dynamic>{};
    data['rawCredential'] = rawCredential;
    data['credentialDisplayData'] = credentialDisplayData;
    return data;
  }

  factory CredentialData.fromJson( Map<String, dynamic> json) {
    return CredentialData(
      rawCredential: json['rawCredential'],
      credentialDisplayData: json['credentialDisplayData'],
    );
  }
}

