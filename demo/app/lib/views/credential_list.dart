import 'dart:developer';

import 'package:app/services/storage_service.dart';
import 'package:app/widgets/credential_card.dart';
import 'package:flutter/material.dart';
import 'package:app/models/store_credential_data.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../models/credential_data_object.dart';

class CredentialList extends StatefulWidget {
  final String title;
  const CredentialList({Key? key, required this.title}) : super(key: key);

  @override
  State<CredentialList> createState() => _CredentialListState();
}

class _CredentialListState extends State<CredentialList> {
  final StorageService _storageService = StorageService();
 // late List<StorageItem> _credentialList;
  late List<CredentialDataObject> _credentialList;
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
      appBar: AppBar(
        title: Text(widget.title),
      ),
      body: Center(
        child: _loading
            ? const CircularProgressIndicator()
            : _credentialList.isEmpty
            ? const Text("Add data in secure storage to display here.")
            : ListView.builder(
            itemCount: _credentialList.length,
            padding: const EdgeInsets.symmetric(horizontal: 8),
            itemBuilder: (_, index) {
              return Dismissible(
                key: Key(_credentialList[index].toString()),
                child: CredentialCard(item: _credentialList[index].value),
                onDismissed: (direction) async {
                  await _storageService.deleteData(_credentialList[index])
                      .then((value) => _credentialList.removeAt(index));
                  initList();
                },
              );
            }),
      ),
      floatingActionButtonLocation: FloatingActionButtonLocation.centerDocked,
    );
  }
}