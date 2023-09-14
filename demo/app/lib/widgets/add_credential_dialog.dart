import 'package:flutter/material.dart';
import 'package:app/models/store_credential_data.dart';

class AddDataDialog extends StatelessWidget {
  AddDataDialog({Key? key}) : super(key: key);

  final TextEditingController _keyController = TextEditingController();
  final TextEditingController _valueController = TextEditingController();

  @override
  Widget build(BuildContext context) {
    return Dialog(
      child: Padding(
        padding: const EdgeInsets.all(8.0),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text('Do you want to save credential?'),
            SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                    onPressed: () {
                      final StorageItem storageItem = StorageItem(_keyController.text, _valueController.text);
                      Navigator.of(context).pop(storageItem);
                    },
                    child: const Text('Save')))
          ],
        ),
      ),
    );
  }
}
