/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

let agent;
let openID4CIInteraction;
let openID4VPInteraction;
let createdDID;

async function jsInitSDK(didResolverURI) {
    const kmsDatabase = await CreateDB("test")
    agent = new Agent.default({assetsPath: "", didResolverURI:didResolverURI, kmsDatabase: kmsDatabase});
    await agent.initialize();
}

async function jsCreateDID(didMethod, keyType) {
    const did = await agent.createDID({
        didMethod: didMethod,
        keyType: keyType
    })

    createdDID = did;

    return {
        did: did.id,
        didDoc: did.content
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

async function jsParseResolvedDisplayData(resolvedCredentialDisplayData) {
    return await agent.parseResolvedDisplayData({
        resolvedCredentialDisplayData: resolvedCredentialDisplayData
    });
}

async function jsCreateOpenID4VPInteraction(authorizationRequest) {
    openID4VPInteraction = await agent.createOpenID4VPInteraction({
        authorizationRequest: authorizationRequest
    })
}

async function jsGetSubmissionRequirements(credentials) {
    let query = await openID4VPInteraction.getQuery();

    let result = await agent.getSubmissionRequirements({
        query: query,
        credentials: credentials,
    });

    return result;
}

async function jsPresentCredential(credentials) {
    return await openID4VPInteraction.presentCredential({
        credentials: credentials
    })
}

async function jsVerifierDisplayData() {
    return await openID4VPInteraction.verifierDisplayData()
}


async function jsVerifyCredentialsStatus(credential) {
    return await agent.verifyCredentialsStatus({
        credential: credential
    })
}

async function jsWellKnownDidConfig(issuerID) {
    return await agent.validateLinkedDomains({
        did: issuerID
    })
}

async function jsGetIssuerMetadata() {
    const issuerMetadata = await openID4CIInteraction.issuerMetadata()
    const result = {
        ...issuerMetadata,
        supportedCredentials: JSON.parse(issuerMetadata.supportedCredentials),
        localizedIssuerDisplays: JSON.parse(issuerMetadata.localizedIssuerDisplays)
    };
    console.log("JS Issuer Metadata:", result);
    return result;
}

// KMS database implementation
function CreateDB(dbName) {
  const keystoreTable = "keyStore";

  return new Promise(function(resolve, reject) {
    let dbReq = indexedDB.open(dbName, 2);

    dbReq.onupgradeneeded = function(event) {
      const db = event.target.result;

      if (!db.objectStoreNames.contains(keystoreTable)) {
        db.createObjectStore(keystoreTable, {keyPath: "key"});
      }
    }

    dbReq.onsuccess = function(event) {
      const db = event.target.result;
      resolve({
        put: (keysetID, data) => put(db, keysetID, data),
        get: (keysetID) => get(db, keysetID),
        delete: (keysetID) => deleteFn(db, keysetID)
      });
    }

    dbReq.onerror = function(event) {
      reject(`error opening database ${event.target.errorCode}`);
    }
  });

  function put(db, keysetID, data) {
    return new Promise((resolve, reject) => {
      let tx = db.transaction(keystoreTable, 'readwrite');
      let store = tx.objectStore(keystoreTable);

      let req = store.put({
        key: keysetID,
        value: data
      });

      req.onsuccess = function() {
          resolve(this.result);
      }
      req.onerror = function(event) {
        reject(`error storing key ${event.target.errorCode}`);
      }
    });
  }

  function get(db, keysetID) {
    return new Promise((resolve, reject) => {
      let tx = db.transaction(keystoreTable, 'readwrite');
      let store = tx.objectStore(keystoreTable);

      let req = store.get(keysetID);

      req.onsuccess = function() {
          resolve(this.result?.value);
      }
      req.onerror = function(event) {
        reject(`error getting key ${event.target.errorCode}`);
      }
    });
  }

  function deleteFn(db, keysetID) {
    return new Promise((resolve, reject) => {
      let tx = db.transaction(keystoreTable, 'readwrite');
      let store = tx.objectStore(keystoreTable);

      let req = store.delete(keysetID);

      req.onsuccess = function() {
          resolve();
      }
      req.onerror = function(event) {
        reject(`error getting key ${event.target.errorCode}`);
      }
    });
  }
}

