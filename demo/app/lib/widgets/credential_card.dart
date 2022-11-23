import 'package:flutter/material.dart';
import 'package:app/models/store_credential_data.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class CredentialCard extends StatefulWidget {
  StorageItem item;

  CredentialCard({required this.item, Key? key}) : super(key: key);

  @override
  State<CredentialCard> createState() => _CredentialCardState();
}

class _CredentialCardState extends State<CredentialCard> {
  bool _visibility = false;
  final _storage = const FlutterSecureStorage();

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Container(
          decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(16),
              boxShadow: [
                BoxShadow(
                    offset: const Offset(3, 3),
                    color: Colors.grey.shade300,
                    blurRadius: 5)
              ]),
          child: ListTile(
            onLongPress: () {
              setState(() {
                _visibility = !_visibility;
              });
            },
            title: Text(
              widget.item.key,
              style: const TextStyle(
                fontSize: 18,
              ),
            ),
            subtitle: Visibility(
              visible: _visibility,
              child: Text(
                widget.item.value,
                style: const TextStyle(fontWeight: FontWeight.bold),
              ),
            ),
            leading: const Icon(Icons.security, size: 30),
            trailing: IconButton(
              icon: const Icon(Icons.edit),
              onPressed: () async {

              },
            ),
          )),
    );
  }
}
