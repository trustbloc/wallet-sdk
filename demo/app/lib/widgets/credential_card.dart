import 'dart:convert';
import 'dart:developer';
import 'package:app/models/credential_data.dart';
import 'package:app/views/credential_details.dart';
import 'package:flutter/material.dart';
import '../main.dart';
import 'credential_metadata_card.dart';
import 'credential_verified_information_view.dart';

class CredentialCard extends StatefulWidget {
  CredentialData credentialData;
  bool isDashboardWidget;
  List<Object?>? activityLogger;
  bool isDetailArrowRequired;

  CredentialCard({required this.credentialData, required this.isDashboardWidget, this.activityLogger, required this.isDetailArrowRequired,  Key? key}) : super(key: key);
  @override
  State<CredentialCard> createState() => _CredentialCardState();
}

class _CredentialCardState extends State<CredentialCard> {
  bool showWidget = false;
  String? credentialDisplayName;
  late String logoURL;
  String? backgroundColor;
  String? textColor;

  @override
  void initState() {
    WalletSDKPlugin.resolveCredDisplayRendering(widget.credentialData.credentialDisplayData).then(
            (response) {
          setState(() {
            var credentialDisplayEncodeData = json.encode(response);
            List<dynamic> responseJson = json.decode(credentialDisplayEncodeData);
            credentialDisplayName = responseJson.first['overviewName'];
            logoURL = responseJson.first['logo'];
            backgroundColor ='0xff${responseJson.first['backgroundColor']!.toString().replaceAll('#', '')}';
            textColor = '0xff${responseJson.first['textColor'].toString().replaceAll('#', '')}';
          });
        });
    super.initState();
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(0, 16, 0, 16),
      child: Column(
           children: [
             Container(
               height: 80,
               alignment: Alignment.center,
               decoration: BoxDecoration(
                   color: backgroundColor!.isNotEmpty ? Color(int.parse(backgroundColor!)) : Colors.white,
                   borderRadius: BorderRadius.circular(12),
                   boxShadow: [
                     BoxShadow(
                         offset: const Offset(3, 3),
                         color: Colors.grey.shade300,
                         blurRadius: 5)
                   ]),
                 child: ListTile(
                   title: Text(
                     credentialDisplayName!,
                     style: TextStyle(
                       fontSize: 14,
                       fontWeight: FontWeight.bold,
                       color: textColor!.isNotEmpty ? Color(int.parse(textColor!)) : const Color(0xff190C21),
                     ),
                     textAlign: TextAlign.start,
                   ),
                   leading: FadeInImage(
                     image: NetworkImage(logoURL),
                     placeholder: AssetImage(logoURL),
                     imageErrorBuilder:(context, error, stackTrace) {
                       return Image.asset('lib/assets/images/genericCredential.png',
                           fit: BoxFit.fitWidth
                       );
                     },
                     fit: BoxFit.fitWidth,
                   ),
                   trailing: widget.isDetailArrowRequired == false ? IconButton(
                     icon: const Icon(Icons.arrow_circle_right, size: 32, color: Color(0xffB6B7C7)),
                     onPressed: () async {
                       Navigator.push(
                         context,
                         MaterialPageRoute(builder: (context) => CredentialDetails(credentialData: widget.credentialData,credentialName: credentialDisplayName!, isDashboardWidget: widget.isDashboardWidget,  activityLogger: widget.activityLogger,)),
                       );
                     },
                   ): IconButton(
                     icon: const Icon(Icons.expand_circle_down_sharp, size: 32, color: Color(0xffB6B7C7)),
                     onPressed: () async {
                       setState(() {
                         showWidget = !showWidget;
                       });
                     },
                   ),
                 )
             ),
             showWidget ? Column(
               children: [
                 CredentialMetaDataCard(credentialData: widget.credentialData),
                 CredentialVerifiedInformation(credentialData: widget.credentialData, height: MediaQuery.of(context).size.height*0.38,),
               ],
             ): Container(),
           ],

      ),
    );
  }
}