class CredentialPreviewData {
  CredentialPreviewData(this.label,this.value, this.valueType);

  final String label;
  final String value;
  final String valueType;

  factory CredentialPreviewData.fromJson(Map<String, dynamic> json) {
    return CredentialPreviewData(
        json["label"],
        json["value"] ?? '',
        json["value_type"] ?? ''
    );}
}