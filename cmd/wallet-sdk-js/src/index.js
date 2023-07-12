/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

'use strict'

import "wasm_exec.js"

const Agent = async function (opts) {
    const go = new Go();
    const assembly = await WebAssembly.instantiateStreaming(fetch(opts.assetsPath + "/wallet-sdk.wasm"), go.importObject);
    await go.run(assembly.instance);

    const goAgent = window.__agentInteropObject;

    await goAgent.initAgent({didResolverURI: opts.didResolverURI})

    return {
        createDID: async function (opts) {
            return await goAgent.createDID({
                didMethod: opts.didMethod,
                keyType: opts.keyType,
                verificationType: opts.verificationType
            });
        },
        createOpenID4CIIssuerInitiatedInteraction: async function (opts) {
            return await goAgent.createOpenID4CIIssuerInitiatedInteraction({
                initiateIssuanceURI: opts.initiateIssuanceURI,
            })
        },
        stop: function () {
            goAgent.stopAssembly()
        }
    }
}