# SDK Usage

Last updated: December 12, 2022 (commit `88e63eaac7e66835e779cb0a718337505bf05dcd`)

This guide explains how to use this SDK in Android or iOS code.

For the sake of readability, the following is omitted in the code examples:
* Error handling
* Optionals unwrapping (in Swift examples)

## API Package

The API package contains various interfaces and types that are used across multiple places in the SDK.
Wallet-SDK contains implementations of these various interfaces that can be used to get an application up and running
fairly quickly.
Implementations of these interfaces can be done in on the mobile side and injected in to the API methods that use them.
A few examples of when you might want to do this:
* Implementing platform-specific credential storage (perhaps in a secure enclave)
* Implementing platform-specific crypto functionality (perhaps leveraging device-specific security chips)

## In-Memory Credential Storage
The credential package contains an in-memory credential storage implementation that can be used to satisfy both
credential reader and credential writer interfaces. As it only uses in-memory storage, you will probably want to
create your own implementation in your mobile code that uses platform-specific storage.

### Examples

#### Kotlin (Android)

```kotlin
import dev.trustbloc.wallet.sdk.api.*
import dev.trustbloc.wallet.sdk.credential.*

val db = credential.newInMemoryDB()

val vc = api.JSONObject()
vc.setData = "VC JSON goes here"

db.add(vc)

retrievedVC = db.get("VC_ID")

retrievedVCs = db.getAll()

db.remove("VC_ID")
```

#### Swift (iOS)

```swift
import Walletsdk

let db = CredentialNewInMemoryDB()

let vc = ApiJSONObject(data: "VC JSON goes here")

db.add(vc)

retrievedVC = db.get("VC_ID")

retrievedVCs = db.getAll()

db.remove("VC_ID")
```

## KMS
This package contains a KMS implementation that uses Google's Tink crypto library.
Private keys may intermittently reside in local memory with this implementation so
keep this consideration in mind when deciding whether to use this or not.
The key store can be injected in by the caller. If the store argument in the constructor is left null/nil,
then an in-memory key store will be used.

### Examples

#### Kotlin (Android)

```kotlin
import dev.trustbloc.wallet.sdk.localkms

val kms = localkms.KMS(null) // The null store argument causes it to use in-memory storage

val keyHandle = kms.create(localkms.KeyTypeED25519)
```

#### Swift (iOS)

```swift
import Walletsdk

let kms = LocalkmsNewKMS(nil) // The nil store argument causes it to use in-memory storage

let keyHandle = kms.create(LocalkmsKeyTypeED25519)
```

## DID Creator

These examples will use an in-memory KMS.

A DID creator can be instantiated in one of two ways: with a key writer or with a key reader.
The behaviour and usage of the DID creator will differ depending on which way you chose.

### With Key Writer

The Keys used for DID documents are created for you automatically by the key writer.

#### Examples

##### Kotlin (Android)

```kotlin
import dev.trustbloc.wallet.sdk.api.CreateDIDOpts
import dev.trustbloc.wallet.sdk.did.*
import dev.trustbloc.wallet.sdk.localkms

val kms = localkms.KMS(null) // The null store argument causes it to use in-memory storage
val didCreator = did.NewCreatorWithKeyWriter(kms)
val didDocResolution = didCreator.create("key", CreateDIDOpts()) // Create a did:key doc
```

##### Swift (iOS)

```swift
import Walletsdk

let kms = LocalkmsNewKMS(nil) // The nil store argument causes it to use in-memory storage
let didCreator = DidNewCreatorWithKeyWriter(kms)
let didDocResolution = didCreator.create("key", ApiCreateDIDOpts()) // Create a did:key doc
```

### With Key Reader
The keys used for DID documents are not automatically generated on the caller's behalf.
They must be passed in by the caller.

#### Examples

##### Kotlin (Android)

```kotlin
import dev.trustbloc.wallet.sdk.api.*
import dev.trustbloc.wallet.sdk.did.*
import dev.trustbloc.wallet.sdk.localkms

val kms = localkms.KMS(null) // The null store argument causes it to use in-memory storage

val keyHandle = kms.create(localkms.KeyTypeED25519)

val didCreator = did.NewCreatorWithKeyReader(kms)

val createDIDOpts = api.CreateDIDOpts()
createDIDOpts.setKeyID = keyHandle.getKeyID()
createDIDOpts.setVerificationType = did.Ed25519VerificationKey2018

val didDocResolution = didCreator.create("key", createDIDOpts) // Create a did:key doc
```

##### Swift (iOS)

```swift
import Walletsdk

let kms = LocalkmsNewKMS(nil) // The nil store argument causes it to use in-memory storage

let keyHandle = kms.create(LocalkmsKeyTypeED25519)

let didCreator = DidNewCreatorWithKeyReader(kms)

let createDIDOpts = ApiCreateDIDOpts(keyID: keyHandle.keyID, verificationType: DidEd25519VerificationKey2018)

let didDocResolution = didCreator.create("key", createDIDOpts) // Create a did:key doc
```

## DID Resolver

### Examples

#### Kotlin (Android)

```kotlin
import dev.trustbloc.wallet.sdk.did.*

val didResolver = did.Resolver()

val didDoc = didResolver.resolve("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
```

#### Swift (iOS)

```swift
import Walletsdk

let didResolver = DidNewResolver()

let didDoc = didResolver.resolve("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
```

### With Key Reader
The keys used for DID documents are not automatically generated on the caller's behalf.
They must be passed in by the caller.

#### Examples

##### Kotlin (Android)

```kotlin
import dev.trustbloc.wallet.sdk.api.*
import dev.trustbloc.wallet.sdk.did.*
import dev.trustbloc.wallet.sdk.localkms

val kms = localkms.KMS(null) // The null store argument causes it to use in-memory storage

val keyHandle = kms.create(localkms.KeyTypeED25519)

val didCreator = did.NewCreatorWithKeyReader(kms)

val createDIDOpts = api.CreateDIDOpts()
createDIDOpts.setKeyID = keyHandle.getKeyID()
createDIDOpts.setVerificationType = did.Ed25519VerificationKey2018

val didDocResolution = didCreator.create("key", createDIDOpts) // Create a did:key doc
```

##### Swift (iOS)

```swift
import Walletsdk

let kms = LocalkmsNewKMS(nil) // The nil store argument causes it to use in-memory storage

let keyHandle = kms.create(LocalkmsKeyTypeED25519)

let didCreator = DidNewCreatorWithKeyReader(kms)

let createDIDOpts = ApiCreateDIDOpts(keyID: keyHandle.keyID, verificationType: DidEd25519VerificationKey2018)

let didDocResolution = didCreator.create("key", createDIDOpts) // Create a did:key doc
```

## DID Service Validation

Note the following limitations:
* The DID document associated with the DID you want to check must specify only a single service.
* If there are multiple URLs for a given service, only the first will be checked.

### Examples

#### Kotlin (Android)

```kotlin
import dev.trustbloc.wallet.sdk.did.*

val didResolver = did.Resolver()

val validationResult = did.ValidateService("YourDIDHere", didResolver)
```

#### Swift (iOS)

```swift
import Walletsdk

let didResolver = DidNewResolver()

let validationResult = DidValidateService("YourDIDHere", didResolver)
```

## OpenID4CI

The OpenID4CI package contains an API that can be used by a [holder](https://www.w3.org/TR/vc-data-model/#dfn-holders)
to go through the [OpenID4CI](https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0-08.html) flow.

Note that the implementation currently only supports the pre-authorized flow.

The general pattern is as follows:

1. Create a `ClientConfig` object. A `ClientConfig` contains the following mandatory parameters:
   * The user's DID.
   * A client ID to use.
   * A `DIDJWTSignerCreator`. See the next section for more detail on this.
   * A DID resolver.
2. Create a new `Interaction` object using an initiate issuance URI obtained from an issuer (e.g. via a QR code)
and your `ClientConfig` from the last step. The `Interaction` object is a stateful object and is meant to be used
for a single interaction of an OpenID4CI flow and then discarded.
3. Call the `Authorize` method on the `Interaction`. Since only the pre-authorized flow is supported currently,
the returned `AuthorizeResult` object can be ignored by the caller. All this step really does right now is ensures that
the initiation request indicates that the user is pre-authorized (and if not, returns an error letting the caller
know that the non-pre-authorized flow isn't implemented).
4. Create a `CredentialRequestOpts` object containing the user's PIN.
5. Call the `RequestCredential` method on the `Interaction`, passing in the `CredentialRequestOpts` object from the
last step. If successful, this method will return `CredentialResponses` to the caller, which contain the issued
credentials.
6. (Optional) - Call the `ResolveDisplay` method (with an optional preferred locale) on the `Interaction` object to
get display information for your new credentials.

### DIDJWTSignerCreator

As mentioned above, the `ClientConfig` object requires an implementation of a `DIDJWTSignerCreator`.
To understand why, let's look at what happens when the `RequestCredential` method is called. Using the DID resolver
(set in the `ClientConfig`), the user's DID (also from the `ClientConfig`) will be resolved, and an appropriate
verification method related to JWT signing will be found. Next, there needs to be some way to map between the
verification method and the corresponding private key to be used for signing. This is what the `DIDJWTSignerCreator`
does - it knows how to create a JWT signer object (that is able to access the private key) based on a verification
method.

Wallet-SDK includes an implementation of this that uses the Google Tink crypto library. Keys may reside in local memory
intermittently. For a production application, you may want to supply your own DIDJWTSignerCreator implementation that
uses a platform-specific crypto implementation.

### Examples

The following examples show how to use the APIs to go through the OpenID4CI flow using the iOS and Android bindings.
They use in-memory key storage and the Tink crypto library.

#### Kotlin (Android)

```kotlin
import dev.trustbloc.wallet.sdk.api.CreateDIDOpts
import dev.trustbloc.wallet.sdk.did.*
import dev.trustbloc.wallet.sdk.localkms
import dev.trustbloc.wallet.sdk.openid4ci

// Setup
val kms = localkms.KMS(null) // The null store argument causes it to use in-memory storage. Will use the Tink crypto library.
val signerCreator = localkms.SignerCreator(kms) // Will use the Tink crypto library
val didResolver = did.Resolver()
val didCreator = did.NewCreatorWithKeyWriter(kms)
val didDocResolution = didCreator.create("key", CreateDIDOpts()) // Create a did:key doc
val cfg = ClientConfig(didDocResolution.id(), "ClientID", signerCreator, didResolver)

// Going through the flow
val interaction = openid4ci.Interaction("YourRequestURIHere", cfg)
interaction.authorize() // Returned object doesn't matter with current implementation limitations
val userPIN = "1234"
val requestCredentialOpts = openid4ci.CredentialRequestOpts(userPIN)
val credentials = interaction.requestCredential(requestCredentialOpts) // Should probably store these somewhere
val displayData = interaction.resolveDisplay("en-US") // Optional (but useful)
```

#### Swift (iOS)

```swift
import Walletsdk

// Setup
let kms = LocalkmsNewKMS(nil) // The nil store argument causes it to use in-memory storage. Will use the Tink crypto library.
let signerCreator = LocalkmsCreateSignerCreator(kms) // Will use the Tink crypto library
let didResolver = DidNewResolver()
let didCreator = DidNewCreatorWithKeyWriter(kms)
let didDocResolution = didCreator.create("key", ApiCreateDIDOpts()) // Create a did:key doc
let cfg =  Openid4ciClientConfig(didDocResolution.id, "ClientID", signerCreator, didResolver)

// Going through the flow
let interaction = Openid4ciNewInteraction("YourRequestURIHere", cfg)
interaction.authorize() // Returned object doesn't matter with current implementation limitations
let userPIN = "1234"
let requestCredentialOpts = Openid4ciNewCredentialRequestOpts(userPIN)
let credentials = interaction.requestCredential(requestCredentialOpts) // Should probably store these somewhere
let displayData = interaction.resolveDisplay("en-US") // Optional (but useful)
```

## OpenID4VP

The OpenID4VP package contains an API that can be used by a [holder](https://www.w3.org/TR/vc-data-model/#dfn-holders)
to go through the [OpenID4VP](https://openid.net/specs/openid-connect-4-verifiable-presentations-1_0-ID1.html) flow.

The general pattern is as follows:

1. Create a new `Interaction` object. An `Interaction` object has the following mandatory parameters:
   * An authorization request URL
   * A key reader
   * A crypto implementation
   * A DID resolver
   * An LD document loader
2. Get the query by calling the `GetQuery` method on the `Interaction`.
3. Create a verifiable presentation from the credentials that match the query from the previous step.
4. Determine the key ID you want to use for signing (e.g. from one of the user's DID docs).
5. Call the `PresentCredential` method on the `Interaction` object.

### Examples

The following examples show how to use the APIs to go through the OpenID4VP flow using the iOS and Android bindings.
They use in-memory key storage and the Tink crypto library.

#### Kotlin (Android)

```kotlin
import dev.trustbloc.wallet.sdk.api.*
import dev.trustbloc.wallet.sdk.did.*
import dev.trustbloc.wallet.sdk.localkms
import dev.trustbloc.wallet.sdk.openid4vp
import dev.trustbloc.wallet.sdk.ld
import dev.trustbloc.wallet.sdk.credential

// Setup
val kms = localkms.KMS(null) // The null store argument causes it to use in-memory storage. Will use the Tink crypto library.
val didResolver = did.Resolver()
val docLoader = ld.DocLoader()
val didCreator = did.NewCreatorWithKeyWriter(kms)
val didDocResolution = didCreator.create("key", api.CreateDIDOpts()) // Create a did:key doc

// Going through the flow
val interaction = openid4vp.Interaction("YourAuthRequestURIHere", kms, kms.getCrypto(), didResolver, docLoader)
val query = interaction.getQuery()
val inquirer = credential.Inquirer(docLoader)
val issuedCredentials = api.VerifiableCredentialsArray() // Would need some actual credentials for this to actually work
val verifiablePres = inquirer.Query(query, credential.CredentialsOpt(issuedCredentials))
val keyID = didDocResolution.assertionMethodKeyID()
val credentials = interaction.presentCredential(verifiablePres, keyID)
```

#### Swift (iOS)

```swift
import Walletsdk

// Setup
let kms = LocalkmsNewKMS(nil) // The nil store argument causes it to use in-memory storage. Will use the Tink crypto library.
let didResolver = DidNewResolver()
let didCreator = DidNewCreatorWithKeyWriter(kms)
let didDocResolution = didCreator.create("key", ApiCreateDIDOpts()) // Create a did:key doc

// Going through the flow
let interaction = Openid4vpNewInteraction("YourAuthRequestURIHere", kms, kms.getCrypto(), didResolver, docLoader)
let query = interaction.getQuery()
let inquirer = CredentialNewinquirer(docLoader)
let issuedCredentials = ApiVerifiableCredentialsArray() // Would need some actual credentials for this to actually work
let verifiablePres = inquirer.Query(query, CredentialNewCredentialsOpt(issuedCredentials))
let keyID = didDocResolution.assertionMethodKeyID()
let credentials = interaction.presentCredential(verifiablePres, keyID)
```
