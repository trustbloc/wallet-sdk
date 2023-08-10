# Packages

The way you use the SDK differs between Android and iOS. This comes down to differences in how the `gomobile` tool
generates the bindings for the two platforms. Due to these differences (and the limitations of `gomobile`),
it's difficult to pick names that look perfect across both platforms and the Go source code.
Thus, the various names used in the Go files were chosen based on achieving a reasonable compromise between all
three versions.


## Android

* `dev.trustbloc.wallet.sdk.api`
* `dev.trustbloc.wallet.sdk.credential`
* `dev.trustbloc.wallet.sdk.didcreator`
* `dev.trustbloc.wallet.sdk.didresolver`
* `dev.trustbloc.wallet.sdk.display`
* `dev.trustbloc.wallet.sdk.localkms`
* `dev.trustbloc.wallet.sdk.mem`
* `dev.trustbloc.wallet.sdk.openid4ci`
* `dev.trustbloc.wallet.sdk.openid4vp`
* `dev.trustbloc.wallet.sdk.otel`
* `dev.trustbloc.wallet.sdk.verifiable`

## iOS

Everything is under a single package named `Walletsdk`.

Interfaces (protocols), types, and functions are all prefixed according to their corresponding Go package names.
Note that this behaviour is enforced by the `gomobile` tool and cannot currently be changed.
See [here](https://github.com/golang/go/issues/32573) for more information.

Prefixes:

* `Api`
* `Credential`
* `Didcreator`
* `Didresolver`
* `Display`
* `Localkms`
* `Mem`
* `Openid4ci`
* `Openid4vp`
* `Otel`
* `Verifiable`

## Package/Module Examples

To better understand how package naming works, it's helpful to look at some examples for each of the packages/modules
and compare how something in Go gets converted to Java (for Android) and Objective-C (for iOS).

These examples are not meant to be full code examples of how you actually use the API.
Rather, these examples are just to help you understand the naming patterns, which should help you in figuring out
the equivalent names for various parts of the API across platforms.

The left-most side in the tables shows Go source code snippets, rather than how they would look when imported in another
file. The middle and right-most side show how the code looks when importing that same functionality in via the bindings.
The reason for comparing between Go and the bindings like this is that the Go code is not intended to be imported
directly - the mobile wrappers exist only to make the Go code compatible with the `gomobile` tool, so seeing how it
would look imported is not so useful. However, the way you import from the generated mobile bindings is important,
and so that's why we show that here. The Go source code is really just meant to help demonstrate the naming patterns
to the reader (as the names used for packages/modules in the bindings are generated automatically by `gomobile` based
on the Go names).

### API

This package contains interface definitions and types that are used in multiple places across the API as a whole.

<table>
<tr>
<td> Go Source Code </td> <td> Java (as called from Kotlin) </td> <td> Obj-C (as called from Swift) </td>
</tr>
<tr>
<td> 

```
package api

type CredentialReader interface {
	Get(...) (*VerifiableCredential, error)
	GetAll() (*VerifiableCredentialsArray, error)
}
```

</td>
<td>

```kotlin
import dev.trustbloc.wallet.sdk.api

api.CredentialReader
```

</td>
<td>

```swift
import Walletsdk

ApiCredentialReader
```

</td>
</tr>
</table>

### Credential Querying

<table>
<tr>
<td> Go Source Code </td> <td> Java (as called from Kotlin) </td> <td> Obj-C (as called from Swift) </td>
</tr>
<tr>
<td> 

```
package credential

func NewInquirer(
	...
) *Inquirer {
	...
}
```

</td>
<td>

```kotlin
import dev.trustbloc.wallet.sdk.credential

val inquirer = credential.newInquirer(...)

inquirer.query(...)
```

</td>
<td>

```swift
import Walletsdk

let inquirer = CredentialNewInquirer(...)

inquirer.query(...)
```

</td>
</tr>
</table>

### DID Creator and Resolver

<table>
<tr>
<td> Go Source Code </td> <td> Java (as called from Kotlin) </td> <td> Obj-C (as called from Swift) </td>
</tr>
<tr>
<td> 

```
package did

func NewCreator(
	...
) (*Creator, error) {
    ...
}

func NewResolver() *Resolver {
	...
}

```

</td>
<td>

```kotlin
import dev.trustbloc.wallet.sdk.did

did.newCreator(...)
did.newResolver()
```

</td>
<td>

```swift
import Walletsdk

DidNewCreator(...)
DidNewResolver()
```

</td>
</tr>
</table>

### In-Memory Credential DB

<table>
<tr>
<td> Go Source Code </td> <td> Java (as called from Kotlin) </td> <td> Obj-C (as called from Swift) </td>
</tr>
<tr>
<td> 

```
package credential

func NewInMemoryDB() *DB {
	...
}
```

</td>
<td>

```kotlin
import dev.trustbloc.wallet.sdk.credential

credential.newInMemoryDB()
```

</td>
<td>

```swift
import Walletsdk

CredentialNewInMemoryDB()
```

</td>
</tr>
</table>

### Credential Status

<table>
<tr>
<td> Go Source Code </td> <td> Java (as called from Kotlin) </td> <td> Obj-C (as called from Swift) </td>
</tr>
<tr>
<td>

```
package credential

func NewStatusVerifier(...) (*StatusVerifier, error) {
	...
}
```

</td>
<td>

```kotlin
import dev.trustbloc.wallet.sdk.credential

credential.newStatusVerifier(...)
```

</td>
<td>

```swift
import Walletsdk

CredentialNewStatusVerifier(...)
```

</td>
</tr>
</table>

### Local KMS

<table>
<tr>
<td> Go Source Code </td> <td> Java (as called from Kotlin) </td> <td> Obj-C (as called from Swift) </td>
</tr>
<tr>
<td> 

```
package localkms

func NewKMS(...) (*KMS, error) {
	...
}
```

</td>
<td>

```kotlin
import dev.trustbloc.wallet.sdk.localkms

localkms.newKMS(...)
```

</td>
<td>

```swift
import Walletsdk

LocalkmsNewKMS(...)
```

</td>
</tr>
</table>

### OpenID4CI

<table>
<tr>
<td> Go Source Code </td> <td> Java (as called from Kotlin) </td> <td> Obj-C (as called from Swift) </td>
</tr>
<tr>
<td> 

```
package openid4ci

func NewInteraction(
	...
) (*Interaction, error) {
```

</td>
<td>

```kotlin
import dev.trustbloc.wallet.sdk.openid4ci

openid4ci.newInteraction(...)
```

</td>
<td>

```swift
import Walletsdk

Openid4ciNewInteraction(...)
```

</td>
</tr>
</table>

### OpenID4VP

<table>
<tr>
<td> Go Source Code </td> <td> Java (as called from Kotlin) </td> <td> Obj-C (as called from Swift) </td>
</tr>
<tr>
<td> 

```
package openid4vp

func NewInteraction(
	...
) *Interaction {
```

</td>
<td>

```kotlin
import dev.trustbloc.wallet.sdk.openid4vp

openid4vp.newInteraction(...)
```

</td>
<td>

```swift
import Walletsdk

Openid4vpNewInteraction(...)
```

</td>
</tr>
</table>
