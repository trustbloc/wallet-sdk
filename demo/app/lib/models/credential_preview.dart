class CredentialPreviewData {
  CredentialPreviewData(this.label,this.value, this.valueType, this.order);

  final String label;
  final String value;
  final String valueType;
  final int order;

  factory CredentialPreviewData.fromJson(Map<String, dynamic> json) {
    return CredentialPreviewData(
        json["label"],
        json["raw_value"] ?? '',
        json["value_type"] ?? '',
        json["order"] ?? -1
    );}
}