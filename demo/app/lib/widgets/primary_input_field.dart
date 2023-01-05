import 'package:flutter/material.dart';

class PrimaryInputField extends StatelessWidget {
  const PrimaryInputField(
      {super.key, this.labelText = '',
        this.titleTextAlign = TextAlign.center,
        required this.textController});

  final String labelText;
  final TextAlign titleTextAlign;
  final TextEditingController textController;

  @override
  Widget build(BuildContext context) {
    final labelText = this.labelText ?? '';
    return TextField(
      controller: textController,
      decoration: InputDecoration(
          floatingLabelStyle: const TextStyle(color: Color(0xff190C21)),
        fillColor: const Color(0xffEEEAEE),
        filled: true,
        enabledBorder: const UnderlineInputBorder( //<-- SEE HERE
          borderSide: BorderSide(
              width: 2, color: Color(0xff8D8A8E)),
        ),
        focusedBorder: const UnderlineInputBorder(
          borderSide: BorderSide(width: 2, color: Color(0xff8D8A8E)),
        ),
        border: const OutlineInputBorder(
          borderRadius: BorderRadius.only(
            topRight: Radius.circular(6),
            topLeft: Radius.circular(6),
          ),
          borderSide: BorderSide(
            width: 0,
            style: BorderStyle.none,
          ),
        ),
        labelText: labelText,
      ),
      style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16, fontFamily: 'SF Pro', color: Color(0xff190C21)),
    );
  }
}