# SDK Usage

This guide explains how to use this SDK in Android or iOS code. The examples in this document demonstrate
how to use the various APIs from a Kotlin and Swift perspective.

The languages for the generated bindings are Java (for Android) and Objective-C (for iOS).

For the sake of readability, the following is omitted in most code examples:
* Error handling
* Nullable type checks (in Kotlin examples)
* Optionals unwrapping (in Swift examples)

## General API Patterns

### Function Arguments

Wallet-SDK APIs follow a certain pattern. Arguments to functions and methods are always mandatory unless
they are in an `Opts` object. If a function or method has an `Opts` argument, it will be the last
argument (after all the mandatory arguments). The exact name of this object may differ depending on the function/method,
but they will all have `Opts` in the name.

`Opts` objects are created by calling their constructors, which always have zero arguments.
To set specific options, use the exposed methods available on the object after you've created it.

If an optional argument is not explicitly set then a default will be used.

If no optional arguments are needed, you can simply pass in `null`/`nil` as the `Opts` argument, which will
result in defaults being used for all the options.

Note that in some circumstances, some options may be required (e.g. if an issuer requires a PIN). See the documentation
for the particular function/method for more information.

### Arrays

Due to a limitation of gomobile, the Wallet-SDK mobile bindings can't use array types directly. As a workaround,
Wallet-SDK instead exposes "array-like" objects that internally contain the desired array while exposing
gomobile-compatible methods for performing array operations.

These "array-like" objects can be identified by their names. They will either have "Array" as a suffix in their names,
or be plural nouns (e.g. `SupportedCredentials`).

These objects will have the following methods:
* `length()`: Returns the number of elements in the array.
* `atIndex(index)`: Returns the element at the given `index`. If `index` is out of bounds, then null/nil/an empty value
is returned instead.

Additionally, some array object types may also support being created directly by the caller (so that they can be used as
a function argument), in which case they will also have constructors and the following method:
* `append(arg)`: appends `arg` to the array. It also returns a reference to the array object to allow for
chaining together multiple `append` calls.

In subsequent sections of this documentation, any reference to "arrays" refer to the "array-like" objects that are
described above.

## Error Handling

Errors from Wallet-SDK come in a structured format that can be (optionally) parsed, allowing for individual fields to
be accessed.

Errors have five fields:

* Category: A short descriptor of the general category of error. This will always be a pre-defined string.
* Code: A short alphanumeric code that is used to broadly group various related errors.
This contains the Category descriptor.
* Details: The full, raw error message. Any lower-level details about the precise cause of the error will be captured
here.
* Message: A short message describing the error that occurred. Only present in certain cases. It will be blank in all
others.
* TraceID: ID of Open Telemetry root trace. Can be used to trace API calls. Only present on errors return from certain
APIs. It will be blank if there is no trace ID available.

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
    try someSDKObject.someMethodHere()
} catch let error as NSError {
    let parsedError = WalleterrorParse(error.localizedDescription)
    // Access parsedError.code, parsedError.category, and parsedError.details as you see fit
}
```

## API Package

The API package contains various interfaces and types that are used across multiple places in the SDK.
Wallet-SDK contains implementations of these various interfaces that can be used to get an application up and running
fairly quickly.
Implementations of these interfaces can be done in on the mobile side and injected in to the API methods that use them.
A few examples of when you might want to do this:
* Implementing platform-specific crypto functionality (perhaps leveraging device-specific security chips)
* Implement a custom LD document loader that uses preloaded local contexts for performance and/or security reasons.
See [LD Document Loading](#ld-document-loading) for more information.

### KeyWriter and KeyReader

The KeyWriter and KeyReader interfaces appear in various places across Wallet-SDK.

The `KeyWriter` interface has a single method called `create` defined on it which accepts a `keyType` argument.
Wallet-SDK expects that `KeyWriter` implementations also store the keys they create for later retrieval.

The `KeyReader` interface has a single method called `exportPubKey` defined on it which accepts a `keyID` argument.
Wallet-SDK expects this method to return the public key associated with the given `keyID` argument.

If either one of these behaviours is not implemented, then certain Wallet-SDK functionality will not work as expected:
* DID creation, credential issuance, and credential presentation will fail since keys won't be found


### Verifiable Credentials

Verifiable Credential objects are present throughout Wallet-SDK. They have a number of useful methods:

* `serialize()`: Returns a serialized representation of this VC. This is useful for storing VCs in a persistent
database. To convert the serialized representation back into a `Credential` object,
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

The `parseCredential` function is located in the `Verifiable` package. It has a few optional arguments that can be
passed in via the methods available on the `Opts` object:
* `disableProofCheck`: Disables the proof check operation when parsing the VC. The proof check can be an expensive
operation - in certain scenarios, it may be appropriate to disable the check. By default, proofs are checked.
* `setDocumentLoader`: Specifies a JSON-LD document loader to use when parsing the VC. If none is specified, then a
network-based loader will be used.

Passing in `null`/`nil` will cause all default options to be used.

### Examples

#### Kotlin (Android)

##### Using Default Options

```kotlin
import dev.trustbloc.wallet.sdk.verifiable.Verifiable

val vc = Verifiable.parseCredential("Serialized VC goes here", null)
```

##### Using Specified Options

```kotlin
import dev.trustbloc.wallet.sdk.verifiable.Opts
import dev.trustbloc.wallet.sdk.verifiable.Verifiable

val opts = Opts().disableProofCheck()
val vc = Verifiable.parseCredential("Serialized VC goes here", opts)
```

#### Swift (iOS)

##### Using Default Options

```swift
import Walletsdk

var error: NSError?
let vc = VerifiableParseCredential("Serialized VC goes here", nil, &error)
```

##### Using Specified Options

```swift
import Walletsdk

var error: NSError?
let opts = VerifiableNewOpts()?.disableProofCheck()
let vc = VerifiableParseCredential("Serialized VC goes here", opts, &error)
```

## LD Document Loading

Several APIs allow for an LD document loader to be specified via the `setDocumentLoader` method on their respective
`Opts` objects.
If no custom LD document loader is specified (or is null/nil), then network-based document loading will be used instead.
For performance and/or security reasons, you may wish to implement a custom LD document loader that uses
preloaded local contexts.

## Network Call Timeout

By default, REST calls in Wallet-SDK have a timeout of 30 seconds. This can be overridden by passing in a
custom timeout via the `setHTTPTimeoutNanoseconds` method, which, if available, will be on the API's `Opts` object.
Passing in 0 will disable timeouts.

## In-Memory Credential Storage

The credential package contains an in-memory credential storage implementation that can be used to store credentials
in memory and also satisfy the `Reader` interface (in the `Credential` package). As it only uses in-memory storage,
it's generally only suitable for testing purposes. You will probably want to create your own implementation in your
mobile code that uses platform-specific storage.

### Examples

#### Kotlin (Android)

```kotlin
import dev.trustbloc.wallet.sdk.credential.InMemoryDB
import dev.trustbloc.wallet.sdk.verifiable.Verifiable

val db = InMemoryDB()

val vc = Verifiable.parseCredential("Serialized VC goes here", null)

db.add(vc)

val retrievedVC = db.get("VC_ID")

val retrievedVCs = db.getAll()

db.remove("VC_ID")
```

#### Swift (iOS)

```swift
import Walletsdk

let db = CredentialNewInMemoryDB()

var error: NSError?
let vc = VerifiableParseCredential("Serialized VC goes here", nil, &error)!

db.add(vc)

let retrievedVC = db.get("VC_ID")

let retrievedVCs = db.getAll()

db.remove("VC_ID")
```

## Local KMS
This package contains a local KMS implementation that uses Google's Tink crypto library. The caller must inject a key
store for the KMS to use. Keys returned by local KMS are returned as `JWK` objects. The key IDs used for storage
are assigned by local KMS and can be determined by using the `id()`  on the `JWK` object. Note: in the iOS
bindings, the method is called `id_()` instead due to a technical limitation.

Wallet-SDK includes an in-memory key store implementation that you can use for getting up and running with local KMS.
It'll allow you to test out its functionality, but in order to ensure your keys aren't lost when your app closes,
you'll want to inject in your own persistent implementation at some point.

Note that regardless of what storage implementation you use, private keys may intermittently reside in local memory
when using local KMS, so keep this limitation in mind.

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

var error: NSError?
let kms = LocalkmsNewKMS(memKMSStore, &error)

let jwk1 = kms.create(LocalkmsKeyTypeED25519)
let jwk2 = kms.create(LocalkmsKeyTypeP384)
```

## DID Creation

Wallet-SDK provides support for creating `key`, `jwk`, or `ion` (longform) DIDs.

To create a DID, call the corresponding `create` function with a key of your choice.

Note: currently, in order to create a DID in Wallet-SDK, you must use a key created by a [local KMS](#local-kms)
instance, which will be in a `JWK` object.

### Examples

The examples below show the full process of creating a DID, including the creation of a LocalKMS
instance (using in-memory storage) and a key. Each example shows how to create `key`, `jwk`, or `ion` (longform) DIDs.

#### Kotlin (Android)

```kotlin
import dev.trustbloc.wallet.sdk.didion.Didkey
import dev.trustbloc.wallet.sdk.didion.Didjwk
import dev.trustbloc.wallet.sdk.didion.Didion
import dev.trustbloc.wallet.sdk.localkms.Localkms
import dev.trustbloc.wallet.sdk.localkms.MemKMSStore

val memKMSStore = MemKMSStore.MemKMSStore()
val kms = Localkms.newKMS(memKMSStore)

val jwk1 = kms.create(Localkms.KeyTypeED25519)
val didDocument1 = Didkey.create(jwk1)

// It's considered a best practice to create a new key for every new DID you want to create.

val jwk2 = kms.create(Localkms.KeyTypeED25519)
val didDocument2 = Didjwk.create(jwk2)

val jwk3 = kms.create(Localkms.KeyTypeED25519)
val didDocument3 = Didion.createLongForm(jwk3)
```

#### Swift (iOS)

```swift
import Walletsdk

let memKMSStore = LocalkmsNewMemKMSStore()

var newKMSError: NSError? //Be sure to actually check this error in real code.
let kms = LocalkmsNewKMS(memKMSStore, &newKMSError)

let jwk1 = try kms.create(LocalkmsKeyTypeED25519)

var didKeyCreateError: NSError? //Be sure to actually check this error in real code.
let didDocument1 = DidkeyCreate(jwk1, &didKeyCreateError)

// It's considered a best practice to create a new key for every new DID you want to create.

let jwk2 = try kms.create(LocalkmsKeyTypeED25519)

var didJWKCreateError: NSError? //Be sure to actually check this error in real code.
let didDocument2 = DidjwkCreate(jwk2, &didJWKCreateError)

let jwk3 = try kms.create(LocalkmsKeyTypeED25519)

var didIONCreateError: NSError? //Be sure to actually check this error in real code.
let didDocument3 = DidionCreateLongForm(jwk3, &didIONCreateError)
```

### Error Codes & Troubleshooting Tips

| Error                             | Possible Reasons                                                                                          |
|-----------------------------------|-----------------------------------------------------------------------------------------------------------|
| CREATE_DID_KEY_FAILED(DID1-0000)  | Failed to create a key during DID creation. Check your KeyWriter implementation.                          |
| CREATE_DID_ION_FAILED(DID1-0001)  | Failed to create a key during DID creation. Check your KeyWriter implementation.                          |
| CREATE_DID_JWK_FAILED(DID1-0002)  | Failed to create a key during DID creation. Check your KeyWriter implementation.                          |
| UNSUPPORTED_DID_METHOD(DID0-0003) | The DID method specified in the `create` call is unsupported. Only `key`, `ion`, and `jwk` are supported. |

## DID Resolver

A DID resolver can be used to resolve DIDs.

A resolver can be created with an optional resolver server URI, which will only be used for resolving DIDs that
require a remote resolver server.

### Examples

#### Kotlin (Android)

##### Using Default Options

```kotlin
import dev.trustbloc.wallet.sdk.did.*

val didResolver = did.Resolver(null)

val didDocument = didResolver.resolve("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
```

##### Using Specified Options

```kotlin
import dev.trustbloc.wallet.sdk.did.*

val opts = ResolverOpts().setResolverServerURI("https://example.com/")
val didResolver = did.Resolver(opts)

val didDocument = didResolver.resolve("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
```

#### Swift (iOS)

##### Using Default Options

```swift
import Walletsdk

let didResolver = DidNewResolver(nil)

let didDocument = didResolver.resolve("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
```

##### Using Specified Options

```swift
import Walletsdk

let resolverOpts = DidNewResolverOpts()?.setResolverServerURI("https://example.com/")
let didResolver = DidNewResolver(resolverOpts)

let didDocument = didResolver.resolve("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
```

### Error Codes & Troubleshooting Tips

| Error                                         | Possible Reasons                                                                                                                                                                                                                                            |
|-----------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
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

val didResolver = Resolver(null)

val validationResult = Did.validateLinkedDomains("YourDIDHere", didResolver)
```

#### Swift (iOS)

```swift
import Walletsdk

let didResolver = DidNewResolver(nil)

let validationResult = ValidateLinkedDomains("YourDIDHere", didResolver)
```

## OpenID4CI

The OpenID4CI package contains an API that can be used by a [holder](https://www.w3.org/TR/vc-data-model/#dfn-holders)
to go through the [OpenID4CI](https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0-11.html) flow.

The way you use the API differs somewhat depending on whether you're doing an issuer-initiated or wallet-initiated
issuance flow. In an issuer-initiated flow, the issuer has already provided a credential offer, while in a wallet-initiated
flow, the wallet only knows the issuer URI. The sections below will elaborate on the API differences where applicable.

### Creating the `Interaction` Object

The first step in any OpenID4CI flow is to create an `Interaction` object. There are two types of `Interaction` objects
in the SDK: `IssuerInitiatedInteraction` and `WalletInitiatedInteraction`. These two types have similar methods on them
and act similarly. From this point on, this documentation will simply refer to both of these types as `Interaction`
objects unless there's a need to differentiate between them.

An `Interaction` object is instantiated using an `IssuerInitiatedInteractionArgs`/`WalletInitiatedInteractionArgs`
object and, optionally, an `InteractionOpts` object. The `Interaction` object is stateful.
Use it for only one single instance of an OpenID4CI flow and then discard it.
Create a new `Interaction` object _every time_ you go through the OpenID4CI flow.

An `IssuerInitiatedInteractionArgs` object contains the following mandatory parameters:
* An initiate issuance URI obtained from an issuer (e.g. via a QR code) which contains a credential offer.
* A crypto implementation.
* A DID resolver.

A `WalletInitiatedInteractionArgs` object contains the following mandatory parameters:
* An issuer URI.
* A crypto implementation.
* A DID resolver.

To set optional arguments, create an `InteractionOpts` object and use the supplied methods available on that object.
The following methods (options) are available:
* `setActivityLogger`: If set, then credential activity logs will be passed to the given `ActivityLogger`
  implementation. Otherwise, no activities will be logged.
* `setDocumentLoader`: Used when parsing JSON-LD VCs. If unspecified, then a network-based loader will be used, which
  can impact performance.
* `setMetricsLogger`: If set, then performance metrics events will be passed to the given `MetricsLogger`
  implementation. Otherwise, no metrics events will be logged.
* `setHTTPTimeoutNanoseconds`: Specifies a timeout for REST calls. Defaults to 30 seconds if not set.
* `disableOpenTelemetry`: Disables the sending of OpenTelemetry headers. By default, they are sent during REST calls
  to the issuer and during OAuth2-related calls (if applicable).
* `addHeader`: Allows you to set an additional header to be sent during REST calls to the issuer and during
OAuth2-related calls (if applicable).
* `addHeaders`: Like `addHeader`, but allows you to specify multiple headers at once.
* `disableHTTPClientTLSVerify`: Disables TLS verification for HTTPS calls. **This should
  be used for testing purposes only**.
* `disableVCProofChecks`: Disables proof checks for Verifiable Credentials received from the issuer. **This should
be used for testing purposes only**.

If you're okay with all defaults being used, you can simply pass in `null`/`nil` in as the `InteractionOpts` argument.

Note that options can also be set by chaining the methods together on a single line.
For example: `newInteractionOpts().setActivityLogger(...).setHeaders(...)`

### Authorization

In this part, the actions you have to take vary greatly depending on whether you're using the Pre-Authorized Code flow
or the Authorization Code flow.

If you are using a `WalletInitiatedInteraction`, then only the Authorization Code flow is available. Skip directly to
the [Authorization Code flow](#authorization-code-flow) section.

If you are using an `IssuerInitiatedInteraction`, then you should first check and see what grant types the issuer says
they support before proceeding. You can do this by checking the issuer's capabilities, which you can do by using the
relevant methods on the `Interaction` object::
* `preAuthorizedCodeGrantTypeSupported`: Indicates whether the issuer supports the pre-authorized code grant type. If it
 does, then you can proceed with the [pre-authorized code flow](#pre-authorized-code-flow).
* `authorizationCodeGrantTypeSupported`: Indicates whether the issuer supports the authorization code grant type. If it
  does, then you can proceed with the [authorization code flow](#authorization-code-flow).

#### Pre-Authorized Code Flow

For the pre-authorized code flow, you need to determine whether the issuer requires a PIN or not. To do this, first get
the `PreAuthorizedCodeGrantParams` object by calling the `preAuthorizedCodeGrantParams` method on your
`Interaction` object. Then, use the `pinRequired` method to determine whether a PIN is needed or not. Once you
know this, you're ready to [request credentials](#request-credential).

#### Authorization Code Flow

If you are using a `WalletInitiatedInteraction`, then you may want to check what credentials the issuer
supports. See [Issuer Metadata](#issuer-metadata) for more information. Note that if you're using an
`IssuerInitiatedInteraction` then the credential format+types are already pre-specified by the issuer in the credential
offer and can't be overridden.

To begin the Authorization Code flow, you need to create an authorization URL. To do this, call the
`createAuthorizationURL` method on the`Interaction` object. You need to provide the following parameters:
* Client ID: This is a value that's specific to your application and/or issuer. See [Client ID](#client-id) for more
information about this parameter.
[Dynamic Client Registration](#dynamic-client-registration), .
* Redirect URI: The URI that you want the service to redirect to after authorization is complete.
You will likely want this to be some sort of deep link to your app. More information on this can be found further below.

Additionally, if you're using a `WalletInitiatedInteraction` object, then there are two additional required
parameters:
* Credential Format: The format of the credential you want to receive from the issuer by the end of the flow.
* Credential Types: The types of the credential you want to receive from the issuer by the end of the flow.

Additionally, some issuers' authorization servers may require scopes to be passed in. In this case, use the `setScopes`
method on the `CreateAuthorizationURLOpts` object to pass in scopes. The `scopes` value is a part of the
OAuth2 specification and needs to be obtained by out-of-band means or via
[Dynamic Client Registration](#dynamic-client-registration).

Once you have your authorization URL, load it in a web browser. The user will then need to log in to the service
(if they are not already) and give permission to share their data with the issuer. The web page will then
redirect the user to the redirect URI that you passed in previously. However, this redirect URI will now have
additional query parameters appended to it. You will need this full URI for the next step, and so you will probably
want the redirect URI (passed in earlier) to be some sort of deep link to your app so that you can easily retrieve this
URI.

You're now ready to [request credentials](#request-credential).

### Request Credential

If you're using a `WalletInitiatedInteraction` object, then there is only one `requestCredential` method, and it has
the same arguments and behaviour as the `requestCredentialWithAuth` method on the `IssuerInitiatedInteraction` object,
which is described below.

If you're using an `IssuerInitiatedInteraction` object, then there are two methods available to request credentials.
Which one you use will depend on your flow:
* `requestCredentialWithPreAuth`: Use this only if you're using the pre-authorized code flow. If a PIN is required,
pass it in using the `setPIN` method on the `RequestCredentialWithPreAuthOpts` object.
* `requestCredentialWithAuth`: Use this only if you're using the Authorization Code flow. This version of the method
  will require the redirect URI (that has additional query parameters) that you got before.

Regardless of which of the two methods you use, if the call succeeds, it will return your issued credentials.
These can then be used in other Wallet-SDK APIs or [serialized for storage](#verifiable-credentials).

### Issuer Metadata

Depending on your use case, you may want to check the issuer's metadata. To do this, call the `issuerMetadata` method
which will return an `IssuerMetadata` object. Note that some of the methods available on this object and its nested
sub-objects, can return null/nil/an empty value if the issuer does not supply the corresponding data in its metadata.
You should ensure that these situations are handled accordingly. Read this section carefully to see which methods may
return null/nil/an empty value.

The `IssuerMetadata` object has the following methods on it:

* `credentialIssuer()`: Returns the issuer's identifier.
* `localizedIssuerDisplays()`: Returns the [LocalizedIssuerDisplays](#LocalizedIssuerDisplays) object, which contains
display information for the issuer.
* `supportedCredentials()`: Returns the [SupportedCredentials](#SupportedCredentials) object, which contains technical
and display data about the credential types that this issuer can issue.

#### LocalizedIssuerDisplays

The `LocalizedIssuerDisplays` object contains display information for the issuer. This is optional data; the issuer
may not provide any display information. If the issuer doesn't provide any display information at all, then the
`length()` method on the `LocalizedIssuerDisplays` object will return 0.
Each element (`LocalizedIssuerDisplay`) in the array corresponds to a different locale and has methods on it that return
various pieces of display information. Note that this display information is all optional, and so any of these methods
may return null/nil/an empty value if the issuer doesn't provide it.

The methods on a `LocalizedIssuerDisplay` are:

* `locale()`: The locale for this `LocalizedIssuerDisplay`.
* `name()`: The issuer's name.
* `url()`: The issuer's URL.
* `logo()`: The issuer's logo.
* `backgroundColor()`: The issuer's background color value.
* `textColor()`: The issuer's text color value.

#### SupportedCredentials

The `SupportedCredentials` object contains information about the credential types that the issuer can issue.
Each element (`SupportedCredential`) in the array corresponds to a different credential type.

The methods on a `SupportedCredential` are:

* `localizedDisplays()`: Returns the [LocalizedCredentialDisplays](#LocalizedCredentialDisplays) object, which contains
display info for this credential type.

The next three methods return lower-level technical data that aren't intended for displaying to an end-user in most
cases, but they may be useful for other Wallet-SDK APIs:
* `id()`: The credential's ID.
* `format()`: The credential format.
* `types()`: The credential's types array. This is the low-level `types` array that appears in the actual issued
Verifiable Credential object. Not to be confused with the general idea of a credential "type" (e.g. A university degree,
driver's license, etc).

Note: while unexpected, if the issuer doesn't specify any credential types at all, then the `length()` method on the
`SupportedCredentials` object will return 0.

##### LocalizedCredentialDisplays

The `LocalizedCredentialDisplays` object contains display information for a credential type. This is optional data;
the issuer may not provide any display information. If the issuer doesn't provide any display information at all, then
the`length()` method on the `LocalizedCredentialDisplays` object will return 0.
Each element (`LocalizedCredentialDisplay`) in the array corresponds to a different locale and has methods on it that return
various pieces of display information. Note that this display information is all optional, and so any of these methods
may return null/nil/an empty value if the issuer doesn't provide it.

The methods on a `LocalizedCredentialDisplay` are:

* `locale()`: The locale for this `LocalizedCredentialDisplay`.
* `name()`: The credential's name.
* `logo()`: The credential's logo.
* `backgroundColor()`: The credential's background color value.
* `textColor()`: The credential's text color value.

### Client ID

This section only applies to the Authorization Code flow. Note that all `WalletInitiatedInteraction` flows use
the Authorization Code flow, but in the case of an `IssuerInitiatedInteraction`, it may or may not use the Authorization
Code Flow depending on whether the issuer and/or your app uses it. See [Authorization](#authorization) for more
information.

In the Authorization Code flow, a client ID must be provided in order to create the authorization URL. The client ID
parameter is a part of the core OAuth 2.0 specification and can be obtained via several different means.
Depending on the method you use, Wallet-SDK may have functionality to help with this process. The following sections
describe some different methods that are explicitly supported by Wallet-SDK. Note that other methods not covered here
may work (depending on how they're designed), but are untested.

#### Dynamic Client Registration

[Dynamic client registration](https://datatracker.ietf.org/doc/html/rfc7591) is a mechanism that allows a client
to dynamically register with an authorization server in order to receive a client ID.

To determine if the issuer supports dynamic client registration, call the `dynamicClientRegistrationSupported` method
on your `Interaction` object. If the issuer does support dynamic client registration, then you will next need to
retrieve the issuer's dynamic client registration endpoint using `dynamicClientRegistrationEndpoint`.

Finally, pass the returned endpoint into the `registerClient` function along with any other parameters as required for
your use case. For Android, this function is located in the`oauth2` package. For iOS, this function and the other types
used with the function are prefixed with `oauth2`.

The `registerClient` function will, upon successful registration, return the registered metadata
along with a client ID. You should then be able to use this client ID when you go to create the authorization URL.
You may also want to (or need to) use the registered scopes from the registered metadata depending on the
issuer and/or authorization server.

#### OAuth Discoverable Client ID Scheme

[The Client ID Scheme](https://mattrglobal.github.io/draft-looker-oauth-client-id-scheme/draft-looker-oauth-client-id-scheme.html)
is an extensibility point to the OAuth 2.0 specification that allows a client to use a Client ID that is not assigned
or managed by the authorization server. The OAuth Discoverable Client ID Scheme is a specific Client ID scheme that
is also described in the same specification.

The OAuth Discoverable Client ID Scheme has certain pre-requisites that the client must meet that are out-of-scope for
Wallet-SDK. For example, the client needs to host an endpoint that can be accessed by the authorization server, as
well as having some requirements on the chosen client ID (that you need to pass into the `createAuthorizationURL`
method). Also, the issuer must support the OAuth Discoverable Client ID Scheme. Note that Wallet-SDK cannot tell you
if an issuer supports the OAuth Discoverable Client ID Scheme.

Assuming these pre-requisites are met, then this scheme can be used with Wallet-SDK. To do so, create a
`CreateAuthorizationURLOpts` object, call the`useOAuthDiscoverableClientIDScheme` method on it, and pass your
`CreateAuthorizationURLOpts` object into the`createAuthorizationURL` method. Be sure to also set your client ID
appropriately as required by the scheme.

To determine if the issuer supports dynamic client registration, call the `dynamicClientRegistrationSupported` method
on your `Interaction` object. If the issuer does support dynamic client registration, then you will next need to
retrieve the issuer's dynamic client registration endpoint using `dynamicClientRegistrationEndpoint`.

#### Pre-Registration (Out-of-Band)

"Traditional" OAuth 2.0 requires that the client ID is obtained by doing some sort of pre-registration process that is
specific to the issuer/authorization server. The pre-registration part is out of scope for Wallet-SDK, but the client ID
you get from that process can be passed into the `createAuthorizationURL` method in order to proceed with the issuance
flow.

### Issuer URI Method

If you're using an `IssuerInitiatedInteraction` object, then there is an additional optional `issuerURI()` method you
may wish to use. The issuer URI can be used to get the display data for your credentials.
See [Credential Display Data](#credential-display-data) for more information.
You may want to save the issuer URI somewhere so that you can get/refresh display data at a later time.

This method isn't required on the `WalletInitiatedInteraction` since the caller already has the issuer URI to
begin with.

### Verifying the Issuer

At any point during an interaction flow, you can call `verifyIssuer()` on your interaction object. This will return a
service URL if the issuer was verified. An error means that either the issuer failed the verification check, or
something went wrong during the process (and so a verification status could not be determined).

### Examples

The following examples show how to use the APIs to go through the OpenID4CI flow using the iOS and Android bindings.
They use in-memory key storage and the Tink crypto library.

#### Kotlin (Android) - Issuer-Initiated - Pre-Authorized Code Flow

```kotlin
import dev.trustbloc.wallet.sdk.didion.Didkey
import dev.trustbloc.wallet.sdk.localkms.Localkms
import dev.trustbloc.wallet.sdk.localkms.MemKMSStore
import dev.trustbloc.wallet.sdk.did.Resolver
import dev.trustbloc.wallet.sdk.openid4ci.*
import dev.trustbloc.wallet.sdk.verifiable.CredentialsArray

// Setup
val memKMSStore = MemKMSStore.MemKMSStore()
val kms = Localkms.newKMS(memKMSStore)
val didResolver = Resolver(null)

val jwk = kms.create(Localkms.KeyTypeED25519)
val didDocument = Didkey.create(jwk) // Create a did:key document

val activityLogger = mem.ActivityLogger()

// Going through the flow
val interactionArgs = IssuerInitiatedInteractionArgs("YourInitiateIssuanceURIHere", kms.getCrypto(), didResolver)
val interactionOpts = InteractionOpts().setActivityLogger(activityLogger) // Optional, but useful for tracking credential activity
val interaction = IssuerInitiatedInteraction(interactionArgs, interactionOpts) // This is a stateful object - we will use this to go through the flow.

// It's a good idea to check the issuer's capabilities first
if (!interaction.preAuthorizedCodeGrantTypeSupported()) {
    // This code example isn't applicable. See the authorization code flow example instead.
    
    return
}

val requestCredentialWithPreAuthOpts = RequestCredentialWithPreAuthOpts()

if (interaction.preAuthorizedCodeGrantParams().pinRequired()) {
    requestCredentialWithPreAuthOpts.setPIN("1234")
}

val credentials = interaction.requestCredentialWithPreAuth(didDocument.assertionMethod(), requestCredentialWithPreAuthOpts)

val issuerURI = interaction.issuerURI() // Optional (but useful)

// Issuance Acknowledgment
interaction.requireAcknowledgment() // returns true, if the issuer requires acknowledgment that credentials are accepted or rejected by the user.
interaction.acknowledgment() // get the Acknowledgment object, if issuer requires ack.

// use this API to get the opaque string, which needs to be submitted to the Acknowledgment APIs
val ackTkn = interaction.acknowledgment().serialize()

val ack = Acknowledgment(ackTkn)

ack.success() // user accepts the credential
ack.reject() // user rejects the credential

// Consider checking the activity log at some point after the interaction
```

#### Swift (iOS) - Issuer-Initiated - Pre-Authorized Code Flow

```swift
import Walletsdk

// Setup
let memKMSStore = LocalkmsNewMemKMSStore()

var newKMSError: NSError? //Be sure to actually check this error in real code
let kms = LocalkmsNewKMS(memKMSStore, &newKMSError)

let didResolver = DidNewResolver(nil)

let jwk = try kms.create(LocalkmsKeyTypeED25519)

var didKeyCreateError: NSError? //Be sure to actually check this error in real code
let didDocument = DidkeyCreate(jwk, &didKeyCreateError)// Create a did:key document

let activityLogger = MemNewActivityLogger()

// Going through the flow
let interactionArgs = Openid4ciNewIssuerInitiatedInteractionArgs("YourInitiateIssuanceURIHere", kms.getCrypto(), didResolver)
let interactionOpts = Openid4ciNewInteractionOpts().setActivityLogger(activityLogger) // Optional, but useful for tracking credential activity
var newInteractionError: NSError?

// This is a stateful object - we will use this to go through the flow.
let interaction = Openid4ciNewIssuerInitiatedInteraction(interactionArgs, interactionOpts, &newInteractionError)

// It's a good idea to check the issuer's capabilities first
if !interaction.preAuthorizedCodeGrantTypeSupported() {
    // This code example isn't applicable. See the authorization code flow example instead.
    
    return
}

let requestCredentialWithPreAuthOpts = Openid4ciRequestCredentialWithPreAuthOpts()

if (issuerCapabilities.preAuthorizedCodeGrantParams().pinRequired()) {
    requestCredentialWithPreAuthOpts!.setPIN("1234")
}

let credentials = interaction.requestCredential(withPreAuth: didDocument.assertionMethod(), opts: requestCredentialWithPreAuthOpts)

let issuerURI = interaction.issuerURI() // Optional (but useful)

// Issuance Acknowledgment
interaction.requireAcknowledgment() // returns true, if the issuer requires acknowledgment that credentials are accepted or rejected by the user.

// use this API to get the opaque string, which needs to be submitted to the Acknowledgment APIs
let ackTkn = interaction.Acknowledgment().Serialize()

let ack = Openid4ciNewAcknowledgment(ackTkn)

try ack.success() // user accepts the credential
try ack.reject() // user rejects the credential

// Consider checking the activity log at some point after the interaction
```

#### Kotlin (Android) - Issuer-Initiated - Authorization Code Flow

```kotlin
import dev.trustbloc.wallet.sdk.api.*
import dev.trustbloc.wallet.sdk.didion.Didkey
import dev.trustbloc.wallet.sdk.localkms.Localkms
import dev.trustbloc.wallet.sdk.localkms.MemKMSStore
import dev.trustbloc.wallet.sdk.did.Resolver
import dev.trustbloc.wallet.sdk.openid4ci.*
import dev.trustbloc.wallet.sdk.verifiable.CredentialsArray

// Setup
val memKMSStore = MemKMSStore.MemKMSStore()
val kms = Localkms.newKMS(memKMSStore)
val didResolver = Resolver(null)

val jwk = kms.create(Localkms.KeyTypeED25519)
val didDocument = Didkey.create(jwk) // Create a did:key document

val activityLogger = mem.ActivityLogger()

// Going through the flow
val interactionArgs = IssuerInitiatedInteractionArgs("YourInitiateIssuanceURIHere", kms.getCrypto(), didResolver)
val interactionOpts = InteractionOpts().setActivityLogger(activityLogger) // Optional, but useful for tracking credential activity
val interaction = IssuerInitiatedInteraction(interactionArgs, interactionOpts) // This is a stateful object - we will use this to go through the flow.

// It's a good idea to check the issuer's capabilities first
if (!interaction.authorizationCodeGrantTypeSupported()) {
    // This code example isn't applicable. See the pre-authorized code flow example instead.
    
    return
}

val scopes = StringArray()
scopes.append("scope1").append("scope2")

val createAuthorizationURLOpts = CreateAuthorizationURLOpts().setScopes(scopesArr)

// If scopes aren't needed, you can pass in null in place of the opts argument.
val authorizationLink := interaction.createAuthorizationURL("clientID", "redirect URI", createAuthorizationURLOpts)

// Open authorizationLink in a browser. Once the user has finished logging in, call requestCredentialWithAuth()
// with the full redirect URI (including query parameters) that the login service sent the user to.
// The code below assumes this has already been done somehow and that the URI is in the redirectURIWithParams variable.
// In actual code, the call to requestCredentialWithAuth() couldn't be immediately after the
// createAuthorizationURL() call like in this example since control has to flow back to the user first.
val redirectURIWithParams = "Put the redirect URI with params here"

val credentials = interaction.requestCredentialWithAuth(didDocument.assertionMethod(), redirectURIWithParams, null)

val issuerURI = interaction.issuerURI() // Optional (but useful)

// Consider checking the activity log at some point after the interaction
```

#### Swift (iOS) - Issuer-Initiated - Authorization Code Flow

```swift
import Walletsdk

// Setup
let memKMSStore = LocalkmsNewMemKMSStore()

var newKMSError: NSError? //Be sure to actually check this error in real code
let kms = LocalkmsNewKMS(memKMSStore, &newKMSError)

let didResolver = DidNewResolver(nil)

let jwk = try kms.create(LocalkmsKeyTypeED25519)

var didKeyCreateError: NSError? //Be sure to actually check this error in real code
let didDocument = DidkeyCreate(jwk, &didKeyCreateError)// Create a did:key document

let activityLogger = MemNewActivityLogger()

// Going through the flow
let interactionArgs = Openid4ciNewIssuerInitiatedInteractionArgs("YourInitiateIssuanceURIHere", kms.getCrypto(), didResolver)
let interactionOpts = Openid4ciNewInteractionOpts().setActivityLogger(activityLogger) // Optional, but useful for tracking credential activity
var newInteractionError: NSError?

// This is a stateful object - we will use this to go through the flow.
let interaction = Openid4ciNewIssuerInitiatedInteraction(interactionArgs, interactionOpts, &newInteractionError)

// It's a good idea to check the issuer's capabilities first
if !interaction.authorizationCodeGrantTypeSupported() {
    // This code example isn't applicable. See the authorization code flow example instead.
    
    return
}

let scopes = ApiStringArray()
scopes.append("scope1").append("scope2")

let opts = Openid4ciNewCreateAuthorizationURLOpts()!.setScopes(scopes)

var createAuthURLError: NSError?

// If scopes aren't needed, you can pass in nil in place of the opts argument.
let authorizationLink = interaction.createAuthorizationURL("clientID", redirectURI: "redirect URI", opts: opts, error: &createAuthURLError)

// Open authorizationLink in a browser. Once the user has finished logging in, call requestCredential()
// with the full redirect URI (including query parameters) that the login service sent the user to.
// The code below assumes this has already been done somehow and that the URI is in the redirectURIWithParams variable.
// In actual code, the call to requestCredential() couldn't be immediately after the
// createAuthorizationURL() call like in this example since control has to flow back to the user first.
let redirectURIWithParams = "Put the redirect URI with params here"

let credentials = interaction.requestCredential(withAuth: didDocument.assertionMethod(), redirectURIWithParams: redirectURIWithParams, opts: nil)

let issuerURI = interaction.issuerURI() // Optional (but useful)

// Consider checking the activity log at some point after the interaction
```

#### Kotlin (Android) - Wallet-Initiated

```kotlin
import dev.trustbloc.wallet.sdk.api.*
import dev.trustbloc.wallet.sdk.didion.Didkey
import dev.trustbloc.wallet.sdk.localkms.Localkms
import dev.trustbloc.wallet.sdk.localkms.MemKMSStore
import dev.trustbloc.wallet.sdk.did.Resolver
import dev.trustbloc.wallet.sdk.openid4ci.*
import dev.trustbloc.wallet.sdk.verifiable.CredentialsArray

// Setup
val memKMSStore = MemKMSStore.MemKMSStore()
val kms = Localkms.newKMS(memKMSStore)
val didResolver = Resolver(null)

val jwk = kms.create(Localkms.KeyTypeED25519)
val didDocument = Didkey.create(jwk) // Create a did:key document

val activityLogger = mem.ActivityLogger()

// Going through the flow
val interactionArgs = WalletInitiatedInteractionArgs("IssuerURIHere", kms.getCrypto(), didResolver)
val interactionOpts = InteractionOpts().setActivityLogger(activityLogger) // Optional, but useful for tracking credential activity
val interaction = WalletInitiatedInteraction(interactionArgs, interactionOpts) // This is a stateful object - we will use this to go through the flow.

val scopes = StringArray()
scopes.append("scope1").append("scope2")

val createAuthorizationURLOpts = CreateAuthorizationURLOpts().setScopes(scopesArr)

val credentialTypes = StringArray()
credentialTypes.append("VerifiableCredential").append("UniversityDegree")

// If scopes aren't needed, you can pass in null in place of the opts argument.
val authorizationLink := interaction.createAuthorizationURL("clientID", "redirect URI", "jwt_vc_json",credentialTypes, createAuthorizationURLOpts)

// Open authorizationLink in a browser. Once the user has finished logging in, call requestCredential()
// with the full redirect URI (including query parameters) that the login service sent the user to.
// The code below assumes this has already been done somehow and that the URI is in the redirectURIWithParams variable.
// In actual code, the call to requestCredential() couldn't be immediately after the
// createAuthorizationURL() call like in this example since control has to flow back to the user first.
val redirectURIWithParams = "Put the redirect URI with params here"

val credentials = interaction.requestCredential(didDocument.assertionMethod(), redirectURIWithParams, null)

// Consider checking the activity log at some point after the interaction
```

#### Swift (iOS) - Wallet-Initiated

```kotlin
import Walletsdk

// Setup
let memKMSStore = LocalkmsNewMemKMSStore()

var newKMSError: NSError? //Be sure to actually check this error in real code
let kms = LocalkmsNewKMS(memKMSStore, &newKMSError)

let didResolver = DidNewResolver(nil)

let jwk = try kms.create(LocalkmsKeyTypeED25519)

var didKeyCreateError: NSError? //Be sure to actually check this error in real code
let didDocument = DidkeyCreate(jwk, &didKeyCreateError)// Create a did:key document

let activityLogger = MemNewActivityLogger()

// Going through the flow
let interactionArgs = Openid4ciNewWalletInitiatedInteractionArgs("IssuerURIHere", kms.getCrypto(), didResolver)
let interactionOpts = Openid4ciNewInteractionOpts().setActivityLogger(activityLogger) // Optional, but useful for tracking credential activity
var newInteractionError: NSError?

// This is a stateful object - we will use this to go through the flow.
let interaction = Openid4ciNewWalletInitiatedInteraction(interactionArgs, interactionOpts, &newInteractionError)

let scopes = ApiStringArray()
scopes.append("scope1").append("scope2")

let opts = Openid4ciNewCreateAuthorizationURLOpts()!.setScopes(scopes)

val credentialTypes = ApiStringArray()
credentialTypes.append("VerifiableCredential").append("UniversityDegree")

var createAuthURLError: NSError?

// If scopes aren't needed, you can pass in nil in place of the opts argument.
let authorizationLink = interaction.createAuthorizationURL("clientID", redirectURI: "redirect URI", credentialFormat: "jwt_vc_json", credentialTypes: credentialTypes, opts: opts, error: &createAuthURLError)

// Open authorizationLink in a browser. Once the user has finished logging in, call requestCredential()
// with the full redirect URI (including query parameters) that the login service sent the user to.
// The code below assumes this has already been done somehow and that the URI is in the redirectURIWithParams variable.
// In actual code, the call to requestCredential() couldn't be immediately after the
// createAuthorizationURL() call like in this example since control has to flow back to the user first.
let redirectURIWithParams = "Put the redirect URI with params here"

let credentials = interaction.requestCredential(vm: didDocument.assertionMethod(), redirectURIWithParams: redirectURIWithParams, opts: nil)

// Consider checking the activity log at some point after the interaction
```

### Error Codes & Troubleshooting Tips

#### Creating New Interaction Object

| Error                                      | Possible Reasons                                                                                                                                                                                                                 |
|--------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| INVALID_ISSUANCE_URI(OCI0-0000)            | The issuance URI used to initiate the OpenID4CI flow isn't a valid URL.<br/><br/>The issuance URI doesn't specify a credential offer.                                                                                            |
| INVALID_CREDENTIAL_OFFER(OCI0-0001)        | The credential offer object is malformed.<br/><br/>The issuance URI specified an endpoint for retrieving the credential offer, but there was an error during the GET call. The server may be down or have a configuration issue. |
| UNSUPPORTED_ISSUANCE_URI_SCHEME(OCI0-0019) | The issuance URI used to initiate the OpenID4CI flow uses an unsupported scheme. Wallet-SDK only supports the "openid-credential-offer" scheme.                                                                                  |

##### Requesting Credential

| Error                                        | Possible Reasons                                                                                                                                                                                                                                                                                                                                                                                                                                    |
|----------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| ISSUER_OPENID_CONFIG_FETCH_FAILED(OCI1-0003) | An error occurred while doing an GET call on the issuer's OpenID configuration endpoint. The server may be down or have a configuration issue.<br/><br/>The issuer's OpenID configuration object is malformed.                                                                                                                                                                                                                                      |
| JWT_SIGNING_FAILED(OCI1-0005)                | The KMS is missing the key that was used to create the DID that you're using. This could happen if your KMS storage is not storing and retrieving keys as expected, or if there is a mismatch between the KMS you used to create the DID (whose verification method you pass into the `requestCredential` function) and the `crypto` object (which should be obtained via the `getCrypto()` method on the KMS) passed in to the interaction object. |
| KEY_ID_MISSING_DID_PART(OCI1-0006)           | The DID is incompatible with Wallet-SDK.                                                                                                                                                                                                                                                                                                                                                                                                            |
| METADATA_FETCH_FAILED(OCI1-0004)             | An error occurred while doing an GET call on the issuer's OpenID credential issuer endpoint. The server may be down or have a configuration issue.<br/><br/>The issuer metadata object from the server is malformed.                                                                                                                                                                                                                                |
| CREDENTIAL_PARSE_FAILED(OCI1-0012)           | The issued credential is invalid, signed incorrectly, or could not be verified.                                                                                                                                                                                                                                                                                                                                                                     |
| INVALID_TOKEN_REQUEST(OCI1-0009)             | A PIN was provided but the authorization server was not expecting one.                                                                                                                                                                                                                                                                                                                                                                              |
| INVALID_GRANT(OCI1-0010)                     | Incorrect PIN.<br/><br/>The pre-authorized code has expired, or the refresh token has expired or been revoked. Try restarting the flow.                                                                                                                                                                                                                                                                                                             |
| INVALID_CLIENT(OCI1-0011)                    | The issuer requires a client ID for the pre-authorized code flow, but none was provided. Note: Wallet-SDK does not currently have support for this.                                                                                                                                                                                                                                                                                                 |
| INVALID_TOKEN(OCI1-0015)                     | The access token has expired or been revoked. Try restarting the flow.                                                                                                                                                                                                                                                                                                                                                                              |
| UNSUPPORTED_CREDENTIAL_FORMAT(OCI1-0016)     | In the wallet-initiated flow, an unsupported credential format was specified.                                                                                                                                                                                                                                                                                                                                                                       |
| UNSUPPORTED_CREDENTIAL_TYPE(OCI1-0017)       | In the wallet-initiated flow, an unsupported credential type was specified.                                                                                                                                                                                                                                                                                                                                                                         |

## Credential Display Data

After completing the `RequestCredential` step of the OpenID4CI flow, you will have your issued Verifiable Credential
objects. These objects contain the data needed for various wallet operations, but they don't tell you how you can
display the credential data in an easily-understandable way via a user interface. This is where the credential display
data comes in.

To get display data, call the `resolveDisplay` function with your VCs and the issuer URI. An issuer URI can be obtained by
calling the `issuerURI` method on an OpenID4CI interaction object after it's been instantiated. It's a good idea to
store the issuer URI somewhere in persistent storage after going through the OpenID4CI flow. This way, you can call the
`resolveDisplay` function later if/when you need to refresh your display data based on the latest display information
from the issuer. Note that if the issuer uses signed metadata, then you'll need to also pass a DID resolver into
the `resolveDisplay` function. See the [Options](#options) section for more information.

Display data objects can be serialized using the `serialize()` method (useful for storage) and parsed from serialized
form back into display data objects using the `parseDisplayData()` function.

## Options

The `resolveDisplay` function has a number of different options available. For a full list of available options, check
the associated `Opts` object. This section will highlight some especially notable ones:

### Set DID Resolver

If the issuer makes use of signed issuer metadata, then you'll need to pass in a DID resolver using the 
`setDIDResolver(didResolver)` option. Otherwise, the`resolveDisplay` call will fail.

### Set Masking String

The `setMaskingString(maskingString)` option allows you to specify the string to be used when creating masked values for
display. The substitution is done on a character-by-character basis, whereby each individual character to be masked
will be replaced by the entire string. See the examples below to better understand exactly how the substitution works.

#### Examples

Note that any quote characters used in these examples are only there for readability reasons - they're not actually
part of the values.

Scenario: The unmasked display value is 12345, and the issuer's metadata specifies that the first 3 characters are
to be masked. The most common use-case is to substitute every masked character with a single character. This is
achieved by specifying just a single character in the `maskingString`. Here's what the masked value would look like
with different `maskingString` choices:
```
maskingString="•"    -->    •••45
maskingString="*"    -->    ***45
```

It's also possible to specify multiple characters in the `maskingString`, or even an empty string if so desired.
Here's what the masked value would like in such cases:

```
maskingString="???"  -->    ?????????45
maskingString=""     -->    45
```

If this option isn't used, then by default "•" characters (without the quotes) will be used for masking.

### Set Preferred Locale

Use the `setPreferredLocale` method to specify what locale to use for resolving display values. The effectiveness of
this option is contingent on what information the issuer provides. If the issuer specifies display value localizations
for your preferred locale, then those will be used whenever possible. If localized values aren't available (for your
preferred locale), then the first locale listed by the issuer will be used instead. Note that some pieces of display
data may be localized more or less than other pieces.

To determine what locales the resolved display values are in, use the various `locale()` methods available in the
display data. You can use this to figure out which display values are actually in your preferred locale and which
used a fallback default locale instead.

### The Display Object Structure

The structure of the display data object is as follows:

#### `Data`

* The root object.
* Contains display information for the issuer and each of the credentials passed in to the `resolveDisplay`
function.
* Can be serialized using the `serialize()` method and parsed using the `parseDisplayData()` function.
* The `issuerDisplay()` method returns the `IssuerDisplay` object.
* Use the `credentialDisplaysLength()` and `credentialDisplayAtIndex()` methods to iterate over the `CredentialDisplay`
  objects. There's one for each credential passed in to the `resolveDisplay` function, and they're in the same order.

#### `IssuerDisplay`

* Describes display information about the issuer.
* Can be serialized using the `serialize()` method and parsed using the `parseIssuerDisplay()` function.
* Has `name()` and `locale()` methods.

#### `CredentialDisplay`

* Describes display information about a credential.
* Can be serialized using the `serialize()` method and parsed using the `parseCredentialDisplay()` function.
* The `overview()` method returns the `CredentialOverview` object.
* Use the `claimsLength()` and `claimAtIndex()` methods to iterate over the `Claim` objects.

#### `CredentialOverview`

* Describes display information for a credential as a whole.
* Has `name()`, `logo()`, `backgroundColor()`, `textColor()`, and `locale()` methods. The `logo()` method returns
  a `Logo` object.

#### `Logo`

* Describes display information for a logo.
* Has `url()` and `altText()` methods.

#### `Claim`

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

### Resolve Display Examples

The following examples show how the `ResolveDisplay` function can be used. The function requires you to pass in
one or more VCs and the issuer URI.

#### Kotlin (Android)

##### Using Default Options

```kotlin
import dev.trustbloc.wallet.sdk.display.*
import dev.trustbloc.wallet.sdk.verifiable.CredentialsArray

val vcArray = CredentialsArray()
vcArray.add(yourVCHere)

val displayData = Display.resolve(vcArray, "Issuer_URI_Goes_Here", null)
```

##### Using Specified Options

```kotlin
import dev.trustbloc.wallet.sdk.display.*
import dev.trustbloc.wallet.sdk.verifiable.CredentialsArray

val vcArray = CredentialsArray()
vcArray.add(yourVCHere)

val opts = Opts().setPreferredLocale("en-us")
val displayData = Display.resolve(vcArray, "Issuer_URI_Goes_Here", opts)
```

#### Swift (iOS)

##### Using Default Options

```swift
import Walletsdk

let vcArray = VerifiableCredentialsArray()

vcArray.add(yourVCHere)

var error: NSError?
let displayData = DisplayResolve(vcArray, "Issuer_URI_Goes_Here", nil, &error)
```

##### Using Specified Options

```swift
import Walletsdk

let vcArray = VerifiableCredentialsArray()

vcArray.add(yourVCHere)

var error: NSError?
let opts = DisplayNewOpts().setPreferredLocale("en-us")
let displayData = DisplayResolve(vcArray, "Issuer_URI_Goes_Here", opts, &error)
```

## Credential Status

After a Verifiable Credential is issued, many credentials support the ability for the issuer to later _revoke_ the
credential, making it no longer valid. If your use case expects that a certain type of Verifiable Credential should
support credential status verification, then you can verify these credentials using a `StatusVerifier`.

A `StatusVerifier` can be instantiated once and used to verify the status of multiple Verifiable Credentials,
using the `verify` API. This will return an error if:
- The provided credential does not support status verification. Your use case will determine whether you should expect
  credentials to support status verification, and your application should handle any logic if you want to handle this
  on a case-by-case basis.
- The status verification process fails, e.g. due to the issuer server being unavailable to report status.
- The credential has been revoked.

By default, `StatusVerifier` only supports status APIs that fetch status metadata via a URL to the issuer's
status endpoint. If you need to also support status APIs that use DID-URL resolution, create a `StatusVerifier`
using the `NewStatusVerifierWithDIDResolver` constructor.

### Examples

#### Kotlin

##### Without DID Resolver

```kotlin
import dev.trustbloc.wallet.sdk.credential.StatusVerifier
import dev.trustbloc.wallet.sdk.verifiable.Verifiable

val cred = Verifiable.parseCredential("Your VC here", opts)

// The StatusVerifierOpts argument is a placeholder for future options.
// There are currently no options that can be set, so pass in null for now.
val statusVerifier = StatusVerifier(null)

try {
    // If no exception is thrown, then status verification succeeded.
    statusVerifier.verify(cred)
} catch (e: Exception) {
    // Status verification failed. Check the exception for details.
}
```

##### With DID Resolver

```kotlin
import dev.trustbloc.wallet.sdk.did.Resolver
import dev.trustbloc.wallet.sdk.credential.StatusVerifier
import dev.trustbloc.wallet.sdk.verifiable.Verifiable

val cred = Verifiable.parseCredential("Your VC here", opts)

val didResolver = Resolver(null)

// In this case, we pass in the DID Resolver, meaning the status verifier supports
// DID-URL resolution for fetching status metadata.
val statusVerifier = StatusVerifier(didResolver, null)

try {
    // If no exception is thrown, then status verification succeeded.
    statusVerifier.verify(cred)
} catch (e: Exception) {
    // Status verification failed. Check the exception for details.
}
```

#### Swift

##### Without DID Resolver

```swift
import Walletsdk

var parseError: NSError?
let cred = VerifiableParseCredential("yourVCHere", nil, &parseError)

var newStatusVerifierError: NSError?
let statusVerifier = CredentialNewStatusVerifier(nil, &newStatusVerifierError)

do {
    // If no exception is thrown, then status verification succeeded.
    try statusVerifier?.verify(cred)
} catch let verifyError as NSError {
    // Status verification failed. Check the error for details.
}
```

##### With DID Resolver

```swift
import Walletsdk

var parseError: NSError?
let cred = VerifiableParseCredential("yourVCHere", nil, &parseError)

var newResolverError: NSError?
let didResolver = DidNewResolver(nil, &newResolverError)

var newStatusVerifierError: NSError?
let statusVerifier = CredentialNewStatusVerifierWithDIDResolver(didResolver, nil, &newStatusVerifierError)

do {
    // If no error is thrown, then status verification succeeded.
    try statusVerifier?.verify(cred)
} catch let verifyError as NSError {
    // Status verification failed. Check the error for details.
}
```

## OpenID4VP

The OpenID4VP package contains an API that can be used by a [holder](https://www.w3.org/TR/vc-data-model/#dfn-holders)
to go through the [OpenID4VP](https://openid.net/specs/openid-connect-4-verifiable-presentations-1_0-ID1.html) flow.

The general pattern is as follows:

1. Create a new `Args` object. An `Args` contains the following mandatory
parameters:
   * An authorization request URI obtained from a verifier (e.g. via a QR code).
   * A crypto implementation.
   * A DID resolver.
2. (optional) Create an `opts` object. To set optional arguments, use the supplied methods available
   on the `Opts` object. For example:
   * `setActivityLogger`: Used to log credential activities.
   * `addHeaders`: Allows you to set additional headers to be sent to the issuer.

   Options can be chained together if you wish (e.g. `newOpts().setActivityLogger(...).setHeaders(...)`).
3. Create a new `Interaction` object using your `Args` and`Opts` objects.
If none of the optional arguments are needed, then a nil `Opts` can be passed in instead.
The `Interaction` object is a stateful object and is meant to be used for a single instance of an OpenID4VP flow and
then discarded.
4. Get the query by calling the `GetQuery` method on the `Interaction`.
5. Select the credentials that match the query from the previous step.
6. Determine the key ID you want to use for signing (e.g. from one of the user's DID docs).
7. Call the `PresentCredential` method on the `Interaction` object with the selected credentials.

In certain use cases, you might want the user to be able to provide a credential
even if no credential matches the requirements. In that case, you can call the
`PresentCredentialUnsafe` method on the `Interaction` object instead of step 7,
passing in a single credential instead of a credential array.

### Examples

The following examples show how to use the APIs to go through the OpenID4VP flow using the iOS and Android bindings.
They use in-memory key storage and the Tink crypto library.

#### Kotlin (Android)

```kotlin
import dev.trustbloc.wallet.sdk.localkms.Localkms
import dev.trustbloc.wallet.sdk.localkms.MemKMSStore
import dev.trustbloc.wallet.sdk.did.Resolver
import dev.trustbloc.wallet.sdk.openid4vp.*
import dev.trustbloc.wallet.sdk.credential.*
import dev.trustbloc.wallet.sdk.verifiable.CredentialsArray

// Setup
val memKMSStore = MemKMSStore.MemKMSStore()
val kms = Localkms.newKMS(memKMSStore)
val didResolver = Resolver(null)

val args = Args("YourAuthRequestURIHere", kms.getCrypto(), didResolver)

val activityLogger = mem.ActivityLogger()
val opts = Opts().setActivityLogger(activityLogger) // Optional, but useful for tracking credential activity

// Going through the flow
val interaction = Interaction(args, opts)
val query = interaction.getQuery()
val inquirer = Inquirer(null)
val savedCredentials = CredentialsArray() // Would need some actual credentials for this to work

// Use this code to display information about the verifier.
val verifierDisplayData = openID4VP.getVerifierDisplayData()
val verifierDID = verifierDisplayData.did()
val verifierName = verifierDisplayData.name()
val verifierLogoURI = verifierDisplayData.logoURI()
val verifierPurpose = verifierDisplayData.purpose()

// Use this code to display the list of VCs to select which of them to send.
val matchedRequirements = inquirer.getSubmissionRequirements(query, savedCredentials) 
val matchedRequirement = matchedRequirements.atIndex(0) // Usually we will have one requirement
val requirementDesc = matchedRequirement.descriptorAtIndex(0) // Usually requirement will contain one descriptor
val selectedVCs = CredentialsArray()
selectedVCs.add(requirementDesc.matchedVCs.atIndex(0)) // Users should select one VC for each descriptor from the matched list and confirm that they want to share it

interaction.presentCredential(selectedVCs)
// Consider checking the activity log at some point after the interaction

// To force a specific credential to be presented, even if it doesn't match all
// requirements, use this instead of presentCredential:
val preferredVC = savedCredentials.atIndex(0)
interaction.presentCredentialUnsafe(preferredVC)
```

###### Read scope and add custom scope claims

```kotlin

val scope = interaction.customScope()

interaction.PresentCredentialOpts(PresentCredentialOpts().addScopeClaim(
        scope.atIndex(0), """{"email":"test@example.com"}"""), selectedCredentials)

```

###### Evaluate issuance trust info

```kotlin
val issuanceRequest = IssuanceRequest()

val trustInfo = newInteraction.issuerTrustInfo()
issuanceRequest.issuerDID = trustInfo.did
issuanceRequest.issuerDomain = trustInfo.domain
issuanceRequest.credentialFormat = trustInfo.credentialFormat
issuanceRequest.credentialType = trustInfo.credentialType

val config = RegistryConfig()
config.evaluateIssuanceURL = evaluateIssuanceURL

val evaluationResult = Registry(config).evaluateIssuance(issuanceRequest)
```

###### Evaluate presentation trust info

```kotlin

val presentationRequest = PresentationRequest()

for (rInd in 0 until submissionRequirements.len()) {
    val requirement = submissionRequirements.atIndex(rInd)
    for (dInd in 0 until requirement.descriptorLen()) {
        val descriptor = requirement.descriptorAtIndex(dInd)
        for (credInd in 0 until descriptor.matchedVCs.length()) {
            val cred = descriptor.matchedVCs.atIndex(credInd)

            val claimsToCheck = CredentialClaimsToCheck();
            claimsToCheck.credentialID = cred.id()
            claimsToCheck.issuerID = cred.issuerID()
            claimsToCheck.credentialTypes = cred.types()
            claimsToCheck.expirationDate = cred.expirationDate()
            claimsToCheck.issuanceDate = cred.issuanceDate()

            presentationRequest.addCredentialClaims(claimsToCheck)
        }
    }
}

val trustInfo = initiatedInteraction.trustInfo()
presentationRequest.verifierDID = trustInfo.did
presentationRequest.verifierDomain = trustInfo.domain

val config = RegistryConfig()
config.evaluatePresentationURL = evaluatePresentationURL

val evaluationResult = Registry(config).evaluatePresentation(presentationRequest)

```

#### Swift (iOS)

```swift
import Walletsdk

// Setup
let memKMSStore = LocalkmsNewMemKMSStore()

var error: NSError?
let kms = LocalkmsNewKMS(memKMSStore, &error)

let didResolver = DidNewResolver(nil)

let args = Openid4vpNewArgs("YourAuthRequestURIHere", kms.getCrypto(), didResolver)

let activityLogger = mem.ActivityLogger()
let opts = Openid4vpNewOpts().setActivityLogger(activityLogger) // Optional, but useful for tracking credential activity

// Going through the flow
var newInteractionError: NSError?
let interaction = Openid4vpNewInteraction(args, opts, &newInteractionError)
let query = interaction.getQuery()
var newInquirerError: NSError?
let inquirer = CredentialNewInquirer(nil, &newInquirerError)
let savedCredentials = VerifiableCredentialsArray() // Would need some actual credentials for this to work

// Use this code to display information about the verifier.
let verifierDisplayData =  try openID4VP.getVerifierDisplayData()
let verifierDID = verifierDisplayData.did(),
let verifierLogoURI = verifierDisplayData.logoURI(),
let verifierName = verifierDisplayData.name(),
let verifierPurpose = verifierDisplayData.purpose()

// Use this code to display the list of VCs to select which of them to send.
let matchedRequirements = inquirer.getSubmissionRequirements(query, savedCredentials) 
let matchedRequirement = matchedRequirements.atIndex(0) // Usually we will have one requirement
let requirementDesc = matchedRequirement.descriptorAtIndex(0) // Usually requirement will contain one descriptor
let selectedVCs = VerifiableCredentialsArray()
selectedVCs.add(requirementDesc.matchedVCs.atIndex(0)) // Users should select one VC for each descriptor from the matched list and confirm that they want to share it

let credentials = interaction.presentCredential(selectedVCs)
// Consider checking the activity log at some point after the interaction

// To force a specific credential to be presented, even if it doesn't match all
// requirements, use this instead of presentCredential:
let preferredVC = savedCredentials.atIndex(0)
interaction.presentCredentialUnsafe(preferredVC)
```

###### Read scope and add custom scope claims

```swift

val scope = interaction.customScope()

try interaction.presentCredentialOpts(
    selectedCreds,
    opts: Openid4vpNewPresentCredentialOpts()?.addScopeClaim(scope.atIndex(0), claimJSON:#"{"email":"test@example.com"}"#)     
)

```

###### Evaluate issuance trust info

```swift
let issuanceRequest = TrustregistryIssuanceRequest()

let trustInfo = try initiatedInteraction.issuerTrustInfo()
issuanceRequest.issuerDID = trustInfo.did
issuanceRequest.issuerDomain = trustInfo.domain
issuanceRequest.credentialFormat = trustInfo.credentialFormat
issuanceRequest.credentialType = trustInfo.credentialType

let config = TrustregistryRegistryConfig()
config.evaluateIssuanceURL = evaluateIssuanceURL

let evaluationResult = try TrustregistryRegistry(config)!.evaluateIssuance(issuanceRequest)
```

###### Evaluate presentation trust info
```swift

let presentationRequest = TrustregistryPresentationRequest()

for rInd in 0..<submissionRequirements.len() {
    let requirement = submissionRequirements.atIndex(rInd)!
    for dInd in 0..<requirement.descriptorLen() {
        let descriptor = requirement.descriptor(at: dInd)!
        for credInd in 0..<descriptor.matchedVCs!.length() {
            let cred = descriptor.matchedVCs!.atIndex(credInd)!

            let claimsToCheck = TrustregistryCredentialClaimsToCheck();
            claimsToCheck.credentialID = cred.id_()
            claimsToCheck.issuerID = cred.issuerID()
            claimsToCheck.credentialTypes = cred.types()
            claimsToCheck.expirationDate = cred.expirationDate()
            claimsToCheck.issuanceDate = cred.issuanceDate()

            presentationRequest.addCredentialClaims(claimsToCheck)
        }
    }
}

let trustInfo = try initiatedInteraction.trustInfo()
presentationRequest.verifierDID = trustInfo.did
presentationRequest.verifierDomain = trustInfo.domain

let config = TrustregistryRegistryConfig()
config.evaluatePresentationURL = evaluatePresentationURL

let evaluationResult = try TrustregistryRegistry(config)!.evaluatePresentation(presentationRequest)

```

### Error Codes & Troubleshooting Tips

| Error                                             | Possible Reasons                                                                                                                                                                                                                                                                                                                                                               |
|---------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| INVALID_AUTHORIZATION_REQUEST(OVP1-0000)          | The authorization request is a URI but specifies a scheme other than "openid-vc".<br/><br/>The authorization request is a URI and is missing the request_uri parameter.<br/><br/>The request object's signature is invalid.<br/><br/>The request object is malformed.<br/><br/>Wallet-SDK does not support the format/type of the authorization request and/or request object. |
| REQUEST_OBJECT_FETCH_FAILED(OVP1-0001)            | The authorization request is a URI and the request URI endpoint that it specifies cannot be reached.                                                                                                                                                                                                                                                                           |
| FAIL_TO_GET_MATCH_REQUIREMENTS_RESULTS(CRQ0-0004) | Invalid presentation definition received from the verifier.                                                                                                                                                                                                                                                                                                                    |
| CREATE_AUTHORIZED_RESPONSE(OVP1-0002)             | No credentials provided in the `presentCredential` method call.                                                                                                                                                                                                                                                                                                                |
| SEND_AUTHORIZED_RESPONSE(OVP1-0003)               | The verifier server rejected your credentials (couldn't be verified, wrong type, etc).<br/><br/>The verifier server is down or incorrectly configured.                                                                                                                                                                                                                         |

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
                              
Note that the sum of all sub-event durations may not add up to the duration of the parent event, as not every possible operation during the parent event will be timed.
Generally, short/trivial operations are not tracked.
