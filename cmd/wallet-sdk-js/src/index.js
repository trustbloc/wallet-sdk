/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

'use strict'

export default class WalletSDKAgent {
    constructor(opts) {
        this.go = new Go();
        this.opts = opts;
    }

    async initialize() {
        const assembly = await WebAssembly.instantiateStreaming(fetch(this.opts.assetsPath + "/wallet-sdk.wasm"), this.go.importObject);
        this.go.run(assembly.instance);
        this.goAgent = window.__agentInteropObject;
        await this.goAgent.initAgent({didResolverURI: this.opts.didResolverURI, kmsDatabase: this.opts.kmsDatabase})
    }

    async createDID(opts) {
        return await this.goAgent.createDID({
            didMethod: opts.didMethod,
            keyType: opts.keyType,
            verificationType: opts.verificationType
        });
    };

    async createOpenID4CIIssuerInitiatedInteraction(opts) {
        return await this.goAgent.createOpenID4CIIssuerInitiatedInteraction({
            initiateIssuanceURI: opts.initiateIssuanceURI,
        })
    };

    async resolveDisplayData(opts) {
        return await this.goAgent.resolveDisplayData({
            issuerURI: opts.issuerURI,
            credentials: opts.credentials
        })
    };

    async getCredentialID(opts) {
        return await this.goAgent.getCredentialID({
            credential: opts.credential
        })
    }

    stop() {
        this.goAgent.stopAssembly()
    }
}
