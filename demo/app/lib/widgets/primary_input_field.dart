import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

class PrimaryInputField extends StatelessWidget {
  const PrimaryInputField(
      {super.key,
      this.labelText = '',
      this.textInputType = TextInputType.text,
      this.maxLength = 12,
      required this.textInputFormatter,
      this.titleTextAlign = TextAlign.center,
      required this.textController});

  final String labelText;
  final int maxLength;
  final TextAlign titleTextAlign;
  final TextEditingController textController;
  final TextInputType textInputType;
  final TextInputFormatter textInputFormatter;

  @override
  Widget build(BuildContext context) {
    final labelText = this.labelText ?? '';
    return TextField(
      maxLength: maxLength,
      controller: textController,
      keyboardType: textInputType,
      inputFormatters: <TextInputFormatter>[
        textInputFormatter,
      ],
      decoration: InputDecoration(
        floatingLabelStyle: const TextStyle(color: Color(0xff190C21), fontWeight: FontWeight.w700, fontSize: 16),
        fillColor: const Color(0xffEEEAEE),
        filled: true,
        enabledBorder: const UnderlineInputBorder(
          //<-- SEE HERE
          borderSide: BorderSide(width: 2, color: Color(0xff8D8A8E)),
        ),
        focusedBorder: const UnderlineInputBorder(
          borderSide: BorderSide(width: 2, color: Color(0xff8D8A8E)),
        ),
        border: const OutlineInputBorder(
          borderRadius: BorderRadius.only(
            topRight: Radius.circular(12),
            topLeft: Radius.circular(12),
          ),
          borderSide: BorderSide(
            width: 0,
            style: BorderStyle.none,
          ),
        ),
        labelText: labelText,
      ),
      style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 16, fontFamily: 'SF Pro', color: Color(0xff190C21)),
    );
  }
}
