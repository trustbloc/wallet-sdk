# JS SDK Usage

Last updated: Jul 18, 2023 (commit `0f0e847a345a979d216f1e03313a00b630e7678e`)

This guide explains how to use this SDK in Web. The examples in this document demonstrate
how to use the various APIs from a JavaScript perspective.

### Examples

This example shows the full OpenID4CI flow.

```javascript
// initialize the Wallet SDK Agent
let pathWhereWalletsWasmLocated = ""

let agent = new Agent({
    assetsPath: pathWhereWalletsWasmLocated, 
    didResolverURI:didResolverURI,
    kmsDatabase: kmsDatabase
});

await agent.initialize()

// Create Decentralized Identifier (DID)
const userDID = await agent.createDID({
    didMethod: didMethod,
    keyType: keyType
})

// get the issuance init url
let initiateIssuanceURI = "URI from scanned QR code."

// Start OpenID4CI Issaunce flow
let openID4CIInteraction = await agent.createOpenID4CIIssuerInitiatedInteraction({
    initiateIssuanceURI: initiateIssuanceURI
})

// check if the issuance requires a PIN
let userPINRequired = (await openID4CIInteraction.preAuthorizedCodeGrantParams()).userPINRequired;

let issuedCrednetials = await openID4CIInteraction.requestCredentialWithPreAuth({
    pin: userPinEntered,
    didDoc: userDID
});

// steps to show credential display on UI
let issuerURI = openID4CIInteraction.issuerURI()

let rawDisplayData = await agent.resolveDisplayData({
    issuerURI: issuerURI,
    credentials: issuedCrednetials
})

let parsedDisplayData = await agent.parseResolvedDisplayData({
    resolvedCredentialDisplayData: rawDisplayData
})
```

### KMS Database

kmsDatabase should implement the next 3 methods:

```
  async function put(db, keysetID, data)
  async function get(db, keysetID)
  async function delete(db, keysetID)
```

#### Sample KMS Database implementation 

```javascript

// KMS database implementation using IndexedDB
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
```
