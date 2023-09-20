/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import 'package:app/views/credential_list.dart';
import 'package:app/views/wallet_initiated_connect.dart';
import 'package:app/views/settings.dart';
import 'package:flutter/material.dart';
import 'scanner.dart';

class Dashboard extends StatefulWidget {
  const Dashboard({super.key});

  @override
  State<Dashboard> createState() => _DashboardState();
}

class _DashboardState extends State<Dashboard> {
  int _selectedIndex = 0;

  static final _widgetOptions = <Widget>[
    const CredentialList(),
    const QRScanner(),
    const ConnectIssuerList(),
    const Settings(),
  ];

  void _onItemTapped(int index) {
    setState(() {
      _selectedIndex = index;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: _widgetOptions.elementAt(_selectedIndex),
      ),
      bottomNavigationBar: BottomNavigationBar(
        type: BottomNavigationBarType.fixed,
        items: const <BottomNavigationBarItem>[
          BottomNavigationBarItem(
            icon: Icon(Icons.home),
            label: 'Home',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.qr_code),
            label: 'Scan QR',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.list),
            label: 'Connect',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.settings),
            label: 'Settings',
          ),
        ],
        currentIndex: _selectedIndex,
        selectedLabelStyle: const TextStyle(fontFamily: 'SF Pro', fontSize: 12, fontWeight: FontWeight.bold),
        selectedItemColor: const Color(0xffEC857C),
        onTap: _onItemTapped,
      ),
    );
  }
}
