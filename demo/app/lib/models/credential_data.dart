/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

class CredentialData {
  final String rawCredential;
  final String resolvedCredentialData;
  final String issuerDisplayData;
  final String issuerURL;
  final String credentialDID;
  final String credID;

  CredentialData(
      {required this.rawCredential,
      required this.issuerURL,
      required this.resolvedCredentialData,
      required this.issuerDisplayData,
      required this.credentialDID,
      required this.credID});

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> data = <String, dynamic>{};
    data['rawCredential'] = rawCredential;
    data['resolvedCredentialData'] = resolvedCredentialData;
    data['issuerDisplayData'] = issuerDisplayData;
    data['issuerURL'] = issuerURL;
    data['credentialDID'] = credentialDID;
    data['credID'] = credID;
    return data;
  }

  factory CredentialData.fromJson(Map<String, dynamic> json) {
    return CredentialData(
      rawCredential: json['rawCredential'],
      resolvedCredentialData: json['resolvedCredentialData'],
      issuerURL: json['issuerURL'],
      issuerDisplayData: json['issuerDisplayData'],
      credentialDID: json['credentialDID'],
      credID: json['credID'],
    );
  }
}
