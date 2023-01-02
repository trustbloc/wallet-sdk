import 'package:app/models/store_credential_data.dart';
import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

class Settings extends StatefulWidget {
  const Settings({super.key});

  @override
  SettingsState createState() => SettingsState();
}

class SettingsState extends State<Settings> {
  final TextEditingController _usernameController = TextEditingController();
  final TextEditingController _userDIDController = TextEditingController();
  @override
  Widget build(BuildContext context) {
    getUserDetails();
    return Scaffold(
      appBar: AppBar(
        automaticallyImplyLeading: false,
        title: const Text('Settings'),
        backgroundColor: const Color(0xffEEEAEE),
        flexibleSpace: Container(
          decoration: const BoxDecoration(
              gradient: LinearGradient(
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                  stops: [0.0, 1.0],
                  colors: <Color>[
                    Color(0xff261131),
                    Color(0xff100716),
                  ])
          ),
        ),
      ),
      body: Center(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Flexible(
           child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 24),
              child: TextFormField(
                  enabled: false,
                  controller: _usernameController,
                  decoration: const InputDecoration(
                    fillColor:  Color(0xff8D8A8E),
                  border: UnderlineInputBorder(),
                  labelText: 'Username',
                  labelStyle: TextStyle(color: Color(0xff190C21), fontWeight: FontWeight.w700,
                      fontFamily: 'SF Pro', fontSize: 16, fontStyle: FontStyle.normal ),
                  )
              ),
            ),
    ),
            Flexible(
              child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical:0),
        child: TextFormField(
            enabled: false,
            controller: _userDIDController,
            keyboardType: TextInputType.multiline,
            maxLines: null,
            decoration: const InputDecoration(
              fillColor:  Color(0xff8D8A8E),
              border: UnderlineInputBorder(),
              labelText: 'DID ID',
              labelStyle: TextStyle(color: Color(0xff190C21), fontWeight: FontWeight.w700,
                  fontFamily: 'SF Pro', fontSize: 16, fontStyle: FontStyle.normal ),
            )
        ),
       ),
       ),
      ],
     ),
    ),
    );
  }
  getUserDetails() async {
   UserLoginDetails userLoginDetails =  await getUser();
   _usernameController.text = userLoginDetails.username!;
   _userDIDController.text = userLoginDetails.userDID!;
  }

}