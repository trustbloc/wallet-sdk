class CredentialPreviewData {
  CredentialPreviewData(this.label,this.rawValue, this.valueType, this.value );

  final String label;
  final String rawValue;
  final String valueType;
  final String value;


  factory CredentialPreviewData.fromJson(Map<String, dynamic> json) {
    return CredentialPreviewData(
        json["label"],
        json["rawValue"] ?? '',
        json["valueType"],
        json["value"] ?? '',
    );}
}