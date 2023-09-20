/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'package:flutter/material.dart';

import 'package:app/widgets/common_title_appbar.dart';

class CustomError extends StatefulWidget {
  String? titleBar;
  final String requestErrorTitleMsg;
  final String requestErrorSubTitleMsg;

  CustomError({super.key, required this.requestErrorTitleMsg, required this.requestErrorSubTitleMsg, this.titleBar});

  @override
  State<CustomError> createState() => CustomErrorPage();
}

class CustomErrorPage extends State<CustomError> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: CustomTitleAppBar(pageTitle: widget.titleBar, addCloseIcon: true, height: 60),
      body: Center(
        child: Container(
          padding: const EdgeInsets.all(12),
          alignment: Alignment.center,
          child: ListTile(
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(12),
            ),
            tileColor: const Color(0xffFBF8FC),
            title: SelectableText(
              widget.requestErrorTitleMsg ?? '',
              style: const TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                color: Color(0xff190C21),
              ),
              textAlign: TextAlign.start,
            ),
            subtitle: SelectableText(
              widget.requestErrorSubTitleMsg ?? '',
              style: const TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: Color(0xff6C6D7C),
              ),
              textAlign: TextAlign.start,
            ),
            //TODO need to add fallback and network image url
            leading: const SizedBox(
                height: 24,
                width: 24,
                child: Image(
                  image: AssetImage('lib/assets/images/errorVector.png'),
                  width: 24,
                  height: 24,
                  fit: BoxFit.cover,
                )),
          ),
        ),
      ),
    );
  }
}
