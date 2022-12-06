import 'package:app/views/credential_list.dart';
import 'package:flutter/material.dart';
import 'scanner.dart';

class Dashboard extends StatefulWidget {
  final String? user;
  const Dashboard({super.key, this.user});

  @override
  State<Dashboard> createState() => _DashboardState();
}

class _DashboardState extends State<Dashboard> {
  int _selectedIndex = 0;
  static String userIDLoggedIn = '';
  @override
  void initState() {
    super.initState();
    userIDLoggedIn = widget.user!;
  }
  static const TextStyle optionStyle =
  TextStyle(fontSize: 28, fontWeight: FontWeight.bold);
  static final _widgetOptions = <Widget>[
   CredentialList(title: 'Saved Credentials List', user: userIDLoggedIn),
    const QRScanner(),
   const Text(
      'Preferences',
      style: optionStyle,
    ),
  ];

  void _onItemTapped(int index) {
    setState(() {
      _selectedIndex = index;
      if (_selectedIndex ==0){

      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('Welcome $userIDLoggedIn'),
      ),
      body: Center(
        child: _widgetOptions.elementAt(_selectedIndex),
      ),
      bottomNavigationBar: BottomNavigationBar(
        items: const <BottomNavigationBarItem>[
          BottomNavigationBarItem(
            icon: Icon(Icons.credit_card),
            label: 'Credentials',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.camera_enhance_outlined),
            label: 'Scan QR Code',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.settings),
            label: 'Settings',
          ),
        ],
        currentIndex: _selectedIndex,
        selectedItemColor: Colors.amber[800],
        onTap: _onItemTapped,
      ),
    );
  }
}