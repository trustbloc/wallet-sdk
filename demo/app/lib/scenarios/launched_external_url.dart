import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:app/views/auth_code_redirect.dart';
import 'dart:developer';

class LaunchExternalURL extends StatefulWidget {
  late final Uri uri;
  LaunchExternalURL(this.uri);

  @override
  State<LaunchExternalURL> createState() => LaunchExternalURLState();
}

class LaunchExternalURLState extends State<LaunchExternalURL> {

  Future<void> _launchUrl(Uri uri) async {
    if (!await launch(uri.toString(), forceWebView: true, forceSafariVC: false)) {
      throw 'Could not launch $uri';
    }
  }
  @override
  Widget build(BuildContext context) {
    return Center(
      child: FutureBuilder(
        future: _launchUrl(widget.uri),
        builder: (context, snapshot) {
          if (snapshot.connectionState == ConnectionState.done) {
            Future.delayed(Duration.zero, () {
              Navigator.push(
                context,
                MaterialPageRoute(builder: (context) => RedirectPathTextBox()),
              );
            });
          }
          return CircularProgressIndicator();
        },
      ),
    );
  }
}
