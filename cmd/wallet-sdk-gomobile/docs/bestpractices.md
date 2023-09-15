# Best Practices (Mobile Bindings)

This document outlines some best practices for using the mobile bindings.

For information on best practices for higher-level concepts that appear in Wallet-SDK,
see [here](../docs/bestpractices.md). While this document is only applicable to the mobile bindings specifically,
those higher-level best practices that are applicable here as well since they're common to all Wallet-SDK bindings.

## Using Serializable Objects

Wallet-SDK has various objects that have `serialize` methods on them. These methods can be used to convert them into a
form that can be easily stored in a persistent database. Wallet-SDK also has corresponding `parse` functions for
these serializable objects.

The serialized versions of these objects should be treated as opaque objects from the caller's perspective.
They should not be programmatically inspected or altered in this form, and no assumptions about what serialized format
they're in should be made.

To view or edit data contained within these objects, always parse the serialized data first and use the supplied methods
in Wallet-SDK.

## Error and Nullable Type/Optionals Handling

Some Wallet-SDK functions/methods can return errors/throw exceptions. These errors should never be ignored, unless
the [usage documentation](usage.md) states that it's safe to do so (e.g. calling `order()` on a claim after
ensuring that `hasOrder()` returns true).

In the mobile bindings, non-primitive types are always nullable types due to a limitation with `gomobile`.
For functions/methods that return errors/throw exceptions, if you've ensured that no error has occurred, then it's
safe to assume that the returned object is not null/nil.
Object constructors that don't return an error also never return null/nil.
This assumption may be useful in Swift or Kotlin code, where the compiler forces null/nil checks.
