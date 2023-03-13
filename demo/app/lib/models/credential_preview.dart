class CredentialPreviewData {
  CredentialPreviewData(this.label,this.value, this.valueType, this.rawValue);

  final String label;
  final String value;
  final String rawValue;
  final String valueType;

  factory CredentialPreviewData.fromJson(Map<String, dynamic> json) {
    return CredentialPreviewData(
        json["label"],
        json["rawValue"] ?? '',
        json["valueType"] ?? '',
        json["value"] ?? '',
    );}
}