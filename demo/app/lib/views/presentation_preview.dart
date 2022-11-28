import 'package:app/views/dashboard.dart';
import 'package:flutter/material.dart';

class PresentationPreview extends StatefulWidget {
  const PresentationPreview({super.key});

  @override
  State<PresentationPreview> createState() => PresentationPreviewState();
}

class PresentationPreviewState extends State<PresentationPreview> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Presentation Preview'),
      ),
      body: Center(
          child:  Column (
            children: [
              const SizedBox(height: 50),
              const Text(
                'Share with Trustbloc Demo App',
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
                    // TODO Fetch dynamically saved credential from the store
                    // If no credential found show Credential Not found error
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
      floatingActionButton: SizedBox(
        width: double.infinity,
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 8),
          child: Row(
            children: [
              Expanded(
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(primary: Colors.red),
                  onPressed: () async {
                    _navigateToDashboard();
                  },
                  child: const Text("Cancel"),
                ),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: ElevatedButton(
                  onPressed: () async {
                    // TODO Add logic to present the presentation on the click of the share button
                  },
                  child: const Text("Share"),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
  _navigateToDashboard() async {
    Navigator.push(context, MaterialPageRoute(builder: (context) => const Dashboard()));
  }
}