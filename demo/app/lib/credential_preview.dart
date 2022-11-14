import 'package:app/dashboard.dart';
import 'package:flutter/material.dart';

void main() => runApp(const CreatePreview());

class CreatePreview extends StatelessWidget {
  const CreatePreview({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      theme: ThemeData(
          primarySwatch: Colors.deepPurple,
          appBarTheme: const AppBarTheme(
            centerTitle: true,
            toolbarHeight: 70,
            titleTextStyle: TextStyle(fontSize: 28),
            iconTheme: IconThemeData(color: Colors.white),
            foregroundColor: Colors.white, //<-- SEE HERE
          )),

      home: const CreatePreviewStatefulWidget(),
    );
  }
}

class CreatePreviewStatefulWidget extends StatefulWidget {
  const CreatePreviewStatefulWidget({super.key});

  @override
  State<CreatePreviewStatefulWidget> createState() => CredentialPreviewState();
}

class CredentialPreviewState extends State<CreatePreviewStatefulWidget> {
  int _selectedIndex = 0;
  static const TextStyle optionStyle =
  TextStyle(fontSize: 28, fontWeight: FontWeight.bold);
  static const List<Widget> _widgetOptions = <Widget>[
    Text(
      'Operation is Cancelled',
      style: optionStyle,
    ),
    // TODO Move to save credential page
    Text(
      'Credentials',
      style: optionStyle,
    ),
  ];

  void _onItemTapped(int index) {
    setState(() {
      _selectedIndex = index;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Credential Preview'),
      ),
        body: Center(
            child:  Column (
              children: [
                const SizedBox(height: 50),
                const Text(
                  'Add a Verified ID',
                  textAlign: TextAlign.center,
                  style: TextStyle(fontSize: 22, color: Colors.black),
                ),
                const SizedBox(height: 30),
                MaterialButton(
                  onPressed: () { },
                  padding: const EdgeInsets.all(0.0),
                  child: Ink(
                    decoration: const BoxDecoration(
                      gradient: LinearGradient(
                        colors: <Color>[Colors.lightBlueAccent, Colors.blueAccent],
                      ),
                    ),
                    child: Container(
                      constraints: const BoxConstraints(maxWidth: 350, minHeight: 120.0), // min sizes for Material buttons
                      alignment: Alignment.center,
                      child: const Text(
                        'Permanent Resident Card',
                        textAlign: TextAlign.center,
                      ),
                    ),
                  ),
                ),
              ],
            )
        ),
      bottomNavigationBar: BottomNavigationBar(
        items: const <BottomNavigationBarItem>[
          BottomNavigationBarItem(
            icon: Icon(Icons.cancel),
            label: 'Cancel',
          ),
          BottomNavigationBarItem(
            backgroundColor: Colors.green,
            icon: Icon(Icons.add),
            label: 'Add',
          ),
        ],
        currentIndex: _selectedIndex,
        selectedItemColor: Colors.amber[800],
        onTap: _onItemTapped,
      ),
    );
  }
}