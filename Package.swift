// swift-tools-version:5.5
// The swift-tools-version declares the minimum version of Swift required to build this package.

/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import PackageDescription

let version = "2.0.0-rc7-swift-pm-swift-pm"
let moduleName = "walletsdk"
let checksum = "06c7ae86812cbf91be701da466783d39604d045298246273275e2c05e95e9be8"

let package = Package(
    name: moduleName,
    products: [
        .library(
            name: moduleName,
            targets: [moduleName]
        )
    ],
    targets: [
        .binaryTarget(
            name: moduleName,
            url: "https://github.com/trustbloc/wallet-sdk/releases/download/\(version)/\(moduleName).xcframework.zip",
            checksum: checksum
        )
    ]
)