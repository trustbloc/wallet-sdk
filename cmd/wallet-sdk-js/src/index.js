/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

'use strict'


const Agent = async function (opts) {
    const go = new Go();
    const assembly = await WebAssembly.instantiateStreaming(fetch(opts.assetsPath + "/wallet-sdk.wasm"), go.importObject);
    go.run(assembly.instance);

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
        resolveDisplayData: async function (opts) {
            return await goAgent.resolveDisplayData({
                issuerURI: opts.issuerURI,
                credentials: opts.credentials
            })
        },
        parseResolvedDisplayData: async function(opts) {
            return await  goAgent.parseResolvedDisplayData({
                resolvedCredentialDisplayData: opts.resolvedCredentialDisplayData,
            })
        },
        getCredentialID: async function (opts) {
            return await goAgent.getCredentialID({
                credential: opts.credential
            })
        },
        stop: function () {
            goAgent.stopAssembly()
        }
    }
}