import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

class CustomTitleAppBar extends AppBar {
  final String? pageTitle;
  CustomTitleAppBar({super.key, required this.pageTitle})
      : super(
      systemOverlayStyle: SystemUiOverlayStyle.light, // 2
      automaticallyImplyLeading:false,
      title: Text(pageTitle!, textAlign: TextAlign.center, style: const TextStyle(fontSize: 22, fontStyle: FontStyle.normal, fontWeight: FontWeight.w700, fontFamily: 'SF Pro')),
      backgroundColor: const Color(0xffEEEAEE),
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