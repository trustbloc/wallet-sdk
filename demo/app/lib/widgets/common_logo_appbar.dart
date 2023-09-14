import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

class CustomLogoAppBar extends AppBar {
  CustomLogoAppBar({super.key})
      : super(
          systemOverlayStyle: SystemUiOverlayStyle.light, // 2
          automaticallyImplyLeading: false,
          title: Image.asset(
            'lib/assets/images/logo.png',
            fit: BoxFit.contain,
            height: 24,
            width: 144.6,
          ),
          backgroundColor: const Color(0xffEEEAEE),
          flexibleSpace: Container(
            height: 130,
            decoration: const BoxDecoration(
                image: DecorationImage(
                  image: ExactAssetImage('lib/assets/images/glow.png'),
                  opacity: 0.6,
                  alignment: Alignment.topCenter,
                  fit: BoxFit.fill,
                ),
                gradient: LinearGradient(begin: Alignment.topCenter, end: Alignment.bottomCenter, colors: <Color>[
                  Color(0xff100716),
                  Color(0xff261131),
                ])),
          ),
        );
}
