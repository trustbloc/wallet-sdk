# SDK Usage

Last updated: March 14, 2023 (commit `7936266814445f20bcd840b94399cebaf50872c7`)

This guide explains how to use this SDK in Android or iOS code.

Note that the exact type/function/method names differ depending on whether you're using the Android bindings or iOS
bindings. See [Packages](packages.md) for more information. Additionally, the examples in this document demonstrate
how to use the various APIs from a Kotlin and Swift perspective. 

For the sake of readability, the following is omitted in most code examples:
* Error handling
* Nullable type checks (in Kotlin examples)
* Optionals unwrapping (in Swift examples)

## Error Handling

Errors from Wallet-SDK come in a structured format that can be (optionally) parsed, allowing for individual fields to be accessed.

Errors have three fields:

* Category: A short descriptor of the general category of error.
* Code: A short alphanumeric code that is used to broadly group various related errors.
This contains the Category descriptor.
* Details: Lower-level details about the precise cause of the error.

### Examples

#### Kotlin (Android)

```kotlin
import dev.trustbloc.wallet.sdk.walleterror.Walleterror

try {
    // Some code here
} catch (e: Exception) {
    val parsedError = Walleterror.parse(e.message)
    // Access parsedError.code, parsedError.category, and parsedError.details as you see fit
}
```

#### Swift (iOS)

Note: For error parsing, make sure you pass `localizedDescription` from the `NSError` object into the
`WalleterrorParse` function. Passing in `description` will cause the parsing to not work correctly.

```swift
import Walletsdk

do {
    try someSDKMethodHere()
} catch let error as NSError {
    let parsedError = WalleterrorParse(error.localizedDescription)
    // Access parsedError.code, parsedError.category, and parsedError.details as you see fit
}
```

For non-primitive return types from functions/methods, they will always be optionals (in Swift) or nullable
(in Kotlin/Java). Some of these functions/methods may also throw exceptions (or return errors)


## API Package

The API package contains various interfaces and types that are used across multiple places in the SDK.
Wallet-SDK contains implementations of these various interfaces that can be used to get an application up and running
fairly quickly.
Implementations of these interfaces can be done in on the mobile side and injected in to the API methods that use them.
A few examples of when you might want to do this:
* Implementing platform-specific credential storage (perhaps in a secure enclave)
* Implementing platform-specific crypto functionality (perhaps leveraging device-specific security chips)

### Verifiable Credentials

Verifiable Credential objects are present throughout Wallet-SDK. They have a number of useful methods:

* `serialize()`: Returns a serialized representation of this VC. This is useful for storing VCs in a persistent
database. To convert the serialized representation back into a `VerifiableCredential` object,
see [Parsing Credentials](#parsing-credentials).
* `id()`: Returns this VC's ID.
* `issuerID()`: Returns the ID of this VC's issuer (typically a DID).
* `name()`: Returns this VC's name.
If the VC doesn't provide a name (or the name is not a string), then an empty string is returned.
* `types()`: Returns the types of this VC. At a minimum, one of the types will be "VerifiableCredential".
There may be additional more specific credential types as well.
* `claimTypes()`: Returns the types specified in the claims of this VC. If none are found, then null/nil is returned.
* `issuanceDate()`: Returns this VC's issuance date as a Unix timestamp.
* `expirationDate()`: Returns this VC's expiration date as a Unix timestamp.

## Parsing Credentials

Serialized credentials cannot be used directly in Wallet-SDK. Instead, they must be parsed first.

The `parse` function is located in the `vcparse` package. It can be provided with a number of parameters, all optional:
* `disableProofCheck`: Disables the proof check operation when parsing the VC. The proof check can be an expensive
operation - in certain scenarios, it may be appropriate to disable the check. By default, proofs are checked.
* `documentLoader`: Specifies a JSON-LD document loader to use when parsing the VC. If none is specified, then a
network-based loader will be used.

### Examples

#### Kotlin (Android)

```kotlin
import dev.trustbloc.wallet.sdk.vcparse.Vcparse

// Specified options below: Don't disable the proof check, and use the default network-based document loader.
// Equivalent to just passing null directly into the parse function instead of this object below.
val opts = Opts(false, null)
val vc = Vcparse.parse("Serialized VC goes here", opts)
```

#### Swift (iOS)

```swift
import Walletsdk

// Specified options below: Don't disable the proof check, and use the default network-based document loader.
// Equivalent to just passing null directly into the parse function instead of this object below.
let opts = VcparseNewOpts(false, nil)
let vc = VcparseParse("Serialized VC goes here", opts, nil)
```

## In-Memory Credential Storage
The credential package contains an in-memory credential storage implementation that can be used to store credentials
in memory and also satisfy the credential reader interface. As it only uses in-memory storage, you will probably want to
create your own implementation in your mobile code that uses platform-specific storage.

### Examples

#### Kotlin (Android)

```kotlin
import dev.trustbloc.wallet.sdk.api.*
import dev.trustbloc.wallet.sdk.credential.*
import dev.trustbloc.wallet.sdk.vcparse.Vcparse

val db = credential.newInMemoryDB()

val opts = Opts(false, null)
val vc = Vcparse.parse("Serialized VC goes here", opts)

db.add(vc)

val retrievedVC = db.get("VC_ID")

val retrievedVCs = db.getAll()

db.remove("VC_ID")
```

#### Swift (iOS)

```swift
import Walletsdk

let db = CredentialNewInMemoryDB()

let opts = VcparseNewOpts(false, nil)
let vc = VcparseParse("Serialized VC goes here", opts, nil)!

db.add(vc)

let retrievedVC = db.get("VC_ID")

let retrievedVCs = db.getAll()

db.remove("VC_ID")
```

## Local KMS
This package contains a local KMS implementation that uses Google's Tink crypto library.
Private keys may intermittently reside in local memory with this implementation so
keep this consideration in mind when deciding whether to use this or not.
The caller must inject a key store for the KMS to use. This package includes an in-memory key store implementation
that can be used, but you will likely want to inject in your own implementation at some point so your keys are
persisted.

### Examples

#### Kotlin (Android)

```kotlin
import dev.trustbloc.wallet.sdk.localkms.Localkms
import dev.trustbloc.wallet.sdk.localkms.MemKMSStore

val memKMSStore = MemKMSStore.MemKMSStore()
val kms = Localkms.newKMS(memKMSStore)

val jwk1 = kms.create(localkms.KeyTypeED25519)
val jwk2 = kms.create(localkms.KeyTypeP384)
```

#### Swift (iOS)

```swift
import Walletsdk

let memKMSStore = LocalkmsNewMemKMSStore()
let kms = LocalkmsNewKMS(memKMSStore, nil)

let jwk1 = kms.create(LocalkmsKeyTypeED25519)
let jwk2 = kms.create(LocalkmsKeyTypeP384)
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
import dev.trustbloc.wallet.sdk.did.Creator
import dev.trustbloc.wallet.sdk.localkms.Localkms
import dev.trustbloc.wallet.sdk.localkms.MemKMSStore

val memKMSStore = MemKMSStore.MemKMSStore()
val kms = Localkms.newKMS(memKMSStore)
val didCreator = Creator(kms as KeyWriter)
val didDocResolution = didCreator.create("key", CreateDIDOpts()) // Create a did:key doc
```

##### Swift (iOS)

```swift
import Walletsdk

let memKMSStore = LocalkmsNewMemKMSStore()
let kms = LocalkmsNewKMS(memKMSStore, nil)
let didCreator = DidNewCreatorWithKeyWriter(kms, nil)
let didDocResolution = didCreator.create("key", ApiCreateDIDOpts()) // Create a did:key doc
```

### With Key Reader
The keys used for DID documents are not automatically generated on the caller's behalf.
They must be passed in by the caller.

#### Examples

##### Kotlin (Android)

```kotlin

import dev.trustbloc.wallet.sdk.api.CreateDIDOpts
import dev.trustbloc.wallet.sdk.did.Creator
import dev.trustbloc.wallet.sdk.did.Did.Ed25519VerificationKey2018
import dev.trustbloc.wallet.sdk.localkms.Localkms
import dev.trustbloc.wallet.sdk.localkms.MemKMSStore

val memKMSStore = MemKMSStore.MemKMSStore()
val kms = Localkms.newKMS(memKMSStore)

val keyHandle = kms.create(localkms.KeyTypeED25519)

val didCreator = Creator(kms as KeyReader)

val createDIDOpts = api.CreateDIDOpts()
createDIDOpts.setKeyID = keyHandle.getKeyID()
createDIDOpts.setVerificationType = Ed25519VerificationKey2018

val didDocResolution = didCreator.create("key", createDIDOpts) // Create a did:key doc
```

##### Swift (iOS)

```swift
import Walletsdk

let memKMSStore = LocalkmsNewMemKMSStore()
let kms = LocalkmsNewKMS(memKMSStore)

let keyHandle = kms.create(LocalkmsKeyTypeED25519)

let didCreator = DidNewCreatorWithKeyReader(kms, nil)

let createDIDOpts = ApiCreateDIDOpts(keyID: keyHandle.keyID, verificationType: DidEd25519VerificationKey2018)

let didDocResolution = didCreator.create("key", createDIDOpts) // Create a did:key doc
```

### Error Codes & Troubleshooting Tips

| Error                             | Possible Reasons                                                                                          |
|-----------------------------------|-----------------------------------------------------------------------------------------------------------|
| CREATE_DID_KEY_FAILED(DID1-0000)  | No verification type specified in the createDIDOpts object.                                               |
| CREATE_DID_ION_FAILED(DID1-0001)  | No verification type specified in the createDIDOpts object.                                               |
| CREATE_DID_JWK_FAILED(DID1-0002)  | No verification type specified in the createDIDOpts object.                                               |
| UNSUPPORTED_DID_METHOD(DID0-0003) | The DID method specified in the `create` call is unsupported. Only `key`, `ion`, and `jwk` are supported. |

## DID Resolver

### Examples

#### Kotlin (Android)

```kotlin
import dev.trustbloc.wallet.sdk.did.*

val didResolver = did.Resolver("") // argument is resolverServerURI, can be empty.

val didDoc = didResolver.resolve("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
```

#### Swift (iOS)

```swift
import Walletsdk

let didResolver = DidNewResolver("", nil)

let didDoc = didResolver.resolve("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
```

### Error Codes & Troubleshooting Tips

| Error                                         | Possible Reasons                                                                                                                                                                                                                                            |
|-----------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| CREATE_DID_KEY_FAILED(DID1-0000)              | No verification type specified in the createDIDOpts object.                                                                                                                                                                                                 |
| CREATE_DID_ION_FAILED(DID1-0001)              | No verification type specified in the createDIDOpts object.                                                                                                                                                                                                 |
| CREATE_DID_JWK_FAILED(DID1-0002)              | No verification type specified in the createDIDOpts object.                                                                                                                                                                                                 |
| UNSUPPORTED_DID_METHOD(DID0-0003)             | The DID method specified in the `create` call is unsupported. Only `key`, `ion`, and `jwk` are supported.                                                                                                                                                   |
| DID_RESOLVER_INITIALIZATION_FAILED(DID1-0005) | An invalid resolver server URI was specified. Note that the resolver server URI is optional, so if it's not required then leave this blank.                                                                                                                 |
| DID_RESOLUTION_FAILED(DID1-0004)              | The specified DID doesn't exist.<br/><br/>The specified DID is incorrectly formatted.<br/><br/>The specified DID isn't supported by Wallet-SDK.<br/><br/>The specified DID requires remote resolution, but there was a communication error with the server. |

## DID Service Validation

Note the following limitations:
* The DID document associated with the DID you want to check must specify only a single service.
* If there are multiple URLs for a given service, only the first will be checked.

### Examples

#### Kotlin (Android)

```kotlin
import dev.trustbloc.wallet.sdk.did.*

val didResolver = Resolver()

val validationResult = Did.validateLinkedDomains("YourDIDHere", didResolver)
```

#### Swift (iOS)

```swift
import Walletsdk

let didResolver = DidNewResolver("", nil)

let validationResult = ValidateLinkedDomains("YourDIDHere", didResolver)
```

## OpenID4CI

The OpenID4CI package contains an API that can be used by a [holder](https://www.w3.org/TR/vc-data-model/#dfn-holders)
to go through the [OpenID4CI](https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0-08.html) flow.

Note that the implementation currently only supports the pre-authorized flow.

The general pattern is as follows:

1. Create a `ClientConfig` object. A `ClientConfig` contains the following parameters:
   * A client ID to use.
   * A crypto implementation
   * A DID resolver.
   * An activity logger (optional, but if set then this will be used to log credential activities)

   Optionally, after creating a `ClientConfig` object, you can set additional headers to be sent to the issuer by
calling the `addHeaders` method on the `ClientConfig`.
2. Create a new `Interaction` object using an initiate issuance URI obtained from an issuer (e.g. via a QR code)
and your `ClientConfig` from the last step. The `Interaction` object is a stateful object and is meant to be used
for a single instance of an OpenID4CI flow and then discarded.
3. Call the `Authorize` method on the `Interaction`. Since only the pre-authorized flow is supported currently,
the returned `AuthorizeResult` object can be ignored by the caller. All this step really does right now is ensures that
the initiation request indicates that the user is pre-authorized (and if not, returns an error letting the caller
know that the non-pre-authorized flow isn't implemented).
4. Create a `CredentialRequestOpts` object containing the user's PIN.
5. Call the `RequestCredential` method on the `Interaction`, passing in the `CredentialRequestOpts` object from the
last step. If successful, this method will return `CredentialResponses` to the caller, which contain the issued
credentials.
6. (Optional, can be called at any point after step 2) - Call the `IssuerURI` method on the `Interaction` object to
get the issuer URI. The issuer URI should be stored somewhere for later use, since it can be used to get the
display data. See [Credential Display Data](#credential-display-data) for more information.


### Examples

The following examples show how to use the APIs to go through the OpenID4CI flow using the iOS and Android bindings.
They use in-memory key storage and the Tink crypto library.

#### Kotlin (Android)

```kotlin
import dev.trustbloc.wallet.sdk.localkms.Localkms
import dev.trustbloc.wallet.sdk.localkms.MemKMSStore
import dev.trustbloc.wallet.sdk.did.Resolver
import dev.trustbloc.wallet.sdk.did.Creator
import dev.trustbloc.wallet.sdk.openid4ci.Interaction
import dev.trustbloc.wallet.sdk.openid4ci.ClientConfig
import dev.trustbloc.wallet.sdk.openid4ci.CredentialRequestOpts
import dev.trustbloc.wallet.sdk.openid4ci.mem

// Setup
val memKMSStore = MemKMSStore.MemKMSStore()
val kms = Localkms.newKMS(memKMSStore)
val didResolver = Resolver("")
val didCreator = Creator(kms as KeyWriter)
val didDocResolution = didCreator.create("key", CreateDIDOpts()) // Create a did:key doc
val activityLogger = mem.ActivityLogger() // Optional, but useful for tracking credential activities
val cfg = ClientConfig("ClientID", kms.crypto, didResolver, activityLogger)

// Going through the flow
val interaction = Interaction("YourRequestURIHere", cfg)
interaction.authorize() // Returned object doesn't matter with current implementation limitations
val userPIN = "1234"
val requestCredentialOpts = CredentialRequestOpts(userPIN)
val credentials = interaction.requestCredential(requestCredentialOpts, didDocResolution.assertionMethod()) // Should probably store these somewhere
val issuerURI = interaction.issuerURI() // Optional (but useful)
// Consider checking the activity log at some point after the interaction
```

#### Swift (iOS)

```swift
import Walletsdk

// Setup
let memKMSStore = LocalkmsNewMemKMSStore()
let kms = LocalkmsNewKMS(memKMSStore, nil)
let didResolver = DidNewResolver("", nil)
let didCreator = DidNewCreatorWithKeyWriter(kms, nil)
let didDocResolution = didCreator.create("key", ApiCreateDIDOpts()) // Create a did:key doc
let activityLogger = MemNewActivityLogger() // Optional, but useful for tracking credential activities
let cfg =  Openid4ciClientConfig(didDocResolution.id, clientID: "ClientID", didRes: didResolver, activityLogger: activityLogger)

// Going through the flow
let interaction = Openid4ciNewInteraction("YourRequestURIHere", cfg, nil)
interaction.authorize() // Returned object doesn't matter with current implementation limitations
let userPIN = "1234"
let requestCredentialOpts = Openid4ciNewCredentialRequestOpts(userPIN)
let credentials = interaction.requestCredential(requestCredentialOpts, didDocResolution.assertionMethod()) // Should probably store these somewhere
let issuerURI = interaction.issuerURI() // Optional (but useful)
// Consider checking the activity log at some point after the interaction
```

### Error Codes & Troubleshooting Tips

#### Creating New Interaction Object

| Error                                             | Possible Reasons                                                                                                                                                                                                                      |
|---------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| NO_CLIENT_CONFIG_PROVIDED(OCI0-0000)              | A nil/null client configuration object was passed in.                                                                                                                                                                                 |
| CLIENT_CONFIG_NO_CLIENT_ID_PROVIDED(OCI0-0001)    | The client configuration doesn't specify a client ID).                                                                                                                                                                                |
| CLIENT_CONFIG_NO_DID_RESOLVER_PROVIDED(OCI0-0002) | The client configuration doesn't specify a DID resolver (or it's nil).                                                                                                                                                                |
| INVALID_ISSUANCE_URI(OCI0-0004)                   | The issuance URI used to initiate the OpenID4CI flow isn't a valid URL.<br/><br/>The issuance URI doesn't specify a credential offer.                                                                                                 |
| INVALID_CREDENTIAL_OFFER(OCI0-0005)               | The credential offer object is malformed.<br/><br/>The issuance URI specified an endpoint for retrieving the credential offer, but there was an error during the HTTP GET call. The server may be down or have a configuration issue. |
| PRE_AUTHORIZED_GRANT_TYPE_REQUIRED(OCI0-0003)     | The credential offer doesn't specify the pre-authorized grant type (which is the only grant type this SDK currently supports).                                                                                                        |

##### Requesting Credential

| Error                                  | Possible Reasons                                                                                                                                                                                                                                                                                                                                                                                                                                    |
|----------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| PIN_CODE_REQUIRED(OCI0-0007)           | The credential offer requires a user PIN but none was provided via the requestCredentialOpts object.                                                                                                                                                                                                                                                                                                                                                |
| ISSUER_OPENID_FETCH_FAILED(OCI1-0008)  | An error occurred while doing an HTTP GET call on the issuer's OpenID configuration endpoint. The server may be down or have a configuration issue.<br/><br/>The issuer's OpenID configuration object is malformed.                                                                                                                                                                                                                                 |
| TOKEN_FETCH_FAILED(OCI1-0010)          | The user PIN is incorrect.<br/><br/>An error occurred while doing an HTTP POST call on the issuer's token endpoint. The server may be down or have a configuration issue.<br/><br/>The token response object from the server is malformed.                                                                                                                                                                                                          |
| JWT_SIGNING_FAILED(OCI1-0011)          | The KMS is missing the key that was used to create the DID that you're using. This could happen if your KMS storage is not storing and retrieving keys as expected, or if there is a mismatch between the KMS you used to create the DID (whose verification method you pass into the `requestCredential` function) and the `crypto` object (which should be obtained via the `getCrypto()` method on the KMS) passed in to the interaction object. |
| KEY_ID_NOT_CONTAIN_DID_PART(OCI1-0013) | The DID is incompatible with Wallet-SDK.                                                                                                                                                                                                                                                                                                                                                                                                            |
| METADATA_FETCH_FAILED(OCI1-0009)       | An error occurred while doing an HTTP GET call on the issuer's OpenID credential issuer endpoint. The server may be down or have a configuration issue.<br/><br/>The issuer metadata object from the server is malformed.                                                                                                                                                                                                                           |
| CREDENTIAL_FETCH_FAILED(OCI1-0012)     | An error occurred while doing an HTTP GET call on the issuer's credential endpoint. The server may be down or have a configuration issue.<br/><br/>The credential response object from the server is malformed.                                                                                                                                                                                                                                     |
| CREDENTIAL_PARSE_FAILED(OCI1-0014)     | The issued credential is invalid, signed incorrectly, or could not be verified.                                                                                                                                                                                                                                                                                                                                                                     |

## Credential Display Data

After completing the `RequestCredential` step of the OpenID4CI flow, you will have your issued Verifiable Credential
objects. These objects contain the data needed for various wallet operations, but they don't tell you how you can
display the credential data in an easily-understandable way via a user interface. This is where the credential display
data comes in.

To get display data, call the `resolveDisplay` function with your VCs and the issuer URI. An issuer URI can be obtained by
calling the `issuerURI` method on an OpenID4CI interaction object after it's been instantiated. It's a good idea to
store the issuer URI somewhere in persistent storage after going through the OpenID4CI flow. This way, you can call the
`resolveDisplay` function later if/when you need to refresh your display data based on the latest display information
from the issuer. See [Resolve Display](#resolve-display) for more information.

Display data objects can be serialized using the `serialize()` method (useful for storage) and parsed from serialized
form back into display data objects using the `parseDisplayData()` function.

The structure of the display data object is as follows:

### `Data`

* The root object.
* Contains display information for the issuer and each of the credentials passed in to the `resolveDisplay`
function.
* Can be serialized using the `serialize()` method and parsed using the `parseDisplayData()` function.
* The `issuerDisplay()` method returns the `IssuerDisplay` object.
* Use the `credentialDisplaysLength()` and `credentialDisplayAtIndex()` methods to iterate over the `CredentialDisplay`
  objects. There's one for each credential passed in to the `resolveDisplay` function, and they're in the same order.

### `IssuerDisplay`

* Describes display information about the issuer.
* Can be serialized using the `serialize()` method and parsed using the `parseIssuerDisplay()` function.
* Has `name()` and `locale()` methods.

### `CredentialDisplay`

* Describes display information about a credential.
* Can be serialized using the `serialize()` method and parsed using the `parseCredentialDisplay()` function.
* The `overview()` method returns the `CredentialOverview` object.
* Use the `claimsLength()` and `claimAtIndex()` methods to iterate over the `Claim` objects.

### `CredentialOverview`

* Describes display information for a credential as a whole.
* Has `name()`, `logo()`, `backgroundColor()`, `textColor()`, and `locale()` methods. The `logo()` method returns
  a `Logo` object.

### `Logo`

* Describes display information for a logo.
* Has `url()` and `altText()` methods.

### `Claim`

* Describes display information for a specific claim within a credential.
* Has `label()`, `rawID()`, `valueType()`, `value()`, `rawValue()`, `isMasked()`, `hasOrder()`, `order()`,
`pattern()`, and `locale()` methods.
* For example, if the UI were to display "Given Name: Alice", then `label()` would correspond to "Given Name" while
  `value()` would correspond to "Alice".
* Display order data is optional and will only exist if the issuer provided it. Use the `hasOrder()` method
to determine if there is a specified order before attempting to retrieve the order, since `order()` will return an
error/throw an exception if the claim has no order information. If you've ensured that the claim has an order
(by using `hadOrder()`), then you can safely ignore the error/exception from the `order()` method.
* `IsMasked()` indicates whether this claim's value is masked. If this method returns true, then the `value()` method
will return the masked value while the `rawValue()` method will return the unmasked version.
* `rawValue()` returns the raw display value for this claim without any formatting.
For example, if this claim is masked, this method will return the unmasked version.
If no special formatting was applied to the display value, then this method will be equivalent to calling Value.
* `rawID()` returns the claim's ID, which is the raw field name (key) from the VC associated with this claim.
It's not localized or formatted for display.
* `valueType()` returns the value type for this claim - when it's "image", then you should expect the value data to be
formatted using the [data URL scheme](https://www.rfc-editor.org/rfc/rfc2397).

### A Note about `locale()`

The locale returned by the various `locale()` methods may not be the same as the preferred locale you passed into the
`ResolveDisplay` function under certain circumstances. For instance, if the locale you passed in wasn't available,
then a default locale may get used instead.

### Resolve Display

The following examples show how the `ResolveDisplay` function can be used. The function accepts a `ResolveOpts` object,
which contains the VCs and the issuer URI (both are mandatory parameters). Optionally, after creating a`ResolveOpts`
object, you can set additional headers to be sent to the issuer by calling the `addHeaders` method on the
`ResolveOpts` object.

##### Kotlin (Android)

```kotlin
import dev.trustbloc.wallet.sdk.api.VerifiableCredentialsArray
import dev.trustbloc.wallet.sdk.display.*
import dev.trustbloc.wallet.sdk.openid4ci.Openid4ci

val vcArray = VerifiableCredentialsArray()

vcArray.add(yourVCHere)

val resolveOpts = ResolveOpts(vcCredentials, "Issuer_URI_Goes_Here")

val displayData = Display.resolve(resolveOpts)
```

##### Swift (iOS)

```kotlin
import Walletsdk

let vcArray = ApiVerifiableCredentialsArray()

vcArray.add(yourVCHere)

let resolveOpts = DisplayNewResolveOpts(vcArray, "Issuer_URI_Goes_Here")

let displayData = DisplayResolve(vcArray, "Issuer_URI_Goes_Here")
```

## OpenID4VP

The OpenID4VP package contains an API that can be used by a [holder](https://www.w3.org/TR/vc-data-model/#dfn-holders)
to go through the [OpenID4VP](https://openid.net/specs/openid-connect-4-verifiable-presentations-1_0-ID1.html) flow.

The general pattern is as follows:

1. Create a new `ClientConfig` object. A `ClientConfig` object has the following parameters:
   * A key reader
   * A crypto implementation
   * A DID resolver
   * An LD document loader
   * An activity logger (optional, but if set then this will be used to log credential activities)

   Optionally, after creating a `ClientConfig` object, you can set additional headers to be sent to the verifier by
calling the `addHeaders` method on the `ClientConfig`.
2. Create a new `Interaction` object using an authorization request URI obtained from a verifier (e.g. via a QR code)
and your `ClientConfig` from the last step. The `Interaction` object is a stateful object and is meant to be used
for a single instance of an OpenID4VP flow and then discarded.
3. Get the query by calling the `GetQuery` method on the `Interaction`.
4. Select the credentials that match the query from the previous step.
5. Determine the key ID you want to use for signing (e.g. from one of the user's DID docs).
6. Call the `PresentCredential` method on the `Interaction` object with the selected credentials.

### Examples

The following examples show how to use the APIs to go through the OpenID4VP flow using the iOS and Android bindings.
They use in-memory key storage and the Tink crypto library.

#### Kotlin (Android)

```kotlin
import dev.trustbloc.wallet.sdk.localkms.Localkms
import dev.trustbloc.wallet.sdk.localkms.MemKMSStore
import dev.trustbloc.wallet.sdk.did.Resolver
import dev.trustbloc.wallet.sdk.did.Creator
import dev.trustbloc.wallet.sdk.ld.DocLoader
import dev.trustbloc.wallet.sdk.localkms
import dev.trustbloc.wallet.sdk.openid4vp
import dev.trustbloc.wallet.sdk.ld
import dev.trustbloc.wallet.sdk.credential
import dev.trustbloc.wallet.sdk.openid4ci.mem
import dev.trustbloc.wallet.sdk.openid4vp.ClientConfig
import dev.trustbloc.wallet.sdk.openid4vp.Interaction

// Setup
val memKMSStore = MemKMSStore.MemKMSStore()
val kms = Localkms.newKMS(memKMSStore)
val didResolver = Resolver("")
val didCreator = Creator(kms as KeyWriter)
val documentLoader = DocLoader()
val activityLogger = mem.ActivityLogger() // Optional, but useful for tracking credential activities
val cfg = ClientConfig(kms, kms.getCrypto(), didResolver, documentLoader, activityLogger)

// Going through the flow
val interaction = openid4vp.Interaction("YourAuthRequestURIHere", cfg)
val query = interaction.getQuery()
val inquirer = credential.Inquirer(docLoader)
val savedCredentials = api.VerifiableCredentialsArray() // Would need some actual credentials for this to actually work

// Use this code to display the list of VCs to select which of them to send.
val matchedRequirements = inquirer.getSubmissionRequirements(query, savedCredentials) 
val matchedRequirement = matchedRequirements.atIndex(0) // Usually we will have one requirement
val requirementDesc = matchedRequirement.descriptorAtIndex(0) // Usually requirement will contain one descriptor
val selectedVCs = api.VerifiableCredentialsArray()
selectedVCs.add(requirementDesc.matchedVCs.atIndex(0)) // Users should select one VC for each descriptor from the matched list and confirm that they want to share it

interaction.presentCredential(selectedVCs)
// Consider checking the activity log at some point after the interaction
```

#### Swift (iOS)

```swift
import Walletsdk

// Setup
let memKMSStore = LocalkmsNewMemKMSStore()
let kms = LocalkmsNewKMS(memKMSStore, nil)
let didResolver = DidNewResolver("", nil)
let documentLoader = LdNewDocLoader()
let activityLogger = MemNewActivityLogger() // Optional, but useful for tracking credential activities
let clientConfig = Openid4vpClientConfig(keyHandleReader: kms, crypto: kms.getCrypto(), didResolver: didResolver, ldDocumentLoader: documentLoader, activityLogger: activityLogger)

// Going through the flow
let interaction = Openid4vpInteraction("YourAuthRequestURIHere", config: clientConfig)
let query = interaction.getQuery()
let inquirer = CredentialNewInquirer(docLoader)
let savedCredentials = ApiVerifiableCredentialsArray() // Would need some actual credentials for this to actually work

// Use this code to display the list of VCs to select which of them to send.
let matchedRequirements = inquirer.getSubmissionRequirements(query, savedCredentials) 
let matchedRequirement = matchedRequirements.atIndex(0) // Usually we will have one requirement
let requirementDesc = matchedRequirement.descriptorAtIndex(0) // Usually requirement will contain one descriptor
let selectedVCs = ApiVerifiableCredentialsArray()
selectedVCs.add(requirementDesc.matchedVCs.atIndex(0)) // Users should select one VC for each descriptor from the matched list and confirm that they want to share it

let credentials = interaction.presentCredential(selectedVCs)
// Consider checking the activity log at some point after the interaction
```

### Error Codes & Troubleshooting Tips

| Error                                             | Possible Reasons                                                                                                                                                                                |
|---------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| REQUEST_OBJECT_FETCH_FAILED(OVP1-0000)            | An incorrect authorization request URI was specified.<br/><br/>The verifier server is down or incorrectly configured.                                                                           |
| VERIFY_AUTHORIZATION_REQUEST_FAILED(OVP1-0001)    | The signature in the JWT received from the verified JWT is invalid.<br/><br/>Malformed request object received from the verifier server.                                                        |
| FAIL_TO_GET_MATCH_REQUIREMENTS_RESULTS(CRQ0-0004) | Invalid presentation definition received from the verifier.                                                                                                                                     |
| NO_CREDENTIAL_SATISFY_REQUIREMENTS(CRQ0-0003)     | None of your supplied credentials satisfy the requirements set by the verifier. Make sure you've gone through the full credential matching process correctly. See the OpenID4VP examples above. |
| CREATE_AUTHORIZED_RESPONSE(OVP1-0002)             | No credentials provided in the `presentCredential` method call.                                                                                                                                 |
| SEND_AUTHORIZED_RESPONSE(OVP1-0003)               | The verifier server rejected your credentials (couldn't be verified, wrong type, etc).<br/><br/>The verifier server is down or incorrectly configured.                                          |

## Metrics

Certain Wallet-SDK functionality is able to report back performance metrics to the caller.

Specifically:
* DID creation
* OpenID4CI
* OpenID4VP

To enable metrics reporting, you must pass in a MetricsLogger implementation into the various config/opts objects.
There is an included MetricsLogger implementation that logs messages to standard error (probably will end up in the console) using pre-determined formatting.
This implementation can be used or a custom implementation can be injected.

The object that gets logged when a metrics event occurs is as follows:

* Event: The name of the event that occurred.
* Parent event: The name of the event that encompasses this event. Some longer Wallet-SDK operations log a larger event
that captures the overall time of the method, and during that method some sub-events are also logged. If the parent
event info is empty, then this event is a "root" event. Sub-events always have a duration that is <= the duration of the parent event.
* Duration: How long the event/operation took.

Example metrics event:

* Event: Retrieving token via an HTTP POST request to example.com
* Parent Event: Requesting credential(s) from issuer
* Duration: 6.37ms

### Example event timeline - OpenID4CI flow

Note: performance numbers given below are illustrative only and are not intended to be representative of real-world performance.
```
                                                     Dashed line indicates event duration in timeline, towards the right is further ahead in time - not to scale
                                
 Request credential(s) from issuer (parent event) --------------------------------------------------------- (5.14s)
 Fetch OpenID config (GET)                (event)     ----(2.92ms)
 Fetch token (POST)                       (event)             ----(6.37ms)
 Fetch issuer metadata (GET)              (event)                 ---(919.15µs)
 Fetch credential (POST)                  (event)                      ------(28.87ms)
 Parsing and checking proof               (event)                            ------------------------------ (5.10s)
```
                              
Note that the sum of all sub-event durations may not add to the duration of the parent event, as not every possible operation during the parent event will be timed.
Generally, short/trivial operations are not tracked.
