import 'dart:convert';
import 'dart:developer';
import 'package:app/widgets/credential_metadata_card.dart';
import 'package:app/widgets/credential_verified_information_view.dart';
import 'package:flutter/material.dart';
import 'package:app/models/credential_data.dart';
import 'package:intl/intl.dart';
import 'package:app/widgets/credential_card.dart';
import 'package:shared_preferences/shared_preferences.dart';

class CredentialDetails extends  StatefulWidget {
  CredentialData credentialData;
  bool isDashboardWidget = true;
  String credentialName;
  List<Object?>? activityLogger;

  CredentialDetails(
      {required this.credentialData, required this.isDashboardWidget, required this.credentialName, this.activityLogger, Key? key})
      : super(key: key);

  @override
  State<CredentialDetails> createState() => CredentialDetailsState();
}
class CredentialDetailsState extends State<CredentialDetails> {
  final ScrollController credDataController = ScrollController();
  final ScrollController rawDataController = ScrollController();
  final ScrollController activityController = ScrollController();
  final ScrollController credentialDIDController = ScrollController();

  bool isSwitched = false;
  String didDoc = '';

  checkDevMode() async {
    SharedPreferences preferences = await SharedPreferences.getInstance();
    setState(() {
      isSwitched = preferences.getBool('devmode') ?? false;
    });
  }
  @override
  void initState() {
    checkDevMode();
    getDidDocument();
    super.initState();
  }

  getCurrentDate() {
    final now = DateTime.now();
    // Todo instead of today's date always it will have to persist the date in the storage in the shared preference.
    String formatter = DateFormat('yMMMMd').format(now);// 28/03/2020
    return  formatter;
  }

  prettifyRawJson(){
    final parsedJson = json.decode(widget.credentialData.credentialDisplayData);
    final prettyString = const JsonEncoder.withIndent('  ').convert(parsedJson);
    return Text(prettyString);
  }

  activityLogDetails() {
    if (widget.activityLogger != null){
      var activities = widget.activityLogger!;
      return listViewWidget(activities.asMap().values);
    }
  }

  getCredentialDID(){
    return Text(widget.credentialData.credentialDID!);
  }

  getDidDocument() async {
    final Future<SharedPreferences> prefs = SharedPreferences.getInstance();
    final SharedPreferences pref = await prefs;
    var userDIDDoc = pref.getString("userDIDDoc");
    final parsedJson = json.decode(userDIDDoc!);
    didDoc = const JsonEncoder.withIndent('  ').convert(parsedJson);
  }

  Widget listViewWidget(Iterable<Object?> activitiesValue) {
    return ListView.builder(
        itemCount: activitiesValue.length,
        scrollDirection: Axis.vertical,
        controller: credDataController,
        shrinkWrap: true,
        itemBuilder: (context, index)
    {
      var value = const JsonEncoder.withIndent('  ').convert(activitiesValue.toList().elementAt(index));
      return Row(
        children: [
          const Divider(
            thickness: 2,
            color: Color(0xffDBD7DC),
          ),
          Expanded(
            child: ListTile(
              title: const Text(
                  "",
                  style: TextStyle(
                      fontSize: 14, fontFamily: 'SF Pro', fontWeight: FontWeight.w400, color: Color(0xff6C6D7C))
              ),
              subtitle: Text(
                value.toString(),
                style: const TextStyle(
                    fontSize: 16,
                    color: Color(0xff190C21),
                    fontFamily: 'SF Pro',
                    fontWeight: FontWeight.normal),
              ),
            ),
          ),
        ]
      );
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Credential Details', textAlign: TextAlign.center, style: TextStyle(fontSize: 18, fontStyle: FontStyle.normal, fontWeight: FontWeight.w700, fontFamily: 'SF Pro')),
      backgroundColor: const Color(0xff261131),
      ),
      body: SingleChildScrollView(
        child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
          const SizedBox(height: 24.0),
          DefaultTabController(
              length: 4, // length of tabs
              initialIndex: 0,
              child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                TabBar(
                  labelColor: const Color(0xff190C21),
                  labelStyle: const TextStyle(fontWeight: FontWeight.w700, fontSize: 14),
                  unselectedLabelColor: const Color(0xff6C6D7C),
                  indicatorColor: const Color(0xff8A35B7),
                  padding: const EdgeInsets.fromLTRB(24, 0, 0, 0),
                  tabs: [
                    const Tab(
                      child: SizedBox(
                        width: 50,
                        child: Text(
                          "Details",
                          textAlign: TextAlign.start,
                        ),
                      ),
                    ),
                    Visibility(
                      visible: isSwitched,
                      child:const Tab(
                      child: SizedBox(
                        width: 150,
                        child: Text(
                          "Raw",
                          textAlign: TextAlign.start,
                        ),
                      ),
                     )
                    ),
                    const Tab(
                      child: SizedBox(
                      width: 60,
                      child: Text(
                        "Activity",
                        textAlign: TextAlign.start,
                        textScaleFactor:1.0,
                      ),
                    ),
                    ),
                   Visibility(
                    visible: isSwitched!=null?isSwitched!:false,
                     child: const Tab(  child: SizedBox(
                      width: 50,
                      child: Text(
                        "DID",
                        textAlign: TextAlign.start,
                      ),
                     ),
                    ),
                   ),
                  ],
                ),
                Container(
                    height: MediaQuery.of(context).size.height*0.8,
                    padding: const EdgeInsets.all(24),
                    alignment: Alignment.center,
                    child:  TabBarView(children: <Widget>[
                      SingleChildScrollView(
                      child: Column(
                        children: [
                      Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                          CredentialCard(credentialData: widget.credentialData, isDashboardWidget: false, isDetailArrowRequired: false),
                          widget.isDashboardWidget?
                          CredentialMetaDataCard(credentialData: widget.credentialData): Container(),
                          CredentialVerifiedInformation(credentialData: widget.credentialData, height: MediaQuery.of(context).size.height*0.42)
                      ],
                    )
                 ])),
                      SizedBox(
                          height: 450,
                          child: Scrollbar(
                            thumbVisibility: true,
                            controller: rawDataController,
                            child: ListView.builder(
                                controller: rawDataController,
                                itemCount: 1,
                                itemBuilder: (BuildContext context, int index) {
                                  return Padding(
                                    padding: const EdgeInsets.all(8.0),
                                    child: prettifyRawJson(),
                                  );
                                }),
                          )),
                      SizedBox(
                          height: 450,
                          child: Scrollbar(
                            thumbVisibility: true,
                            controller: activityController,
                            child: ListView.builder(
                                controller: activityController,
                                itemCount: 1,
                                itemBuilder: (BuildContext context, int index) {
                                  return Padding(
                                    padding: const EdgeInsets.all(8.0),
                                    child: activityLogDetails()
                                  );
                                }),
                          )),
                      SizedBox(
                          height: 450,
                          child: Scrollbar(
                            thumbVisibility: true,
                            controller: credentialDIDController,
                            child: ListView.builder(
                                controller: credentialDIDController,
                                itemCount: 1,
                                itemBuilder: (BuildContext context, int index) {
                                  return Column(
                                    children: [
                                      const Padding(
                                          padding:EdgeInsets.all(8.0),
                                          child:  Text(
                                            "DID",
                                            style: TextStyle(
                                              fontSize: 16,
                                              color: Color(0xff190C21),
                                              fontFamily: 'SF Pro',
                                              fontWeight: FontWeight.bold),)
                                      ),
                                        Padding(
                                          padding:const EdgeInsets.all(8.0),
                                            child: getCredentialDID()
                                        ),
                                      const Padding(
                                          padding:EdgeInsets.all(8.0),
                                          child:  Text(
                                            "DidDocument",
                                            style: TextStyle(
                                              fontSize: 16,
                                              color: Color(0xff190C21),
                                              fontFamily: 'SF Pro',
                                              fontWeight: FontWeight.bold),)
                                      ),
                                      Padding(
                                          padding:const EdgeInsets.all(8.0),
                                          child: Text(didDoc, style: const TextStyle(
                                              fontSize: 12,
                                              color: Color(0xff190C21),
                                              fontFamily: 'SF Pro',
                                              fontWeight: FontWeight.normal),)
                                       ),
                                      ],

                                  );
                                }),
                          ))
                    ])
                )]),
          ),
        ])
      ),
    );
  }
}