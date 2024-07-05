/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

class WellKnownDidConfig {
  final bool isValid;
  final String serviceURL;

  const WellKnownDidConfig({
    required this.isValid,
    required this.serviceURL,
  });

  Map<String, dynamic> toMap() {
    return {
      'isValid': isValid,
      'serviceURL': serviceURL,
    };
  }

  factory WellKnownDidConfig.fromMap(Map<String, dynamic> map) {
    return WellKnownDidConfig(
      isValid: map['isValid'] as bool,
      serviceURL: map['serviceURL'] as String,
    );
  }
}

class SubmissionRequirement {
  final String rule;
  final String name;
  final int min;
  final int max;
  final int count;
  final List<SubmissionRequirement> nested;
  final List<InputDescriptor> inputDescriptors;

//<editor-fold desc="Data Methods">
  const SubmissionRequirement({
    required this.rule,
    required this.name,
    required this.min,
    required this.max,
    required this.count,
    required this.nested,
    required this.inputDescriptors,
  });

  @override
  String toString() {
    return 'SubmissionRequirement{ rule: $rule, name: $name, min: $min, max: $max, count: $count, nested: $nested, inputDescriptors: $inputDescriptors,}';
  }

  factory SubmissionRequirement.fromMap(Map<String, dynamic> map) {
    return SubmissionRequirement(
      rule: map['rule'] as String,
      name: map['name'] as String,
      min: map['min'] as int,
      max: map['max'] as int,
      count: map['count'] as int,
      nested: (map['nested'] as List<dynamic>)
          .map((obj) => SubmissionRequirement.fromMap(obj.cast<String, dynamic>()))
          .toList(),
      inputDescriptors: (map['inputDescriptors'] as List<dynamic>)
          .map((obj) => InputDescriptor.fromMap(obj.cast<String, dynamic>()))
          .toList(),
    );
  }

//</editor-fold>
}

class InputDescriptor {
  final String id;
  final String name;
  final String purpose;
  final List<String> matchedVCsID;
  final List<String> matchedVCs;

  const InputDescriptor({
    required this.id,
    required this.name,
    required this.purpose,
    required this.matchedVCsID,
    required this.matchedVCs,
  });

  @override
  String toString() {
    return 'InputDescriptor{ id: $id, name: $name, purpose: $purpose, matchedVCs: $matchedVCs, matchedVCsID: $matchedVCsID }';
  }

  factory InputDescriptor.fromMap(Map<String, dynamic> map) {
    return InputDescriptor(
      id: map['id'] as String,
      name: map['name'] as String,
      purpose: map['purpose'] as String,
      matchedVCsID: map['matchedVCsID'].cast<String>(),
      matchedVCs: map['matchedVCs'].cast<String>(),
    );
  }
}

class CreateDID {
  final String did;
  final String didDoc;

  const CreateDID({
    required this.did,
    required this.didDoc,
  });

  @override
  String toString() {
    return 'CreateDID { did: $did, didDoc: $didDoc }';
  }

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> data = <String, dynamic>{};
    data['did'] = did;
    data['didDoc'] = didDoc;
    return data;
  }

  factory CreateDID.fromJson(Map<String, dynamic> json) {
    return CreateDID(
      did: json['did'],
      didDoc: json['didDoc'],
    );
  }
}

class CredentialOfferDisplayData {
  final List<CredentialDisplayData> offeredCredentials;
  final IssuerDisplayData issuer;

  const CredentialOfferDisplayData({
    required this.issuer,
    required this.offeredCredentials,
  });

  @override
  String toString() {
    return 'IssuerMetaData{ localizedIssuerDisplay: $issuer, offeredCredentials: $offeredCredentials}';
  }

  Map<String, dynamic> toMap() {
    return {
      'localizedIssuerDisplay': issuer,
      'offeredCredentials': offeredCredentials,
    };
  }

  factory CredentialOfferDisplayData.fromMap(Map<String, dynamic> map) {
    return CredentialOfferDisplayData(
        offeredCredentials: (map['offeredCredentials'] as List<dynamic>)
            .map((obj) => CredentialDisplayData.fromMap(obj.cast<String, dynamic>()))
            .toList(),
        issuer: IssuerDisplayData.fromMap(map['localizedIssuerDisplay'].cast<String, dynamic>()));
  }
}

class SupportedCredentials {
  final String format;
  final List<String> types;
  final List<SupportedCredentialDisplayData> display;

  const SupportedCredentials({required this.format, required this.types, required this.display});

  @override
  String toString() {
    return 'SupportedCredentials{ format: $format, types: $types, display: $display }';
  }

  Map<String, dynamic> toMap() {
    return {'format': format, 'types': types, 'display': display};
  }

  factory SupportedCredentials.fromMap(Map<String, dynamic> map) {
    List<dynamic> display = map['display'];
    return SupportedCredentials(
        format: map['format'] as String,
        types: map['types'].cast<String>(),
        display: display.map((c) => SupportedCredentialDisplayData.fromMap(c.cast<String, dynamic>())).toList());
  }
}

class SupportedCredentialDisplayData {
  final String name;
  final String locale;
  final String? logo;
  final String? textColor;
  final String? backgroundColor;

  const SupportedCredentialDisplayData({
    required this.name,
    required this.locale,
    this.logo,
    this.textColor,
    this.backgroundColor,
  });

  @override
  String toString() {
    return 'SupportedCredentialDisplayData { name: $name, locale: $locale, logo: $logo, textColor: $textColor, backgroundColor: $backgroundColor }';
  }

  Map<String, dynamic> toMap() {
    return {'name': name, 'locale': locale, 'textColor': textColor, 'logo': logo, 'backgroundColor': backgroundColor};
  }

  factory SupportedCredentialDisplayData.fromMap(Map<String, dynamic> map) {
    return SupportedCredentialDisplayData(
      name: map['name'] as String,
      locale: map['locale'] as String,
      logo: map['logo'] as String,
      textColor: map['textColor'] as String,
      backgroundColor: map['backgroundColor'] as String,
    );
  }
}

class IssuerDisplayData {
  final String name;
  final String locale;
  final String url;
  String? logo;
  String? backgroundColor;
  String? textColor;

  IssuerDisplayData({
    required this.name,
    required this.locale,
    required this.url,
    this.logo,
    this.textColor,
    this.backgroundColor,
  });

  @override
  String toString() {
    return 'IssuerDisplayData { name: $name, locale: $locale, url: $url , logo: $logo, textColor: $textColor, backgroundColor: $backgroundColor}';
  }

  Map<String, dynamic> toMap() {
    return {
      'name': name,
      'locale': locale,
      'url': url,
      'logo': logo,
      'textColor': textColor,
      'backgroundColor': backgroundColor
    };
  }

  factory IssuerDisplayData.fromMap(Map<String, dynamic> map) {
    return IssuerDisplayData(
        name: map['name'] as String,
        locale: map['locale'] as String,
        url: map['url'] as String,
        logo: map['logo'] as String?,
        textColor: map['textColor'] as String,
        backgroundColor: map['backgroundColor'] as String);
  }
}

class WalletSDKError {
  final String code;
  final String category;
  final String details;
  final String traceID;

  const WalletSDKError({
    required this.code,
    required this.category,
    required this.details,
    required this.traceID,
  });

  @override
  String toString() {
    return 'WalletSDKError { code: $code, category: $category, details: $details, traceID: $traceID }';
  }

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> data = <String, dynamic>{};
    data['code'] = code;
    data['category'] = category;
    data['details'] = details;
    data['traceID'] = traceID;
    return data;
  }

  factory WalletSDKError.fromJson(Map<String, dynamic> json) {
    return WalletSDKError(
      code: json['code'],
      category: json['category'],
      details: json['details'],
      traceID: json['traceID'],
    );
  }
}

class CredentialDisplayData {
  final String overviewName;
  final String logo;
  final String backgroundColor;
  final String textColor;
  final List<CredentialDisplayClaim> claims;

//<editor-fold desc="Data Methods">
  const CredentialDisplayData({
    required this.overviewName,
    required this.logo,
    required this.backgroundColor,
    required this.textColor,
    required this.claims,
  });

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is CredentialDisplayData &&
          runtimeType == other.runtimeType &&
          overviewName == other.overviewName &&
          logo == other.logo &&
          backgroundColor == other.backgroundColor &&
          textColor == other.textColor &&
          claims == other.claims);

  @override
  int get hashCode =>
      overviewName.hashCode ^
      logo.hashCode ^
      backgroundColor.hashCode ^
      textColor.hashCode ^
      claims.hashCode;

  @override
  String toString() {
    return 'CredentialDisplayData{ overviewName: $overviewName, logo: $logo, backgroundColor: $backgroundColor, textColor: $textColor,}';
  }

  CredentialDisplayData copyWith({
    String? issuerName,
    String? overviewName,
    String? logo,
    String? backgroundColor,
    String? textColor,
    List<CredentialDisplayClaim>? claims,
  }) {
    return CredentialDisplayData(
      overviewName: overviewName ?? this.overviewName,
      logo: logo ?? this.logo,
      backgroundColor: backgroundColor ?? this.backgroundColor,
      textColor: textColor ?? this.textColor,
      claims: claims ?? this.claims,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'overviewName': overviewName,
      'logo': logo,
      'backgroundColor': backgroundColor,
      'textColor': textColor,
      'claims': claims.map((c) => c.toMap())
    };
  }

  factory CredentialDisplayData.fromMap(Map<String, dynamic> map) {
    List<dynamic> claims = map['claims'] ?? [];

    return CredentialDisplayData(
        overviewName: map['overviewName'] as String,
        logo: map['logo'] as String,
        backgroundColor: map['backgroundColor'] as String,
        textColor: map['textColor'] as String,
        claims: claims.map((c) => CredentialDisplayClaim.fromMap(c.cast<String, dynamic>())).toList());
  }

//</editor-fold>
}

class CredentialDisplayClaim {
  final String rawValue;
  final String valueType;
  final String? uri;
  final String label;
  final String? value;
  final int? order;

//<editor-fold desc="Data Methods">
  const CredentialDisplayClaim({
    required this.rawValue,
    required this.valueType,
    required this.uri,
    required this.label,
    this.value,
    this.order,
  });

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is CredentialDisplayClaim &&
          runtimeType == other.runtimeType &&
          rawValue == other.rawValue &&
          valueType == other.valueType &&
          uri == other.uri &&
          label == other.label &&
          value == other.value &&
          order == other.order);

  @override
  int get hashCode => rawValue.hashCode ^ valueType.hashCode ^ label.hashCode ^ value.hashCode ^ order.hashCode ^ uri.hashCode;

  @override
  String toString() {
    return 'CredentialDisplayClaim{ rawValue: $rawValue, valueType: $valueType, label: $label, value: $value, order: $order, uri: $uri}';
  }

  CredentialDisplayClaim copyWith({
    String? rawValue,
    String? valueType,
    String? label,
    String? uri,
    String? value,
    int? order,
  }) {
    return CredentialDisplayClaim(
      rawValue: rawValue ?? this.rawValue,
      valueType: valueType ?? this.valueType,
      label: label ?? this.label,
      uri: uri ?? this.uri,
      value: value ?? this.value,
      order: order ?? this.order,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'rawValue': rawValue,
      'valueType': valueType,
      'label': label,
      'value': value,
      'uri': uri,
      'order': order,
    };
  }

  factory CredentialDisplayClaim.fromMap(Map<String, dynamic> map) {
    return CredentialDisplayClaim(
      rawValue: map['rawValue'] as String,
      valueType: map['valueType'] as String,
      label: map['label'] as String,
      uri: map['uri'] as String?,
      value: map['value'] as String?,
      order: map['order'] as int?,
    );
  }

//</editor-fold>
}

class VerifierDisplayData {
  final String name;
  final String did;
  final String purpose;
  final String logoURI;

//<editor-fold desc="Data Methods">
  const VerifierDisplayData({
    required this.name,
    required this.did,
    required this.purpose,
    required this.logoURI,
  });

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is VerifierDisplayData &&
          runtimeType == other.runtimeType &&
          name == other.name &&
          did == other.did &&
          purpose == other.purpose &&
          logoURI == other.logoURI);

  @override
  int get hashCode => name.hashCode ^ did.hashCode ^ purpose.hashCode ^ logoURI.hashCode;

  @override
  String toString() {
    return 'VerifierDisplayData{ name: $name, did: $did, purpose: $purpose, logoURI: $logoURI,}';
  }

  VerifierDisplayData copyWith({
    String? name,
    String? did,
    String? purpose,
    String? logoURI,
  }) {
    return VerifierDisplayData(
      name: name ?? this.name,
      did: did ?? this.did,
      purpose: purpose ?? this.purpose,
      logoURI: logoURI ?? this.logoURI,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'name': name,
      'did': did,
      'purpose': purpose,
      'logoURI': logoURI,
    };
  }

  factory VerifierDisplayData.fromMap(Map<String, dynamic> map) {
    return VerifierDisplayData(
      name: map['name'] as String,
      did: map['did'] as String,
      purpose: map['purpose'] as String,
      logoURI: map['logoURI'] as String,
    );
  }

//</editor-fold>
}

class CredentialWithId {
  final String id;
  final String content;

  const CredentialWithId({
    required this.id,
    required this.content,
  });

  Map<String, dynamic> toMap() {
    return {
      'id': id,
      'content': content,
    };
  }

  factory CredentialWithId.fromMap(Map<String, dynamic> map) {
    return CredentialWithId(
      id: map['id'] as String,
      content: map['content'] as String,
    );
  }
}

class CredentialsDisplayData {
  final List<String> credentialsDisplay;
  final String issuerDisplay;

  const CredentialsDisplayData({
    required this.credentialsDisplay,
    required this.issuerDisplay,
  });

  Map<String, dynamic> toMap() {
    return {
      'credentialsDisplay': credentialsDisplay,
      'issuerDisplay': issuerDisplay,
    };
  }

  factory CredentialsDisplayData.fromMap(Map<String, dynamic> map) {
    return CredentialsDisplayData(
      credentialsDisplay: map['credentialsDisplay']?.cast<String>(),
      issuerDisplay: map['issuerDisplay'] as String,
    );
  }
}

class EvaluationResult {
  final bool allowed;
  final String errorCode;
  final String errorMessage;
  final List<String> requestedAttestations;
  bool multipleCredentialAllowed;
  final String denyReason;

//<editor-fold desc="Data Methods">
  EvaluationResult({
    required this.allowed,
    required this.errorCode,
    required this.errorMessage,
    required this.requestedAttestations,
    required this.multipleCredentialAllowed,
    required this.denyReason
  });

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is EvaluationResult &&
          runtimeType == other.runtimeType &&
          allowed == other.allowed &&
          multipleCredentialAllowed == other.multipleCredentialAllowed &&
          errorCode == other.errorCode &&
          errorMessage == other.errorMessage &&
          denyReason == other.denyReason
      );

  @override
  int get hashCode => allowed.hashCode ^ multipleCredentialAllowed.hashCode ^ errorCode.hashCode ^ errorMessage.hashCode ^ denyReason.hashCode ;

  @override
  String toString() {
    return 'EvaluationResult{ allowed: $allowed, multipleCredentialAllowed: $multipleCredentialAllowed, errorCode: $errorCode, errorMessage: $errorMessage, requestedAttestations: $requestedAttestations, denyReason: $denyReason}';
  }

  EvaluationResult copyWith({
    bool? allowed,
    bool? multipleCredentialAllowed,
    String? errorCode,
    String? errorMessage,
    String? denyReason,
  }) {
    return EvaluationResult(
      allowed: allowed ?? this.allowed,
      errorCode: errorCode ?? this.errorCode,
      errorMessage: errorMessage ?? this.errorMessage,
      requestedAttestations: requestedAttestations,
      denyReason:  denyReason ?? this.denyReason,
      multipleCredentialAllowed: multipleCredentialAllowed ?? this.multipleCredentialAllowed,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'allowed': allowed,
      'multipleCredentialAllowed' : multipleCredentialAllowed,
      'errorCode': errorCode,
      'errorMessage': errorMessage,
      'requestedAttestations': requestedAttestations,
      'denyReason': denyReason
    };
  }

  factory EvaluationResult.fromMap(Map<String, dynamic> map) {
    return EvaluationResult(
      allowed: map['allowed'] as bool,
      multipleCredentialAllowed: map['multipleCredentialAllowed'] as bool,
      errorCode: map['errorCode'] as String,
      errorMessage: map['errorMessage'] as String,
      requestedAttestations: map['requestedAttestations'].cast<String>(),
      denyReason: map['denyReason'] as String
    );
  }

//</editor-fold>
}
