import 'dart:developer';

import 'package:app/wallet_sdk/wallet_sdk_model.dart';
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:app/models/credential_data.dart';
import 'package:app/main.dart';

class CredentialMetaDataCard extends StatefulWidget {
  CredentialData credentialData;

  CredentialMetaDataCard({required this.credentialData, Key? key}) : super(key: key);
  @override
  State<CredentialMetaDataCard> createState() => CredentialMetaDataCardState();
}

class CredentialMetaDataCardState extends State<CredentialMetaDataCard> {
  String issueDate = '';
  String expiryDate = '';
  List<CredentialDisplayClaim> credentialClaimsData = [];
  bool isLoading = false;
  bool verifiedStatus = true;

  @override
  void initState() {
    setState(() {
      isLoading = true;
    });
    super.initState();
    WalletSDKPlugin.parseCredentialDisplayData(widget.credentialData.credentialDisplayData).then((response) {
      setState(() {
        if(response.claims.isNotEmpty){
        credentialClaimsData = response.claims;
        isLoading = false;
      }});
    });
    WalletSDKPlugin.credentialStatusVerifier(widget.credentialData.rawCredential).then((response) => {
          setState(() {
            log('status $response');
            verifiedStatus = response;
          })
        });
  }

  getIssuanceDate() {
    var claimsList = credentialClaimsData;
    for (var claims in claimsList) {
      if (claims.label.contains('Issue Date')) {
        var issueDate = claims.rawValue;
        return issueDate;
      }
    }
    final now = DateTime.now();
    String formatter = DateFormat('yMMMMd').format(now); // 28/03/2020
    return formatter;
  }

  getExpiryDate() {
    var claimsList = credentialClaimsData;
    for (var claims in claimsList) {
      if (claims.label.contains('Expiry Date')) {
        var expiryDate = claims.rawValue;
        return expiryDate;
      }
    }
    return 'Never';
  }

  @override
  Widget build(BuildContext context) {
    return isLoading
        ? const Center(child: LinearProgressIndicator())
        : Container(
            decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: const BorderRadius.only(
                  bottomLeft: Radius.circular(12),
                  bottomRight: Radius.circular(12),
                ),
                boxShadow: [
                  BoxShadow(
                    color: Colors.grey.shade300,
                    blurRadius: 4,
                    offset: const Offset(4, 4),
                  )
                ]),
            padding: const EdgeInsets.fromLTRB(0, 0, 0, 16),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Flexible(
                    child: SizedBox(
                  height: 60,
                  child: ListTile(
                    title: const Text(
                      'Added on',
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.bold,
                        color: Color(0xff190C21),
                      ),
                      textAlign: TextAlign.start,
                    ),
                    subtitle: Text(
                      getIssuanceDate(),
                      style: const TextStyle(
                        fontSize: 14,
                        color: Color(0xff6C6D7C),
                      ),
                      textAlign: TextAlign.start,
                    ),
                    trailing: verifiedStatus
                        ? const Text.rich(
                            TextSpan(
                              children: [
                                WidgetSpan(
                                    child: Icon(
                                  Icons.verified_user_outlined,
                                  color: Colors.lightGreen,
                                  size: 18,
                                )),
                                TextSpan(
                                  text: 'Active',
                                  style: TextStyle(
                                    fontSize: 14,
                                    fontWeight: FontWeight.bold,
                                    color: Colors.lightGreen,
                                  ),
                                ),
                              ],
                            ),
                          )
                        : const Text.rich(
                            TextSpan(
                              children: [
                                WidgetSpan(
                                    child: Icon(
                                  Icons.dangerous_outlined,
                                  color: Colors.redAccent,
                                  size: 18,
                                )),
                                TextSpan(
                                  text: 'Revoked',
                                  style: TextStyle(
                                    fontSize: 14,
                                    fontWeight: FontWeight.bold,
                                    color: Colors.redAccent,
                                  ),
                                ),
                              ],
                            ),
                          ),
                  ),
                )),
                Flexible(
                    child: SizedBox(
                        height: 60,
                        child: ListTile(
                            title: const Text(
                              'Expires on',
                              style: TextStyle(
                                fontSize: 14,
                                fontWeight: FontWeight.bold,
                                color: Color(0xff190C21),
                              ),
                              textAlign: TextAlign.start,
                            ),
                            //TODO need to add fallback and network image url
                            subtitle: Text(
                              getExpiryDate(),
                              style: const TextStyle(
                                fontSize: 14,
                                color: Color(0xff6C6D7C),
                              ),
                              textAlign: TextAlign.start,
                            )))),
              ],
            ));
  }
}
