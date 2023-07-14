/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

var agent;
var openID4CIInteraction;
var createdDID;

async function jsInitSDK(didResolverURI) {
    agent = await Agent({assetsPath: "", didResolverURI:didResolverURI});
}

async function jsCreateDID(didMethod, keyType) {
    const did = await agent.createDID({
        didMethod: didMethod,
        keyType: keyType
    })

    createdDID = did;

    return {
        id: did.id,
        content: did.content
    }
}

async function jsCreateOpenID4CIInteraction(initiateIssuanceURI) {
    openID4CIInteraction = await agent.createOpenID4CIIssuerInitiatedInteraction({
        initiateIssuanceURI: initiateIssuanceURI
    })

    let userPINRequired = (await openID4CIInteraction.preAuthorizedCodeGrantParams()).userPINRequired;

    return {
        userPINRequired: userPINRequired
    };
}

async function jsRequestCredentialWithPreAuth(userPinEntered) {
    var creds = await openID4CIInteraction.requestCredentialWithPreAuth({
        pin: userPinEntered,
        didDoc: createdDID
    });

    return creds[0];
}

function jsIssuerURI() {
    return openID4CIInteraction.issuerURI()
}

async function jsResolveDisplayData(issuerURI, credentials) {
    let data = await agent.resolveDisplayData({
        issuerURI: issuerURI,
        credentials:credentials
    })

    return data;
}

async function jsGetCredentialID(credential) {
    return await agent.getCredentialID({
        credential: credential
    });
}