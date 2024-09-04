// swift-tools-version:5.5
// The swift-tools-version declares the minimum version of Swift required to build this package.

/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import PackageDescription

let version = "2.2.0-swift-pm"
let moduleName = "walletsdk"
let checksum = "c2d0655919a10501ab1315c1518b3286267819a8cd3b70d21224ab405842b8d5"

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