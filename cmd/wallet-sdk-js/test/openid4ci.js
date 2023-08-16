/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import '../src/wasm_exec.js';
import Agent from '../src/index.js';


describe("OpenID4CI and OpenID4VP integrated flow.", function () {
    it("issue credentials and verify", async function () {
        const env = __karma__.config;

        const kmsDatabase = await CreateDB("test")
        const agent = new Agent({
            assetsPath: "base/dist",
            kmsDatabase: kmsDatabase,
            didResolverURI: "http://localhost:8072/1.0/identifiers",
        });
        await agent.initialize()

        const didDoc = await agent.createDID({
            didMethod: "ion",
        })

        console.log("INITIATE_ISSUANCE_URL: ", env.INITIATE_ISSUANCE_URL);

        const openID4CIInteraction = await agent.createOpenID4CIIssuerInitiatedInteraction({
            initiateIssuanceURI: env.INITIATE_ISSUANCE_URL
        });

        const userPINRequired = (await openID4CIInteraction.preAuthorizedCodeGrantParams()).userPINRequired;

        expect(userPINRequired).toBe(false);

        const credentials = await openID4CIInteraction.requestCredentialWithPreAuth({
            pin: "",
            didDoc: didDoc
        });

        expect(credentials.length).toBe(1);


        const openID4VPInteraction = await agent.createOpenID4VPInteraction({
            authorizationRequest: env.INITIATE_VERIFICATION_URL
        })

        const query = await openID4VPInteraction.getQuery();

        const requirements = await agent.getSubmissionRequirements({
            query: query,
            credentials: credentials,
        });

        expect(requirements.length).toBe(1);
        expect(requirements[0].descriptors.length).toBe(1);
        expect(requirements[0].descriptors[0].matchedVCs.length).toBe(1);

        await openID4VPInteraction.presentCredential({
            credentials: requirements[0].descriptors[0].matchedVCs
        })

    });
});

// KMS database implementation
function CreateDB(dbName) {
    const keystoreTable = "keyStore";

    return new Promise(function (resolve, reject) {
        const dbReq = indexedDB.open(dbName, 2);

        dbReq.onupgradeneeded = function (event) {
            const db = event.target.result;

            if (!db.objectStoreNames.contains(keystoreTable)) {
                db.createObjectStore(keystoreTable, {keyPath: "key"});
            }
        }

        dbReq.onsuccess = function (event) {
            const db = event.target.result;
            resolve({
                put: (keysetID, data) => put(db, keysetID, data),
                get: (keysetID) => get(db, keysetID),
                delete: (keysetID) => deleteFn(db, keysetID)
            });
        }

        dbReq.onerror = function (event) {
            reject(`error opening database ${event.target.errorCode}`);
        }
    });

    function put(db, keysetID, data) {
        return new Promise((resolve, reject) => {
            const tx = db.transaction(keystoreTable, 'readwrite');
            const store = tx.objectStore(keystoreTable);

            const req = store.put({
                key: keysetID,
                value: data
            });

            req.onsuccess = function () {
                resolve(this.result);
            }
            req.onerror = function (event) {
                reject(`error storing key ${event.target.errorCode}`);
            }
        });
    }

    function get(db, keysetID) {
        return new Promise((resolve, reject) => {
            const tx = db.transaction(keystoreTable, 'readwrite');
            const store = tx.objectStore(keystoreTable);

            const req = store.get(keysetID);

            req.onsuccess = function () {
                resolve(this.result?.value);
            }
            req.onerror = function (event) {
                reject(`error getting key ${event.target.errorCode}`);
            }
        });
    }

    function deleteFn(db, keysetID) {
        return new Promise((resolve, reject) => {
            const tx = db.transaction(keystoreTable, 'readwrite');
            const store = tx.objectStore(keystoreTable);

            const req = store.delete(keysetID);

            req.onsuccess = function () {
                resolve();
            }
            req.onerror = function (event) {
                reject(`error getting key ${event.target.errorCode}`);
            }
        });
    }
}

