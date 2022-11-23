import 'dart:convert';
import 'dart:developer';

import 'package:app/credential_preview.dart';
import 'package:flutter/material.dart';
import 'package:flutter/scheduler.dart';
import 'package:qr_code_scanner/qr_code_scanner.dart';
import 'OTP.dart';
import 'demo_method_channel.dart';

class QRScanner extends StatefulWidget {
  const QRScanner({Key? key}) : super(key: key);

  @override
  State<StatefulWidget> createState() {
    return QRScannerState();
  }
}

var WalletSDKPlugin = MethodChannelWallet();

class QRScannerState extends State<QRScanner> {
  Barcode? result;
  QRViewController? controller;
  final GlobalKey qrKey = GlobalKey(debugLabel: 'QR');

  @override
  Widget build(BuildContext context) {
    readQr();
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
      ),
    );
  }
  void _onQRViewCreated(QRViewController controller) {
    setState(() {
      this.controller = controller;
    });
    controller.scannedDataStream.listen((scanData) {
      setState(() {
        result = scanData;
      });
    });
  }

  void readQr() async {
    if (result != null) {
      controller!.pauseCamera();
      controller!.dispose();
      SchedulerBinding.instance.addPostFrameCallback((_) {
        _authorize(result!.code!);
      });
    } else {
      const Text("Scan QR Code");
    }
  }

  _navigateToOTPScreen() async {
    Navigator.push(context, MaterialPageRoute(builder: (context) => const OTP()));
  }

  _navigateToCredPreviewScreen() async {
    Navigator.push(context, MaterialPageRoute(builder: (context) => const CreatePreview(credentialResponse: '',)));
  }

   _authorize(String qrCode) async {
    var authorizeResultPinRequired = await WalletSDKPlugin.authorize(qrCode);
    if (authorizeResultPinRequired == true){
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


