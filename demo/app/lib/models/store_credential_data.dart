import 'package:shared_preferences/shared_preferences.dart';

class StorageItem {
  StorageItem(this.key, this.value);

  final String key;
  final String value;
}

class UserLoginDetails {
  UserLoginDetails(this.username, this.userDID);
  String username;
  String userDID;
}


Future<UserLoginDetails> getUser() async {
  SharedPreferences prefs = await SharedPreferences.getInstance();
  final SharedPreferences p = prefs;
  String? userLoggedIn =  p.getString("userLoggedIn");
  String? userDID=  p.getString("userDID");
  return UserLoginDetails(userLoggedIn!, userDID!);
}
