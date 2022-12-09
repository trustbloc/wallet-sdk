class CredentialPreviewData {
  CredentialPreviewData(this.label, this.value);

  final String label;
  final String value;

  factory CredentialPreviewData.fromJson(Map<String, dynamic> json) {
    return CredentialPreviewData(
        json["label"],
        json["value"]
    );}
}