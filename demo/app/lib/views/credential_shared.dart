import 'package:flutter/material.dart';
import 'package:app/models/store_credential_data.dart';
import 'dashboard.dart';

class CredentialShared extends StatefulWidget {
  final String? verifierName;
  const CredentialShared({super.key, this.verifierName});

  @override
  State<CredentialShared> createState() => CredentialSharedState ();
}

class CredentialSharedState extends State<CredentialShared> {
  late final String userLoggedIn;

  @override
  void initState() {
    super.initState();
    /// Await your Future here (This function only called once after the layout is Complete)
    WidgetsBinding.instance?.addPostFrameCallback((timeStamp) async {
      UserLoginDetails userLoginDetails =  await getUser();
      userLoggedIn = userLoginDetails.username!;
    });
  }

  static const TextStyle optionStyle =
  TextStyle(fontSize: 28, fontWeight: FontWeight.bold);

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Credential Shared'),
        actions: [
          IconButton(
            onPressed: () {
              _navigateToDashboard(userLoggedIn);
          },
            icon: const Icon(Icons.close),
          ),
        ],
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
          mainAxisAlignment: MainAxisAlignment.center,
          crossAxisAlignment: CrossAxisAlignment.center,
          children: <Widget>[
            RichText(
              text: const TextSpan(
                children: [
                  WidgetSpan(
                    child: Icon(Icons.check_circle, size: 30, color: Colors.green),
                  ),
                  TextSpan(
                    text: "Success",
                    style: TextStyle(
                      color: Colors.black,
                      fontWeight: FontWeight.w600,
                      fontSize: 28,
                    ),
                  )
                ],
              ),
            ),
            Text(
              "Credential has been shared with ${widget.verifierName}",
              style: const TextStyle(
                color: Colors.black54,
                fontWeight: FontWeight.w400,
                fontSize: 22,
              ),
            ),
          ],
        ),
      ),
    );
  }
  _navigateToDashboard(String userLoggedIn) async {
    Navigator.push(context, MaterialPageRoute(builder: (context) => const Dashboard()));
  }
}