/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'package:shared_preferences/shared_preferences.dart';

class StorageItem {
  StorageItem(this.key, this.value);

  final String key;
  final String value;
}

class UserLoginDetails {
  UserLoginDetails(this.username);
  String? username;
}

Future<UserLoginDetails> getUser() async {
  SharedPreferences prefs = await SharedPreferences.getInstance();
  final SharedPreferences p = prefs;
  String? userLoggedIn = p.getString('userLoggedIn');
  print('userLoginDetails -> $userLoggedIn');
  return UserLoginDetails(userLoggedIn);
}
