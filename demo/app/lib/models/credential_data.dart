/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

class CredentialData {
  final String rawCredential;
  final String credentialDisplayData;
  final String issuerDisplayData;
  final String issuerURL;
  final String credentialDID;
  final String credID;

  CredentialData(
      {required this.rawCredential,
      required this.issuerURL,
      required this.issuerDisplayData,
      required this.credentialDisplayData,
      required this.credentialDID,
      required this.credID});

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> data = <String, dynamic>{};
    data['rawCredential'] = rawCredential;
    data['credentialDisplayData'] = credentialDisplayData;
    data['issuerDisplayData'] = issuerDisplayData;
    data['issuerURL'] = issuerURL;
    data['credentialDID'] = credentialDID;
    data['credID'] = credID;
    return data;
  }

  factory CredentialData.fromJson(Map<String, dynamic> json) {
    return CredentialData(
      rawCredential: json['rawCredential'],
      credentialDisplayData: json['credentialDisplayData'],
      issuerURL: json['issuerURL'],
      issuerDisplayData: json['issuerDisplayData'],
      credentialDID: json['credentialDID'],
      credID: json['credID'],
    );
  }
}
