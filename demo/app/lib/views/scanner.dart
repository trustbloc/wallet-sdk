import 'dart:developer';
import 'package:app/scenarios/handle_openid_url.dart';
import 'package:app/widgets/common_title_appbar.dart';
import 'package:flutter/material.dart';
import 'package:qr_code_scanner/qr_code_scanner.dart';
import 'package:safe_device/safe_device.dart';

import 'package:app/widgets/primary_button.dart';

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
  final textController = TextEditingController();
  bool? isRealDevice;

  @override
  void initState(){
    super.initState();
    isDeviceReal();
  }

  isDeviceReal ()async{
    bool isRealDeviceResp = await SafeDevice.isRealDevice;
    setState((){
      isRealDevice = isRealDeviceResp;
    });
  }

  @override
  Widget build(BuildContext context) {
    return isRealDevice! ? Scaffold(
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
    ):  Scaffold(
      appBar: const CustomTitleAppBar(
        pageTitle: 'QR Code Simulator',
        addCloseIcon: true,
        height: 50,
      ),
    body: Column(
      crossAxisAlignment: CrossAxisAlignment.center,
      children: <Widget>[
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 32),
          child: GestureDetector(
            onDoubleTap:() {
              if(textController.text.isNotEmpty) {
                textController.selection = TextSelection(baseOffset: 0, extentOffset:textController.text.length);
              }
            },
         child: TextField(
            controller: textController,
            keyboardType: TextInputType.multiline,
            minLines: 1,
            maxLines: 7,
            autofocus: true,
            decoration: const InputDecoration(
              border: OutlineInputBorder(),
              labelText: 'Paste the qr code url',
            ),
          ),
          )),
        const Padding(
          padding: EdgeInsets.fromLTRB(24, 0, 24, 0),
        ),
        PrimaryButton(
            onPressed: () async {
              handleOpenIDUrl(context, textController.text.toString());
            },
            width: double.infinity,
            child: const Text('Submit', style: TextStyle(fontSize: 16, color: Colors.white))
        ),
      ],
     )
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
