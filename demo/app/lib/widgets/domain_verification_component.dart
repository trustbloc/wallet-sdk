import 'package:flutter/material.dart';
import 'package:flutter_svg/svg.dart';

class DomainVerificationComponent extends StatefulWidget {
  final String status;
  final String imagePath;

  const DomainVerificationComponent({
    super.key,
    required this.status,
    required this.imagePath,
  });

  @override
  State<DomainVerificationComponent> createState() =>
      DomainVerificationComponentState();
}

class DomainVerificationComponentState
    extends State<DomainVerificationComponent> {
  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(
        horizontal: 8,
        vertical: 4,
      ),
      decoration: ShapeDecoration(
        shape: RoundedRectangleBorder(
          side: const BorderSide(
            width: 1,
            color: Color(0xff190C21),
          ),
          borderRadius: BorderRadius.circular(100),
        ),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        mainAxisAlignment: MainAxisAlignment.start,
        crossAxisAlignment: CrossAxisAlignment.center,
        children: [
          SvgPicture.asset(
            widget.imagePath,
            height: 16,
            width: 16,
            allowDrawingOutsideViewBox: true,
          ),
          const SizedBox(width: 4),
          Text(
            widget.status,
            style: const TextStyle(
              color: Color(0xff190C21),
              fontSize: 13,
              fontFamily: 'Poppins',
              fontWeight: FontWeight.w700,
            ),
          ),
        ],
      ),
    );
  }
}
