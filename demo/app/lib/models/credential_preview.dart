class CredentialPreviewData {
  CredentialPreviewData(this.label,this.rawValue, this.valueType, this.value , this.order);

  final String label;
  final String rawValue;
  final String valueType;
  final String value;
  final int order;


  factory CredentialPreviewData.fromJson(Map<String, dynamic> json) {
    return CredentialPreviewData(
        json["label"],
        json["rawValue"] ?? '',
        json["valueType"],
        json["value"] ?? '',
        json["order"] ?? -1,
    );}
}