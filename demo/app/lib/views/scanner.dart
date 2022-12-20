import 'dart:developer';
import 'package:app/models/credential_data_object.dart';
import 'package:app/views/presentation_preview.dart';
import 'package:flutter/material.dart';
import 'package:flutter/scheduler.dart';
import 'package:qr_code_scanner/qr_code_scanner.dart';
import 'package:app/views/otp.dart';
import 'package:app/demo_method_channel.dart';
import '../services/storage_service.dart';
import 'dashboard.dart';

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
    controller.scannedDataStream.listen((scanData) async {
      controller.dispose();
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

  // Skip the otp if user pin required is false
  _navigateToDashboard(String userLoggedIn) async {
    Navigator.push(context, MaterialPageRoute(builder: (context) => Dashboard(user: userLoggedIn)));
  }

  _authorize(String qrCodeURL) async {
    final StorageService storageService = StorageService();
    late List<CredentialDataObject> storedCredentials;
    log('received qr code url - $qrCodeURL');
    if (!qrCodeURL.contains("openid-vc")) {
      var authorizeResultPinRequired = await WalletSDKPlugin.authorize(qrCodeURL);
      if (authorizeResultPinRequired == true) {
        _navigateToOTPScreen();
        return;
      }
    } else {
    // Check if the flow is for the verifiable presentation or for issuance.
      var username = await storageService.retrieve("username");
      storedCredentials = await storageService.retrieveCredentials(username!);
      var credentials = storedCredentials.map((e) => e.value.rawCredential).toList();
      await WalletSDKPlugin.processAuthorizationRequest(
          authorizationRequest: qrCodeURL, storedCredentials: credentials);
      // TODO if the creds returned in the process authorize request matches anything in the retrieved credentials
      await WalletSDKPlugin.presentCredential();
      _navigateToPresentationPreviewScreen();
      return;
      }
    }
  }
