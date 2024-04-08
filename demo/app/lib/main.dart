/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'package:app/scenarios/handle_openid_url.dart';
import 'package:app/services/config_service.dart';
import 'package:app/widgets/common_logo_appbar.dart';
import 'package:app/widgets/primary_input_field.dart';
import 'package:app/widgets/primary_button.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'wallet_sdk/wallet_sdk.dart';
import 'views/dashboard.dart';
import 'package:uni_links/uni_links.dart';

final WalletSDKPlugin = WalletSDK();

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await ConfigService.init();
  await WalletSDKPlugin.initSDK(ConfigService.config.didResolverURI);

  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  // This widget is the root of your application.
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      home: Scaffold(
        appBar: CustomLogoAppBar(),
        body: Container(
          decoration: const BoxDecoration(
            image: DecorationImage(
                image: ExactAssetImage('lib/assets/images/background.png'), alignment: Alignment.topLeft),
          ),
          child: const MainWidget(),
        ),
        backgroundColor: const Color(0xffF4F1F5),
      ),
      debugShowCheckedModeBanner: false, //Removing Debug Banner
    );
  }
}

class MainWidget extends StatefulWidget {
  const MainWidget({Key? key}) : super(key: key);

  @override
  State<MainWidget> createState() => _MainWidgetState();
}

class _MainWidgetState extends State<MainWidget> {
  final TextEditingController _usernameController = TextEditingController();
  final Future<SharedPreferences> prefs = SharedPreferences.getInstance();
  String? initialLink;

  @override
  void initState() {
    super.initState();

    getInitialLink().then((value) {
      initialLink = value;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Center(
        child: ListView(
      shrinkWrap: true,
      children: <Widget>[
        Container(
          padding: const EdgeInsets.fromLTRB(0, 0, 0, 24),
          child: const Text('TrustBloc Sign In',
              textAlign: TextAlign.center,
              style: TextStyle(fontWeight: FontWeight.bold, fontSize: 24, fontFamily: 'SF Pro')),
        ),
        Container(
          padding: const EdgeInsets.fromLTRB(24, 0, 24, 0),
          child: PrimaryInputField(
            textController: _usernameController,
            titleTextAlign: TextAlign.center,
            labelText: 'Username',
            textInputFormatter: FilteringTextInputFormatter.singleLineFormatter,
          ),
        ),
        Container(
          padding: const EdgeInsets.fromLTRB(24, 24, 24, 0),
          child: PrimaryButton(
            width: double.infinity,
            onPressed: () async {
              final SharedPreferences pref = await prefs;
              pref.setString('userLoggedIn', _usernameController.text);
              _loginCompleted();
            },
            child: const Text('Sign In ', style: TextStyle(fontSize: 16, color: Colors.white)),
          ),
        ),
        Container(
          padding: const EdgeInsets.fromLTRB(24, 24, 24, 0),
          child: const Text('This is a reference app and not to be used for production use cases.',
              textAlign: TextAlign.center,
              style: TextStyle(
                  fontWeight: FontWeight.normal, fontSize: 14, fontFamily: 'SF Pro', color: Color(0xff6C6D7C))),
        ),
      ],
    ));
  }

  _loginCompleted() async {
    try {
      if (initialLink != null && !kIsWeb) {
        handleOpenIDUrl(context, initialLink!);
        return;
      }
    } on PlatformException {}

    Navigator.push(context, MaterialPageRoute(builder: (context) => const Dashboard()));
  }
}
