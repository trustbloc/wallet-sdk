import 'dart:developer';

import 'package:app/main.dart';
import 'package:app/services/storage_service.dart';
import 'package:app/widgets/credential_card.dart';
import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:app/models/credential_data_object.dart';
import 'package:app/widgets/common_logo_appbar.dart';

class CredentialList extends StatefulWidget {
  const CredentialList({super.key});


  @override
  State<CredentialList> createState() => _CredentialListState();
}

class _CredentialListState extends State<CredentialList> {
  final StorageService _storageService = StorageService();
  late List<CredentialDataObject> _credentialList;
  late List<Object?> activityLogger;
  late List<Object?> resolveCredDisplay;
  String? credentialDisplayData;
  bool _loading = true;
  static String? username = '';
  @override
  void initState() {
    super.initState();
    initList();
  }

  void initList() async {
    SharedPreferences prefs = await SharedPreferences.getInstance();
    final SharedPreferences p = prefs;
    username =  p.getString("userLoggedIn");
    log("list - $username");
      _credentialList = await _storageService.retrieveCredentials(username!);
      if (_credentialList.isNotEmpty) {
        for (var cred in _credentialList) {
          var credID = await WalletSDKPlugin.getCredID([cred.value.rawCredential]);
          var activities = await _storageService.retrieveActivities(credID!);
          activityLogger = await WalletSDKPlugin.parseActivities(activities);
        }
      }
    if (_credentialList.isEmpty) {
      _loading = true;
      _credentialList.clear();
    }
    _loading = false;
    setState(() {});
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: CustomLogoAppBar(),
      body: Center(
        child: Stack(
          children: <Widget>[
            Container(
                padding: const EdgeInsets.fromLTRB(36, 24, 24, 24),
                alignment: Alignment.topLeft,
                child: const Text(
                  'Credentials',
                  style: TextStyle(color: Color(0xff190C21), fontWeight: FontWeight.bold, fontSize: 16),
                )),
            Container(
                padding: const EdgeInsets.fromLTRB(24, 40, 16, 24),
                alignment: Alignment.center,
                child: _loading
                  ? const CircularProgressIndicator()
                  : _credentialList.isEmpty
                  ? const Text("No credentials found")
                  : ListView.builder(
                  itemCount: _credentialList.length,
                  padding: const EdgeInsets.symmetric(horizontal: 8),
                  itemBuilder: (_, index) {
                    return Dismissible(
                      key: Key(_credentialList[index].toString()),
                      direction: DismissDirection.endToStart,
                      onDismissed: (direction) async {
                        if (direction == DismissDirection.endToStart) {
                          await _storageService.deleteData(_credentialList[index])
                              .then((value) => _credentialList.removeAt(index));
                          initList();
                        }
                      },
                      background: const ColoredBox(
                        color: Colors.white,
                        child: Align(
                          alignment: Alignment.centerRight,
                          child: Padding(
                            padding: EdgeInsets.all(0),
                            child: Icon(Icons.delete, color: Colors.red),
                          ),
                        ),
                      ),
                      confirmDismiss: (DismissDirection direction) async {
                        final confirmed = await showDialog<bool>(
                          context: context,
                          builder: (context) {
                            return AlertDialog(
                              title: const Text('Are you sure you want to delete?',  style: TextStyle(fontSize: 12)),
                              actions: [
                                TextButton(
                                  onPressed: () => Navigator.pop(context, false),
                                  child: const Text('No'),
                                ),
                                TextButton(
                                  onPressed: () => Navigator.pop(context, true),
                                  child: const Text('Yes'),
                                )
                              ],
                            );
                          },
                        );
                        log('Deletion confirmed: $confirmed');
                        return confirmed;
                      },
                      child: CredentialCard(credentialData: _credentialList[index].value, activityLogger: activityLogger, isDashboardWidget: true, isDetailArrowRequired: false,),
                    );
                  }),
            ),
          ],
        ),
      ),
      floatingActionButtonLocation: FloatingActionButtonLocation.centerDocked,
    );
  }
}