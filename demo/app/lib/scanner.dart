import 'dart:developer';

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
      // Todo Issue 63 - Integrate wallet sdk new interaction function here
      _invokeNewInteraction(result?.code);
      controller!.pauseCamera();
      controller!.dispose();
      // TODO Call Authorize function
      // TODO Check if the user pin is required here
      SchedulerBinding.instance.addPostFrameCallback((_) {
        _loadDataAndNavigate();
      });
    } else {
      const Text("Scan QR Code");
    }
  }

  _loadDataAndNavigate() async {
    // fetch data | await this.service.fetch(x,y)
    Navigator.push(context, MaterialPageRoute(builder: (context) => const OTP()));
  }

   _invokeNewInteraction(String? qrCode){
    print("Scanned qr code => $qrCode");
    return;
     // Todo #63 invoke sdk open vc api function to pass the qr code
    /*  var interaction = await WalletSDKPlugin.newInteraction(qrCode);*/
  }

    @override
     void dispose() {
      controller!.stopCamera();
      controller!.dispose();
      super.dispose();
    }
  }


