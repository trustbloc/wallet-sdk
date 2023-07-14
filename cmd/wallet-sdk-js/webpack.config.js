/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

const path = require("path")

const srcDir = path.join(__dirname, "src/")
const buildDir = path.join(__dirname, "dist/")

const PATHS = {
    srcDir: srcDir,
    buildDir: buildDir
}

module.exports = {
    entry: path.join(PATHS.srcDir, "index.js"),
    target: 'web',
    output: {
        path: PATHS.buildDir,
        publicPath: "dist",
        libraryTarget: "umd",
        filename: 'agent.js',
        library: 'Agent'
    },
}
