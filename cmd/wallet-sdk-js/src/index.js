/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

'use strict'

export default class Agent {
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
            keyType: opts.keyType
        });
    };

    async createOpenID4CIIssuerInitiatedInteraction(opts) {
        return await this.goAgent.createOpenID4CIIssuerInitiatedInteraction({
            initiateIssuanceURI: opts.initiateIssuanceURI,
        })
    };

    async createOpenID4VPInteraction(opts) {
        return await this.goAgent.createOpenID4VPInteraction({
            authorizationRequest: opts.authorizationRequest
        })
    }

    async getSubmissionRequirements(opts) {
        return await this.goAgent.getSubmissionRequirements({
            query: opts.query,
            credentials: opts.credentials
        })
    }

    async resolveDisplayData(opts) {
        return await this.goAgent.resolveDisplayData({
            issuerURI: opts.issuerURI,
            credentials: opts.credentials
        })
    };

    async parseResolvedDisplayData(opts) {
        return await this.goAgent.parseResolvedDisplayData({
            resolvedCredentialDisplayData: opts.resolvedCredentialDisplayData,
        })
    };

    async verifyCredentialsStatus(opts) {
        return await this.goAgent.verifyCredentialsStatus({
            credential: opts.credential
        })
    }

    async validateLinkedDomains(opts) {
        return await this.goAgent.validateLinkedDomains({
            did: opts.did
        })
    }

    async getCredentialID(opts) {
        return await this.goAgent.getCredentialID({
            credential: opts.credential
        })
    }

    stop() {
        this.goAgent.stopAssembly()
    }
}
