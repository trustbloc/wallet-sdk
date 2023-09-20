/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

class ActivityLogger {
  final String status;
  final String operation;
  final String date;
  final String issuedBy;
  final String activityType;

  ActivityLogger(
      {required this.status,
      required this.operation,
      required this.date,
      required this.issuedBy,
      required this.activityType});

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> data = <String, dynamic>{};
    data['Status'] = status;
    data['Operation'] = operation;
    data['Date'] = date;
    data['Issued By'] = issuedBy;
    data['Activity Type'] = activityType;
    return data;
  }

  factory ActivityLogger.fromJson(Map<String, dynamic> json) {
    return ActivityLogger(
      status: json['Status'],
      operation: json['Operation'],
      date: json['Date'],
      issuedBy: json['Issued By'],
      activityType: json['Activity Type'],
    );
  }
}
