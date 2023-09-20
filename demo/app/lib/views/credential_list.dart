/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'dart:developer';

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
    username = p.getString('userLoggedIn');
    log('list - $username');
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
                      ? const Text('No credentials found')
                      : ListView.builder(
                          itemCount: _credentialList.length,
                          padding: const EdgeInsets.symmetric(horizontal: 8),
                          itemBuilder: (_, index) {
                            return CredentialCard(
                                credentialData: _credentialList[index].value,
                                isDashboardWidget: true,
                                delete: () async {
                                  await _storageService
                                      .deleteData(_credentialList[index])
                                      .then((value) => _credentialList.removeAt(index));
                                  setState(() {});
                                },
                                isDetailArrowRequired: false);
                          }),
            ),
          ],
        ),
      ),
    );
  }
}
