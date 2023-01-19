import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

class CredentialMetaDataCard extends StatelessWidget {
  const CredentialMetaDataCard({super.key});


  getCurrentDate() {
    final now = DateTime.now();
    String formatter = DateFormat('yMMMMd').format(now);// 28/03/2020
    return  formatter;
  }

  @override
  Widget build(BuildContext context) {
    return Container(
        decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: const BorderRadius.only(
              bottomLeft: Radius.circular(12),
              bottomRight: Radius.circular(12),
            ),
            boxShadow: [
              BoxShadow(
                color: Colors.grey.shade300,
                blurRadius: 4,
                offset: const Offset(4, 4),
              )
            ]),
        padding: const EdgeInsets.fromLTRB(0, 0, 0, 16),
      child: Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Flexible(
            child: SizedBox(
              height: 60,
              child: ListTile(
                  title: const Text(
                    'Added on',
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                      color: Color(0xff190C21),
                    ),
                    textAlign: TextAlign.start,
                  ),
                  //TODO need to add fallback and network image url
                  subtitle: Text(
                    getCurrentDate(),
                    style: const TextStyle(
                      fontSize: 14,
                      color: Color(0xff6C6D7C),
                    ),
                    textAlign: TextAlign.start,
                  )
              ),
            )
        ),
        const Flexible(
            child: SizedBox(
                height: 60,
                child: ListTile(
                    title: Text(
                      'Expires on',
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.bold,
                        color: Color(0xff190C21),
                      ),
                      textAlign: TextAlign.start,
                    ),
                    //TODO need to add fallback and network image url
                    subtitle: Text(
                      'Never',
                      style: TextStyle(
                        fontSize: 14,
                        color: Color(0xff6C6D7C),
                      ),
                      textAlign: TextAlign.start,
                    )
                )
            )
        ),
        const Flexible(
            child: SizedBox(
                height: 60,
                child: ListTile(
                    title: Text(
                      'Last used',
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.bold,
                        color: Color(0xff190C21),
                      ),
                      textAlign: TextAlign.start,
                    ),
                    //TODO need to add fallback and network image url
                    subtitle: Text(
                      'Never',
                      style: TextStyle(
                        fontSize: 14,
                        color: Color(0xff6C6D7C),
                      ),
                      textAlign: TextAlign.start,
                    )
                )
            )
        )
      ],
      )
    );
  }
}