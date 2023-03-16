# Best Practices (High-Level Concepts)

This document outlines some best practices for using DIDs, Verifiable Credentials, and other concepts that appear in
Wallet-SDK.

For information on best practices for using the mobile bindings,
see [here](../cmd/wallet-sdk-gomobile/docs/bestpractices.md). This document deals only with higher-level concepts that
are applicable to Wallet-SDK, regardless of the chosen bindings.

## DIDs

* It is encouraged to generate a new DID for each credential issuance for holder binding.
This is considered a best practice since the DID cannot be used as a correlating identifier.

## Tracing IDs

* Some APIs allow for customizable HTTP client behaviour - for example, in the gomobile SDK, some APIs allow for
additional headers to be injected in. These headers could be used to add in tracing IDs for debugging purposes. In
this case, it's recommended to use a different tracing ID for every interaction to help avoid potential privacy issues
that could be caused by the issuer/verifier correlating them to specific users.
