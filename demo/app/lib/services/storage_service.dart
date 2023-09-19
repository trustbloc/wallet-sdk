/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'dart:convert';
import 'dart:core';

import 'package:app/models/activity_data_object.dart';
import 'package:app/models/credential_data_object.dart';
import 'package:app/models/credential_data.dart';
import 'package:flutter/cupertino.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:app/models/store_credential_data.dart';

class StorageService {
  final _secureStorage = const FlutterSecureStorage();

  AndroidOptions _getAndroidOptions() => const AndroidOptions(
        encryptedSharedPreferences: true,
      );

  Future<void> addCredential(CredentialDataObject newItem) async {
    debugPrint('Adding new data having key ${newItem.key}');
    await _secureStorage.write(
        key: newItem.key, value: json.encode(newItem.value.toJson()), aOptions: _getAndroidOptions());
  }

  Future<List<CredentialDataObject>> retrieveCredentials(String username) async {
    debugPrint('Retrieve all secured data');
    var allData = await _secureStorage.readAll(aOptions: _getAndroidOptions());
    List<CredentialDataObject> list = allData.entries
        .where((e) => e.key.contains(username))
        .map((e) => CredentialDataObject(e.key, CredentialData.fromJson(jsonDecode(e.value))))
        .toList();
    return list;
  }

  Future<void> addActivities(ActivityDataObj activityObj) async {
    debugPrint('Adding new data having key ${activityObj.key}');
    await _secureStorage.write(
        key: activityObj.key, value: jsonEncode(activityObj.value), aOptions: _getAndroidOptions());
  }

  Future<List> retrieveActivities(String credID) async {
    debugPrint('Retrieve stored activities $credID');
    var allData = await _secureStorage.readAll(aOptions: _getAndroidOptions());
    List list = allData.entries.where((e) => e.key.contains(credID)).map((e) => jsonDecode(e.value)).toList();
    return list.first;
  }

  Future<void> deleteData(CredentialDataObject item) async {
    debugPrint('Deleting data having key ${item.key}');
    await _secureStorage.delete(key: item.key, aOptions: _getAndroidOptions());
  }

  Future<List<StorageItem>> retrieveAll() async {
    debugPrint('Retrieve all secured data');
    var allData = await _secureStorage.readAll(aOptions: _getAndroidOptions());
    List<StorageItem> list = allData.entries.map((e) => StorageItem(e.key, e.value)).toList();
    return list;
  }

  Future<void> deleteAllData() async {
    debugPrint('Deleting all secured data');
    await _secureStorage.deleteAll(aOptions: _getAndroidOptions());
  }

  Future<bool> containsKeyInSecureData(String key) async {
    debugPrint('Checking data for the key $key');
    var containsKey = await _secureStorage.containsKey(key: key, aOptions: _getAndroidOptions());
    return containsKey;
  }
}
