/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

export 'wallet_sdk_mobile.dart' // By default
if (dart.library.js) 'wallet_sdk_js.dart'
if (dart.library.io) 'wallet_sdk_mobile.dart';
