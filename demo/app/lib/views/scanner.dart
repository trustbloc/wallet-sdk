import 'dart:developer';
import 'package:app/scenarios/handle_openid_url.dart';
import 'package:app/widgets/common_title_appbar.dart';
import 'package:flutter/material.dart';
import 'package:qr_code_scanner/qr_code_scanner.dart';

class QRScanner extends StatefulWidget {
  const QRScanner({Key? key}) : super(key: key);

  @override
  State<StatefulWidget> createState() {
    return QRScannerState();
  }
}

class QRScannerState extends State<QRScanner> {
  QRViewController? controller;
  final GlobalKey qrKey = GlobalKey(debugLabel: 'QR');

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: const CustomTitleAppBar(
        pageTitle: 'Scan QR',
        addCloseIcon: true,
        height: 50,
      ),
      body: QRView(
        key: qrKey,
        onQRViewCreated: _onQRViewCreated,
        overlay: QrScannerOverlayShape(
          borderColor: const Color(0xffE74577),
          borderRadius: 24,
          borderLength: 30,
          borderWidth: 10,
          cutOutSize: 256,
        ),
        onPermissionSet: (ctrl, p) => _onPermissionSet(context, ctrl, p),
      ),
    );
  }

  void _onQRViewCreated(QRViewController controller) {
    this.controller = controller;
    controller.scannedDataStream.listen((scanData) async {
      controller.dispose();
      handleOpenIDUrl(context, scanData.code!);
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
}
