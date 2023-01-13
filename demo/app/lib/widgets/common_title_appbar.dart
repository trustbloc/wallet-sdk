import 'dart:ffi';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:app/views/dashboard.dart';

class CustomTitleAppBar extends AppBar {
  final String? pageTitle;
  final bool? addCloseIcon;
  CustomTitleAppBar({super.key, required this.pageTitle, this.addCloseIcon})
      : super(
      systemOverlayStyle: SystemUiOverlayStyle.light, // 2
      automaticallyImplyLeading:false,
      title: Text(pageTitle!, textAlign: TextAlign.center, style: const TextStyle(fontSize: 18, fontStyle: FontStyle.normal, fontWeight: FontWeight.w700, fontFamily: 'SF Pro')),
      backgroundColor: const Color(0xffEEEAEE),
    actions: addCloseIcon == true ? [
      IconButton(
        // todo implement to go to dashboard
        onPressed: () {},
        icon: const Icon(Icons.close),
      ),
    ] : [],
      flexibleSpace: Container(
        height: 130,
        padding: const EdgeInsets.all(18),
        decoration: const BoxDecoration(
            image: DecorationImage(
              image: ExactAssetImage('lib/assets/images/glow.png'),
              opacity: 0.6,
              alignment: Alignment.topCenter,
              fit: BoxFit.fill,
            ),
            gradient: LinearGradient(
                begin: Alignment.topCenter,
                end: Alignment.bottomCenter,
                colors: <Color>[
                  Color(0xff100716),
                  Color(0xff261131),
                ])
        ),
      ),
  );
}