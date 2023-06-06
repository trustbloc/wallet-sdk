class CredentialOfferObject {
  final String credentialIssuer;

  const CredentialOfferObject({
    required this.credentialIssuer,
  });

  factory CredentialOfferObject.fromJson(Map<String, dynamic> json) {
    return CredentialOfferObject(
      credentialIssuer: json['credential_issuer'],
    );
  }
}