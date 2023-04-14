# Network Operations

At a high-level, here are the network operations that get performed in Wallet-SDK for various flows/operations:

## DID Creation

* Depends on the DID method (for did:key or did:jwk: none)

## OpenID4CI

* Fetch credential offer
* Fetch issuer’s OpenID configuration
* Fetch token
* Fetch issuer metadata
* Fetch credential
* DID resolution (if remote)
* Document loading (if required)

## OpenID4VP
* Fetch request object
* DID resolution (if remote)
* Send authorized response (AKA present credential)

## Credential Display Resolution
* Fetch issuer metadata