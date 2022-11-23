import 'package:app/services/storage_service.dart';
import 'package:app/widgets/credential_card.dart';
import 'package:flutter/material.dart';
import 'package:app/models/store_credential_data.dart';

class CredentialList extends StatefulWidget {
  final String title;

  const CredentialList({Key? key, required this.title}) : super(key: key);

  @override
  State<CredentialList> createState() => _CredentialListState();
}

class _CredentialListState extends State<CredentialList> {
  final StorageService _storageService = StorageService();
  late List<StorageItem> _items;
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    initList();
  }

  void initList() async {
    _items = await _storageService.readAllSecureData();
    print("dashboard credential_list, what are all item $_items");
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
            : _items.isEmpty
            ? const Text("Add data in secure storage to display here.")
            : ListView.builder(
            itemCount: _items.length,
            padding: const EdgeInsets.symmetric(horizontal: 8),
            itemBuilder: (_, index) {
              return Dismissible(
                key: Key(_items[index].toString()),
                child: CredentialCard(item: _items[index]),
                onDismissed: (direction) async {
                  await _storageService.deleteSecureData(_items[index])
                      .then((value) => _items.removeAt(index));
                  initList();
                },
              );
            }),
      ),
      floatingActionButtonLocation: FloatingActionButtonLocation.centerDocked,
    );
  }
}