/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'dart:async';
import 'dart:developer';
import 'package:app/scenarios/handle_openid_url.dart';
import 'package:app/widgets/common_title_appbar.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:mobile_scanner/mobile_scanner.dart';
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
  final GlobalKey qrKey = GlobalKey(debugLabel: 'QR');
  final textController = TextEditingController();
  bool isRealDevice = false;

  final MobileScannerController controller = MobileScannerController(
      detectionSpeed: DetectionSpeed.noDuplicates,
      returnImage: true
  );

  @override
  void initState() {
    super.initState();
    isDeviceReal();
  }

  isDeviceReal() async {
    bool isRealDeviceResp = !kIsWeb && await SafeDevice.isRealDevice;
    setState(() {
      isRealDevice = isRealDeviceResp;
    });
  }

  @override
  Widget build(BuildContext context) {
    return isRealDevice
        ? Scaffold(
            appBar: const CustomTitleAppBar(
              pageTitle: 'Scan QR',
              addCloseIcon: true,
              height: 50,
            ),
            body: MobileScanner(
                controller: controller,
                onDetect: (capture) {
                  final List<Barcode> barcodes = capture.barcodes;
                  for (final barcode in barcodes) {
                    handleOpenIDUrl(context, barcode.rawValue ?? 'No Data found in QR');
                  }
                  controller.dispose();
                }),
          )
        : Scaffold(
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
                      onDoubleTap: () {
                        if (textController.text.isNotEmpty) {
                          textController.selection =
                              TextSelection(baseOffset: 0, extentOffset: textController.text.length);
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
                    child: const Text('Submit', style: TextStyle(fontSize: 16, color: Colors.white))),
              ],
            ));
  }

}
