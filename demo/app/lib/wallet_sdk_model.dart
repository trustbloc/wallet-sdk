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