import 'dart:developer';

import 'package:app/views/credential_preview.dart';
import 'package:app/views/presentation_preview.dart';
import 'package:flutter/material.dart';
import 'package:flutter/scheduler.dart';
import 'package:qr_code_scanner/qr_code_scanner.dart';
import 'package:app/views/otp.dart';
import 'package:app/demo_method_channel.dart';

class QRScanner extends StatefulWidget {
  const QRScanner({Key? key}) : super(key: key);

  @override
  State<StatefulWidget> createState() {
    return QRScannerState();
  }
}

var WalletSDKPlugin = MethodChannelWallet();

class QRScannerState extends State<QRScanner> {
  QRViewController? controller;
  final GlobalKey qrKey = GlobalKey(debugLabel: 'QR');

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: QRView(
        key: qrKey,
        onQRViewCreated: _onQRViewCreated,
        overlay: QrScannerOverlayShape(
          borderColor: Colors.green,
          borderRadius: 10,
          borderLength: 30,
          borderWidth: 10,
          cutOutSize: 250,
        ),
        onPermissionSet: (ctrl, p) => _onPermissionSet(context, ctrl, p),
      ),
    );
  }

  void _onQRViewCreated(QRViewController controller) {
    this.controller = controller;
    controller.scannedDataStream.listen((scanData) {
      _authorize(scanData.code!);
    });
  }

  void _onPermissionSet(BuildContext context, QRViewController ctrl, bool p) {
    log('${DateTime.now().toIso8601String()}_onPermissionSet $p');
    if (!p) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('no Permission')),
      );
    }
  }

  _navigateToOTPScreen() async {
    Navigator.push(context, MaterialPageRoute(builder: (context) => const OTP()));
  }

  _navigateToPresentationPreviewScreen() async {
    Navigator.push(context, MaterialPageRoute(builder: (context) => const PresentationPreview()));
  }

  _navigateToCredPreviewScreen() async {
    Navigator.push(
        context,
        MaterialPageRoute(
            builder: (context) => const CredentialPreview(
                  credentialResponse: '',
                )));
  }

  _authorize(String qrCode) async {
    print("what is qr code $qrCode");
    // Check if the flow is for the verifiable presentation or for issuance.
    if (qrCode.contains("presentation_definition") || qrCode.contains("presentationdefs")) {
      _navigateToPresentationPreviewScreen();
      return;
    }
    var authorizeResultPinRequired = await WalletSDKPlugin.authorize(qrCode);
    if (authorizeResultPinRequired == true) {
      _navigateToOTPScreen();
    } else {
      // Skip the otp if user pin required is false
      _navigateToCredPreviewScreen();
    }
  }

  @override
  void dispose() {
    controller!.stopCamera();
    controller!.dispose();
    super.dispose();
  }
}
